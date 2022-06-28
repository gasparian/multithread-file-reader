package main

import (
	"flag"
	"log"
	"github.com/gasparian/clickhouse-test-file-reader/internal/ranker"
	"github.com/gasparian/clickhouse-test-file-reader/internal/io"
)

func main() {
	nWorkers := flag.Int("workers", 10, "number of workers to process lines")
	topK := flag.Int("topk", 10, "number of top k elements to return")
	bufSize := flag.Int("buf", 128*1024, "size of buffer to reaed lines from file")
	flag.Parse()

	path, _ := io.ParseInputPath()
	res, err := ranker.ProcessFile(
		path,
		*bufSize,
		*nWorkers,
		*topK,
	)
	if err != nil {
		log.Fatal(err)
	}
	io.PrintResult(res)
}
