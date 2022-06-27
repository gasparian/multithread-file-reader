package ranker

import (
	"sync"
	"github.com/gasparian/clickhouse-test-file-reader/pkg/heap"
	"github.com/gasparian/clickhouse-test-file-reader/internal/record"
)

type rankerConfig struct {
	sync.RWMutex
	topK int
	nWorkers int
}

type Ranker struct {
	inputChan chan string
	heapsChan chan *heap.Heap[*record.Record]
	config rankerConfig
}

func (r *Ranker) GetTopK() int {
	r.config.RLock()
	defer r.config.RUnlock()
    return r.config.topK
}

func (r *Ranker) GetNWorkers() int {
	r.config.RLock()
	defer r.config.RUnlock()
    return r.config.nWorkers
}

func (r *Ranker) mapper(wg *sync.WaitGroup) {
	h := heap.NewHeap(record.CompareRecords, nil)
	for str := range r.inputChan {
		record, err := record.ParseRecord(str)
		if err != nil {
			continue
		}
		h.Push(record)
	}
	r.heapsChan <- h
	wg.Done()
}

func NewRanker(nWorkers, topK int) *Ranker {
	r := &Ranker{
	    inputChan: make(chan string),
	    heapsChan: make(chan *heap.Heap[*record.Record], nWorkers),
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

func (r *Ranker) ProcessLine(line string) {
	r.inputChan <- line
}

func (r *Ranker) CloseInputChan() {
	close(r.inputChan)
}

func (r *Ranker) GetRank() []string {
	finalHeap := heap.NewHeap(record.CompareRecords, nil)
	topK := r.GetTopK()
	for h := range r.heapsChan {
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