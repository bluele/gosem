package gosem

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/redis.v3"
)

var (
	ExistsKey      = "EXISTS"
	GrabbedKey     = "GRABBED"
	AvailableKey   = "AVAILABLE"
	ReleaseLockKey = "RELEASE_LOCK"

	ResourceValue    = "1"
	DefaultNameSpace = "GOSEM::"
)

type RedisSemaphore struct {
	Client             *redis.Client
	Permits            int
	Namespace          string
	UseLocalTime       bool
	StaleClientTimeout time.Duration

	mu          sync.RWMutex
	localTokens []string
}

func NewRedisSemaphore(client *redis.Client, permits int) *RedisSemaphore {
	return &RedisSemaphore{
		Client:    client,
		Permits:   permits,
		Namespace: DefaultNameSpace,
	}
}

func (sm *RedisSemaphore) init() error {
	err := sm.Client.Expire(sm.ExistsKey(), 10*time.Second).Err()
	if err != nil {
		return err
	}
	multi := sm.Client.Multi()
	_, err = multi.Exec(func() error {
		multi.Del(sm.GrabbedKey(), sm.AvailableKey())
		multi.RPush(sm.AvailableKey(), makeRangeSlice(0, sm.Permits)...)
		return nil
	})
	if err != nil {
		return err
	}
	_, err = sm.Client.Persist(sm.ExistsKey()).Result()
	return err
}

func (sm *RedisSemaphore) Reset() error {
	return sm.init()
}

func (sm *RedisSemaphore) existsOrInit() error {
	_, err := sm.Client.GetSet(sm.ExistsKey(), ResourceValue).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	if err == redis.Nil {
		return sm.init()
	}
	return nil
}

func (sm *RedisSemaphore) Acquire() error {
	return sm.AcquireWithTimeout(0)
}

func (sm *RedisSemaphore) AcquireWithTimeout(timeout time.Duration) error {
	return sm.acquire(func() (string, error) {
		results, err := sm.Client.BLPop(timeout, sm.AvailableKey()).Result()
		if err != nil {
			if err == redis.Nil {
				err = TimeoutError
			}
			return "", err
		}
		return results[1], nil
	})
}

func (sm *RedisSemaphore) acquire(getToken func() (string, error)) error {
	err := sm.existsOrInit()
	if err != nil {
		return err
	}
	if sm.StaleClientTimeout != 0 {
		sm.ReleaseStaleLocks(10 * time.Second)
	}
	token, err := getToken()
	if err != nil {
		return err
	}
	sm.mu.Lock()
	sm.localTokens = append(sm.localTokens, token)
	sm.mu.Unlock()

	t, err := sm.CurrentTime()
	if err != nil {
		return err
	}

	if err = sm.Client.HSet(sm.GrabbedKey(), token, fmt.Sprintf("%v", t)).Err(); err != nil {
		return err
	}
	return nil
}

func (sm *RedisSemaphore) Release() error {
	ok, err := sm.hasLock()
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.signal(sm.localTokens[len(sm.localTokens)-1])
}

func (sm *RedisSemaphore) Available() (int, error) {
	av, err := sm.Client.LLen(sm.AvailableKey()).Result()
	return int(av), err
}

func (sm *RedisSemaphore) hasLock() (bool, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for _, token := range sm.localTokens {
		ok, err := sm.isLocked(token)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func (sm *RedisSemaphore) isLocked(token string) (bool, error) {
	return sm.Client.HExists(sm.GrabbedKey(), token).Result()
}

func (sm *RedisSemaphore) ReleaseStaleLocks(expires time.Duration) error {
	_, err := sm.Client.GetSet(sm.ReleaseLockKey(), ResourceValue).Result()
	if err == redis.Nil {
		return nil
	} else if err != nil {
		return err
	}
	err = sm.Client.Expire(sm.ReleaseLockKey(), expires).Err()
	if err != nil {
		return err
	}
	defer sm.Client.Del(sm.ReleaseLockKey())
	m, err := sm.Client.HGetAllMap(sm.GrabbedKey()).Result()
	if err != nil {
		return err
	}
	current, err := sm.CurrentTime()
	if err != nil {
		return err
	}
	for token, lookedAt := range m {
		v, err := strconv.ParseFloat(lookedAt, 64)
		if err != nil {
			return err
		}
		if (v-current)*float64(time.Second) < float64(sm.StaleClientTimeout) {
			sm.signal(token)
		}
	}
	return nil
}

func (sm *RedisSemaphore) signal(token string) error {
	multi := sm.Client.Multi()
	_, err := multi.Exec(func() error {
		multi.HDel(sm.GrabbedKey(), token)
		multi.LPush(sm.AvailableKey(), token)
		return nil
	})
	return err
}

func (sm *RedisSemaphore) ExistsKey() string {
	return fmt.Sprintf("%v%v", sm.Namespace, ExistsKey)
}

func (sm *RedisSemaphore) GrabbedKey() string {
	return fmt.Sprintf("%v%v", sm.Namespace, GrabbedKey)
}

func (sm *RedisSemaphore) AvailableKey() string {
	return fmt.Sprintf("%v%v", sm.Namespace, AvailableKey)
}

func (sm *RedisSemaphore) ReleaseLockKey() string {
	return fmt.Sprintf("%v%v", sm.Namespace, ReleaseLockKey)
}

func (sm *RedisSemaphore) CurrentTime() (float64, error) {
	if sm.UseLocalTime {
		return float64(time.Now().UnixNano()) / 1000000000, nil
	}

	t, err := sm.Client.Time().Result()
	if err != nil {
		return 0.0, err
	}
	return strconv.ParseFloat(strings.Join(t, "."), 64)
}

func makeRangeSlice(start, end int) []string {
	sl := []string{}
	for i := start; i < end; i++ {
		sl = append(sl, strconv.Itoa(i))
	}
	return sl
}
