package file

import (
	"os"
	"bufio"
	"github.com/gasparian/clickhouse-test-file-reader/internal/ranker"
)

func Process(path string, bufSize, nWorkers, topK int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	s := bufio.NewScanner(f)
	buf := make([]byte, 0)
	s.Buffer(buf, bufSize)

	r := ranker.NewRanker(nWorkers, topK)
	for s.Scan() {
		text := s.Text()
		if len(text) > 0 {
			r.ProcessLine(text)
		}
		if err := s.Err(); err != nil {
			return nil, err
		}
	}
	r.CloseInputChan()
	return r.GetRank(), nil
}