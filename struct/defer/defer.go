package main

import "fmt"

func printArray(array *[3]int) {
	for i := range array {
		fmt.Println(array[i])
	}
}
func deferFuncParameter() {
	var aArray = [3]int{1, 2, 3}
	defer printArray(&aArray)
	aArray[0] = 10
	return
}

func deferFuncReturn() (result int) {
	i := 1
	defer func() {
		result++
		fmt.Println(result)
	}()
	return i
}

func foo() int {
	var i int
	defer func() {
		i++
		fmt.Println(i)
	}()
	return 1
}

func main() {
	foo()
}
