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
	}
	for i := 0; i < permits; i++ {
		sm.channel <- resource
	}
	go sm.gc()
	return sm
}

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

func (sm *Semaphore) Release() error {
	select {
	case sm.channel <- resource:
	default:
		return TooManyReleaseError
	}
	return nil
}

func (sm *Semaphore) Available() int {
	return len(sm.channel)
}

func (sm *TimeoutSemaphore) Aquire() error {
	<-sm.channel
	return nil
}

func (sm *TimeoutSemaphore) AquireWithTimeout(timeout time.Duration) error {
	ch := time.After(timeout)
	select {
	case <-sm.channel:
		return nil
	case <-ch:
		return TimeoutError
	}
}

func (sm *TimeoutSemaphore) Release() error {
	select {
	case sm.buffer <- resource:
	default:
		return TooManyReleaseError
	}
	return nil
}

func (sm *TimeoutSemaphore) Available() int {
	return len(sm.channel)
}

// Destroy close each channels and stop observing resources on channels
func (sm *TimeoutSemaphore) Destroy() {
	select {
	case sm.destroy <- resource:
		close(sm.channel)
		close(sm.buffer)
	default:
		return
	}
}

func (sm *TimeoutSemaphore) gc() {
	tick := time.NewTicker(sm.per)
	defer tick.Stop()
	for {
	Main:
		select {
		case <-tick.C:
			for {
				select {
				case b := <-sm.buffer:
					sm.channel <- b
				default:
					break Main
				}
			}
		case <-sm.destroy:
			close(sm.destroy)
			return
		}
	}
}
