package main

import (
	"log"
	"sync"
	"time"

	"github.com/bluele/gosem"
)

func task(workerNumber, permit int) chan struct{} {
	ch := make(chan struct{})
	go func() {
		var wg sync.WaitGroup
		sem := gosem.NewTimeSemaphore(permit, time.Second)
		for i := 0; i < workerNumber; i++ {
			wg.Add(1)
			sem.Acquire()
			go func(i int) {
				defer wg.Done()
				time.Sleep(time.Millisecond)
				log.Printf("Done: task-%v\n", i)
				sem.Release()
			}(i)
		}
		wg.Wait()
		ch <- struct{}{}
		close(ch)
	}()
	return ch
}

func main() {
	tick := time.NewTicker(time.Second)
	ch := task(20, 6)
Main:
	for {
		select {
		case <-ch:
			break Main
		case <-tick.C:
			log.Println("tick")
		}
	}
}
