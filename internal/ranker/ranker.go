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

func (r *Ranker) processSegment(fileSegment *io.FileSegmentPointer) (*heap.InvertedBoundedHeap[*record.Record], error) {
	f, err := os.Open(fileSegment.Fpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	f.Seek(fileSegment.Start, 0)
	s := bufio.NewScanner(f)
	buf := make([]byte, 0)
	s.Buffer(buf, fileSegment.BufSize)
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

func (r *Ranker) worker(wg *sync.WaitGroup) {
	for fileSegmentPointer := range r.inputChan {
		h, err := r.processSegment(fileSegmentPointer)
		if err != nil {
			log.Println("Error: cannot process file segment: ", err)
			continue
		}
		r.heapsChan <- h
	}
	wg.Done()
}

func validateRankerParams(nWorkers, topK int) error {
	if topK < 1 {
		return fmt.Errorf("error: `topK` should be >= 1")
	}
	if nWorkers <= 0 {
		return fmt.Errorf("error: `nWorkers` should be a non-zero positive number")
	}
	if nWorkers > 1023 {
		nWorkers = 1023
		log.Printf("info: number of workers decreased from %v to 1023, since 1024 is a soft limit (for Linux)\n", nWorkers)
	}
	return nil
}

// NewRanker creates new instance of the ranker
func NewRanker(nWorkers, topK int) (*Ranker, error) {
	err := validateRankerParams(nWorkers, topK)
	if err != nil {
		return nil, err
	}
	r := &Ranker{
		inputChan: make(chan *io.FileSegmentPointer),
		heapsChan: make(chan *heap.InvertedBoundedHeap[*record.Record]),
		config: rankerConfig{
			topK:     topK,
			nWorkers: nWorkers,
		},
	}
	go func() {
		wg := &sync.WaitGroup{}
		for i := 0; i < nWorkers; i++ {
			wg.Add(1)
			go r.worker(wg)
		}
		wg.Wait()
		close(r.heapsChan)
	}()
	return r, nil
}

// GetRankedList merges heaps produced by mappers and
// outputs slice of topk ranked urls
func (r *Ranker) GetRankedList() []string {
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
	// invert an order of elements, since we're maintaining min heap
	// but we need highest values first in result
	for i := topK - 1; i >= 0; i-- {
		v := finalHeap.Pop()
		result[i] = v.Url
	}
	return result
}

// EmitFileSegments starts parsing the file and emits found segments
// one by one to the input channel, then closes it to stop the workers
func (r *Ranker) EmitFileSegments(fpath string, bufSize int, segmentSize int64) error {
	segmentsChan, err := io.GetFileSegments(fpath, bufSize, segmentSize, '\n')
	if err != nil {
		return err
	}
	go func() {
		for segment := range segmentsChan {
			r.inputChan <- segment
		}
		close(r.inputChan)
	}()
	return nil
}

// ProcessFile reads file, splits it in segments and sends segments to ranker workers;
// Then it waits for the final aggregated result and returns it;
// if `segmentSize` is zero - file will not be splitted in chunks
func ProcessFile(fpath string, bufSize, nWorkers, topK int, segmentSize int64) ([]string, error) {
	if int64(bufSize) > segmentSize && segmentSize != 0 {
		return nil, errors.New("error: segment size should be larger than buffer size")
	}
	r, err := NewRanker(nWorkers, topK)
	if err != nil {
		return nil, err
	}
	err = r.EmitFileSegments(fpath, bufSize, segmentSize)
	if err != nil {
		return nil, err
	}
	rank := r.GetRankedList()
	return rank, nil
}
