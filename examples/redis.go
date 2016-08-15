package main

import (
	"log"
	"sync"
	"time"

	"gopkg.in/redis.v3"

	"github.com/bluele/gosem"
)

func createSemaphore(permit int) *gosem.RedisSemaphore {
	return &gosem.RedisSemaphore{
		Client:  redis.NewClient(&redis.Options{Network: "tcp", Addr: "127.0.0.1:6379", PoolSize: 20}),
		Permits: permit,
	}
}

func task(workerNumber, permit int) chan struct{} {
	ch := make(chan struct{})
	go func() {
		var wg sync.WaitGroup
		sem := createSemaphore(permit)
		sem.Reset()
		for i := 0; i < workerNumber; i++ {
			wg.Add(1)
			av, _ := sem.Available()
			log.Println("Available: ", av)
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
	ch := task(20, 6)
	for {
		select {
		case <-ch:
			return
		}
	}
}
