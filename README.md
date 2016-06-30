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
    sem.Aquire(1)
    go func() {
      defer sem.Release()
      do(task)
    }()
  }
}
```

# Features

## Memory-based semaphore

It can be used between processes that are on the same process.

### Counting semaphore

```go
sem := gosem.NewSemaphore(5)
```

### Timeout semaphore

```go
sem := gosem.NewTimeoutSemaphore(5)
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
