package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	runtime2()
}

func runtime2() {
	//runtime.GOMAXPROCS(1)
	//for i := 0; i < 10; i++ {
	//	i:= i
	//	go func() {
	//		fmt.Println("A: " , i)
	//	}()
	//}
	//time.Sleep(time.Hour)
	var ch = make(chan int, 1)
	ch <- 1
	fmt.Println(<-ch)
	ch1 := make(chan string, 1)
	ch1 <- "hello world"
	fmt.Println(<-ch1)
}

func runtime1() {
	var x int
	threads := runtime.GOMAXPROCS(0) - 1
	for i := 0; i < threads; i++ {
		go func() {
			for {
				x++
			}
		}()
	}
	time.Sleep(time.Second)
	fmt.Println("x =", x)
}
