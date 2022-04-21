package stream

import "github.com/not2dim/gostream/iterator"

type pipeline interface {
	GetSource() iterator.Iterable[any]
	GetUpstream() pipeline
	WrapSink(down sink) sink
}

type sink interface {
	Begin(size uint64, known bool)
	Accept(v any)
	Rejecting() bool
	Close()
}

type base[E any] struct {
	Meta *meta
	Prev pipeline
	Curr pipeline
}

func (b *base[E]) GetSource() iterator.Iterable[any] {
	return nil
}

func (b *base[E]) GetUpstream() pipeline {
	return b.Prev
}

func (b *base[E]) WrapSink(down sink) sink {
	return baseSink{down: down}
}

type baseSink struct {
	down sink
}

// region baseSink impl

func (b baseSink) Begin(size uint64, known bool) {
	b.down.Begin(size, known)
}

func (b baseSink) Accept(v any) {
	b.down.Accept(v)
}

func (b baseSink) Rejecting() bool {
	return b.down.Rejecting()
}

func (b baseSink) Close() {
	b.down.Close()
}

// endregion

func (b *base[E]) Skip(n uint64) Stream[E] {
	if b.Meta.MaxSize() == 0 {
		return b
	} else if b.Meta.MaxSize() <= n {
		return newEmptyHeader[E]()
	}
	return newOpSkip[E](b.Meta.Copy(), b.Curr, n)
}

func (b *base[E]) Limit(n uint64) Stream[E] {
	if b.Meta.MaxSize() <= n {
		return b
	} else if b.Meta.MaxSize() == 0 {
		return newEmptyHeader[E]()
	}
	return newOpLimit[E](b.Meta.Copy(), b.Curr, n)
}

func (b *base[E]) Filter(pred func(v E) bool) Stream[E] {
	if b.Meta.MaxSize() == 0 {
		return b
	}
	return newOpFilter(b.Meta.Copy(), b.Curr, pred)
}

func (b *base[E]) Peek(act func(v E)) Stream[E] {
	if b.Meta.MaxSize() == 0 {
		return b
	}
	return newOpPeek(b.Meta.Copy(), b.Curr, act)
}

func (b *base[E]) Cond(cond func(v E) bool) Stream[E] {
	if b.Meta.MaxSize() == 0 {
		return b
	}
	return newOpCond(b.Meta.Copy(), b.Curr, cond)
}

func (b *base[E]) Distinct() Stream[E] {
	if b.Meta.Distinct() || b.Meta.MaxSize() == 0 {
		return b
	}
	return newOpDistinct[E](b.Meta.Copy(), b.Curr)
}

func (b *base[E]) DistinctBy(id func(v E) any) Stream[E] {
	if b.Meta.MaxSize() == 0 {
		return b
	}
	return newOpDistinctBy[E](b.Meta.Copy(), b.Curr, id)
}

func (b *base[E]) MinBy(cmp func(u E, v E) int) Nullable[E] {
	var min Nullable[E]
	if b.Meta.MaxSize() == 0 {
		return min
	}
	newOpForeach(b.Curr, func(v E) {
		if !min.OK {
			min.OK = true
			min.Val = v
			return
		}
		if cmp(min.Val, v) > 0 {
			min.Val = v
		}
	}, nil, nil).Terminate()
	return min
}

func (b *base[E]) MaxBy(cmp func(u E, v E) int) Nullable[E] {
	var max Nullable[E]
	if b.Meta.MaxSize() == 0 {
		return max
	}
	newOpForeach(b.Curr, func(v E) {
		if !max.OK {
			max.OK = true
			max.Val = v
			return
		}
		if cmp(max.Val, v) < 0 {
			max.Val = v
		}
	}, nil, nil).Terminate()
	return max
}

func (b *base[E]) First() Nullable[E] {
	var first Nullable[E]
	if b.Meta.MaxSize() == 0 {
		return first
	}
	newOpForCond(b.Curr, func(v E) bool {
		first = Nullable[E]{Val: v, OK: true}
		return true
	}, nil, nil).Terminate()
	return first
}

func (b *base[E]) Last() Nullable[E] {
	var last Nullable[E]
	if b.Meta.MaxSize() == 0 {
		return last
	}
	newOpForeach(b.Curr, func(v E) {
		last = Nullable[E]{Val: v, OK: true}
	}, nil, nil).Terminate()
	return last
}

func (b *base[E]) SortBy(cmp func(u E, v E) int) Stream[E] {
	if b.Meta.MaxSize() == 0 {
		return b
	}
	return newOpSortBy(b.Meta.Copy(), b.Curr, cmp)
}

func (b *base[E]) Map(mapper func(v E) E) Stream[E] {
	if b.Meta.MaxSize() == 0 {
		return b
	}
	return newOpMap(b.Meta.Copy(), b.Curr, mapper)
}

func mapToAny[S any, T any](up Stream[S], mapper func(v S) T) (down Stream[T]) {
	meta := reflectBaseMeta(up)
	if meta.MaxSize() == 0 {
		return newEmptyHeader[T]()
	}
	return newOpMapToAny(meta.Copy(), reflectBaseCurr(up), mapper)
}

func (b *base[E]) FlatMap(mapper func(v E) Stream[E]) Stream[E] {
	if b.Meta.MaxSize() == 0 {
		return b
	}
	return newOpFlatMap(b.Meta.Copy(), b.Curr, mapper)
}

func flatMapToAny[S any, T any](up Stream[S], mapper func(v S) Stream[T]) (down Stream[T]) {
	meta := reflectBaseMeta(up)
	if meta.MaxSize() == 0 {
		return newEmptyHeader[T]()
	}
	return newOpFlatMapToAny(meta.Copy(), reflectBaseCurr(up), mapper)
}

func (b *base[E]) Count() uint64 {
	var cnt uint64
	if b.Meta.MaxSize() == 0 {
		return 0
	}
	var sizeKnown bool
	newOpForCond(b.Curr,
		func(_ E) bool {
			cnt++
			return sizeKnown
		},
		func(size uint64, known bool) {
			if known {
				sizeKnown = true
				cnt = size
			}
		}, nil).Terminate()
	if sizeKnown {
		cnt--
	}
	return cnt
}

func (b *base[E]) Collect() []E {
	if b.Meta.MaxSize() == 0 {
		return nil
	}
	var ret []E
	newOpForeach(b.Curr,
		func(v E) { ret = append(ret, v) },
		func(size uint64, known bool) {
			if known {
				ret = make([]E, 0, size)
			}
		}, nil).Terminate()
	return ret
}

func collectToAny[C any, E any, R any](up Stream[E],
	supplier func(size uint64, known bool) C,
	accumulator func(b C, a E) C,
	finisher func(b C) R) R {
	meta := reflectBaseMeta(up)
	if meta.MaxSize() == 0 {
		return finisher(supplier(0, true))
	}
	var container C
	newOpForeach(reflectBaseCurr(up),
		func(v E) {
			container = accumulator(container, v)
		},
		func(size uint64, known bool) {
			container = supplier(size, known)
		}, nil).Terminate()
	return finisher(container)
}

func (b *base[E]) Reduce(id E, accum func(b, a E) E) E {
	var bs = id
	if b.Meta.MaxSize() == 0 {
		return bs
	}
	newOpForeach(b.Curr, func(v E) { bs = accum(bs, v) }, nil, nil).Terminate()
	return bs
}

func (b *base[E]) Foreach(act func(v E)) {
	if b.Meta.MaxSize() == 0 {
		return
	}
	newOpForeach(b.Curr, act, nil, nil).Terminate()
}

func (b *base[E]) ForCond(cond func(v E) bool) {
	if b.Meta.MaxSize() == 0 {
		return
	}
	newOpForCond(b.Curr, cond, nil, nil).Terminate()
}

func (b *base[E]) Iterator() iterator.Iterator[E] {
	if b.Meta.MaxSize() == 0 {
		return iterator.EmptyIterator[E]{}
	}
	return newOpIterator[E](b.Meta, b.Curr).Build()
}
