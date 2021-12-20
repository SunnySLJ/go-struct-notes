package main

import (
	"sync"
	"testing"
)

// BenchmarkWithoutPool 每次都重新申请[]string
func BenchmarkWithoutPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		list := make([]string, 0)
		for i := 0; i < 4; i++ {
			list = append(list, "1")
		}
	}
}

// BenchmarkWithPool 优先使用Pool中的[]string，使用完后将[]string放入Pool
func BenchmarkWithPool(b *testing.B) {
	var slicePool = sync.Pool{New: func() interface{} {
		slice := make([]string, 0)
		return &(slice)
	}}
	for i := 0; i < b.N; i++ {
		list := *(slicePool.Get().(*[]string))
		for i := 0; i < 4; i++ {
			list = append(list, "1")
		}
		list = list[:0] // 重置切片的len
		slicePool.Put(&list)
	}
}

/**
go test -bench=. -benchmem -benchtime=1s
go test -bench=. benchmarkpool_test.go -benchmem -benchtime=1s
方法名                    循环次数  耗时	         内存使用    内存分配次数
BenchmarkWithoutPool-8   7637380  151 ns/op    112 B/op  3 allocs/op
BenchmarkWithPool-8     22991926   49.8 ns/op   32 B/op  1 allocs/op
*/
