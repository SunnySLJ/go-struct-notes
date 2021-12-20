package main

import (
	"fmt"
	"sync"
	"time"
)

type People struct {
	Name string
	Age  int
}

// PeoplePool New指定对象创建方法(当Pool为空时，Get方法会调用New方法创建对象)
var PeoplePool = sync.Pool{New: func() interface{} {
	return &People{
		Name: "people",
		Age:  18,
	}
}}

type structR6 struct {
	B1 [100000]int
}

var r6Pool = sync.Pool{
	New: func() interface{} {
		return new(structR6)
	},
}

func usePool() {
	startTime := time.Now()
	for i := 0; i < 10000; i++ {
		sr6 := r6Pool.Get().(*structR6)
		sr6.B1[0] = 0
		//r6Pool.Put(sr6)
	}
	fmt.Println("pool Used:", time.Since(startTime))
}
func standard() {
	startTime := time.Now()
	for i := 0; i < 10000; i++ {
		var sr6 structR6
		sr6.B1[0] = 0
	}
	fmt.Println("standard Used:", time.Since(startTime))
}

func main() {
	standard()
	usePool()

	//p1 := PeoplePool.Get().(*People)
	//fmt.Printf("address:%p, value:%+v\n", p1, p1)
	//PeoplePool.Put(p1)
	//
	//p := PeoplePool.Get()
	//p4 := p.(*People)
	//p2 := PeoplePool.Get().(*People)
	//fmt.Printf("address:%p, value:%+v\n", p4, p4)
	//
	//fmt.Printf("address:%p, value:%+v\n", p2, p2)
	//
	//p3 := PeoplePool.Get().(*People)
	//fmt.Printf("address:%p, value:%+v\n", p3, p3)
}

/**
------ 执行结果 ------
address:0xc00008a020, value:&{Name:people Age:18}	// 新建的People
address:0xc00008a020, value:&{Name:people Age:18}	// Pool中缓存的People
address:0xc00008a080, value:&{Name:people Age:18} // 新建的People
*/
