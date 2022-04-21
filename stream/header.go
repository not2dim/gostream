package stream

import "github.com/not2dim/gostream/iterator"

type header[E any] struct {
	base[E]
	src iterator.Iterable[E]
}

func (h *header[E]) GetSource() iterator.Iterable[any] {
	return wrapIterable(h.src)
}

func newHeader[E any](meta *meta, source iterator.Iterable[E]) *header[E] {
	ret := &header[E]{src: source}
	ret.base = base[E]{
		Meta: meta,
		Prev: nil,
	}
	ret.Curr = ret
	return ret
}

type emptyHeader[E any] struct {
	base[E]
}

func (h *emptyHeader[E]) GetSource() iterator.Iterable[any] {
	return wrapIterable[E](iterator.EmptyIterable[E]{})
}

func newEmptyHeader[E any]() *emptyHeader[E] {
	return &emptyHeader[E]{
		base[E]{
			Meta: defaultMeta.Copy().SetDistinct(true).SetMaxSize(0),
		},
	}
}
