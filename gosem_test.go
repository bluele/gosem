package gosem_test

import (
	"testing"
	"time"

	"github.com/bluele/gosem"
)

func TestSemaphore(t *testing.T) {
	sem := gosem.NewSemaphore(3)
	permits := sem.Permits
	sem.Aquire()
	if sem.Available() != permits-1 {
		t.Errorf("sem.Available() should be %v", permits-1)
	}
	sem.Release()
	if sem.Available() != permits {
		t.Errorf("sem.Available() should be %v", permits)
	}

	if err := sem.AquireWithTimeout(time.Millisecond); err != nil {
		t.Errorf("sem.AquireWithTimeout(time.Millisecond) should not return err: %v", err)
	}
	sem.Release()
	for i := 0; i < permits; i++ {
		sem.Aquire()
	}
	if err := sem.AquireWithTimeout(time.Millisecond); err == nil {
		t.Error("sem.AquireWithTimeout(time.Millisecond) should return error")
	}
}

func TestTimeoutSemaphore(t *testing.T) {
	permits := 3
	sem := gosem.NewTimeoutSemaphore(permits, time.Second)
	sem.Aquire()
	if sem.Available() != permits-1 {
		t.Errorf("sem.Available() should be %v", permits-1)
	}
	sem.Release()
	if sem.Available() != permits-1 {
		t.Errorf("sem.Available() should be %v", permits-1)
	}
	time.Sleep(2 * time.Second)
	if sem.Available() != permits {
		t.Errorf("sem.Available() should be %v", permits)
	}

	if err := sem.AquireWithTimeout(time.Millisecond); err != nil {
		t.Errorf("sem.AquireWithTimeout(time.Millisecond) should not return err: %v", err)
	}
	sem.Release()
	for i := 0; i < permits; i++ {
		sem.Aquire()
	}
	if err := sem.AquireWithTimeout(time.Millisecond); err == nil {
		t.Error("sem.AquireWithTimeout(time.Millisecond) should return error")
	}

	if !sem.Destroy() {
		t.Error("sem should be destroyed")
	}
	// After destroy semaphore, you can release the resources, but cannot aquire new resources.
	for i := 0; i < permits; i++ {
		sem.Release()
	}
	if err := sem.AquireWithTimeout(time.Millisecond); err == nil {
		t.Error("sem.AquireWithTimeout(time.Millisecond) should return error")
	}
}
