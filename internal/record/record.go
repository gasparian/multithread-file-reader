package record

import (
	"fmt"
	"errors"
	"strings"
	"strconv"
)

// Record holds url data presented in files
type Record struct {
	Url string
	Value int64
}

// ParseRecord parses input string and creates Record object from it
func ParseRecord(str string) (*Record, error) {
	strSlice := strings.Fields(str)
	strSliceLen := len(strSlice)
	if strSliceLen != 2 {
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

// Equal small helper function to compare two Records
func Equal(a, b *Record) bool {
	return strings.Compare(a.Url, b.Url) == 0 && 
	    a.Value == b.Value
}