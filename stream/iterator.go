package stream

import "github.com/not2dim/gostream/iterator"

// region anyIterable

type anyIterable[E any] struct {
	iterator.EmptyIterator[E]
	inner iterator.Iterable[E]
}

func (i anyIterable[E]) Iterator() iterator.Iterator[any] {
	return wrapIterator[E](i.inner.Iterator())
}

func (i anyIterable[E]) Size() (n uint64, known bool) {
	return i.inner.Size()
}

func wrapIterable[E any](itera iterator.Iterable[E]) iterator.Iterable[any] {
	return anyIterable[E]{inner: itera}
}

func unwrapIterable[E any](anyItera iterator.Iterable[any]) iterator.Iterable[E] {
	iter, ok := anyItera.(anyIterable[E])
	if !ok {
		panic("failed to unwrap anyIterator[E]")
	}
	return iter.inner
}

type anyIterator[E any] struct {
	iterator.EmptyIterator[E]
	inner iterator.Iterator[E]
}

func (i anyIterator[E]) MoveNext() bool {
	return i.inner.MoveNext()
}

func (i anyIterator[E]) Current() any {
	return i.inner.Current()
}

func wrapIterator[E any](iter iterator.Iterator[E]) iterator.Iterator[any] {
	return anyIterator[E]{inner: iter}
}

func unwrapIterator[E any](anyIter iterator.Iterator[any]) iterator.Iterator[E] {
	iter, ok := anyIter.(anyIterator[E])
	if !ok {
		panic("failed to unwrap anyIterator[E]")
	}
	return iter.inner
}

// endregion

// region rangeIterable

type rangeIterable[E integer | uinteger] struct {
	from, to E
}

func newRangeIterable[E integer | uinteger](from, to E) iterator.Iterable[E] {
	if to < from {
		to = from
	}
	return rangeIterable[E]{from, to}
}

func (r rangeIterable[E]) Iterator() iterator.Iterator[E] {
	return &rangeIterator[E]{
		from: r.from,
		to:   r.to,
		curr: r.from - 1,
	}
}

func (r rangeIterable[E]) Size() (n uint64, known bool) {
	return uint64(r.to - r.from), true
}

type rangeIterator[E integer | uinteger] struct {
	iterator.EmptyIterator[E]
	from, to, curr E
}

func (r *rangeIterator[E]) MoveNext() bool {
	if r.curr+1 >= r.to {
		return false
	}
	r.curr++
	return true
}

func (r *rangeIterator[E]) Current() E {
	return r.curr
}

// endregion
