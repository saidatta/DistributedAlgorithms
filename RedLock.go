package main

import (
	"errors"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"
)

var uuidCounter int64

const (
	// ClockDriftFactor is the maximum amount of time that a clock can be
	// drifted before the Redlock algorithm considers it invalid.
	ClockDriftFactor = 0.01

	// RedisConnectTimeout is the maximum amount of time that the Redlock
	// algorithm will wait for a connection to a Redis server.
	RedisConnectTimeout = time.Second
)

// RedisClient is an interface that represents a Redis client.
type RedisClient interface {
	// Set sets the value of a key.
	Set(key string, value interface{}, expiration time.Duration) error

	// Get gets the value of a key.
	Get(key string) (interface{}, error)

	// Del deletes a key.
	Del(key string) error

	// Eval runs a Redis script.
	Eval(script string, keys []string, args ...interface{}) (interface{}, error)
}

// Redlock is a distributed lock manager that uses the Redlock algorithm.
type Redlock struct {
	// ClockDriftFactor is the maximum amount of time that a clock can be
	// drifted before the Redlock algorithm considers it invalid.
	ClockDriftFactor float64

	// RedisConnectTimeout is the maximum amount of time that the Redlock
	// algorithm will wait for a connection to a Redis server.
	RedisConnectTimeout time.Duration

	// retryCount is the number of times that the Redlock algorithm will
	// try to acquire the lock before giving up.
	retryCount int

	// retryDelay is the amount of time that the Redlock algorithm will
	// wait between lock acquisition attempts.
	retryDelay time.Duration

	// quorum is the number of Redis servers that must agree on the lock
	// status in order for the lock to be considered valid.
	quorum int

	// servers is a list of Redis servers that the Redlock algorithm will
	// use to manage the lock.
	servers []RedisClient
}

// NewRedlock creates a new Redlock with the given Redis servers and
// configuration options.
func NewRedlock(servers []RedisClient, opts ...func(*Redlock)) *Redlock {
	r := &Redlock{
		ClockDriftFactor:    ClockDriftFactor,
		RedisConnectTimeout: RedisConnectTimeout,
		retryCount:          10,
		retryDelay:          200 * time.Millisecond,
		quorum:              len(servers)/2 + 1,
		servers:             servers,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// LockOptions are the options for acquiring a lock.
type LockOptions struct {
	// Validity is the duration that the lock is valid for.
	Validity time.Duration

	// LockScript is the Redis script used to acquire the lock.
	LockScript string
}

// Lock is a distributed lock.
//type Lock struct {
//	name       string
//	resource   string
//	validity   int64
//	lockScript string
//	servers    []RedisClient
//}

// IsLocked returns true if the lock is still valid.
//func (l *Lock) IsLocked() bool {
//	for _, server := range l.servers {
//		val, err := server.Get(l.name)
//		if err != nil {
//			continue
//		}
//
//		s, ok := val.(string)
//		if !ok {
//			continue
//		}
//
//		if s == l.resource {
//			return true
//		}
//	}
//
//	return false
//}

//
//func setLock(server RedisClient, name string, resource string, validity int64) (bool, error) {
//	return server.Set(name, resource, time.Duration(validity)*time.Millisecond)
//}

//func (r *Redlock) tryLock(name string, resource string, validity int64) (*Lock, error) {
//	var lock *Lock
//	var mu sync.Mutex
//	var wg sync.WaitGroup
//	wg.Add(len(r.servers))
//
//	for _, server := range r.servers {
//		go func(server RedisClient) {
//			defer wg.Done()
//			locked, err := setLock(server, name, resource, validity)
//			if err != nil {
//				return
//			}
//
//			if !locked {
//				return
//			}
//
//			mu.Lock()
//			defer mu.Unlock()
//
//			if lock != nil {
//				return
//			}
//
//			lock = &Lock{
//				name:       name,
//				resource:   resource,
//				validity:   validity,
//				lockScript: unlockScript,
//				servers:    r.servers,
//			}
//		}(server)
//	}
//
//	wg.Wait()
//
//	if lock == nil {
//		return nil, errors.New("failed to acquire lock")
//	}
//
//	return lock, nil
//}

// Lock attempts to acquire a lock with the given name and options.
func (r *Redlock) Lock(name string, opts ...func(*LockOptions)) (*Lock, error) {
	options := &LockOptions{
		Validity:   time.Second,
		LockScript: "if redis.call('setnx', KEYS[1], ARGV[1]) == 1 then return redis.call('pexpire', KEYS[1], ARGV[2]) else return 0 end",
	}

	for _, opt := range opts {
		opt(options)
	}

	uuid := generateUUID()
	lock := &Lock{
		Name:      name,
		Value:     uuid,
		Validity:  options.Validity,
		Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
	}

	for i := 0; i < r.retryCount; i++ {
		startTime := time.Now()

		n := 0
		for _, server := range r.servers {
			ok, err := lockAcquired(server, lock, options)
			if err != nil {
				continue
			}
			if ok {
				n++
			}
		}

		if n >= r.quorum {
			return lock, nil
		}

		drift := float64(time.Now().UnixNano()-startTime.UnixNano()) / float64(time.Millisecond) / float64(options.Validity/time.Millisecond)
		if drift > r.ClockDriftFactor {
			return nil, errors.New("clock drift is too high")
		}

		time.Sleep(r.retryDelay)
	}

	return nil, errors.New("unable to acquire lock")
}

func lockAcquired(client RedisClient, lock *Lock, options *LockOptions) (bool, error) {
	value, err := client.Eval(options.LockScript, []string{lock.Name}, lock.Value, int(options.Validity/time.Millisecond))
	if err != nil {
		return false, err
	}
	acquired, ok := value.(int64)
	if !ok {
		return false, errors.New("invalid response from Redis")
	}
	return acquired == 1, nil
}

// Lock is a distributed lock.
type Lock struct {
	// Name is the name of the lock.
	Name string

	// Value is the value of the lock.
	Value string

	// Validity is the duration that the lock will be held for.
	Validity time.Duration

	// Timestamp is the timestamp of when the lock was acquired.
	Timestamp int64
}

// Unlock releases the lock.
func (r *Redlock) Unlock(lock *Lock) error {
	n := 0
	for _, server := range r.servers {
		ok, err := lockReleased(server, lock)
		if err != nil {
			continue
		}
		if ok {
			n++
		}
	}

	if n >= r.quorum {
		return nil
	}

	return errors.New("unable to release lock")
}

func lockReleased(client RedisClient, lock *Lock) (bool, error) {
	value, err := client.Eval("if redis.call('get', KEYS[1]) == ARGV[1] then return redis.call('del', KEYS[1]) else return 0 end", []string{lock.Name}, lock.Value)
	if err != nil {
		return false, err
	}
	released, ok := value.(int64)
	if !ok {
		return false, errors.New("invalid response from Redis")
	}
	return released == 1, nil
}

// generateUUID generates a unique identifier (UUID).
func generateUUID() string {
	uuid := atomic.AddInt64(&uuidCounter, 1)
	return time.Now().Format("20060102150405.000000") + "-" + strconv.FormatInt(uuid, 10)
}

func main() {
	// Connect to Redis servers.
	var servers []RedisClient

	// Create a Redlock instance.
	redlock := NewRedlock(servers)

	// Try to acquire a lock.
	lock, err := redlock.Lock("my-lock", func(options *LockOptions) {
		options.Validity = time.Minute
	})
	if err != nil {
		fmt.Println("unable to acquire lock:", err)
		return
	}
	fmt.Println("lock acquired")

	// Do something with the lock.
	time.Sleep(time.Second)

	// Release the lock.
	if err := redlock.Unlock(lock); err != nil {
		fmt.Println("unable to release lock:", err)
		return
	}
	fmt.Println("lock released")
}
