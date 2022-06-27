package main

import (
	"flag"
	"log"
	"github.com/gasparian/clickhouse-test-file-reader/internal/file"
	"github.com/gasparian/clickhouse-test-file-reader/internal/io"
)

func main() {
	nWorkers := flag.Int("workers", 4, "number of workers to process lines")
	topK := flag.Int("topk", 2, "number of top k elements to return")
	bufSize := flag.Int("buf", 1024*1024, "size of buffer to reaed lines from file")
	flag.Parse()

	// path, _ := io.ParseInputPath() // UNCOMMENT BEFORE SENDING
	res, err := file.Process(
		"./data/file1",
		*bufSize,
		*nWorkers,
		*topK,
	)
	if err != nil {
		log.Fatal(err)
	}
	io.PrintResult(res)
	// For default example, the output should be: 
    // http://api.tech.com/item/122345
    // http://api.tech.com/item/124345

	// http://api.tech.com/item/121345  9
    // http://api.tech.com/item/122345  350
    // http://api.tech.com/item/123345  25
    // http://api.tech.com/item/124345  231
    // http://api.tech.com/item/125345  111
}

// TODO: 
//     - add size limit to the heap
//     - implement simple unit tests
//     - implement performance tests for large file (generate 1 mln random lines on the fly)
//     - update README with a brief solution decsription
//     - send ZIP with solution with the proposal of adding devs to the private github repo
