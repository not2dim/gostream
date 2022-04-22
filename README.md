# Gostream

Gostream is a type-safe, Java-like Stream API library written in Go.

## Requirement

- Go 1.18 or higher.

## Installation

```shell
go get github.com/not2dim/gostream
```

## Interfaces

This library makes 3 basic abstractions: **Iterator**, **Iterable** and **Stream**.

### Iterator

It's suggested to call **Close** after iteration done. This library provides **Iterator** implementations for slice, map
key, map value, []byte and string.

```go
type Iterator[E any] interface {
    MoveNext() bool
    Current() E
    Close()
}
```

### Iterable

**Iterable** should be stateless and its method **Iterator** can be called multiple times. Any iterable data structures
can be adapted to **Iterable**. This library provides **Iterable** implementations for slice, map key, map value, []byte
and string.

```go
type Iterable[E any] interface {
    Iterator() Iterator[E]
    Size() (n uint64, known bool)
}
```

### Stream

**Stream** supports most of the operations in Java Stream. But due
to [some limitation of current generics](https://tip.golang.org/doc/go1.18#generics), this library implements some
operations including **Map** and **FlatMap** as functions.

## Usage

For more details about Gostream, you should check at [pkg.go.dev](https://pkg.go.dev/github.com/not2dim/gostream/).

### Example 1: Basic Stream operations

```go
func Example1() {
    n := stream.Slice([]int{8, 7, 4, 9, 1, 2, 3, 6, 0, 5})
        Filter(func (v int) bool { return v%2 == 1 }).
        SortBy(stream.CmpRealNum[int]).
        First()
    fmt.Printf("%+v\n", n) // {Val:1 OK:true}
}
```

### Example 2: Basic Stream operations

```go
func Example2() {
    sum := stream.Slice([]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}).
        FlatMap(func (cap int) stream.Stream[int] {
            var slc []int
            for i := 0; i < 10; i++ {
                slc = append(slc, cap*10+i)
            }
            return stream.Slice(slc)
        }).Reduce(0, func (b, a int) int {
            return b + a
        })
    fmt.Println(sum) // 4950
}
```

### Example 3: Use Cond to emulate Limit

```go
func Example3() {
    cnt, lmt := 0, 3
    slc := stream.Of[any](1, "a", (*int)(nil), []byte{}, 0.6, 1+2i, '😔').
        Cond(func (e any) bool {
            if cnt >= lmt {
                return true
            }
            cnt++
            return false
        }).Collect()
    fmt.Println(slc) // [1 a <nil>]
}
```

### Example 4: Map and Joining

```go
func Example4() {
    stm := stream.Map(
        stream.Of(0, 1, 2, 3, 4, 5, 6, 7, 8, 9),
        func (i int) string {
            return fmt.Sprintf("%v", i)
        },
    )
    fmt.Println(stream.Joining(stm, " ", 10))
    // 0 1 2 3 4 5 6 7 8 9
}
```

### Example 5: Range and Counting

```go
func Example5() {
    dict := map[string]int{
        "apple": 7, "banana": 3,
        "cherry": 5, "elderberry": 1,
        "fig": 2, "milk": 0,
        "tomato": 10, "ham": 6,
    }
    words := stream.MapKeys(dict).FlatMap(func (w string) stream.Stream[string] {
        return stream.Map(
                stream.Range(0, dict[w]),
                    func (i int) string {
                        return w
                    },
                )
        })
    counts := stream.Counting(words)
    fmt.Println(counts)
    // map[apple:7 banana:3 cherry:5 elderberry:1 fig:2 ham:6 tomato:10]
}
```

### Example 6: Iterate rune and Joining

```go
func Example6() {
    var emojis = "😀😃😄😁😆😅😂😊😇🙂🙃"
    itera := iterator.StringRuneIterable(emojis)
    joined := stream.Joining(
        stream.Iterable(itera),
        "➔", uint64(len(emojis)*2),
    )
    fmt.Println(joined) // 😀➔😃➔😄➔😁➔😆➔😅➔😂➔😊➔😇➔🙂➔🙃
}
```
