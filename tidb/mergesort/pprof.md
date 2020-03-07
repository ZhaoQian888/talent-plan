# 解题思路及优化过程

## 解题思路

### 1. Grouting 并行工作可以尽可能的提高cpu的利用率

### 2.并归排序的合并过程会占用大量的内存和cpu

### 3.快速排序速度可以提高很多

### 4.使用归并树

## 优化记录

1. 比NormalSort操作更快
2. 由于采取了递归的方式，导致占有了大量的内存

``` shell
(base) zhaoqiandeMacBook-Pro:mergesort zhaoqian$ make bench
go test -bench Benchmark -run xx -count 5 -benchmem
goos: darwin
goarch: amd64
pkg: pingcap/talentplan/tidb/mergesort
BenchmarkMergeSort-8                   1        2894949988 ns/op        3087009176 B/op 16777221 allocs/op
BenchmarkMergeSort-8                   1        2715994714 ns/op        3087009016 B/op 16777221 allocs/op
BenchmarkMergeSort-8                   1        2704834934 ns/op        3087007840 B/op 16777215 allocs/op
BenchmarkMergeSort-8                   1        2707028820 ns/op        3087007936 B/op 16777216 allocs/op
BenchmarkMergeSort-8                   1        2672504573 ns/op        3087007840 B/op 16777215 allocs/op
BenchmarkNormalSort-8                  1        3685581766 ns/op              64 B/op          2 allocs/op
BenchmarkNormalSort-8                  1        3704931451 ns/op              64 B/op          2 allocs/op
BenchmarkNormalSort-8                  1        3692188061 ns/op              64 B/op          2 allocs/op
BenchmarkNormalSort-8                  1        3733914598 ns/op              64 B/op          2 allocs/op
BenchmarkNormalSort-8                  1        3761529397 ns/op              64 B/op          2 allocs/op
PASS
ok      pingcap/talentplan/tidb/mergesort       36.109s
```

``` go 
func MergeSort(src []int64) {
	n := len(src)
	if n == 2 {
		if src[0] < src[1] {
			return
		}
		src[0], src[1] = src[1], src[0]
		return
	} else if n == 1 {
		return
	} else {
		len1 := len(src) / 2
		len2 := len(src) - (len(src) / 2)
		MergeSort(src[:len1])
		MergeSort(src[len1:])
		arr1 := make([]int64, len1)
		arr2 := make([]int64, len2)
		copy(arr1, src[:len1])
		copy(arr2, src[len1:])
		i := 0
		j := 0
		n := 0
		for {
			if i >= len1 || j >= len2 {
				break
			}
			if arr1[i] > arr2[j] {
				src[n] = arr2[j]
				j++
				n++
			} else {
				src[n] = arr1[i]
				i++
				n++
			}
		}
		if i < len1 {
			for i < len1 {
				src[n] = arr1[i]
				i++
				n++
			}
		} else {
			for j < len2 {
				src[n] = arr2[j]
				j++
				n++
			}
		}
	}
}
```

## 版本二 并行
``` go
func MergeSort(src []int64) {
	tag := make(chan int, 2)
	subMergeSort(src, tag)
}
func subMergeSort(src []int64, tag chan int) {
	n := len(src)
	if n == 2 {
		if src[0] < src[1] {
			tag <- 1
			return
		}
		src[0], src[1] = src[1], src[0]
		tag <- 1
		return
	} else if n == 1 {
		tag <- 1
		return
	} else {
		subtag := make(chan int, 2)
		len1 := len(src) / 2
		len2 := len(src) - (len(src) / 2)
		go subMergeSort(src[:len1], subtag)
		go subMergeSort(src[len1:], subtag)
		_ = <-subtag
		_ = <-subtag
		arr1 := make([]int64, len1)
		arr2 := make([]int64, len2)
		copy(arr1, src[:len1])
		copy(arr2, src[len1:])
		i := 0
		j := 0
		n := 0
		for {
			if i >= len1 || j >= len2 {
				break
			}
			if arr1[i] > arr2[j] {
				src[n] = arr2[j]
				j++
				n++
			} else {
				src[n] = arr1[i]
				i++
				n++
			}
		}
		if i < len1 {
			for i < len1 {
				src[n] = arr1[i]
				i++
				n++
			}
		} else {
			for j < len2 {
				src[n] = arr2[j]
				j++
				n++
			}
		}
	}
	tag <- 1
	return
}
```
``` shell
go test -bench Benchmark -run xx -count 5 -benchmem
goos: darwin
goarch: amd64
pkg: pingcap/talentplan/tidb/mergesort
BenchmarkMergeSort-8                   1        5145646499 ns/op        4344033440 B/op 26225357 allocs/op
BenchmarkMergeSort-8                   1        4707741418 ns/op        4108602336 B/op 25697766 allocs/op
BenchmarkMergeSort-8                   1        4422213684 ns/op        4062730176 B/op 25542888 allocs/op
BenchmarkMergeSort-8                   1        4572316885 ns/op        4066274496 B/op 25579808 allocs/op
BenchmarkMergeSort-8                   1        4531702013 ns/op        4073060928 B/op 25635500 allocs/op
BenchmarkNormalSort-8                  1        3464384914 ns/op              64 B/op          2 allocs/op
BenchmarkNormalSort-8                  1        3450776641 ns/op              64 B/op          2 allocs/op
BenchmarkNormalSort-8                  1        3427838297 ns/op              64 B/op          2 allocs/op
BenchmarkNormalSort-8                  1        3522915165 ns/op              64 B/op          2 allocs/op
BenchmarkNormalSort-8                  1        3592085567 ns/op              64 B/op          2 allocs/op
PASS
ok      pingcap/talentplan/tidb/mergesort       46.063s
```
因为携程之间的等待关系，实际速度比原来还慢了  - -! 

又想了一种办法，还是很慢
``` go
func MergeSort(src []int64) {
	l := len(src)
	if l <= 1 {
		return
	}
	cb := make(chan slice, l)
	go subMergeSort(src, 0, l-1, cb)
	combine(src, cb)
	close(cb)

}

// start 数组的开头，end 数组的结尾   [1,3,4,5] start 0,end 3
func subMergeSort(src []int64, start int, end int, cb chan slice) {

	len := len(src)
	if len == 2 {
		if src[0] > src[1] {
			src[0], src[1] = src[1], src[0]
		}
		return
	} else if len == 1 {
		return
	} else {
		len1 := len / 2

		subMergeSort(src[:len1], start, start+len1-1, cb)
		subMergeSort(src[len1:], start+len1, end, cb)
		cb <- slice{
			start: start,
			mid:   start + len1,
			end:   end,
		}
	}
}

func combine(src []int64, cb chan slice) {
	var arr slice
	for {
		arr = <-cb

		combineArr(src, arr)
		if arr.start == 0 && arr.end == len(src)-1 {
			break
		}
	}
}
func combineArr(src []int64, arr slice) {
	len1 := arr.mid - arr.start
	len2 := arr.end - arr.mid + 1
	arr1 := make([]int64, len1)
	arr2 := make([]int64, len2)

	copy(arr1, src[arr.start:arr.mid])
	copy(arr2, src[arr.mid:arr.end+1])
	i := 0
	j := 0
	n := arr.start
	for {
		if i >= len1 || j >= len2 {
			break
		}
		if arr1[i] > arr2[j] {
			src[n] = arr2[j]
			j++
			n++
		} else {
			src[n] = arr1[i]
			i++
			n++
		}
	}
	if i < len1 {
		for i < len1 {
			src[n] = arr1[i]
			i++
			n++
		}
	} else {
		for j < len2 {
			src[n] = arr2[j]
			j++
			n++
		}
	}

}
```

## 版本三引入combine树

思路：每当一个线程完成，就看看是否能够合并。

``` go
func (t *combineTreeNode) rightCombine(src []int64) {
	if t.rightchild == nil {
		return
	}
	if t.rightchild.val.start == t.val.end {
		combineArr(src, comArr{
			arr1: t.val,
			arr2: t.rightchild.val,
		})
		t.val = arr{
			start: t.val.start,
			end:   t.rightchild.val.end,
		}
		t.rightchild = t.rightchild.rightchild
		if t.rightchild != nil {
			t.rightchild.father = t
		}
	}
}
func (t *combineTreeNode) leftCombine(src []int64) {
	if t.leftchild == nil {
		return
	}
	if t.leftchild.val.end == t.val.start {
		combineArr(src, comArr{
			arr1: t.leftchild.val,
			arr2: t.val,
		})
		t.val = arr{
			start: t.leftchild.val.start,
			end:   t.val.end,
		}
		t.leftchild = t.leftchild.leftchild
		if t.leftchild != nil {
			t.leftchild.father = t
		}
	}
}

func checkAndCombine(src []int64, root **combineTreeNode, element arr) {
	if *root == nil {
		*root = &combineTreeNode{
			val: element,
		}
		return
	}
	var p *combineTreeNode
	p = *root
	for p != nil {
		if p.val.end < element.start {
			if p.rightchild == nil {
				p.rightchild = &combineTreeNode{
					val:    element,
					father: p,
				}
				break
			} else {
				p = p.rightchild
			}
		} else if p.val.end == element.start {
			combineArr(src, comArr{
				arr1: p.val,
				arr2: element,
			})
			p.val = arr{start: p.val.start, end: element.end}
			p.rightCombine(src)
			if p.father != nil {
				if p.father.leftchild == p {
					p.father.leftCombine(src)
				} else {
					p.father.rightCombine(src)
				}
			}
			break
		} else if element.end < p.val.start {
			if p.leftchild == nil {
				p.leftchild = &combineTreeNode{
					val:    element,
					father: p,
				}
				break
			} else {
				p = p.leftchild
			}
		} else if element.end == p.val.start {
			combineArr(src, comArr{
				arr1: element,
				arr2: p.val,
			})
			p.val = arr{start: element.start, end: p.val.end}
			p.leftCombine(src)
			if p.father != nil {
				if p.father.leftchild == p {
					p.father.leftCombine(src)
				} else {
					p.father.rightCombine(src)
				}
			}
			break
		}
	}
}
```

引入归并树后，并把归并排序的递归形式改成栈的形式。性能还是没有提升。我开始反思到底是什么影响了速度，后发现是因为如果每个切片都采用归并，会导致每次的合并都会消耗大量内存并影响速度。

把自排序换成sort.Slice 再次测试性能。多goroutine确实提高了性能。加快了速度并降低了内存消耗。但还不够理想

``` shell
BenchmarkMergeSort-8                   4         275001326 ns/op        301991348 B/op        20 allocs/op
BenchmarkMergeSort-8                   4         251930556 ns/op        268436760 B/op        19 allocs/op
BenchmarkMergeSort-8                   4         261677683 ns/op        301991448 B/op        19 allocs/op
BenchmarkMergeSort-8                   4         262392407 ns/op        301990988 B/op        18 allocs/op
BenchmarkMergeSort-8                   4         258068330 ns/op        301991000 B/op        18 allocs/op
BenchmarkNormalSort-8                  2         647006307 ns/op              64 B/op          2 allocs/op
BenchmarkNormalSort-8                  2         668865890 ns/op              64 B/op          2 allocs/op
BenchmarkNormalSort-8                  2         631543894 ns/op              64 B/op          2 allocs/op
BenchmarkNormalSort-8                  2         670301114 ns/op              64 B/op          2 allocs/op
BenchmarkNormalSort-8                  2         697946620 ns/op              64 B/op          2 allocs/op
PASS
ok      pingcap/talentplan/tidb/mergesort       30.386s
```

于是我决定引入快排作为各自goroutine的切片的排序方法

果然速度大幅度提高

```shell
pkg: pingcap/talentplan/tidb/mergesort
BenchmarkMergeSort-8                   2         670408204 ns/op        301992272 B/op        18 allocs/op
BenchmarkMergeSort-8                   2         669716934 ns/op        301991048 B/op        11 allocs/op
BenchmarkMergeSort-8                   2         645073403 ns/op        301991392 B/op        13 allocs/op
BenchmarkMergeSort-8                   2         670683402 ns/op        301990792 B/op        11 allocs/op
BenchmarkMergeSort-8                   2         682199114 ns/op        268436816 B/op        13 allocs/op
BenchmarkNormalSort-8                  1        3481941904 ns/op              64 B/op          2 allocs/op
BenchmarkNormalSort-8                  1        3478714484 ns/op              64 B/op          2 allocs/op
BenchmarkNormalSort-8                  1        3477961054 ns/op              64 B/op          2 allocs/op
BenchmarkNormalSort-8                  1        3522772548 ns/op              64 B/op          2 allocs/op
BenchmarkNormalSort-8                  1        3800350215 ns/op              64 B/op          2 allocs/op
PASS
ok      pingcap/talentplan/tidb/mergesort       33.906s
```



