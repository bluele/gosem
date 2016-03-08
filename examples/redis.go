package main

import (
	"fmt"
	"time"

	"gopkg.in/redis.v3"

	"github.com/bluele/gosem"
)

func createSemaphore(permit int) *gosem.RedisSemaphore {
	return &gosem.RedisSemaphore{
		Client: redis.NewClient(&redis.Options{Network: "tcp", Addr: "127.0.0.1:6379"}),
		Option: &gosem.RedisSemaphoreOption{
			Permit: permit,
		},
	}
}

func task(i int) {
	time.Sleep(time.Millisecond)
	fmt.Printf("done: task-%v\n", i)
}

func getMainCh() chan byte {
	ch := make(chan byte)
	go func() {
		taskNumber := 20
		permit := 6
		sem := createSemaphore(permit)

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
