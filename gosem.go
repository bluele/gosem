package gosem

import (
	"time"
)

const (
	resource byte = 0
)

type Semaphore struct {
	Permits int

	channel chan byte
}

type TimeoutSemaphore struct {
	Permits int

	channel chan byte
	buffer  chan byte
	per     time.Duration

	destroy chan byte
}

func NewSemaphore(permits int) *Semaphore {
	sm := &Semaphore{
		Permits: permits,
		channel: make(chan byte, permits),
	}
	for i := 0; i < permits; i++ {
		sm.channel <- resource
	}
	return sm
}

func NewTimeoutSemaphore(permits int, per time.Duration) *TimeoutSemaphore {
	sm := &TimeoutSemaphore{
		Permits: permits,
		per:     per,
		channel: make(chan byte, permits),
		buffer:  make(chan byte, permits),
		destroy: make(chan byte, 1),
	}
	for i := 0; i < permits; i++ {
		sm.channel <- resource
	}
	go sm.gc()
	return sm
}

// Aquire gets a new resource from buffer
func (sm *Semaphore) Aquire() error {
	<-sm.channel
	return nil
}

func (sm *Semaphore) AquireWithTimeout(timeout time.Duration) error {
	ch := time.After(timeout)
	select {
	case <-sm.channel:
		return nil
	case <-ch:
		return TimeoutError
	}
}

// Release gives a new resource to buffer
func (sm *Semaphore) Release() error {
	select {
	case sm.channel <- resource:
	}
	return nil
}

// Available gets the number of available resource
func (sm *Semaphore) Available() int {
	return len(sm.channel)
}

// Aquire gets a new resource from buffer
func (sm *TimeoutSemaphore) Aquire() error {
	<-sm.channel
	return nil
}

// AquireWithTimeout gets a new resource from buffer, but this operation will be timeout after specified time
func (sm *TimeoutSemaphore) AquireWithTimeout(timeout time.Duration) error {
	ch := time.After(timeout)
	select {
	case <-sm.channel:
		return nil
	case <-ch:
		return TimeoutError
	}
}

// Release gives a new resource to buffer
func (sm *TimeoutSemaphore) Release() error {
	select {
	case sm.buffer <- resource:
	default:
	}
	return nil
}

// Available gets the number of available resource
func (sm *TimeoutSemaphore) Available() int {
	return len(sm.channel)
}

// Destroy stops observing resources on channels
// After call this method, this semaphore object wouldn't be controllable.
func (sm *TimeoutSemaphore) Destroy() bool {
	select {
	case sm.destroy <- resource:
		return true
	default:
		return false
	}
}

func (sm *TimeoutSemaphore) gc() {
	for {
		select {
		case b := <-sm.buffer:
			<-time.After(sm.per)
			sm.flushBuffer(b)
		case <-sm.destroy:
			close(sm.destroy)
			return
		}
	}
}

func (sm *TimeoutSemaphore) flushBuffer(b byte) {
	select {
	case sm.channel <- b:
	default:
		return
	}

	for {
		select {
		case b = <-sm.buffer:
			select {
			case sm.channel <- b:
			default:
				return
			}
		default:
			return
		}
	}
}
