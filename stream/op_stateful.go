package stream

import (
	"sort"
)

// region Skip

type opSkip[E any] struct {
	*base[E]
	n uint64
}

func newOpSkip[E any](meta *meta, upstream pipeline, n uint64) (ret *opSkip[E]) {
	ret = &opSkip[E]{n: n}
	ret.base = &base[E]{Meta: meta.DecrSize(n), Prev: upstream, Curr: ret}
	return
}

type skipSink struct {
	baseSink
	n uint64
	i uint64
}

func (s *skipSink) Begin(size uint64, known bool) {
	if known {
		s.down.Begin(Max(size-s.n, 0), true)
	} else {
		s.down.Begin(0, false)
	}
}

func (s *skipSink) Accept(v any) {
	if s.i < s.n {
		s.i++
		return
	}
	s.down.Accept(v)
}

func (f *opSkip[E]) WrapSink(down sink) sink {
	return &skipSink{baseSink: baseSink{down: down}, n: f.n, i: 0}
}

// endregion

// region Limit

type opLimit[E any] struct {
	base[E]
	n uint64
}

func newOpLimit[E any](meta *meta, upstream pipeline, n uint64) (ret *opLimit[E]) {
	ret = &opLimit[E]{n: n}
	ret.base = base[E]{Meta: meta.LimitSize(n), Prev: upstream, Curr: ret}
	return
}

type limitSink struct {
	baseSink
	n uint64
	i uint64
}

func (s *limitSink) Begin(size uint64, known bool) {
	if known {
		s.down.Begin(Min(s.n, size), true)
	} else {
		s.down.Begin(size, false)
	}
}

func (s *limitSink) Accept(v any) {
	if s.i < s.n {
		s.down.Accept(v)
		s.i++
	}
}

func (s *limitSink) Rejecting() bool {
	return s.i >= s.n
}

func (f *opLimit[E]) WrapSink(down sink) sink {
	return &limitSink{baseSink: baseSink{down: down}, n: f.n, i: 0}
}

// endregion

// region Cond

type opCond[E any] struct {
	base[E]
	cond func(v E) bool
}

func newOpCond[E any](meta *meta, upstream pipeline, cond func(v E) bool) (ret *opCond[E]) {
	ret = &opCond[E]{cond: cond}
	ret.base = base[E]{
		Meta: meta,
		Prev: upstream, Curr: ret,
	}
	return
}

type condSink[E any] struct {
	baseSink
	rejecting bool
	cond      func(v E) bool
}

func (c *condSink[E]) Accept(v any) {
	c.rejecting = c.rejecting || c.cond(v.(E))
	if !c.rejecting {
		c.down.Accept(v)
	}
}

func (c *condSink[E]) Rejecting() bool {
	return c.rejecting
}

func (f *opCond[E]) WrapSink(down sink) sink {
	return &condSink[E]{baseSink: baseSink{down: down}, cond: f.cond}
}

// endregion

// region Distinct

type opDistinct[E any] struct {
	base[E]
}

func newOpDistinct[E any](meta *meta, upstream pipeline) (ret *opDistinct[E]) {
	ret = &opDistinct[E]{}
	ret.base = base[E]{
		Meta: meta.SetDistinct(true).SetSinkIterable(false),
		Prev: upstream, Curr: ret,
	}
	return
}

type distinctSink struct {
	baseSink
	m map[any]struct{}
}

func (s *distinctSink) Begin(size uint64, _ bool) {
	s.m = make(map[any]struct{}, size)
}

func (s *distinctSink) Accept(v any) {
	s.m[v] = struct{}{}
}

func (s *distinctSink) Close() {
	s.down.Begin(uint64(len(s.m)), true)
	for v := range s.m {
		if s.down.Rejecting() {
			break
		}
		s.down.Accept(v)
	}
	s.down.Close()
}

func (f *opDistinct[E]) WrapSink(down sink) sink {
	return &distinctSink{baseSink: baseSink{down: down}}
}

// endregion

// region DistinctBy

type opDistinctBy[E any] struct {
	base[E]
	id func(E) any
}

func newOpDistinctBy[E any](meta *meta, upstream pipeline, id func(E) any) (ret *opDistinctBy[E]) {
	ret = &opDistinctBy[E]{id: id}
	ret.base = base[E]{Meta: meta.SetSinkIterable(false), Prev: upstream, Curr: ret}
	return
}

type distinctBySink[E any] struct {
	baseSink
	m  map[any]any
	id func(v E) any
}

func (s *distinctBySink[E]) Begin(size uint64, _ bool) {
	s.m = make(map[any]any, size)
}

func (s *distinctBySink[E]) Accept(v any) {
	s.m[s.id(v.(E))] = v
}

func (s *distinctBySink[E]) Close() {
	s.down.Begin(uint64(len(s.m)), true)
	for _, v := range s.m {
		if s.down.Rejecting() {
			break
		}
		s.down.Accept(v)
	}
	s.down.Close()
}

func (f *opDistinctBy[E]) WrapSink(down sink) sink {
	return &distinctBySink[E]{baseSink: baseSink{down: down}, id: f.id}
}

// endregion

// region SortBy

type opSortBy[E any] struct {
	base[E]
	cmp func(u E, v E) int
}

func newOpSortBy[E any](meta *meta, upstream pipeline, cmp func(u E, v E) int) (ret *opSortBy[E]) {
	ret = &opSortBy[E]{cmp: cmp}
	ret.base = base[E]{Meta: meta.SetSinkIterable(false), Prev: upstream, Curr: ret}
	return
}

type sortBySink[E any] struct {
	baseSink
	slc []E
	cmp func(u E, v E) int
}

func (s *sortBySink[E]) Begin(size uint64, _ bool) {
	s.slc = make([]E, 0, size)
}

func (s *sortBySink[E]) Accept(v any) {
	s.slc = append(s.slc, v.(E))
}

func (s *sortBySink[E]) Close() {
	s.down.Begin(uint64(len(s.slc)), true)
	sort.Slice(s.slc, func(i, j int) bool {
		return s.cmp(s.slc[i], s.slc[j]) < 0
	})
	for _, v := range s.slc {
		if s.down.Rejecting() {
			break
		}
		s.down.Accept(v)
	}
	s.down.Close()
}

func (f *opSortBy[E]) WrapSink(down sink) sink {
	return &sortBySink[E]{baseSink: baseSink{down: down}, cmp: f.cmp}
}

// endregion
