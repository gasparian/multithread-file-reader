package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/gasparian/clickhouse-test-file-reader/internal/ranker"
)

var (
	MAXPROCS = 4
)

func generateRandomLine() string {
	line := fmt.Sprintf(
		"http://api.tech.com/item/%v %v",
		rand.Intn(math.MaxInt),
		rand.Intn(math.MaxInt),
	)
	return line
}

func createTempFile(nLines int) (string, error) {
	f, err := os.CreateTemp("/tmp", "filereader-perf-*")
	if err != nil {
		return "", err
	}
	defer f.Close()
	for i := 0; i < nLines; i++ {
		fmt.Fprintln(f, generateRandomLine())
	}
	fi, err := f.Stat()
	if err != nil {
		return "", err
	}
	log.Printf("Generated file size in bytes: %v\n", fi.Size())
	return f.Name(), nil
}

func processFile(fname string, topK, buffSize, nworkers int, segmentSize int64) int64 {
	start := time.Now()
	_, err := ranker.ProcessFile(fname, buffSize, nworkers, topK, segmentSize)
	duration := time.Since(start)
	if err != nil {
		log.Fatal(err)
	}
	return int64(duration)
}

func init() {
	numCPU := runtime.NumCPU()
	if numCPU < MAXPROCS {
		MAXPROCS = numCPU
	}
	runtime.GOMAXPROCS(MAXPROCS)
	log.Printf("MAXPROCS set to %v", MAXPROCS)
}

func main() {
	nLines := 2500000
	topK := 10
	bufSize := 512 * 1024
	nRuns := 10
	var defaultSegmentSize int64 = 1024 * 1024
	var avrgDuration float64 = 0
	fname, err := createTempFile(nLines)
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(fname)
	for i := 0; i <= 5; i++ {
		segmentSize := int64(math.Pow(2, float64(i))) * defaultSegmentSize
		log.Printf("--- Segment size: %v\n", segmentSize)
		for j := 0; j <= 3; j++ {
			nWorkers := math.Pow(2, float64(j))
			avrgDuration = 0
			log.Printf(">>> %v workers \n", nWorkers)
			for k := 0; k < nRuns; k++ {
				duration := processFile(fname, topK, bufSize, int(nWorkers), segmentSize)
				avrgDuration += float64(duration) / 1e6
			}
			avrgDuration /= float64(nRuns)
			log.Printf("Average elapsed time: %v ms\n", int(avrgDuration))
			log.Println("---------------------")
		}
		log.Println()
	}
}
