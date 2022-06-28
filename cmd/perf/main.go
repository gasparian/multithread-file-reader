package main

import (
	"os"
	"time"
	"fmt"
	"log"
	"strings"
	"runtime"
	"math/rand"

	"github.com/gasparian/clickhouse-test-file-reader/internal/ranker"
)

const (
	TOPK = 10
	NWORKERS = 3
	BUFSIZE = 128*1024
	FPATH = "/tmp/clickhouse-file-reader-test-ranker-large"
)

var (
    MAXPROCS = 4
	LINES = generateLines(5000000)
)

func generateLines(n int) []string {
	lines := make([]string, n)
	for i := range lines {
		randomID := rand.Intn(n)
		randomValue := rand.Intn(n)
		lines[i] = fmt.Sprintf("http://api.tech.com/item/%v %v", randomID, randomValue)
	}
	return lines
}

func testRanker(nWorkers int) {
	r := ranker.NewRanker(nWorkers, TOPK)
       start := time.Now()
	for _, str := range LINES {
		r.ProcessLine(str)
	}
	r.GetRankedList()
    duration := time.Since(start)
	log.Println("Elapsed time:", duration)
	log.Println("---------------------")
}

func TestRankerSingleWorkers() {
	log.Println(">>> Single worker test")
	testRanker(1)
}

func TestRankerMultipleWorkers() {
	log.Println(">>> Multiple workers test")
	testRanker(NWORKERS)
}

func TestFileParserWorker() {
	defer os.RemoveAll(FPATH)
	builder := strings.Builder{}
	for _, l := range LINES {
		_, err := builder.WriteString(l+"\n")
		if err != nil {
			log.Fatal(err)
		}
	}
	err := os.WriteFile(FPATH, []byte(builder.String()), 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(">>> Single worker test")
    start := time.Now()
	_, err = ranker.ProcessFile(FPATH, BUFSIZE, 1, TOPK)
    duration := time.Since(start)
	log.Println("Elapsed time:", duration)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("---------------------")

	log.Println(">>> Multiple workers test")
    start = time.Now()
	_, err = ranker.ProcessFile(FPATH, BUFSIZE, NWORKERS, TOPK)
    duration = time.Since(start)
	log.Println("Elapsed time:", duration)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("---------------------")
}

func init() {
    numCPU := runtime.NumCPU()
    if numCPU < MAXPROCS {
    	MAXPROCS = numCPU
    }
	runtime.GOMAXPROCS(MAXPROCS)
}

func main() {
    TestRankerSingleWorkers()
    TestRankerMultipleWorkers()
	fmt.Println()
    TestFileParserWorker()
}