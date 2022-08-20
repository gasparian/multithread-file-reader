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

func TestGetFileSegments(t *testing.T) {
	path := "/tmp/clickhouse-file-reader-test-io"
	defer os.RemoveAll(path)
	data := []byte(`
http://api.tech.com/item/121345  9
http://api.tech.com/item/122345  350
http://api.tech.com/item/123345  25
http://api.tech.com/item/124345  231

`)
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	segmentsSizes := []int64{72, 73}

	segments, err := GetFileSegments(f, 64, 64, '\n')
	if err != nil {
		t.Fatal(err)
	}
	t.Log(segments)
	if len(segments) > 2 {
		t.Fatalf("Segments array size should be = 2, but got %v", len(segments))
	}
	for i := range segments {
		if segments[i].Len != segmentsSizes[i] {
			t.Fatalf("Segment length should be = %v, but got %v", segmentsSizes[i], segments[i].Len)
		}
	}
}
