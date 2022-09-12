package heap

import (
	"sort"
	"testing"
)

func TestMinInvertedHeap(t *testing.T) {
	h := NewHeap(func(a, b int) bool { return a > b }, 3, nil)
	h.Push(1)
	h.Push(2)
	if h.Len() != 2 {
		t.Fatalf("Expected: %v, but got: %v\n", 2, h.Len())
	}
	v := h.Pop()
	if v != 2 {
		t.Fatalf("Expected: %v, but got: %v\n", 2, v)
	}
	h.Push(2)
	h.Push(3)
	v = h.Push(4)
	if v != 4 {
		t.Fatalf("Expected: %v, but got: %v\n", 3, v)
	}
	if h.Len() > 3 {
		t.Fatalf("Expected: %v, but got: %v\n", 3, h.Len())
	}
}

func TestMinInvertedHeapFromSlice(t *testing.T) {
	data := []int{1, 3, 0, 2, 12, 10}
	trueOrder := make([]int, len(data))
	copy(trueOrder, data)
	sort.Slice(
		trueOrder, func(i, j int) bool {
			return trueOrder[i] > trueOrder[j]
		},
	)
	h := NewHeap(func(a, b int) bool { return a > b }, 100, data)
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

func TestMergeMinInvertedHeaps(t *testing.T) {
	data := []int{1, 3, 0, 2, 12, 10, 99, -2}
	trueOrder := make([]int, len(data))
	copy(trueOrder, data)
	sort.Slice(
		trueOrder, func(i, j int) bool {
			return trueOrder[i] > trueOrder[j]
		},
	)
	comp := func(a, b int) bool { return a > b }
	maxSize := 100
	h1 := NewHeap(comp, maxSize, data[:len(data)/2])
	h2 := NewHeap(comp, maxSize, data[len(data)/2:])
	h1.Merge(h2)
	for i := range trueOrder {
		v := h1.Pop()
		if v != trueOrder[i] {
			t.Fatal()
		}
	}
}

func TestMergeMinInvertedHeapsBounded(t *testing.T) {
	data := []int{1, 3, 0, 2, 12, 10, 99, -2}
	trueOrder := make([]int, len(data))
	copy(trueOrder, data)
	sort.Slice(
		trueOrder, func(i, j int) bool {
			return trueOrder[i] > trueOrder[j]
		},
	)
	maxSize := 6
	trueOrder = trueOrder[len(trueOrder)-maxSize:]
	comp := func(a, b int) bool { return a > b }
	h1 := NewHeap(comp, maxSize, data[:len(data)/2])
	h2 := NewHeap(comp, maxSize, data[len(data)/2:])
	h1.Merge(h2)
	for i := range trueOrder {
		v := h1.Pop()
		if v != trueOrder[i] {
			t.Fatalf("Expected: %v, but got: %v\n", trueOrder[i], v)
		}
	}
}

func TestMergeMaxInvertedHeapsBoundedSmall(t *testing.T) {
	data := []int{9, 350, 25, 231, 111}
	trueOrder := []int{350, 231}
	maxSize := 2
	comp := func(a, b int) bool { return a < b }
	h1 := NewHeap(comp, maxSize, data[:2])
	h2 := NewHeap(comp, maxSize, data[2:4])
	h3 := NewHeap(comp, maxSize, data[4:])
	h1.Merge(h2)
	h1.Merge(h3)
	// 2 highest values has been dropped during the merge
	for i := maxSize - 1; i >= 0; i-- {
		v := h1.Pop()
		if v != trueOrder[i] {
			t.Fatalf("Expected: %v, but got: %v\n", trueOrder[i], v)
		}
	}
}

func TestMergeMaxInvertedHeapsBoundedSmallPush(t *testing.T) {
	data := []int{9, 350, 25, 231, 111}
	trueOrder := []int{350, 231}
	maxSize := 2
	comp := func(a, b int) bool { return a < b }
	h1 := NewHeap(comp, maxSize, nil)
	for _, d := range data {
		h1.Push(d)
	}
	h2 := NewHeap(comp, maxSize, nil)
	h1.Merge(h2)
	// 2 highest values has been dropped during the merge
	for i := maxSize - 1; i >= 0; i-- {
		v := h1.Pop()
		if v != trueOrder[i] {
			t.Fatalf("Expected: %v, but got: %v\n", trueOrder[i], v)
		}
	}
}
