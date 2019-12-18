package main

import (
	"fmt"
	"sync"
)

var Ch = map[string]chan bool{"1": make(chan bool, 1)}

func main() {
	var wg sync.WaitGroup
	wg.Add(10)

	for index := 0; index < 10; index++ {
		go func(index int) {
			defer wg.Add(-1)
			for {
				ret := <-Ch["1"]
				// delete(Ch, "1")
				fmt.Println(index)
				fmt.Printf("ret: %v\n", ret)
				return
			}
		}(index)
	}

	for index := 0; index < 10; index++ {
		Ch["1"] <- true
	}
	wg.Wait()
}
