package heap

// Ref.: https://gist.github.com/nwillc/554847806891a41e7bd32041308dfb40#file-go_generics_heap-go

// InvertedBoundedHeap holds generic heap implementation
// Main idea: use min heap as max heap, since it's very convinient to
// drop smallest values when the maxSize exceeded
type InvertedBoundedHeap[T any] struct {
	data    []T
	comp    func(a, b T) bool
	maxSize int
}

// NewHeap creates new instance of InvertedBoundedHeap, by comparator and optionally provided array
func NewHeap[T any](comp func(a, b T) bool, maxSize int, data []T) *InvertedBoundedHeap[T] {
	h := &InvertedBoundedHeap[T]{comp: comp, maxSize: maxSize}
	if data != nil {
		h.data = data
		h.build()
	}
	return h
}

// Len returns size of InvertedBoundedHeap
func (h *InvertedBoundedHeap[T]) Len() int { return len(h.data) }

// Push adds new element to InvertedBoundedHeap
// it will return value from the top of the heap if size limit exceeded (be careful!)
func (h *InvertedBoundedHeap[T]) Push(v T) T {
	h.data = append(h.data, v)
	h.up(h.Len() - 1)
	var v_ T
	if h.Len() > h.maxSize {
		v_ = h.Pop()
	}
	return v_
}

// Pop removes and returns top element from InvertedBoundedHeap
func (h *InvertedBoundedHeap[T]) Pop() T {
	n := h.Len() - 1
	if n > 0 {
		h.swap(0, n)
		h.down()
	}
	v := h.data[n]
	h.data = h.data[0:n]
	return v
}

// Merge merges current heap with the provided one
func (h *InvertedBoundedHeap[T]) Merge(inputHeap *InvertedBoundedHeap[T]) []T {
	h.data = append(h.data, inputHeap.data...)
	return h.build()
}

func (h *InvertedBoundedHeap[T]) swap(i, j int) {
	h.data[i], h.data[j] = h.data[j], h.data[i]
}

func (h *InvertedBoundedHeap[T]) up(jj int) {
	for {
		i := parent(jj)
		if i == jj || !h.comp(h.data[jj], h.data[i]) {
			break
		}
		h.swap(i, jj)
		jj = i
	}
}

func (h *InvertedBoundedHeap[T]) down() {
	n := h.Len() - 1
	i1 := 0
	for {
		j1 := left(i1)
		if j1 >= n || j1 < 0 {
			break
		}
		j := j1
		j2 := right(i1)
		if j2 < n && h.comp(h.data[j2], h.data[j1]) {
			j = j2
		}
		if !h.comp(h.data[j], h.data[i1]) {
			break
		}
		h.swap(i1, j)
		i1 = j
	}
}

func (h *InvertedBoundedHeap[T]) heapify(i int) {
	largest := i
	left := left(i)
	right := right(i)
	len := h.Len()
	if left < len && h.comp(h.data[left], h.data[largest]) {
		largest = left
	}
	if right < len && h.comp(h.data[right], h.data[largest]) {
		largest = right
	}
	if largest != i {
		h.swap(i, largest)
		h.heapify(largest)
	}
}

func (h *InvertedBoundedHeap[T]) build() []T {
	remaining := make([]T, 0)
	startIdx := h.Len()/2 - 1
	for i := startIdx; i >= 0; i-- {
		h.heapify(i)
	}
	for h.Len() > h.maxSize {
		remaining = append(remaining, h.Pop())
	}
	return remaining
}

func parent(i int) int { return (i - 1) / 2 }
func left(i int) int   { return (i * 2) + 1 }
func right(i int) int  { return left(i) + 1 }
