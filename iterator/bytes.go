package iterator

import (
	"reflect"
	"unsafe"
)

type bytesIterable[E comparable] struct {
	data   uintptr
	len    int
	elemSz int
}

type bytesIterator[E comparable] struct {
	EmptyIterator[E]
	data   uintptr
	len    int
	elemSz int
	offset int
	curr   E
}

func BytesIterable[E comparable, BS ~[]byte](bytes BS) Iterable[E] {
	if len(bytes) == 0 {
		return EmptyIterable[E]{}
	}
	slcH := *(*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	var e E
	elemSz := int(unsafe.Sizeof(e))
	return &bytesIterable[E]{
		data:   slcH.Data,
		len:    slcH.Len,
		elemSz: elemSz,
	}
}

func StringIterable[E comparable](str string) Iterable[E] {
	if len(str) == 0 {
		return EmptyIterable[E]{}
	}
	strH := *(*reflect.SliceHeader)(unsafe.Pointer(&str))
	var e E
	elemSz := int(unsafe.Sizeof(e))
	return &bytesIterable[E]{
		data:   strH.Data,
		len:    strH.Len,
		elemSz: elemSz,
	}
}

func (b bytesIterable[E]) Iterator() Iterator[E] {
	return &bytesIterator[E]{
		data:   b.data,
		len:    b.len,
		elemSz: b.elemSz,
		offset: 0,
	}
}

func (b bytesIterable[E]) Size() (n uint64, known bool) {
	return uint64(b.len / b.elemSz), true
}

// BytesIterator returns an Iterator[E] that iterates sizeof(E) elements sequentially from the given []byte.
// Note that, BytesIterator[rune] does not iterate elements decoded in utf-8 rune.
// For that case, you should use BytesRuneIterator.
func BytesIterator[E comparable, BS ~[]byte](bytes BS) Iterator[E] {
	if len(bytes) == 0 {
		return EmptyIterator[E]{}
	}
	slcH := *(*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	var e E
	elemSz := int(unsafe.Sizeof(e))
	return &bytesIterator[E]{
		data:   slcH.Data,
		len:    slcH.Len,
		elemSz: elemSz,
		offset: 0,
	}
}

// StringIterator returns an Iterator[E] that iterates sizeof(E) elements sequentially from the given string.
// Note that, StringIterator[rune] does not iterate elements decoded in utf-8 rune.
// For that case, you should use StringRuneIterator.
func StringIterator[E comparable](str string) Iterator[E] {
	strH := *(*reflect.SliceHeader)(unsafe.Pointer(&str))
	var e E
	elemSz := int(unsafe.Sizeof(e))
	return &bytesIterator[E]{
		data:   strH.Data,
		len:    strH.Len,
		elemSz: elemSz,
		offset: 0,
	}
}

func (s *bytesIterator[E]) MoveNext() bool {
	if s.offset >= s.len {
		return false
	}
	//goland:noinspection GoVetUnsafePointer
	s.curr = *(*E)(unsafe.Pointer(s.data + uintptr(s.offset)))
	s.offset += s.elemSz
	return true
}

func (s *bytesIterator[E]) Current() E {
	return s.curr
}
