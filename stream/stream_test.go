package stream

import (
	"sort"
	"strconv"
	"testing"
)

func TestStreamSkip(t *testing.T) {
	var slc = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	var size = uint64(len(slc))
	stm := Slice(slc)
	for _, step := range [...]uint64{1, 2, 3, 4, 5} {
		stm = stm.Skip(step)
		if size >= step {
			size -= step
		} else {
			size = 0
		}
		if stm.Count() != Max(size, 0) {
			t.Fail()
		}
	}
}

func TestStreamLimit(t *testing.T) {
	var slc = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	stm := Slice(slc)
	for _, lmt := range [...]uint64{5, 4, 3, 2, 1, 0} {
		stm = stm.Limit(lmt)
		if stm.Count() > lmt {
			t.Fail()
		}
	}
}

func TestStreamFilter(t *testing.T) {
	var m = map[int]struct{}{}
	for i := 0; i < 10; i++ {
		m[i] = struct{}{}
	}
	var size = uint64(len(m))
	stm := MapKeys(m)
	for _, tgt := range [...]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9} {
		if _, ok := m[tgt]; ok {
			size--
		}
		var j = tgt
		stm = stm.Filter(func(i int) bool { return i != j })
		if stm.Count() != size {
			t.Fail()
		}
	}
}

func TestStreamPeek(t *testing.T) {
	stm := Of(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
	var cnt = stm.Count()
	var accum uint64 = 0
	stm.Peek(func(i int) { accum++ }).Foreach(func(i int) {})
	if cnt != accum {
		t.Fail()
	}
}

func TestStreamDistinct(t *testing.T) {
	stm := Of(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
	if stm.Distinct().Count() != stm.Count() {
		t.Fail()
	}
	if stm.DistinctBy(func(i int) any {
		return i % 5
	}).Count() != 5 {
		t.Fail()
	}
}

func FuzzStreamMinAndMax(f *testing.F) {
	for _, tc := range [][]byte{{}, {1}, {1, 2, 3}} {
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, slc []byte) {
		var min, max Nullable[byte]
		for _, b := range slc {
			if !min.OK {
				min.OK = true
				min.Val = b
			}
			if !max.OK {
				max.OK = true
				max.Val = b
			}
			if min.Val > b {
				min.Val = b
			}
			if max.Val < b {
				max.Val = b
			}
		}
		sMin := Slice(slc).MinBy(CmpRealNum[byte])
		sMax := Slice(slc).MaxBy(CmpRealNum[byte])
		if min != sMin {
			t.Fatalf("expected: %v, actual: %v\n", min, sMin)
		}
		if max != sMax {
			t.Fatalf("expected: %v, actual: %v\n", max, sMax)
		}
	})
}

func FuzzStreamSort(f *testing.F) {
	for _, tc := range [][]byte{{}, {1}, {3, 2, 1, 0}, {1, 2, 3, 4, 5, 6}, {8, 4, 1, 3, 6, 9, 7, 5, 2, 0}} {
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, slc []byte) {
		var sorted []byte
		Slice(slc).SortBy(CmpRealNum[byte]).Foreach(func(i byte) { sorted = append(sorted, i) })
		sort.Slice(slc, func(i, j int) bool {
			return slc[i] < slc[j]
		})
		if len(sorted) != len(slc) {
			t.Fatalf("expected: %v, actual: %v\n", slc, sorted)
		}
		for i := 0; i < len(slc); i++ {
			if sorted[i] != slc[i] {
				t.Fatalf("expected: %v, actual: %v\n", slc, sorted)
			}
		}
	})
}

func FuzzStreamCount(f *testing.F) {
	for _, tc := range [][]byte{{}, {0}, {0, 1}} {
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, slc []byte) {
		cnt := Slice(slc).Count()

		if uint64(len(slc)) != cnt {
			t.Fatalf("expected: %v, actual: %v\n", len(slc), cnt)
		}
	})
}

func TestStreamMap(t *testing.T) {
	if Of[int]().Map(func(i int) int { return i }).Count() != 0 {
		t.Fail()
	}
}

func TestMap(t *testing.T) {
	for _, tc := range [][]byte{{}, {0}, {0, 1, 2, 3, 4, 5, 6, 7, 8, 9}} {
		var mapper = func(v byte) string { return string(v) }
		var mapped []string
		Map(Slice(tc), mapper).Foreach(func(str string) {
			mapped = append(mapped, str)
		})
		for i := 0; i < len(tc); i++ {
			if mapped[i] != mapper(tc[i]) {
				t.Fatalf("expected: %v, actual: %v\n", mapper(tc[i]), mapped[i])
			}
		}
	}
}

func TestMapVals(t *testing.T) {
	if MapVals(map[int]int{0: 0, 1: 1, 2: 2, 3: 3}).Reduce(0, func(b, a int) int { return b + a }) != 6 {
		t.Fail()
	}
}

func TestStreamFlatMap(t *testing.T) {
	var idx = 0
	Slice([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}).FlatMap(func(cap int) Stream[int] {
		var slc []int
		for i := 0; i < 10; i++ {
			slc = append(slc, cap*10+i)
		}
		return Slice(slc)
	}).SortBy(CmpRealNum[int]).Foreach(func(n int) {
		if idx != n {
			t.Fail()
		}
		idx++
	})
}

func TestFlatMap(t *testing.T) {
	var idx = 0
	FlatMap(Slice([]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}),
		func(cap string) Stream[int] {
			var slc []int
			for i := 0; i < 10; i++ {
				var n, _ = strconv.Atoi(cap + strconv.Itoa(i))
				slc = append(slc, n)
			}
			return Slice(slc)
		},
	).SortBy(CmpRealNum[int]).Foreach(func(n int) {
		if idx != n {
			t.Fail()
		}
		idx++
	})
}

func TestStreamIterator_CaseSourceIterator(t *testing.T) {
	var idx int
	slc := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	iter := Slice(slc).Iterator()
	defer iter.Close()
	for iter.MoveNext() {
		if slc[idx] != iter.Current() {
			t.Fatalf("idx: %v, expected: %v, actual: %v\n", idx, slc[idx], iter.Current())
		}
		idx++
	}
	if idx != len(slc) {
		t.Fail()
	}
}

func TestStreamIterator_CaseSinkIterator(t *testing.T) {
	var idx int
	slc := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	mapper := func(i int) int { return i * i }
	iter := Slice(slc).Map(mapper).Iterator()
	defer iter.Close()
	for iter.MoveNext() {
		if mapper(slc[idx]) != iter.Current() {
			t.Fatalf("idx: %v, expected: %v, actual: %v\n", idx, slc[idx], iter.Current())
		}
		idx++
	}
	if idx != len(slc) {
		t.Fail()
	}
}

func TestStreamIterator_CaseChanneledIterator(t *testing.T) {
	var idx int
	slc := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	mapper := func(cap int) Stream[int] {
		var slc []int
		for i := 0; i < 10; i++ {
			slc = append(slc, cap*10+i)
		}
		return Slice(slc)
	}
	iter := Slice(slc).FlatMap(mapper).Limit(50).Iterator()
	defer iter.Close()
	for iter.MoveNext() {
		if idx != iter.Current() {
			t.Fatalf("idx: %v, expected: %v, actual: %v\n", idx, idx, iter.Current())
		}
		idx++
	}
}

func TestRange(t *testing.T) {
	var idx = 0
	from, to := 0, 10
	iter := Range(from, to).Iterator()
	defer iter.Close()
	for iter.MoveNext() {
		if idx != iter.Current() {
			t.Fatalf("idx: %v, expected: %v, actual: %v\n", idx, idx, iter.Current())
		}
		idx++
	}
	if to-from != idx {
		t.Fail()
	}
}

func TestConcat(t *testing.T) {
	slc := Concat(
		Of(0, 1, 2, 3, 4),
		Of(5, 6, 7, 8, 9),
	).Collect()
	for idx, v := range slc {
		if idx != v {
			t.Fail()
		}
	}
}

func TestStreamCollectAndToSlice(t *testing.T) {
	stm := Of(0, 1, 2, 3, 4, 5, 6, 7, 8, 9)
	slc0 := stm.Collect()
	slc1 := ToSlice(stm)
	if len(slc0) != len(slc1) {
		t.Fail()
	}
	for i := 0; i < len(slc0); i++ {
		if slc0[i] != slc1[i] {
			t.Fail()
		}
	}
}

func TestStreamReduceAndJoining(t *testing.T) {
	var words = []string{
		"我", "爱", "日", "本", "！", "\n",
		"I", " ", "love", " ", "Japan", "!", "\n",
		"日", "本", "が", "大", "好", "き", "で", "す", "！", "\n",
	}
	stm := Slice(words)
	para0 := stm.Reduce("", func(b, a string) string {
		return b + a
	})
	para1 := Joining(stm, "", 64)
	if para0 != para1 {
		t.Fail()
	}
}

func TestCounting(t *testing.T) {
	stm := Of(0, 0, 1, 1, 2, 2, 3, 3, 3, 4, 4, 4, 5, 5)
	t.Logf("%v\n", Counting(stm))
}

func TestToSet(t *testing.T) {
	stm := Concat(
		Range(-10, 4),
		Range(1, 10),
		Range(11, 20),
		Range(20, 30),
	)
	if len(ToSet(stm)) != 39 {
		t.Fail()
	}
}

func TestToMap(t *testing.T) {
	stm := Range(0, 100)
	m := ToMap(stm, func(i int) int {
		return i % 5
	})
	for k, slc := range m {
		t.Logf("key: %v, val: %v\n", k, slc)
	}
}
