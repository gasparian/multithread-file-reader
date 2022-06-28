package io

import (
	"os"
    "testing"
)

func TestCheckValidPath(t *testing.T) {
	path := "/tmp/clickhouse-file-reader-test"
	defer os.RemoveAll(path)
	err := os.WriteFile(path, []byte{}, 0644)
    if err != nil {
		t.Fatal(err)
	}
	err = checkValidPath(path)
	if err != nil {
		t.Fatal(err)
	}

	err = checkValidPath("/tmp/this-folder-not-exist")
	if err == nil {
		t.Fatal()
	}
}
