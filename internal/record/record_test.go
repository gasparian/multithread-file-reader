package record

import "testing"

func TestRecordParser(t *testing.T) {
	gt := Record{
		Url:   "http://api.tech.com/item/121345",
		Value: 9,
	}
	rec, err := ParseRecord("http://api.tech.com/item/121345  9")
	if err != nil {
		t.Fatal(err)
	}
	if !Equal(rec, gt) {
		t.Fatalf("`%v` and `%v` expected to be equal", rec, gt)
	}
	_, err = ParseRecord("http://api.tech.com/item/121345  ")
	if err == nil {
		t.Fatal()
	}
}
