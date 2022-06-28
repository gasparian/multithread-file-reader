package ranker

import (
	"os"
	"testing"
	"strings"
)

func TestProcessFile(t *testing.T) {
	path := "/tmp/clickhouse-file-reader-test-ranker"
	defer os.RemoveAll(path)
	data := []byte(`
http://api.tech.com/item/121345  9
http://api.tech.com/item/122345  350
http://api.tech.com/item/123345  25
http://api.tech.com/item/124345  231
http://api.tech.com/item/125345  111

`)
	err := os.WriteFile(path, data, 0644)
	if err != nil {
		t.Fatal(err)
	}
	gt := []string{
		"http://api.tech.com/item/122345",
		"http://api.tech.com/item/124345",
	}

	bufSize := 128*1024
	topK := 2

	t.Run("SingleWorker", func(t *testing.T) {
	    res, err := ProcessFile(path, bufSize, 1, topK)
	    if err != nil {
	    	t.Fatal(err)
	    }
	    for i, r := range res {
	    	if strings.Compare(r, gt[i]) != 0 {
	    		t.Fatalf("Expected `%v` but got `%v`", gt[i], r)
	    	}
	    }
	})

	t.Run("MultipleWorkers", func(t *testing.T) {
	    res, err := ProcessFile(path, bufSize, 4, topK)
	    if err != nil {
	    	t.Fatal(err)
	    }
	    for i, r := range res {
	    	if strings.Compare(r, gt[i]) != 0 {
	    		t.Fatalf("Expected `%v` but got `%v`", gt[i], r)
	    	}
	    }
	})
}

func TestRankerShortSeq(t *testing.T) {
	r := NewRanker(2, 1)
	data := []string{
        "http://api.tech.com/item/121345  9",
        "http://api.tech.com/item/122345  350",
	}
	for _, str := range data {
		r.ProcessLine(str)
	}
	res := r.GetRankedList()
	if len(res) != 1 || strings.Compare(res[0], "http://api.tech.com/item/122345") != 0 {
		t.Fatalf("Output rank is wrong: %v", res)
	}
}

func TestRankerEmptyInput(t *testing.T) {
	r := NewRanker(2, 2)
	data := []string{"\n", ""}
	for _, str := range data {
		r.ProcessLine(str)
	}
	res := r.GetRankedList()
	if len(res) != 0 {
		t.Fatal("Output should be empty slice")
	}
}
