package main

import (
	"log"
	"fmt"
	"os"
	"bufio"
	"errors"
	"strings"
	"strconv"
	"sync"
	"github.com/gasparian/clickhouse-test-file-reader/pkg/heap"
)

func checkValidPath(path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return err
	}
	return err
}

func ParseInputPath() (string, error) {
	reader := bufio.NewReader(os.Stdin)
    fmt.Print("Enter path to file: ")
    path, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	path = strings.TrimRight(path, "\n")
    err = checkValidPath(path)
	if err != nil {
		return "", err
	}
	return path, nil
}

type Record struct {
	Url string
	Value int64
}

func ParseRecord(str string) (*Record, error) {
	strSlice := strings.Fields(str)
	strSliceLen := len(strSlice)
	if strSliceLen > 2 || strSliceLen == 0 {
		return nil, errors.New(fmt.Sprintf("Record should consist of exactly 2 fields, but got %v", strSliceLen))
	}
	parsedVal, err := strconv.ParseInt(strSlice[1], 10, 64)
	if err != nil {
		return nil, err
	}
	record := &Record{
		Url: strSlice[0],
		Value: parsedVal,
	}
	return record, nil
}

func CompareRecords(a, b *Record) bool {
	return a.Value > b.Value
}

func Map(inp chan string, output chan *heap.Heap[*Record], wg *sync.WaitGroup) {
	h := heap.NewHeap(CompareRecords, nil)
	for str := range inp {
		record, err := ParseRecord(str)
		if err != nil {
			continue
		}
		h.Push(record)
	}
	output <- h
	wg.Done()
}

func Reduce(mapOutPut chan *heap.Heap[*Record], topK int) []string {
	finalHeap := heap.NewHeap(CompareRecords, nil)
	for h := range mapOutPut {
		finalHeap.Merge(h)
	}
	if finalHeap.Len() < topK {
		topK = finalHeap.Len()
	}
	result := make([]string, topK)
	for i:=0; i < topK; i++ {
		result[i] = finalHeap.Pop().Url
	}
	return result
}

func PrintResult(res []string) {
	for _, r := range res {
		fmt.Println(r)
	}
}

func ProcessFile(path string, chunkSizeBytes int) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error to read file %v: %v", path, err.Error())
	}
	r := bufio.NewScanner(f)
	buf := make([]byte, 0)
	r.Buffer(buf, chunkSizeBytes)
	nWorkers := 2 // no. of mapper
	inpChan := make(chan string)
	outputChan := make(chan *heap.Heap[*Record], nWorkers)
	go func() {
		wg := &sync.WaitGroup{}
        for i:=0; i < nWorkers; i++ {
			wg.Add(1) 
	        go Map(inpChan, outputChan, wg)
        }
		wg.Wait()
		close(outputChan)
	}()
	for r.Scan() {
		text := r.Text()
		if len(text) > 0 {
			// here the processing happens
		    // log.Println(text, "|", len(text))
			inpChan <- text
		}
		if err := r.Err(); err != nil {
			log.Fatal(err)
		}
	}
	close(inpChan)
	res := Reduce(outputChan, 2)
	PrintResult(res)
	// For default example, the output should be: 
    // http://api.tech.com/item/122345
    // http://api.tech.com/item/124345
}

func main() {
	// path, _ := ParseInputPath()
	ProcessFile("./data/file1", 40)
}