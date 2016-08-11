# gosem

[![CircleCI](https://circleci.com/gh/bluele/gosem/tree/master.svg?style=svg)](https://circleci.com/gh/bluele/gosem/tree/master)

gosem provides multiple semaphore functions.
Currently supports inmemory, redis implementation.

# Example

```go
// You can set a number of available resources.
var sem = gosem.NewSemaphore(3)

func main() {
  for _, task := range tasks {
    // Do task if available resources exists.
    sem.Acquire()
    go func(task Task) {
      defer sem.Release()
      do(task)
    }(task)
  }
}
```

# Features

## Memory-based semaphore

### Simple semaphore

Semaphore is a semaphore that can be used to coordinate the number of accessing shared data from multiple goroutine.

```go
sem := gosem.NewSemaphore(5)
// acquire a new resource
sem.Acquire()
// release the resource
sem.Release()
```

### Time semaphore

TimeSemaphore is a semaphore that can be used to coordinate the number of accessing shared data from multiple goroutine with specified time.

```go
sem := gosem.NewTimeSemaphore(5, time.Second)
// acquire a new resource
sem.Acquire()
// release the resource
sem.Release()
```

## Redis-based semaphore

Implements a semaphore using Redis.

It can be used between processes that are on the multiple host.

### Counting semaphore

```go
import "gopkg.in/redis.v3"

client := redis.NewClient(
  &redis.Options{Network: "tcp", Addr: "127.0.0.1:6379"},
)
sem := gosem.NewRedisSemaphore(client, 5)
```


# Contribution

1. Fork ([https://github.com/bluele/gosem/fork](https://github.com/bluele/gosem/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create new Pull Request

# Author

**Jun Kimura**

* <http://github.com/bluele>
* <junkxdev@gmail.com>
