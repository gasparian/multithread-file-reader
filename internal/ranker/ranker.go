package ranker

import (
	"os"
	"bufio"
	"log"
	"sync"
	"github.com/gasparian/clickhouse-test-file-reader/pkg/heap"
	"github.com/gasparian/clickhouse-test-file-reader/internal/record"
)

// instead of max we do min here, to maintian heap of constant size
// we then have to reverse order of elements that we get from the heap
// to get topk values
func comparator(a, b *record.Record) bool {
	return a.Value < b.Value
}

type rankerConfig struct {
	sync.RWMutex
	topK int
	nWorkers int
}

func (rc *rankerConfig) getTopK() int {
	rc.RLock()
	defer rc.RUnlock()
    return rc.topK
}

func (rc *rankerConfig) getNWorkers() int {
	rc.RLock()
	defer rc.RUnlock()
    return rc.nWorkers
}

// Ranker holds channels for communicating between processing stages
// and methods for parsing and ranking input text data
type Ranker struct {
	inputChan chan string
	heapsChan chan *heap.InvertedBoundedHeap[*record.Record]
	config rankerConfig
}

func (r *Ranker) mapper(wg *sync.WaitGroup) {
	h := heap.NewHeap(comparator, r.config.getTopK(), nil)
	for str := range r.inputChan {
		record, err := record.ParseRecord(str)
		if err != nil {
			log.Println("Warning: line parsing failed with error: ", err)
			continue
		}
		h.Push(record)
	}
	r.heapsChan <- h
	wg.Done()
}

// NewRanker creates new instance of the ranker
func NewRanker(nWorkers, topK int) *Ranker {
	r := &Ranker{
	    inputChan: make(chan string, 1000),
	    heapsChan: make(chan *heap.InvertedBoundedHeap[*record.Record], nWorkers),
		config: rankerConfig{
			topK: topK, 
			nWorkers: nWorkers,
	    },
	}
	go func() {
		wg := &sync.WaitGroup{}
        for i:=0; i < nWorkers; i++ {
			wg.Add(1) 
	        go r.mapper(wg)
        }
		wg.Wait()
		close(r.heapsChan)
	}()
	return r
}

// ProcessLine sends text line to the channel for parsing
func (r *Ranker) ProcessLine(line string) {
	r.inputChan <- line
}

// GetRankedList merges heaps produced by mappers and 
// outputs slice of topk ranked urls
func (r *Ranker) GetRankedList() []string {
    // close inputs channels to stop recieving new 
	// records and gracefully stop mappers
	close(r.inputChan) 
	topK := r.config.getTopK()
	finalHeap := heap.NewHeap(comparator, topK, nil)
	for h := range r.heapsChan {
		finalHeap.Merge(h)
	}
	if finalHeap.Len() == 0 {
		return []string{}
	}
	if finalHeap.Len() < topK {
		topK = finalHeap.Len()
	}
	result := make([]string, topK)
	// invert an order of elements, since we're mainting min heap
	// but we need highest values first in result
	for i:=topK-1; i >= 0; i-- {
		v := finalHeap.Pop()
		result[i] = v.Url
	}
	return result
}

// ProcessFile reads file in chunks, sends text data to ranker
// and waits for the final aggregated result
func ProcessFile(path string, bufSize, nWorkers, topK int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	s := bufio.NewScanner(f)
	buf := make([]byte, 0)
	s.Buffer(buf, bufSize)
	r := NewRanker(nWorkers, topK)
	for s.Scan() {
		text := s.Text()
		if len(text) > 0 {
			r.ProcessLine(text)
		}
		if err := s.Err(); err != nil {
			return nil, err
		}
	}
	return r.GetRankedList(), nil
}