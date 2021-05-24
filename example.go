package stream

import "fmt"

func example1() {
	// 创建一个空的流
	Empty()
	// 使用任意元素创建一个流
	Of(1, "1", 22, "22")
	// 从循环中创建一个流
	From(func(source chan<- interface{}) {
		for i := 0; i < 1000; i++ {
			source <- i
		}
	})
	// 根据通道创建一个流
	ch := make(chan interface{}, 2)
	Range(ch)

}
func example2() {
	// 合并流
	Concat(Empty(), Empty())
	Concat(Empty(), Of(1, 2, 3))
	Empty().Concat(Of(1, 2, 3))

}
func example3() {
	// 遍历
	Of(1, 2, 3, 4).Foreach(func(item interface{}) {
		fmt.Println(item)
	})
	//倒序遍历
	Of(1, 2, 3, 4).ForeachOrdered(func(item interface{}) {
		fmt.Println(item)
	})
}
func example4() {
	// 排序
	Of(1, 4, 2, 3).Sort(func(a, b interface{}) bool {
		return a.(int) < b.(int)
	})
}
func example5() {
	// 跳过前2条
	Of(1, 4, 2, 3).Skip(2)
}
func example6() {
	// 限制2条数
	Of(1, 4, 2, 3).Limit(2)
}
func example7() {
	// 返回最后2条数
	Of(1, 4, 2, 3).Tail(2)
}
func example8() {
	// 返回前2条
	Of(1, 4, 2, 3).Head(2)
}
