package main

import "time"

func main() {
	var ch1 chan int
	var ch2 = make(chan int)
	close(ch2)
	go func() {
		for {
			select {
			case d := <-ch1:
				println("ch1", d)
			case d := <-ch2:
				println("ch2", d)
			}
		}
	}()
	time.Sleep(time.Hour)
}
