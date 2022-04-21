package iterator

import (
	"testing"
)

func FuzzBytesIterator(f *testing.F) {
	f.Add([]byte{'s', 't', 'r', 'e', 'a', 'm'})
	f.Fuzz(func(t *testing.T, bytes []byte) {
		iter := BytesIterator[byte](bytes)
		defer iter.Close()
		for iter.MoveNext() {
			t.Logf("%s\n", string(iter.Current()))
		}
	})
}

func FuzzStringIterator(f *testing.F) {
	f.Add("stream")
	f.Fuzz(func(t *testing.T, str string) {
		iter := StringIterator[byte](str)
		defer iter.Close()
		for iter.MoveNext() {
			t.Logf("%s\n", string(iter.Current()))
		}
	})
}
