package main

import (
	"fmt"
	"time"

	"github.com/bluele/gosem"
)

func task(i int) {
	time.Sleep(time.Millisecond)
	fmt.Printf("done: task-%v\n", i)
}

func getMainCh() chan byte {
	ch := make(chan byte)
	go func() {
		taskNumber := 20
		permit := 6
		sem := gosem.NewTimeoutSemaphore(permit, time.Second)

		for i := 0; i < taskNumber; i++ {
			sem.Aquire(1)
			go func(i int) {
				task(i)
				sem.Release()
			}(i)
		}
		sem.Wait()
		ch <- 0
	}()
	return ch
}

func main() {
	tick := time.NewTicker(time.Second)
	ch := getMainCh()
Main:
	for {
		select {
		case <-ch:
			break Main
		case t := <-tick.C:
			fmt.Println(t.Unix())
		}
	}
}
