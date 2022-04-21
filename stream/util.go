package stream

type integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type uinteger interface {
	~uintptr | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type float interface {
	~float32 | ~float64
}

type realNum interface {
	integer | uinteger | float
}

type complexNum interface {
	~complex64 | ~complex128
}

func CmpRealNum[N realNum](u, v N) int {
	if u < v {
		return -1
	} else if u > v {
		return 1
	}
	return 0
}

func Max[E realNum](a, b E) E {
	if a > b {
		return a
	}
	return b
}

func Min[E realNum](a, b E) E {
	if a < b {
		return a
	}
	return b
}
