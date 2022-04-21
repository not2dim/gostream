package iterator

import "reflect"

type MapKeyIterable[K comparable, V any] map[K]V

func (m MapKeyIterable[K, V]) Iterator() Iterator[K] {
	return MapKeyIterator(m)
}

func (m MapKeyIterable[K, V]) Size() (n uint64, known bool) {
	return uint64(len(m)), true
}

type MapValIterable[K comparable, V any] map[K]V

func (m MapValIterable[K, V]) Iterator() Iterator[V] {
	return MapValIterator(m)
}

func (m MapValIterable[K, V]) Size() (n uint64, known bool) {
	return uint64(len(m)), true
}

func MapKeyIterator[T ~map[K]V, K comparable, V any](m T) Iterator[K] {
	rm := reflect.ValueOf(m)
	return mapKeyIterator[K, V]{
		riter: rm.MapRange(),
	}
}

func MapValIterator[T ~map[K]V, K comparable, V any](m T) Iterator[V] {
	rm := reflect.ValueOf(m)
	return mapValIterator[K, V]{
		riter: rm.MapRange(),
	}
}

type mapKeyIterator[K comparable, V any] struct {
	EmptyIterator[K]
	riter *reflect.MapIter
}

func (m mapKeyIterator[K, V]) MoveNext() bool {
	return m.riter.Next()
}

func (m mapKeyIterator[K, V]) Current() K {
	return m.riter.Key().Interface().(K)
}

type mapValIterator[K comparable, V any] struct {
	EmptyIterator[V]
	riter *reflect.MapIter
}

func (m mapValIterator[K, V]) MoveNext() bool {
	return m.riter.Next()
}

func (m mapValIterator[K, V]) Current() V {
	return m.riter.Value().Interface().(V)
}
