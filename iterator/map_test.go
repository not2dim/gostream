package iterator

import (
	"fmt"
	"testing"
)

func TestMapKeyIter(t *testing.T) {
	iter := MapKeyIterator(map[string]int{"apple": 1, "banana": 2, "cherry": 3})
	for iter.MoveNext() {
		fmt.Println(iter.Current())
	}
}

func TestMapValIter(t *testing.T) {
	iter := MapValIterator(map[string]int{"apple": 1, "banana": 2, "cherry": 3})
	for iter.MoveNext() {
		fmt.Println(iter.Current())
	}
}
