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

func generateRandomURLStat() (string, int) {
	url := fmt.Sprintf("http://api.tech.com/item/%v", rand.Intn(math.MaxInt))
	val := rand.Intn(math.MaxInt)
	return url, val
}

func createTempFile(nLines int) (string, string, error) {
	f, err := os.CreateTemp("/tmp", "filereader-perf-*")
	if err != nil {
		return "", "", err
	}
	defer f.Close()
	var maxVal int = math.MinInt
	var maxValUrl string
	for i := 0; i < nLines; i++ {
		url, val := generateRandomURLStat()
		if val > maxVal {
			maxVal = val
			maxValUrl = url
		}
		line := fmt.Sprintf("%v  %v\n", url, val)
		fmt.Fprintln(f, line)
	}
	fi, err := f.Stat()
	if err != nil {
		return "", "", err
	}
	log.Printf("Generated file size in bytes: %v\n", fi.Size())
	return f.Name(), maxValUrl, nil
}

func processFile(fname string, topK, buffSize, nworkers int, segmentSize int64) (int64, []string) {
	start := time.Now()
	rank, err := ranker.ProcessFile(fname, buffSize, nworkers, topK, segmentSize)
	duration := time.Since(start)
	if err != nil {
		log.Fatal(err)
	}
	return int64(duration), rank
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
	bufSize := 1024 * 1024
	nRuns := 10
	var defaultSegmentSize int64 = 1024 * 1024
	var avrgDuration float64 = 0
	fname, maxValUrl, err := createTempFile(nLines)
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(fname)
	for i := 0; i <= 4; i += 2 {
		segmentSize := int64(math.Pow(2, float64(i))) * defaultSegmentSize
		log.Printf("--- Segment size: %v b\n", segmentSize)
		for j := 0; j <= 3; j++ {
			nWorkers := math.Pow(2, float64(j))
			avrgDuration = 0
			log.Printf(">>> %v workers \n", nWorkers)
			for k := 0; k < nRuns; k++ {
				duration, rank := processFile(fname, topK, bufSize, int(nWorkers), segmentSize)
				if rank[0] != maxValUrl {
					log.Fatalf("%s should be top record, but got %s\n", maxValUrl, rank[0])
				}
				avrgDuration += float64(duration) / 1e6
			}
			avrgDuration /= float64(nRuns)
			log.Printf("Average elapsed time: %v ms\n", int(avrgDuration))
			log.Println("---------------------")
		}
		log.Println()
	}
}
