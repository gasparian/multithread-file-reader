package ranker

import (
	"os"
	"fmt"
	"time"
	"testing"
	"strings"
	"runtime"
	"math/rand"
)

func TestMain(m *testing.M) {
    maxProcs := 4
    numCPU := runtime.NumCPU()
    if numCPU < maxProcs {
    	maxProcs = numCPU
    }
    runtime.GOMAXPROCS(maxProcs)
	os.Exit(m.Run())
}

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

func generateLines(n int) []string {
	lines := make([]string, n)
	for i := range lines {
		randomID := rand.Intn(n)
		randomValue := rand.Intn(n)
		lines[i] = fmt.Sprintf("http://api.tech.com/item/%v %v", randomID, randomValue)
	}
	return lines
}

func TestPerfRanker(t *testing.T) {
	n := 1000000
	topK := 10
	lines := generateLines(n)

	t.Run("SingleWorker", func(t *testing.T) {
	    r := NewRanker(1, topK)
        start := time.Now()
	    for _, str := range lines {
	    	r.ProcessLine(str)
	    }
	    res := r.GetRankedList()
        duration := time.Since(start)
	    t.Log("Elapsed time:", duration)
	    if len(res) != 10 {
	    	t.Fatalf("Output length should be equal to %v", topK)
	    }
	})

	t.Run("MultipleWorkers", func(t *testing.T) {
	    r := NewRanker(10, topK)
        start := time.Now()
	    for _, str := range lines {
	    	r.ProcessLine(str)
	    }
	    res := r.GetRankedList()
        duration := time.Since(start)
	    t.Log("Elapsed time:", duration)
	    if len(res) != 10 {
	    	t.Fatal("Output length should be equal to 10")
	    }
	})
}

func TestPerfFileParser(t *testing.T) {
	n := 1000000
	topK := 10
	lines := generateLines(n)
	bufSize := 128*1024

	path := "/tmp/clickhouse-file-reader-test-ranker-large"
	defer os.RemoveAll(path)

	builder := strings.Builder{}
	for _, l := range lines {
		_, err := builder.WriteString(l+"\n")
		if err != nil {
			t.Fatal(err)
		}
	}
	err := os.WriteFile(path, []byte(builder.String()), 0644)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("SingleWorker", func(t *testing.T) {
        start := time.Now()
	    _, err := ProcessFile(path, bufSize, 1, topK)
        duration := time.Since(start)
	    t.Log("Elapsed time:", duration)
	    if err != nil {
	    	t.Fatal(err)
	    }
	})

	t.Run("MultipleWorkers", func(t *testing.T) {
        start := time.Now()
	    _, err := ProcessFile(path, bufSize, 10, topK)
        duration := time.Since(start)
	    t.Log("Elapsed time:", duration)
	    if err != nil {
	    	t.Fatal(err)
	    }
	})
}