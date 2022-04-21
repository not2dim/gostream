package stream

import "github.com/not2dim/gostream/iterator"

func process(terminal pipeline) (header pipeline, wrapped sink) {
	var pipelines []pipeline
	var curr = terminal
	for curr != nil {
		pipelines = append(pipelines, curr)
		curr = curr.GetUpstream()
	}
	wrapped = nil
	for i := 0; i < len(pipelines)-1; i++ {
		wrapped = pipelines[i].WrapSink(wrapped)
	}
	return pipelines[len(pipelines)-1], wrapped
}

func terminate(terminal pipeline) {
	header, wrapped := process(terminal)
	src := header.GetSource()
	size, known := src.Size()
	iter := src.Iterator()
	defer iter.Close()
	wrapped.Begin(size, known)
	for iter.MoveNext() && !wrapped.Rejecting() {
		wrapped.Accept(iter.Current())
	}
	wrapped.Close()
}

type termSink struct {
	begin func(uint64, bool)
	close func()
}

func (t termSink) Begin(size uint64, unknown bool) {
	if t.begin != nil {
		t.begin(size, unknown)
	}
}

func (t termSink) Accept(_ any) {
}

func (t termSink) Rejecting() bool {
	return false
}

func (t termSink) Close() {
	if t.close != nil {
		t.close()
	}
}

// region Foreach

type opForeach[E any] struct {
	base[E]
	act   func(v E)
	begin func(uint64, bool)
	close func()
}

func newOpForeach[E any](upstream pipeline, act func(v E), begin func(uint64, bool), close func()) (ret *opForeach[E]) {
	ret = &opForeach[E]{act: act, begin: begin, close: close}
	ret.base = base[E]{Prev: upstream, Curr: ret}
	return
}

type foreachSink[E any] struct {
	termSink
	act func(v E)
}

func (f *foreachSink[E]) Accept(v any) {
	f.act(v.(E))
}

func (o *opForeach[E]) WrapSink(_ sink) sink {
	return &foreachSink[E]{act: o.act, termSink: termSink{begin: o.begin, close: o.close}}
}

func (o *opForeach[E]) Terminate() {
	terminate(o)
}

// endregion

// region ForCond

type opForCond[E any] struct {
	base[E]
	cond  func(v E) bool // sink will reject all inputs, once cond returns true.
	begin func(uint64, bool)
	close func()
}

func newOpForCond[E any](upstream pipeline, cond func(v E) bool, begin func(uint64, bool), close func()) (ret *opForCond[E]) {
	ret = &opForCond[E]{cond: cond, close: close}
	ret.base = base[E]{Prev: upstream, Curr: ret}
	return
}

type forCondSink[E any] struct {
	termSink
	rejecting bool
	cond      func(v E) bool
	close     func()
}

func (c *forCondSink[E]) Accept(v any) {
	c.rejecting = c.rejecting || c.cond(v.(E))
}

func (c *forCondSink[E]) Rejecting() bool {
	return c.rejecting
}

func (o *opForCond[E]) WrapSink(_ sink) sink {
	return &forCondSink[E]{cond: o.cond, termSink: termSink{begin: o.begin, close: o.close}}
}

func (o *opForCond[E]) Terminate() {
	terminate(o)
}

// endregion

// region Iterator

type opIterator[E any] struct {
	base[E]
	current E
}

func newOpIterator[E any](meta *meta, upstream pipeline) (ret *opIterator[E]) {
	ret = &opIterator[E]{}
	ret.base = base[E]{Meta: meta, Prev: upstream, Curr: ret}
	return
}

type iterSink[E any] struct {
	termSink
	op *opIterator[E]
}

func (f iterSink[E]) Accept(v any) {
	f.op.current = v.(E)
}

func (f *opIterator[E]) WrapSink(_ sink) sink {
	return iterSink[E]{op: f}
}

type sinkIterator[E any] struct {
	op            *opIterator[E]
	begun, closed bool
	iter          iterator.Iterator[any]
	size          uint64
	known         bool
	wrapped       sink
}

func (f *sinkIterator[E]) MoveNext() bool {
	if !f.begun {
		f.begun = true
		f.wrapped.Begin(f.size, f.known)
	}
	if !f.iter.MoveNext() || f.wrapped.Rejecting() {
		f.closed = true
		f.iter.Close()
		f.wrapped.Close()
		return false
	}
	f.wrapped.Accept(f.iter.Current())
	return true
}

func (f *sinkIterator[E]) Current() E {
	return f.op.current
}

func (f *sinkIterator[E]) Close() {
	if f.begun && !f.closed {
		f.closed = true
		f.iter.Close()
		f.wrapped.Close()
	}
}

func (f *opIterator[E]) Build() iterator.Iterator[E] {
	if f.Prev.GetSource() != nil {
		return unwrapIterable[E](f.Prev.GetSource()).Iterator()
	}
	if f.Meta.SinkIterable() {
		header, wrapped := process(f)
		src := header.GetSource()
		size, known := src.Size()
		return &sinkIterator[E]{
			op:      f,
			iter:    src.Iterator(),
			size:    size,
			known:   known,
			wrapped: wrapped,
		}
	}
	ch := make(chan E, 32)
	stop := make(chan struct{})
	go func() {
		newOpForCond(f.GetUpstream(),
			func(v E) bool {
				select {
				case ch <- v:
					return false
				case <-stop:
					return true
				}
			}, nil, func() {
				close(ch)
			},
		).Terminate()
	}()
	return &channeledIterator[E]{
		stop: stop,
		ch:   ch,
	}
}

type channeledIterator[E any] struct {
	stop chan<- struct{}
	ch   <-chan E
	curr E
}

func (f *channeledIterator[E]) MoveNext() bool {
	tmp, ok := <-f.ch
	if !ok {
		return false
	}
	f.curr = tmp
	return true
}

func (f *channeledIterator[E]) Current() E {
	return f.curr
}

func (f *channeledIterator[E]) Close() {
	close(f.stop)
}

// endregion

// region Collect

// endregion
