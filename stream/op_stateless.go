package stream

import (
	"math"
	"reflect"
)

// region Filter

type opFilter[E any] struct {
	base[E]
	pred func(v E) bool
}

func newOpFilter[E any](meta *meta, upstream pipeline, pred func(v E) bool) (ret *opFilter[E]) {
	ret = &opFilter[E]{pred: pred}
	ret.base = base[E]{Meta: meta, Prev: upstream, Curr: ret}
	return
}

type filterSink[E any] struct {
	baseSink
	pred func(v E) bool
}

func (b filterSink[E]) Accept(v any) {
	if b.pred(v.(E)) {
		b.down.Accept(v)
	}
}

func (f *opFilter[E]) WrapSink(down sink) sink {
	return filterSink[E]{baseSink{down: down}, f.pred}
}

// endregion

// region Peek

type opPeek[E any] struct {
	base[E]
	act func(v E)
}

func newOpPeek[E any](meta *meta, upstream pipeline, act func(v E)) (ret *opPeek[E]) {
	ret = &opPeek[E]{act: act}
	ret.base = base[E]{Meta: meta, Prev: upstream, Curr: ret}
	return
}

type peekSink[E any] struct {
	baseSink
	act func(v E)
}

func (b peekSink[E]) Accept(v any) {
	b.act(v.(E))
	b.down.Accept(v)
}

func (f *opPeek[E]) WrapSink(down sink) sink {
	return peekSink[E]{baseSink{down: down}, f.act}
}

// endregion

// region Map

type opMap[E any] struct {
	base[E]
	mapper func(v E) E
}

func newOpMap[E any](meta *meta, upstream pipeline, mapper func(v E) E) (ret *opMap[E]) {
	ret = &opMap[E]{mapper: mapper}
	ret.base = base[E]{Meta: meta.SetDistinct(false), Prev: upstream, Curr: ret}
	return
}

type mapSink[E any] struct {
	baseSink
	mapper func(v E) E
}

func (b mapSink[E]) Accept(v any) {
	b.down.Accept(b.mapper(v.(E)))
}

func (f *opMap[E]) WrapSink(down sink) sink {
	return mapSink[E]{baseSink{down: down}, f.mapper}
}

// endregion

// region MapToAny

type opMapToAny[T any] struct {
	base[T]
	mapper reflect.Value // func mapper[S any, T any](v S) T
}

func newOpMapToAny[S any, T any](meta *meta, upstream pipeline, mapper func(v S) T) (ret *opMapToAny[T]) {
	ret = &opMapToAny[T]{mapper: reflectMapper(mapper)}
	ret.base = base[T]{Meta: meta.SetDistinct(false), Prev: upstream, Curr: ret}
	return
}

type mapToAnySink[T any] struct {
	baseSink
	mapper reflect.Value
}

func (b mapToAnySink[T]) Accept(v any) {
	b.down.Accept(reflectCallMapper(b.mapper, v))
}

func (f *opMapToAny[T]) WrapSink(down sink) sink {
	return mapToAnySink[T]{baseSink{down: down}, f.mapper}
}

// endregion

// region FlatMap

type opFlatMap[E any] struct {
	base[E]
	mapper func(v E) Stream[E]
}

func newOpFlatMap[E any](meta *meta, upstream pipeline, mapper func(v E) Stream[E]) (ret *opFlatMap[E]) {
	ret = &opFlatMap[E]{mapper: mapper}
	ret.base = base[E]{
		Meta: meta.SetDistinct(false).SetSinkIterable(false).SetMaxSize(math.MaxUint64),
		Prev: upstream, Curr: ret,
	}
	return
}

type flatMapSink[E any] struct {
	baseSink
	mapper func(v E) Stream[E]
}

func (b flatMapSink[E]) Accept(v any) {
	iter := b.mapper(v.(E)).Iterator()
	defer iter.Close()
	for iter.MoveNext() && !b.down.Rejecting() {
		b.down.Accept(iter.Current())
	}
}

func (f *opFlatMap[E]) WrapSink(down sink) sink {
	return flatMapSink[E]{baseSink{down: down}, f.mapper}
}

// endregion

// region FlatMapToAny

type opFlatMapToAny[T any] struct {
	base[T]
	mapper reflect.Value // func mapper[S any, T any](v S) T
}

func newOpFlatMapToAny[S any, T any](meta *meta, upstream pipeline, mapper func(v S) Stream[T]) (ret *opFlatMapToAny[T]) {
	ret = &opFlatMapToAny[T]{mapper: reflectMapper(mapper)}
	ret.base = base[T]{
		Meta: meta.SetDistinct(false).SetSinkIterable(false).SetMaxSize(math.MaxUint64),
		Prev: upstream, Curr: ret,
	}
	return
}

type flatMapToAnySink[T any] struct {
	baseSink
	mapper reflect.Value
}

func (b flatMapToAnySink[T]) Accept(v any) {
	iter := reflectCallMapper(b.mapper, v).(Stream[T]).Iterator()
	defer iter.Close()
	for iter.MoveNext() && !b.down.Rejecting() {
		b.down.Accept(iter.Current())
	}
}

func (f *opFlatMapToAny[T]) WrapSink(down sink) sink {
	return flatMapToAnySink[T]{baseSink{down: down}, f.mapper}
}

// endregion
