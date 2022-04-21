package iterator

import (
	"unicode/utf8"
	"unsafe"
)

type BytesRuneIterable []byte

func (b BytesRuneIterable) Iterator() Iterator[rune] {
	return BytesRuneIterator(b)
}

func (b BytesRuneIterable) Size() (n uint64, known bool) {
	return uint64(len(b)), false
}

func StringRuneIterable(str string) Iterable[rune] {
	return BytesRuneIterable(*(*[]byte)(unsafe.Pointer(&str)))
}

type runeIterator struct {
	EmptyIterator[rune]
	data []byte
	idx  int
	curr rune
}

func BytesRuneIterator(bytes []byte) Iterator[rune] {
	return &runeIterator{
		data: bytes,
		idx:  0,
		curr: utf8.RuneError,
	}
}

func StringRuneIterator(str string) Iterator[rune] {
	return &runeIterator{
		data: *(*[]byte)(unsafe.Pointer(&str)),
		idx:  0,
		curr: utf8.RuneError,
	}
}

func (r *runeIterator) MoveNext() bool {
	if r.idx >= len(r.data) {
		return false
	}
	rn, size := utf8.DecodeRune(r.data[r.idx:])
	r.curr = rn
	r.idx += size
	return true
}

func (r *runeIterator) Current() rune {
	return r.curr
}
