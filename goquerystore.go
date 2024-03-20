package querystore

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type QueryStore struct {
	store map[string]string
}

func (q *QueryStore) Get(key string) string {
	return q.store[key]
}

const (
	prefix = "querykey:"
)

type fileContents struct {
	named    string
	contents string
}

func New(directory string) (*QueryStore, error) {
	files := []string{}
	err := filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
		if filepath.Ext(path) == ".sql" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	cs := make(chan fileContents)
	go fileMap(files, cs)
	m := make(map[string]string)
	for res := range cs {
		if _, ok := m[res.named]; ok {
			return nil, fmt.Errorf("keys cannot be duplicate: %s", res.named)
		}
		m[res.named] = res.contents
	}

	file := &QueryStore{}
	file.store = m
	return file, nil
}

func fileMap(files []string, cs chan fileContents) {
	defer close(cs)
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			return
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		fL, cont, err := getContents(scanner)
		if err != nil {
			return
		}
		named, ok := getNamedKey(fL)
		if !ok {
			continue
		}
		cs <- fileContents{named: named, contents: cont}
	}
}

func getContents(scanner *bufio.Scanner) (string, string, error) {
	var fL string
	var rem string
	for scanner.Scan() {
		if fL == "" {
			fL = scanner.Text()
			continue
		}
		rem += scanner.Text() + "\n"
	}
	return fL, rem, nil
}

func getNamedKey(line string) (string, bool) {
	is := strings.Split(line, " ")
	for _, i := range is {
		if strings.HasPrefix(i, prefix) {
			return strings.TrimPrefix(i, prefix), true
		}
	}
	return "", false
}
