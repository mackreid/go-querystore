package goquerystore

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FileKeyStore struct {
	Store     map[string]string
	FileCount int
}

type result struct {
	named    string
	contents string
}

func fileMap(files []string, cs chan result) {
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
		cs <- result{named: n, contents: c}
	}
}

func (f *FileKeyStore) Load(directory string) error {
	files := []string{}
	err := filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
		if filepath.Ext(path) == ".sql" {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	cs := make(chan result)
	go fileMap(files, cs)
	m := make(map[string]string)
	for res := range cs {
		if _, ok := m[res.named]; ok {
			return fmt.Errorf("keys cannot be duplicate: %s", res.named)
		}
		m[res.named] = res.contents
	}

	f.Store = m
	f.FileCount = len(m)
	return nil
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
