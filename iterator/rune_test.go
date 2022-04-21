package iterator

import (
	"fmt"
	"testing"
)

func TestStringRuneIter(t *testing.T) {
	str := "我爱日本！\nI love Japan! \n日本が大好きです！\n"
	iter := StringRuneIterator(str)
	for iter.MoveNext() {
		fmt.Print(string(iter.Current()))
	}
}
