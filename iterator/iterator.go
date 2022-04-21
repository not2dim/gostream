package iterator

type Iterable[E any] interface {
	// Iterator returns an Iterator
	Iterator() Iterator[E]
	// Size returns 'n' implying the count of elements, and 'known' equals true when the elements are countable.
	Size() (n uint64, known bool)
}

type Iterator[E any] interface {
	// MoveNext moves to the next element, returns true if the next element exists and false otherwise.
	MoveNext() bool
	// Current returns the current element.
	Current() E
	// Close release resources.
	Close()
}

type EmptyIterable[E any] struct{}

func (e EmptyIterable[E]) Iterator() Iterator[E] {
	return EmptyIterator[E]{}
}

func (e EmptyIterable[E]) Size() (n uint64, known bool) {
	return 0, true
}

type EmptyIterator[E any] struct{}

func (e EmptyIterator[E]) MoveNext() bool {
	return false
}

func (e EmptyIterator[E]) Current() E {
	panic("empty iterator contains no values!")
}

func (e EmptyIterator[E]) Close() {}
