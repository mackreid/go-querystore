package goquerystore

import (
	"testing"
)

func TestMain(t *testing.T) {
	f, _ := New("./sql")
	t.Log(f.Get("insert"))
}
