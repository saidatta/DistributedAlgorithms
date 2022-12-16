package chapter5_deadlocks

import "sync"

// lock stealing, lock timeouts, lock escalation
//Lock escalation is a process that automatically converts many fine-grained locks into fewer, coarser-grained locks.
//	This is done to improve performance by reducing lock contention and the amount of memory used to store locks.
//	In a database system, for example, lock escalation may be triggered when a large number of locks are held on a
//single page or table, or when a transaction has held locks for a long time. When lock escalation occurs, the existing
//locks are released and a single, more general lock is acquired in its place. This can help to reduce the likelihood of
//deadlocks, which can occur when two transactions are waiting for locks held by each other.

//Imagine that a transaction has been running for a while and has acquired many locks on individual rows in a table.
//This can happen when the transaction performs a lot of updates, inserts, or deletes on the table. As the transaction
//continues to run, it may acquire more and more locks, which can consume a significant amount of memory and cause lock
//contention with other transactions.
//
//To improve performance and reduce the chances of deadlocks, the database system may perform lock escalation. In this
//case, the existing locks on the individual rows would be released, and a single, coarser-grained lock would be
//acquired on the entire table. This would free up memory and reduce the likelihood of lock contention, allowing the
//whatransaction to continue running more efficiently.

//A coarser-grained lock is a lock that applies to a larger and broader resource than a fine-grained lock. In a database system,
//for example, a coarser-grained lock might be acquired on an entire table, while a fine-grained lock might be acquired
//on a single row in the table. Coarser-grained locks are generally less resource-intensive and can help to improve
//performance by reducing lock contention and the likelihood of deadlocks.
//
//In general, coarser-grained locks are less restrictive than fine-grained locks. Because they apply to a larger resource,
//they allow multiple transactions to access the resource concurrently, as long as they are not trying to modify the same
//data. This can help to improve concurrency and allow more transactions to be processed simultaneously.
//
//However, coarser-grained locks can also cause problems if they are not used carefully. Because they allow multiple
//transactions to access the same resource concurrently, they can lead to inconsistencies if the transactions are
//modifying the same data. For this reason, it is important to use coarser-grained locks carefully and only when they
//will improve performance without causing problems.

// Resource is a type that represents a shared resource that can be locked
type Resource struct {
	sync.RWMutex
}

// Lock is a type that represents a lock held on a resource
type Lock struct {
	resource *Resource
	shared   bool
}

// acquireLock tries to acquire a lock on a resource, escalating locks as necessary
func acquireLock(resource *Resource) *Lock {
	// Try to acquire a shared lock on the resource
	lock := acquireSharedLock(resource)
	if lock == nil {
		// If we can't acquire a shared lock, try to acquire an exclusive lock
		lock = acquireExclusiveLock(resource)
		if lock == nil {
			// If we can't acquire an exclusive lock, escalate the locks held by other processes
			escalateLocks()
			// Try to acquire the exclusive lock again
			lock = acquireExclusiveLock(resource)
		}
	}
	return lock
}

// acquireSharedLock tries to acquire a shared lock on a resource
func acquireSharedLock(resource *Resource) *Lock {
	resource.RLock()
	return &Lock{resource: resource, shared: true}
}

// acquireExclusiveLock tries to acquire an exclusive lock on a resource
func acquireExclusiveLock(resource *Resource) *Lock {
	resource.Lock()
	return &Lock{resource: resource, shared: false}
}

// escalateLocks escalates the locks held by other processes
func escalateLocks() {
	// TODO: implement lock escalation
}

// release unlocks a resource that was previously locked
func (lock *Lock) release() {
	if lock.shared {
		lock.resource.RUnlock()
	} else {
		lock.resource.Unlock()
	}
}
