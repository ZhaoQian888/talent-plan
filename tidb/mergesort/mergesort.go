package main

// start =0 end =len()
type arr struct {
	start int
	end   int
}
type comArr struct {
	arr1 arr
	arr2 arr
}

type combineTreeNode struct {
	val        arr
	leftchild  *combineTreeNode
	rightchild *combineTreeNode
	father     *combineTreeNode
}

var grNum int = 4

// MergeSort performs the merge sort algorithm.
// Please supplement this function to accomplish the home work.
func MergeSort(src []int64) {
	l := len(src)
	//如果数组比较小，采用插排即可
	if l < grNum*100 {
		InsertSort(src)
		return
	}
	// 数组分块，block为块大小
	block := l / grNum
	// 用来通知主goroutine 那些goroutine已经计算完毕
	signal := make(chan arr, grNum)
	// 对所有goroutine 分发任务
	for i := 0; i < grNum; i++ {
		go goSubSort(i, src[i*block:(i+1)*block], signal, block)
	}
	// 主goroutine完成剩余部分
	if l%grNum != 0 {
		goSubSort(grNum, src[grNum*block:], signal, block)
		grNum++
	}
	// 用来判断是否所有的goroutine都完成归并
	count := 0
	// 引入归并树
	var cbTree *combineTreeNode
	for {
		// 所有的都完成归并
		if count == grNum {
			break
		}
		// 去除待归并切片
		el := <-signal
		// 插入归并树
		CheckAndCombine(src, &cbTree, el)
		// 有效插入时 计数++
		if el.end != 0 {
			count++
		}
	}

}
func goSubSort(gr int, src []int64, signal chan arr, block int) {
	QuickSort(src)
	// sort.Slice(src, func(i, j int) bool { return src[i] < src[j] }) //可以是任意排序方式
	signal <- arr{start: gr * block, end: gr*block + len(src)} //排序结束，注入信号
}

// InsertSort 插排
func InsertSort(src []int64) {
	var temp int
	for i := 0; i < len(src); i++ {
		temp = i
		for j := i + 1; j < len(src); j++ {
			if src[j] < src[temp] {
				temp = j
			}
		}
		src[i], src[temp] = src[temp], src[i]
	}
	return
}

// QuickSort 快排
func QuickSort(src []int64) {
	if len(src) <= 8 { // 长度比较小时采用选择排序
		var temp int
		for i := 0; i < len(src); i++ {
			temp = i
			for j := i + 1; j < len(src); j++ {
				if src[j] < src[temp] {
					temp = j
				}
			}
			src[i], src[temp] = src[temp], src[i]
		}
		return
	}
	i := 0
	j := len(src) - 1
	tmp := src[0]
	for i < j {
		for src[j] >= tmp && i < j {
			j--
		}
		src[i] = src[j]
		for src[i] <= tmp && i < j {
			i++
		}
		src[j] = src[i]
	}
	src[i] = tmp
	QuickSort(src[0:i])
	QuickSort(src[i+1:])
}

//合并切片
func combineArr(src []int64, arr comArr) {
	len1 := arr.arr1.end - arr.arr1.start
	len2 := arr.arr2.end - arr.arr2.start
	arr1 := make([]int64, len1)
	arr2 := make([]int64, len2)
	copy(arr1, src[arr.arr1.start:arr.arr1.end])
	copy(arr2, src[arr.arr2.start:arr.arr2.end])
	i := 0
	j := 0
	n := arr.arr1.start
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

// 归并树右结合
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

// 归并树左结合
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

// CheckAndCombine 插入归并树
func CheckAndCombine(src []int64, root **combineTreeNode, element arr) {
	// 第一次创建
	if *root == nil {
		*root = &combineTreeNode{
			val: element,
		}
		return
	}
	// p 遍历指针
	var p *combineTreeNode
	p = *root
	for p != nil {
		if p.val.end < element.start {
			if p.rightchild == nil { //左插入
				p.rightchild = &combineTreeNode{
					val:    element,
					father: p,
				}
				break
			} else {
				p = p.rightchild //指针右移
			}
		} else if p.val.end == element.start { //右合并
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
			if p.leftchild == nil { //左插入
				p.leftchild = &combineTreeNode{
					val:    element,
					father: p,
				}
				break
			} else {
				p = p.leftchild //指针左移
			}
		} else if element.end == p.val.start { //左合并
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
