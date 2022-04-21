package stream

import (
	"testing"
)

func TestReflectBaseMeta(t *testing.T) {
	for _, bs := range []any{base[int]{}, opMapToAny[string]{base: base[string]{Meta: &meta{}}}} {
		t.Logf("%v\n", reflectBaseMeta(bs))
	}
}

func TestReflectBaseCurr(t *testing.T) {
	for _, bs := range []any{
		base[int]{Curr: nil},
		base[int]{Curr: &base[int]{}},
		opMapToAny[string]{base: base[string]{Meta: &meta{}, Curr: &base[int]{}}},
	} {
		t.Logf("%v\n", reflectBaseCurr(bs))
	}
}
