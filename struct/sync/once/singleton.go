package once

import (
	"fmt"
	"sync"
)

var lock = &sync.Mutex{}

type single1 struct{}

var singleInstance1 *single1

func GetInstance1() *single1 {
	if singleInstance1 == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleInstance1 == nil {
			fmt.Println("Creating Single Instance Now")
			singleInstance1 = &single1{}
		} else {
			fmt.Println("Single Instance already created-1")
		}
	} else {
		fmt.Println("Single Instance already created-2")
	}
	return singleInstance1
}
