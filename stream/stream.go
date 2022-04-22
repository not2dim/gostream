package stream

import (
	"bytes"
	"github.com/not2dim/gostream/iterator"
)

type Stream[E any] interface {
	// Skip skips the first n elements.
	Skip(n uint64) Stream[E]
	// Limit limits the Stream size <= n.
	Limit(n uint64) Stream[E]
	// Filter filters elements satisfying pred(v) == true.
	Filter(pred func(v E) bool) Stream[E]
	// Peek applies the provided func act to every element.
	Peek(act func(v E)) Stream[E]
	// Cond applies the provided func cond to each iterated element until cond(v) returns true.
	// Note that the last v letting cond(v) true does not flow into the next Stream.
	Cond(cond func(v E) bool) Stream[E]
	// Distinct filters duplicate elements according to builtin ==.
	Distinct() Stream[E]
	// DistinctBy filters duplicate elements according to the provided func id.
	DistinctBy(id func(v E) any) Stream[E]
	// MinBy returns the minimum elements of all according to the provided func cmp.
	MinBy(cmp func(u, v E) int) Nullable[E]
	// MaxBy returns the maximum elements of all according to the provided func cmp.
	MaxBy(cmp func(u, v E) int) Nullable[E]
	// First returns the first element of the Stream.
	First() Nullable[E]
	// Last returns the last element of the Stream.
	Last() Nullable[E]
	// SortBy sorts elements in the Stream according to the provided func cmp.
	SortBy(cmp func(u, v E) int) Stream[E]
	// Map applies the given func mapper to every element.
	Map(mapper func(v E) E) Stream[E]
	// FlatMap applies the given func mapper to every element.
	FlatMap(mapper func(v E) Stream[E]) Stream[E]
	// Count returns the count of elements in the Stream.
	Count() uint64
	// Collect collects all elements into a slice.
	Collect() []E
	// Reduce performs a reduction of the Stream, using the provided identity value and an associative function,
	// and returns the reduced value.
	Reduce(id E, accum func(b, a E) E) E
	// Foreach applies the provided func act to every element.
	Foreach(act func(v E))
	// ForCond applies the provided func cond to each iterated element until cond(v) returns true.
	ForCond(cond func(v E) bool)
	// Iterator returns an iterator of the Stream.
	Iterator() iterator.Iterator[E]
}

// Nullable denotes a non-existing Val when OK = false.
type Nullable[E any] struct {
	Val E
	OK  bool
}

// Map applies the provided func mapper func(S) T to every element of input Stream[S], and returns a new Stream[T].
func Map[S any, T any](up Stream[S], mapper func(v S) T) (down Stream[T]) {
	return mapToAny(up, mapper)
}

// FlatMap applier the provided mapper func(v S) Stream[T] to every element, and returns a new Stream[T] concatenated
// by all Stream[T]s returned by mapper.
func FlatMap[S any, T any](up Stream[S], mapper func(v S) Stream[T]) (down Stream[T]) {
	return flatMapToAny(up, mapper)
}

// Of returns a new Stream[E] of providing arguments.
func Of[E any](elems ...E) Stream[E] {
	return Slice[[]E, E](elems)
}

// Slice returns a new Stream[E] from the given slice []E.
func Slice[T ~[]E, E any](slc T) Stream[E] {
	return newHeader[E](
		defaultMeta.Copy().SetMaxSize(uint64(len(slc))),
		iterator.SliceIterable[E](slc),
	)
}

// MapKeys returns a new Stream[K] consisting of map[K]V's keys.
func MapKeys[T ~map[K]V, K comparable, V any](m T) Stream[K] {
	return newHeader[K](
		defaultMeta.Copy().SetDistinct(true).SetMaxSize(uint64(len(m))),
		iterator.MapKeyIterable[K, V](m),
	)
}

// MapVals returns a new Stream[V] consisting of map[K]V's vals.
func MapVals[T ~map[K]V, K comparable, V any](m T) Stream[V] {
	return newHeader[V](
		defaultMeta.Copy().SetMaxSize(uint64(len(m))),
		iterator.MapValIterable[K, V](m),
	)
}

// Iterable returns a new Stream[E], whose elements all come from the provided Iterable[E].Iterator().
func Iterable[E any](iterable iterator.Iterable[E]) Stream[E] {
	return newHeader[E](
		defaultMeta.Copy(),
		iterable,
	)
}

// Range returns a new Stream[E], whose elements are all integer or unsigned integer within [from, to).
func Range[E integer | uinteger](from, to E) Stream[E] {
	return Iterable[E](newRangeIterable(from, to))
}

// Concat concatenates multiple Stream[E] into a new Stream[E].
func Concat[E any](ss ...Stream[E]) Stream[E] {
	return FlatMap[int, E](Range(0, len(ss)), func(idx int) Stream[E] { return ss[idx] })
}

// Collect collects all elements of the input Stream[E] and returns a container R.
// In the function header, C is the type of intermediate container, E is the type of Stream element,
// and R is the type of final container returned by Collect.
// The param supplier is the builder function of container C; the param accumulator accumulates elements E to C;
// the param finisher converts the intermediate container C to the final container R.
func Collect[C any, E any, R any](s Stream[E],
	supplier func(size uint64, known bool) C,
	accumulator func(b C, a E) C,
	finisher func(b C) R) R {
	return collectToAny(s, supplier, accumulator, finisher)
}

// Group classifies every element of the Stream[E] by the func identify(v E) K, and returns a container R.
// Group first collects all elements into a map[K]VC, and finally converts the intermediate map into R.
func Group[E any, R any, K comparable, VC any, VE any](s Stream[E],
	identity func(v E) K,
	supplier func() VC,
	mapper func(v E) VE,
	accumulator func(b VC, a VE) VC,
	finisher func(b map[K]VC) R) R {
	return Collect(s,
		func(size uint64, known bool) map[K]VC {
			if !known {
				size = 0
			}
			return make(map[K]VC, size)
		},
		func(b map[K]VC, a E) map[K]VC {
			k := identity(a)
			vc, ok := b[k]
			if !ok {
				vc = supplier()
				b[k] = vc
			}
			b[k] = accumulator(vc, mapper(a))
			return b
		},
		func(b map[K]VC) R {
			return finisher(b)
		})
}

// Identity is a generic identity function.
func Identity[E any](v E) E {
	return v
}

// ToSlice collects elements of Stream[E] into a slice []E.
func ToSlice[E any](s Stream[E]) []E {
	return Collect(s,
		func(size uint64, known bool) []E {
			if !known {
				size = 0
			}
			slc := make([]E, 0, size)
			return slc
		},
		func(b []E, a E) []E {
			b = append(b, a)
			return b
		},
		Identity[[]E],
	)
}

// ToSet collects elements of Stream[E] into a map[E]struct{}.
func ToSet[E comparable](s Stream[E]) map[E]struct{} {
	return Collect(s,
		func(size uint64, known bool) map[E]struct{} {
			if !known {
				size = 0
			}
			return make(map[E]struct{}, size)
		},
		func(b map[E]struct{}, a E) map[E]struct{} {
			b[a] = struct{}{}
			return b
		},
		Identity[map[E]struct{}],
	)
}

// ToMap groups every element by the func identify and returns a map[K][]E,
// collecting all elements E with same identity K into one slice []E.
func ToMap[E any, K comparable](s Stream[E], identity func(v E) K) map[K][]E {
	return Group[E, map[K][]E, K, []E, E](s,
		identity,
		func() []E {
			return nil
		},
		Identity[E],
		func(b []E, a E) []E {
			b = append(b, a)
			return b
		},
		Identity[map[K][]E],
	)
}

// Joining concatenates every string-like element in Stream[E] into a string.
func Joining[E string | rune | byte](s Stream[E], delimiter string, bufSize uint64) string {
	var idx int
	return Collect(s,
		func(size uint64, known bool) *bytes.Buffer {
			if known && size > bufSize {
				bufSize = size
			}
			return bytes.NewBuffer(make([]byte, 0, bufSize))
		},
		func(b *bytes.Buffer, a E) *bytes.Buffer {
			if idx != 0 {
				b.WriteString(delimiter)
			}
			idx++
			b.WriteString(string(a))
			return b
		},
		func(b *bytes.Buffer) string {
			return b.String()
		},
	)
}

// Counting returns a map[E]uint64 counting occurrence of each element.
func Counting[E comparable](s Stream[E]) map[E]uint64 {
	return Group[E, map[E]uint64, E, uint64, uint64](s,
		Identity[E],
		func() uint64 {
			return 0
		},
		func(v E) uint64 {
			return 1
		},
		func(b uint64, a uint64) uint64 {
			return b + a
		},
		Identity[map[E]uint64],
	)
}
