# Steam

> 流处理API,你可以像Java Stream一样使用它

[![Godoc](https://img.shields.io/badge/godoc-reference-brightgreen)](https://pkg.go.dev/github.com/chenquan/stream)
[![Go Report Card](https://goreportcard.com/badge/github.com/chenquan/stream)](https://goreportcard.com/report/github.com/chenquan/stream)
[![codecov](https://codecov.io/gh/chenquan/stream/branch/master/graph/badge.svg?token=MXUK9MSJP1)](https://codecov.io/gh/chenquan/stream)
[![GitHub](https://img.shields.io/github/license/chenquan/stream)](https://github.com/chenquan/stream/blob/master/LICENSE)
[![GitHub stars](https://img.shields.io/github/stars/chenquan/stream)](https://github.com/chenquan/stream/stargazers)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fchenquan%2Fstream.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fchenquan%2Fstream?ref=badge_shield)

## GRATITUDE

**API的部分实现参考`go-zero`中模块[fx](https://github.com/tal-tech/go-zero/blob/master/core/fx/stream.go)**

### EXAMPLE

## 安装

```shell
go get -u github.com/chenquan/stream
```

## 使用案例

### 1.创建流

```go
// 创建一个空的流
Empty()
// 使用任意元素创建一个流
Of(1, "1", 22, "22")
// 从循环中创建一个流
From(func (source chan<- interface{}) {
for i := 0; i < 1000; i++ {
source <- i
}
})
// 根据通道创建一个流
ch := make(chan interface{}, 2)
Range(ch)
```

### 2.合并流

```go
// 合并流
Concat(Empty(), Empty())
Concat(Empty(), Of(1, 2, 3))
Empty().Concat(Of(1, 2, 3))

```

### 3.遍历

```go
// 遍历
Of(1, 2, 3, 4).Foreach(func (item interface{}) {
fmt.Println(item)
})
//倒序遍历
Of(1, 2, 3,4).ForeachOrdered(func (item interface{}) {
fmt.Println(item)
})
```

### 4.排序

```go
// 遍历
Of(1, 4, 2, 3).Sort(func (a, b interface{}) bool {
return a.(int) < b.(int)
})
// 1,2,3,4
```

### 5.跳过

```go
Of(1, 4, 2, 3).Skip(2)
// 2,3
```

### 6.限制条数

```go
// 限制条数
Of(1, 4, 2, 3).Limit(2)
// 1,4
```

### 7.返回最后2条数

```go
// 返回最后2条数
Of(1, 4, 2, 3, 1).Tail(2)
// 3,1
```

**更多使用方式请参考:[stream_test.go](stream_test.go)**



## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fchenquan%2Fstream.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fchenquan%2Fstream?ref=badge_large)