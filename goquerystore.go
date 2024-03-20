package goquerystore

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type QueryStore struct {
	Store     map[string]string
	FileCount int
	FileNames []string
}

func (q *QueryStore) Get(key string) string {
	return q.Store[key]
}

type fileContents struct {
	named    string
	contents string
}

func fileMap(files []string, cs chan fileContents) {
	defer close(cs)
	for _, file := range files {
		f, err := os.Open(file)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		scanner := bufio.NewScanner(f)
		o, ok := getOpening(scanner)
		if !ok {
			continue
		}
		n, ok := getNamed(o)
		if !ok {
			continue
		}
		c := getRemaining(scanner)
		cs <- fileContents{named: n, contents: c}
	}
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
	file.Store = m
	file.FileCount = len(m)
	file.FileNames = files
	return file, nil
}

func getOpening(scanner *bufio.Scanner) (string, bool) {
	if scanner.Scan() {
		return scanner.Text(), true
	} else {
		return "", false
	}
}

func getRemaining(scanner *bufio.Scanner) string {
	var rem string
	for scanner.Scan() {
		rem += scanner.Text() + "\n"
	}
	return rem
}

const prefix = "querykey:"

func getNamed(line string) (string, bool) {
	is := strings.Split(line, " ")
	for _, i := range is {
		if strings.HasPrefix(i, prefix) {
			return strings.TrimPrefix(i, prefix), true
		}
	}
	return "", false
}
