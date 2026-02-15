// Package naive implements a connection pool with INTENTIONAL PROBLEMS.
//
// This implementation demonstrates common mistakes when building resource pools:
// 1. Unbounded Growth - Creates unlimited connections (memory leak)
// 2. No Health Checks - Returns broken connections
// 3. No Timeout - Acquire() can block forever
// 4. Resource Leak - Connections not closed on Release()
// 5. Race Condition - Concurrent map access without mutex
// 6. No Metrics - Can't observe pool behavior
//
// Expected behavior:
// - Light load: Works (lucky!)
// - Heavy load: OOM from unlimited connections
// - Network issues: Returns dead connections, causes errors
//
// HOW TO OBSERVE THESE PROBLEMS:
//
//  1. Unbounded Growth (memory leak):
//     pool := NewPool(factory)
//     for i := 0; i < 10000; i++ {
//     conn, _ := pool.Acquire()
//     // Never Release()
//     }
//     // Result: 10,000 connections created, never cleaned up
//     // Memory usage keeps growing
//
//  2. No Health Checks (broken connections):
//     conn, _ := pool.Acquire()
//     conn.Close() // Simulate network failure
//     pool.Release(conn)
//     conn2, _ := pool.Acquire() // Gets same broken connection!
//     conn2.Query() // FAILS - connection is dead
//
//  3. No Timeout (goroutine leak):
//     pool := NewPool(factory, 1) // Max 1 connection
//     conn1, _ := pool.Acquire()
//     go func() {
//     conn2, _ := pool.Acquire() // Blocks forever!
//     }()
//     // Second goroutine blocked forever
//
//  4. Resource Leak (connections not closed):
//     for i := 0; i < 100; i++ {
//     conn, _ := pool.Acquire()
//     pool.Release(conn)
//     }
//     pool.Close()
//     // Connections never actually closed
//     // Database shows 100 idle connections
//
//  5. Race Condition (concurrent map panic):
//     go pool.Acquire()
//     go pool.Acquire()
//     go pool.Release(conn)
//     // Result: "fatal error: concurrent map writes"
//
//  6. No Metrics (blind operation):
//     // No way to know:
//     // - How many connections active?
//     // - How many waiting?
//     // - Any errors occurring?
package naive

import (
	"errors"
	"fmt"
)

// Conn represents a database connection (simplified).
type Conn interface {
	Query(sql string) error
	Close() error
	IsAlive() bool
}

// Factory creates new connections.
type Factory func() (Conn, error)

// Pool represents a naive connection pool.
//
// PROBLEMS:
// - No size limits (unbounded growth)
// - No health checking
// - No timeout support
// - map without mutex (race condition)
type Pool struct {
	factory Factory
	pool    chan Conn
	conns   map[Conn]bool // RACE: no mutex protection
}

// NewPool creates a new naive pool.
//
// PROBLEM: No configuration options (timeout, min/max size, etc.)
func NewPool(factory Factory) *Pool {
	return &Pool{
		factory: factory,
		pool:    make(chan Conn, 10), // PROBLEM: Fixed buffer, arbitrary size
		conns:   make(map[Conn]bool),
	}
}

// Acquire gets a connection from the pool.
//
// PROBLEMS:
// - No timeout (can block forever)
// - No health check (may return broken connection)
// - Creates unlimited connections if pool empty
// - Race condition on conns map
func (p *Pool) Acquire() (Conn, error) {
	// Try to get from pool
	select {
	case conn := <-p.pool:
		// PROBLEM: No health check!
		// Connection might be closed/broken
		return conn, nil
	default:
		// PROBLEM: Always creates new connection if pool empty
		// No max limit - can create thousands!
		conn, err := p.factory()
		if err != nil {
			return nil, err
		}

		// PROBLEM: Concurrent map write - WILL PANIC!
		p.conns[conn] = true

		fmt.Printf("Created new connection (total: %d)\n", len(p.conns))
		return conn, nil
	}
}

// Release returns a connection to the pool.
//
// PROBLEMS:
// - No validation (could be nil)
// - No health check (could be broken)
// - Connection not closed if pool full (leaked)
// - Race condition on conns map
func (p *Pool) Release(conn Conn) error {
	if conn == nil {
		return errors.New("nil connection")
	}

	// PROBLEM: No health check before returning to pool
	// Broken connections stay in pool

	select {
	case p.pool <- conn:
		// PROBLEM: If pool full, drops connection
		// Should close it!
		return nil
	default:
		// PROBLEM: Connection dropped, never closed
		// Resource leak!
		fmt.Println("Pool full, dropping connection (LEAK)")
		return nil
	}
}

// Close closes the pool.
//
// PROBLEMS:
// - Doesn't wait for in-use connections
// - Doesn't actually close connections
// - Race condition on conns map
func (p *Pool) Close() error {
	// PROBLEM: Just closes channel
	// Doesn't close actual connections!
	close(p.pool)

	// PROBLEM: Connections in use are leaked
	// Should wait for them to be released

	return nil
}

// Stats returns pool statistics.
//
// PROBLEM: Race condition on conns map
func (p *Pool) Stats() (total int, pooled int) {
	// RACE: concurrent map access without mutex
	return len(p.conns), len(p.pool)
}

// Example usage (demonstrates problems):
//
//	func main() {
//		pool := naive.NewPool(func() (Conn, error) {
//			return OpenDBConnection()
//		})
//
//		// PROBLEM 1: Unbounded growth
//		for i := 0; i < 1000; i++ {
//			conn, _ := pool.Acquire()
//			// Never Release() - creates 1000 connections!
//		}
//
//		// PROBLEM 2: Returns broken connections
//		conn, _ := pool.Acquire()
//		conn.Close() // Simulate failure
//		pool.Release(conn)
//		conn2, _ := pool.Acquire() // Gets broken conn!
//		conn2.Query("SELECT 1") // FAILS
//
//		// PROBLEM 3: Blocks forever
//		pool2 := naive.NewPool(factory)
//		conn1, _ := pool2.Acquire()
//		conn2, _ := pool2.Acquire() // Blocks if pool empty and factory slow
//
//		// PROBLEM 4: Concurrent access panics
//		go pool.Acquire()
//		go pool.Acquire()
//		go pool.Release(conn)
//		// Race detector shows: WARNING: DATA RACE
//	}
//
// WHY THIS IS BAD:
//
// 1. Memory leak: Unlimited connection creation
// 2. Broken connections: No health checks
// 3. Goroutine leak: No timeout on Acquire()
// 4. Resource leak: Connections not closed properly
// 5. Crashes: Concurrent map access
// 6. No observability: Can't monitor pool health
//
// REAL-WORLD IMPACT:
// - Database: "too many connections" error
// - Memory: OOM killer kills process
// - Goroutines: Thousands blocked waiting
// - Errors: Random "connection refused" failures
//
// FIX STRATEGY (see improved/):
// 1. Add sync.Mutex for conns map
// 2. Limit max connections (buffered channel)
// 3. Add timeout to Acquire() with context
// 4. Health check connections before reuse
// 5. Close connections properly on Release() if pool full
// 6. Add metrics with atomic counters
