package heap

import (
	"sort"
	"testing"
)

func TestHeap(t *testing.T) {
	h := NewHeap(func(a, b int) bool { return a > b }, nil)
	h.Push(1)
	h.Push(2)
	if h.Len() != 2 {
		t.Fatal()
	}
	v := h.Pop()
	if v != 2 {
		t.Fatalf("Expected: %v, but got: %v\n", 1, v)
	}
}

func TestHeapFromSlice(t *testing.T) {
	data := []int{1, 3, 0, 2, 12, 10}
	trueOrder := make([]int, len(data))
	copy(trueOrder, data)
	sort.Slice(
		trueOrder, func(i, j int) bool {
			return trueOrder[i] > trueOrder[j]
		},
	)
	h := NewHeap(func(a, b int) bool { return a > b }, data)
	if h.Len() != 6 {
		t.Fatal()
	}
	newMaxVal := 13
	h.Push(newMaxVal)
	v := h.Pop()
	if v != newMaxVal {
		t.Fatalf("Expected: %v, but got: %v\n", newMaxVal, v)
	}
	for i := range trueOrder {
		v = h.Pop()
		if v != trueOrder[i] {
		    t.Fatal()
		}
	}
}

func TestMergeHeaps(t *testing.T) {
	data := []int{1, 3, 0, 2, 12, 10, 99, -2}
	trueOrder := make([]int, len(data))
	copy(trueOrder, data)
	sort.Slice(
		trueOrder, func(i, j int) bool {
			return trueOrder[i] > trueOrder[j]
		},
	)
	comp := func(a, b int) bool { return a > b }
	h1 := NewHeap(comp, data[:len(data)/2])
	h2 := NewHeap(comp, data[len(data)/2:])
	h1.Merge(h2)
	for i := range trueOrder {
		v := h1.Pop()
		if v != trueOrder[i] {
		    t.Fatal()
		}
	}
}