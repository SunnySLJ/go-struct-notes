package main

import (
	"fmt"
)

type person struct {
	age  int
	name string
}

/*
	源码文件：runtime/map.go
	初始化：makemap
		1. map中bucket的初始化 makeBucketArray
		2. overflow 的定义为哈希冲突的值，用链表法解决
	赋值：mapassign
		1. 不支持并发操作 h.flags&hashWriting
		2. key的alg算法包括两种，一个是equal，用于对比；另一个是hash，也就是哈希函数
		3. 位操作 h.flags ^= hashWriting 与 h.flags &^= hashWriting
		4. 根据hash找到bucket，遍历其链表下的8个bucket，对比hashtop值；如果key不在map中，判断是否需要扩容
	扩容：hashGrow
		1. 扩容时，会将原来的 buckets 搬运到 oldbuckets
	读取：mapaccess
		1. mapaccess1_fat 与 mapaccess2_fat 分别对应1个与2个返回值
		2. hash 分为低位与高位两部分，先通过低位快速找到bucket，再通过高位进一步查找，对后对比具体的key
		3. 访问到oldbuckets中的数据时，会迁移到buckets
	删除：mapdelete
		1. 引入了emptyOne与emptyRest，后者是为了加速查找
*/

func main() {
	mapCompile()
}

// go tool compile -S map.go
func mapCompile() {
	m := make(map[string]int, 9)
	key := "test"
	m[key] = 1
	_, ok := m[key]
	if ok {
		delete(m, key)
	}
}

func copy1() {
	var m = map[int]person{
		1: person{11, "abc"},
		2: person{22, "def"},
	}
	var mm = map[int]*person{}
	for k, v := range m {
		//value := v
		//mm[k] = &value
		mm[k] = &v
	}

	fmt.Println(mm[1], mm[2])
}

func copy2() {
	originalMap := make(map[string]int)
	originalMap["one"] = 1
	originalMap["two"] = 2

	// Create the target map
	targetMap := make(map[string]int)

	// Copy from the original map to the target map
	for key, value := range originalMap {
		targetMap[key] = value
	}

	fmt.Println(targetMap["one"], targetMap["two"])
}
