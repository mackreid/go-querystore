package goquerystore

import "testing"

func TestMain(t *testing.T) {
	f := &FileKeyStore{}
	f.Load("./sql")
	t.Log(f.Store)
}
