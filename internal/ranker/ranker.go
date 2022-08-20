package ranker

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/gasparian/clickhouse-test-file-reader/internal/io"
	"github.com/gasparian/clickhouse-test-file-reader/internal/record"
	"github.com/gasparian/clickhouse-test-file-reader/pkg/heap"
)

// instead of max we do min here, to maintian heap of constant size
// we then have to reverse order of elements that we get from the heap
// to get topk values
func comparator(a, b *record.Record) bool {
	return a.Value < b.Value
}

type rankerConfig struct {
	sync.RWMutex
	topK     int
	nWorkers int
}

func (rc *rankerConfig) getTopK() int {
	rc.RLock()
	defer rc.RUnlock()
	return rc.topK
}

// Ranker holds channels for communicating between processing stages
// and methods for parsing and ranking input text data
type Ranker struct {
	inputChan chan *io.FileSegmentPointer
	heapsChan chan *heap.InvertedBoundedHeap[*record.Record]
	config    rankerConfig
}

func (r *Ranker) processSegment(fpath string, fileSegment *io.FileSegmentPointer, bufSize int) (*heap.InvertedBoundedHeap[*record.Record], error) {
	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	f.Seek(fileSegment.Start, 0)
	s := bufio.NewScanner(f)
	buf := make([]byte, 0)
	s.Buffer(buf, bufSize)
	var nBytesRead int64 = 0
	h := heap.NewHeap(comparator, r.config.getTopK(), nil)
	for s.Scan() {
		text := s.Text()
		if len(text) > 0 {
			record, err := record.ParseRecord(text)
			if err != nil {
				log.Println("Warning: line parsing failed with error: ", err)
				continue
			}
			h.Push(record)
		}
		if err := s.Err(); err != nil {
			return nil, err
		}
		nBytesRead += int64(len(s.Bytes()))
		if nBytesRead >= fileSegment.Len {
			break
		}
	}
	return h, nil
}

func (r *Ranker) worker(fpath string, bufSize int, wg *sync.WaitGroup) {
	for fileSegmentPointer := range r.inputChan {
		h, err := r.processSegment(fpath, fileSegmentPointer, bufSize)
		if err != nil {
			log.Println("Error: cannot process file segment: ", err)
			continue
		}
		r.heapsChan <- h
	}
	wg.Done()
}

func validatePositiveIntParams(params map[string]int) error {
	for k, v := range params {
		if v <= 0 {
			return errors.New(fmt.Sprintf("error: `%s` should be a non-zero positive number", k))
		}
	}
	return nil
}

// NewRanker creates new instance of the ranker
func NewRanker(fpath string, nWorkers, topK, bufSize int) (*Ranker, error) {
	err := validatePositiveIntParams(
		map[string]int{
			"nWorkers": nWorkers,
			"topK":     topK,
			"bufSize":  bufSize,
		},
	)
	if err != nil {
		return nil, err
	}
	r := &Ranker{
		inputChan: make(chan *io.FileSegmentPointer, 1000),
		heapsChan: make(chan *heap.InvertedBoundedHeap[*record.Record], nWorkers),
		config: rankerConfig{
			topK:     topK,
			nWorkers: nWorkers,
		},
	}
	go func() {
		wg := &sync.WaitGroup{}
		for i := 0; i < nWorkers; i++ {
			wg.Add(1)
			go r.worker(fpath, bufSize, wg)
		}
		wg.Wait()
		close(r.heapsChan)
	}()
	return r, nil
}

// SubmitSegment pushes file segment pointer to the task queue
func (r *Ranker) SubmitSegment(segment io.FileSegmentPointer) {
	r.inputChan <- &segment
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
	for i := topK - 1; i >= 0; i-- {
		v := finalHeap.Pop()
		result[i] = v.Url
	}
	return result
}

// ProcessFile reads file, splits it in segments and sends segments to ranker workers;
// Then it waits for the final aggregated result and returns it;
// if `segmentSize` is zero - file will not be splitted in chunks
func ProcessFile(fpath string, bufSize, nWorkers, topK int, segmentSize int64) ([]string, error) {
	if int64(bufSize) > segmentSize && segmentSize != 0 {
		return nil, errors.New("error: segment size should be larger than buffer size")
	}
	f, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	segments, err := io.GetFileSegments(f, bufSize, segmentSize, '\n')
	f.Close()
	if err != nil {
		return nil, err
	}
	r, err := NewRanker(fpath, nWorkers, topK, bufSize)
	if err != nil {
		return nil, err
	}
	for _, segment := range segments {
		r.SubmitSegment(segment)
	}
	return r.GetRankedList(), nil
}
