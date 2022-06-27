package record

import (
	"fmt"
	"errors"
	"strings"
	"strconv"
)

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