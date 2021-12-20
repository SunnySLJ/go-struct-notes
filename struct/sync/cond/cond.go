package main

import (
	"fmt"
	"sync"
	"time"
)

var mutex = sync.Mutex{}
var cond = sync.NewCond(&mutex)

var queue []int

func producer() {
	i := 0
	for {
		mutex.Lock()
		queue = append(queue, i)
		i++
		fmt.Println("produce", i)
		mutex.Unlock()

		cond.Broadcast()
		time.Sleep(200 * time.Millisecond)
	}
}

func consumer(consumerName string) {
	for {
		mutex.Lock()
		for len(queue) == 0 {
			cond.Wait()
		}

		fmt.Println(consumerName, queue[0])
		queue = queue[1:]
		mutex.Unlock()
	}
}

func main() {
	// 开启一个 producer
	go producer()

	// 开启两个 consumer
	go consumer("consumer-1")
	go consumer("consumer-2")

	for {
		time.Sleep(1 * time.Minute)
	}
}
