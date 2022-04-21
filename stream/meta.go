package stream

import (
	"math"
)

type meta struct {
	maxSize      uint64
	distinct     bool
	sinkIterable bool
}

var defaultMeta *meta = nil

func (m *meta) MaxSize() uint64 {
	if m == nil {
		return math.MaxUint64
	}
	return m.maxSize
}

func (m *meta) SetMaxSize(ms uint64) *meta {
	m.maxSize = ms
	return m
}

func (m *meta) DecrSize(n uint64) *meta {
	if m.maxSize >= n {
		m.maxSize -= n
	} else {
		m.maxSize = 0
	}
	return m
}

func (m *meta) LimitSize(n uint64) *meta {
	m.maxSize = Min(n, m.maxSize)
	return m
}

func (m *meta) Distinct() bool {
	if m == nil {
		return false
	}
	return m.distinct
}

func (m *meta) SetDistinct(distinct bool) *meta {
	m.distinct = distinct
	return m
}

func (m *meta) SinkIterable() bool {
	if m == nil {
		return true
	}
	return m.sinkIterable
}

func (m *meta) SetSinkIterable(able bool) *meta {
	m.sinkIterable = able
	return m
}

func (m *meta) Copy() *meta {
	if m == nil {
		return &meta{
			maxSize:      m.MaxSize(),
			distinct:     m.Distinct(),
			sinkIterable: m.SinkIterable(),
		}
	}
	var cp = *m
	return &cp
}
