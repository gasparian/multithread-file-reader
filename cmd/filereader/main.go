package main

import (
	"flag"
	"log"

	"github.com/gasparian/clickhouse-test-file-reader/internal/io"
	"github.com/gasparian/clickhouse-test-file-reader/internal/ranker"
)

func main() {
	nWorkers := flag.Int("workers", 4, "number of workers to process lines")
	topK := flag.Int("topk", 10, "number of top k elements to return")
	bufSize := flag.Int("buf", 1024*1024, "size of buffer to read lines from file")
	segmentSize := flag.Int64("segment", 1024*1024, "size of the file segment in bytes to be processed by a single worker")
	flag.Parse()

	path, _ := io.ParseInputPath()
	res, err := ranker.ProcessFile(
		path,
		*bufSize,
		*nWorkers,
		*topK,
		*segmentSize,
	)
	if err != nil {
		log.Fatal(err)
	}
	io.PrintResult(res)
}
