package chapter5_deadlocks

import (
	"fmt"
	"sync"
	"time"
)

// Lease represents a lock that can be held by a single holder at a time.
type Lease struct {
	mu sync.Mutex

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
func (l *Lease) Acquire(holder string) (string, error) {
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
func (l *Lease) Renew(holder, token string, ttl time.Duration) error {
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
func (l *Lease) Release(holder, token string) error {
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

func main() {
	// Create a new lease with a TTL of 5 seconds.
	lease := &Lease{ttl: 5 * time.Second}

	// Try to acquire the lease as holder "Alice".
	token, err := lease.Acquire("Alice")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Lease acquired by Alice with token", token)

	// Try to renew the lease as holder "Bob". This should fail because "Bob" is not the current holder.
	err = lease.Renew("Bob", token, 5*time.Second)
	if err != nil {
		fmt.Println(err)
	}

	// Try to renew the lease as holder "Alice" with the correct token.
	err = lease.Renew("Alice", token, 5*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Lease renewed by Alice")

	// Wait for the lease to expire.
	time.Sleep(6 * time.Second)

	// Try to renew the lease as holder "Alice". This should fail because the lease has expired.
	err = lease.Renew("Alice", token, 5*time.Second)
	if err != nil {
		fmt.Println(err)
	}

	// Try to release the lease as holder "Bob". This should fail because "Bob" is not the current holder.
	err = lease.Release("Bob", token)
	if err != nil {
		fmt.Println(err)
	}

	// Try to release the lease as holder "Alice" with the correct token.
	err = lease.Release("Alice", token)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Lease released by Alice")
}
