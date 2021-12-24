package main

import (
	"fmt"
	"time"
)

func main() {
	timer1 := time.NewTimer(2 * time.Second)
	go func() {
		if !timer1.Stop() {
			<-timer1.C
		}
	}()

	select {
	case <-timer1.C:
		fmt.Println("expired")
	default:
	}
	println("done")
}
