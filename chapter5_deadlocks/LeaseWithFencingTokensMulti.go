package chapter5_deadlocks

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type ThreadSafeLease struct {
	mu sync.RWMutex

	// holder is the current holder of the lease.
	holder string

	// token is a unique token that is assigned to the holder of the lease.
	// The token can be used to prove that the holder is the current holder of the lease.
	token string

	// ttl is the time-to-live for the lease. After the lease expires, it can be acquired by another holder.
	ttl time.Duration

	// expiration is the time at which the lease expires.
	expiration time.Time
}

// Acquire acquires the lease and returns a unique token that can be used to prove that the caller is the holder of the lease.
// If the lease is already held by another holder, this function will return an error.
func (l *ThreadSafeLease) Acquire(holder string) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.holder != "" && time.Now().Before(l.expiration) {
		return "", fmt.Errorf("lease already held by %s", l.holder)
	}

	l.holder = holder
	l.token = fmt.Sprintf("%d", time.Now().UnixNano())
	l.expiration = time.Now().Add(l.ttl)

	return l.token, nil
}

// Renew renews the lease and extends the expiration time by the specified time-to-live (TTL).
// The holder must provide the unique token that was returned when the lease was acquired.
// If the token is invalid or the lease has already expired, this function will return an error.
func (l *ThreadSafeLease) Renew(holder, token string, ttl time.Duration) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.holder != holder || l.token != token || time.Now().After(l.expiration) {
		return fmt.Errorf("invalid token or lease has expired")
	}

	l.expiration = time.Now().Add(ttl)

	return nil
}

// Release releases the lease and allows another holder to acquire it.
// The holder must provide the unique token that was returned when the lease was acquired.
// If the token is invalid or the lease has already expired, this function will return an error.
func (l *ThreadSafeLease) Release(holder, token string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.holder != holder || l.token != token || time.Now().After(l.expiration) {
		return fmt.Errorf("invalid token or lease has expired")
	}

	l.holder = ""
	l.token = ""
	l.expiration = time.Time{}

	return nil
}

// IsHeld returns true if the lease is currently held by a holder, and false otherwise.
func (l *ThreadSafeLease) IsHeld() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.holder != "" && time.Now().Before(l.expiration)
}

// Holder returns the current holder of the lease. If the lease is not currently held, this function will return an empty string.
func (l *ThreadSafeLease) Holder() string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.holder == "" || time.Now().After(l.expiration) {
		return ""
	}

	return l.holder
}

// Token returns the unique token that is assigned to the holder of the lease. If the lease is not currently held, this function will return an empty string.
func (l *ThreadSafeLease) Token() string {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if l.holder == "" || time.Now().After(l.expiration) {
		return ""
	}

	return l.token
}

// TTLExpired returns true if the lease has expired, and false otherwise.
func (l *ThreadSafeLease) TTLExpired() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return time.Now().After(l.expiration)
}

// RemainingTTL returns the remaining time-to-live for the lease. If the lease has expired, this function will return a negative value.
func (l *ThreadSafeLease) RemainingTTL() time.Duration {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.expiration.Sub(time.Now())
}

func main() {
	// Create a new lease with a TTL of 5 seconds.
	lease := &ThreadSafeLease{ttl: 5 * time.Second}

	// Start a goroutine that continuously tries to acquire the lease as holder "Alice".
	go func() {
		for {
			token, err := lease.Acquire("Alice")
			if err != nil {
				fmt.Println("Alice failed to acquire lease:", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			fmt.Println("Alice acquired lease with token", token)

			// Sleep for a random duration between 1 and 3 seconds.
			time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)

			// Try to release the lease.
			err = lease.Release("Alice", token)
			if err != nil {
				fmt.Println("Alice failed to release lease:", err)
			} else {
				fmt.Println("Alice released lease")
			}
		}
	}()

	// Start a goroutine that continuously tries to acquire the lease as holder "Bob".
	go func() {
		for {
			token, err := lease.Acquire("Bob")
			if err != nil {
				fmt.Println("Bob failed to acquire lease:", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			fmt.Println("Bob acquired lease with token", token)

			// Sleep for a random duration between 1 and 3 seconds.
			time.Sleep(time.Duration(1+rand.Intn(3)) * time.Second)

			// Try to release the lease.
			err = lease.Release("Bob", token)
			if err != nil {
				fmt.Println("Bob failed to release lease:", err)
			} else {
				fmt.Println("Bob released lease")
			}
		}
	}()

	// Sleep for 20 seconds to allow the goroutines to run.
	time.Sleep(20 * time.Second)
}
