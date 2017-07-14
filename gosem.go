package gosem

import (
	"container/list"
	"log"
	"sync"
	"time"
)

var (
	resource = struct{}{}
)

// Semaphore is a simple semaphore that can be used to coordinate the number of accessing shared data from multiple goroutine.
type Semaphore struct {
	Permits int

	channel chan struct{}
}

// TimeSemaphore is a semaphore that can be used to coordinate the number of accessing shared data from multiple goroutine with specified time.
type TimeSemaphore struct {
	Permits int
	Logger  *log.Logger

	mu        sync.RWMutex
	channel   chan struct{}
	tokenList *list.List
	buffer    chan *list.Element
	per       time.Duration

	destroy chan struct{}
}

// NewSemaphore returns a new Semaphore object
func NewSemaphore(permits int) *Semaphore {
	sm := &Semaphore{
		Permits: permits,
		channel: make(chan struct{}, permits),
	}
	for i := 0; i < permits; i++ {
		sm.channel <- resource
	}
	return sm
}

// NewTimeSemaphore returns a new TimeSemaphore object
func NewTimeSemaphore(permits int, per time.Duration) *TimeSemaphore {
	sm := &TimeSemaphore{
		Permits:   permits,
		per:       per,
		channel:   make(chan struct{}, permits),
		tokenList: list.New(),
		buffer:    make(chan *list.Element, permits),
		destroy:   make(chan struct{}, 1),
	}
	go sm.gc()
	return sm
}

type token struct {
	acquiredAt *time.Time
}

func newToken() *token {
	t := time.Now()
	return &token{
		acquiredAt: &t,
	}
}

// Acquire gets a new resource
func (sm *Semaphore) Acquire() error {
	<-sm.channel
	return nil
}

// AcquireWithTimeout try to get a new resource, but this operation will be timeout after specified time.
func (sm *Semaphore) AcquireWithTimeout(timeout time.Duration) error {
	ch := time.After(timeout)
	select {
	case <-sm.channel:
		return nil
	case <-ch:
		return TimeoutError
	}
}

// Release gives the resource
func (sm *Semaphore) Release() error {
	select {
	case sm.channel <- resource:
	}
	return nil
}

// Available gets the number of available resources
func (sm *Semaphore) Available() int {
	return len(sm.channel)
}

// Acquire gets a new resource
func (sm *TimeSemaphore) Acquire() error {
	sm.channel <- resource
	sm.mu.Lock()
	sm.pushToken(newToken())
	sm.mu.Unlock()
	return nil
}

// AcquireWithTimeout try to get a new resource, but this operation will be timeout after specified time.
func (sm *TimeSemaphore) AcquireWithTimeout(timeout time.Duration) error {
	ch := time.After(timeout)

	select {
	case sm.channel <- resource:
		sm.mu.Lock()
		sm.pushToken(newToken())
		sm.mu.Unlock()
		return nil
	case <-ch:
		return TimeoutError
	}
}

func (sm *TimeSemaphore) pushToken(tk *token) *list.Element {
	return sm.tokenList.PushBack(tk)
}

// returns: token, true if exists
func (sm *TimeSemaphore) getCurrentToken() (*list.Element, bool) {
	elt := sm.tokenList.Front()
	if elt == nil {
		return nil, false
	}
	return elt, true
}

// Release gives the resource
func (sm *TimeSemaphore) Release() error {
	sm.mu.Lock()
	elt, ok := sm.getCurrentToken()
	if ok {
		sm.tokenList.Remove(elt)
	}
	sm.mu.Unlock()
	if ok {
		sm.buffer <- elt
	}
	return nil
}

// Available gets the number of available resource
func (sm *TimeSemaphore) Available() int {
	return sm.Permits - len(sm.channel)
}

// Destroy stops observing resources of channels.
// After call this method, this semaphore object wouldn't be controllable.
func (sm *TimeSemaphore) Destroy() bool {
	select {
	case sm.destroy <- resource:
		return true
	default:
		return false
	}
}

func (sm *TimeSemaphore) gc() {
	for {
		select {
		case elt := <-sm.buffer:
			tk := elt.Value.(*token)
			sub := time.Now().Sub(*tk.acquiredAt)
			if sub <= sm.per {
				sm.logf("gc() waits(%v)\n", sm.per-sub)
				<-time.After(sm.per - sub)
			} else {
				sm.logf("gc() nowait(%v)\n", sm.per-sub)
			}
			<-sm.channel
			sm.log("gc() release channel")
		case <-sm.destroy:
			return
		}
	}
}

func (sm *TimeSemaphore) log(v ...interface{}) {
	if sm.Logger != nil {
		sm.Logger.Println(v...)
	}
}

func (sm *TimeSemaphore) logf(format string, v ...interface{}) {
	if sm.Logger != nil {
		sm.Logger.Printf(format, v...)
	}
}
