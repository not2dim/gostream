package iterator

type SliceIterable[E any] []E

func (s SliceIterable[E]) Iterator() Iterator[E] {
	return SliceIterator(s)
}

func (s SliceIterable[E]) Size() (n uint64, known bool) {
	return uint64(len(s)), true
}

type sliceIterator[E any] struct {
	EmptyIterator[E]
	inner []E
	idx   int
}

func SliceIterator[T ~[]E, E any](slice T) Iterator[E] {
	return &sliceIterator[E]{
		inner: slice,
		idx:   -1,
	}
}

func (s *sliceIterator[E]) MoveNext() bool {
	if s.idx+1 >= len(s.inner) {
		return false
	}
	s.idx++
	return true
}

func (s *sliceIterator[E]) Current() E {
	return s.inner[s.idx]
}
