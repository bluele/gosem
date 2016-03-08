package gosem_test

import (
	"sync"
	"testing"

	"github.com/bluele/gosem"
	"gopkg.in/redis.v3"
)

func createTestRedisSemaphore(permits int) *gosem.RedisSemaphore {
	return &gosem.RedisSemaphore{
		Client:    redis.NewClient(&redis.Options{Network: "tcp", Addr: "127.0.0.1:6379"}),
		Permits:   permits,
		Namespace: "GOSEMTEST::",
	}
}

func TestRedisSemaphore(t *testing.T) {
	sem := createTestRedisSemaphore(3)
	permits := sem.Permits

	if err := sem.Aquire(); err != nil {
		t.Error(err)
		return
	}
	if av, _ := sem.Available(); av != permits-1 {
		t.Errorf("sem.Available() should not be %v", av)
	}
	if err := sem.Release(); err != nil {
		t.Error(err)
		return
	}
	if av, _ := sem.Available(); av != permits {
		t.Errorf("sem.Available() should not be %v", av)
	}
}

func TestRedisSemaphoreWithGoroutine(t *testing.T) {
	var (
		num = 3
		wg  sync.WaitGroup
	)
	parent := createTestRedisSemaphore(num)
	defer parent.Reset()
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			child := createTestRedisSemaphore(num)
			if err := child.Aquire(); err != nil {
				panic(err)
			}
		}()

	}
	wg.Wait()
	if av, _ := parent.Available(); av != 0 {
		t.Error("parent.Available() should return 0.")
		return
	}
}
