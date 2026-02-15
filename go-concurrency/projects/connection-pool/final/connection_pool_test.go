package final

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// mockConn implements Conn interface for testing.
type mockConn struct {
	id     int
	closed atomic.Bool
	alive  atomic.Bool
}

func newMockConn(id int) *mockConn {
	conn := &mockConn{id: id}
	conn.alive.Store(true)
	return conn
}

func (m *mockConn) Query(sql string) error {
	if m.closed.Load() || !m.alive.Load() {
		return errors.New("connection closed")
	}
	return nil
}

func (m *mockConn) Close() error {
	m.closed.Store(true)
	m.alive.Store(false)
	return nil
}

func (m *mockConn) IsAlive() bool {
	return m.alive.Load() && !m.closed.Load()
}

// TestPool_Basic tests basic pool functionality.
func TestPool_Basic(t *testing.T) {
	var connID atomic.Int32
	factory := func() (Conn, error) {
		return newMockConn(int(connID.Add(1))), nil
	}

	pool, err := NewPool(factory, Config{
		MinConns: 2,
		MaxConns: 5,
	})
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	// Check initial state
	metrics := pool.Metrics()
	if metrics.TotalConns != 2 {
		t.Errorf("Expected 2 initial connections, got %d", metrics.TotalConns)
	}

	// Acquire and release
	ctx := context.Background()
	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("Failed to acquire: %v", err)
	}

	if err := pool.Release(conn); err != nil {
		t.Errorf("Failed to release: %v", err)
	}

	// Verify metrics
	metrics = pool.Metrics()
	if metrics.AcquireCount != 1 {
		t.Errorf("Expected 1 acquire, got %d", metrics.AcquireCount)
	}
	if metrics.ReleaseCount != 1 {
		t.Errorf("Expected 1 release, got %d", metrics.ReleaseCount)
	}
}

// TestPool_MaxConnections tests max connection limit.
func TestPool_MaxConnections(t *testing.T) {
	var connID atomic.Int32
	factory := func() (Conn, error) {
		return newMockConn(int(connID.Add(1))), nil
	}

	pool, err := NewPool(factory, Config{
		MinConns: 1,
		MaxConns: 3,
	})
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Acquire max connections
	var conns []Conn
	for i := 0; i < 3; i++ {
		conn, err := pool.Acquire(ctx)
		if err != nil {
			t.Fatalf("Failed to acquire connection %d: %v", i, err)
		}
		conns = append(conns, conn)
	}

	// Try to acquire one more (should timeout)
	ctx2, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = pool.Acquire(ctx2)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected timeout, got %v", err)
	}

	// Release one
	if err := pool.Release(conns[0]); err != nil {
		t.Fatalf("Failed to release: %v", err)
	}

	// Now should be able to acquire
	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Errorf("Failed to acquire after release: %v", err)
	}

	// Cleanup
	pool.Release(conn)
	for i := 1; i < len(conns); i++ {
		pool.Release(conns[i])
	}
}

// TestPool_HealthCheck tests connection validation.
func TestPool_HealthCheck(t *testing.T) {
	var connID atomic.Int32
	factory := func() (Conn, error) {
		return newMockConn(int(connID.Add(1))), nil
	}

	pool, err := NewPool(factory, Config{
		MinConns:    1,
		MaxConns:    3,
		IdleTimeout: 100 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Acquire connection
	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("Failed to acquire: %v", err)
	}

	mockConn := conn.(*mockConn)

	// Mark as dead
	mockConn.alive.Store(false)

	// Release (should close)
	if err := pool.Release(conn); err != nil {
		t.Fatalf("Failed to release: %v", err)
	}

	// Acquire new connection (should get fresh one)
	conn2, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("Failed to acquire after dead conn: %v", err)
	}

	// Should be different connection
	mockConn2 := conn2.(*mockConn)
	if mockConn.id == mockConn2.id {
		t.Error("Expected new connection after dead one")
	}

	pool.Release(conn2)
}

// TestPool_IdleTimeout tests idle connection cleanup.
func TestPool_IdleTimeout(t *testing.T) {
	var connID atomic.Int32
	factory := func() (Conn, error) {
		return newMockConn(int(connID.Add(1))), nil
	}

	pool, err := NewPool(factory, Config{
		MinConns:    1,
		MaxConns:    3,
		IdleTimeout: 200 * time.Millisecond,
		HealthCheck: 100 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Acquire and release
	conn, _ := pool.Acquire(ctx)
	pool.Release(conn)

	// Wait for idle timeout + maintenance
	time.Sleep(400 * time.Millisecond)

	// Metrics should show connection was closed
	metrics := pool.Metrics()
	if metrics.CloseCount == 0 {
		t.Error("Expected idle connections to be closed")
	}

	t.Logf("Closed %d idle connections", metrics.CloseCount)
}

// TestPool_MaxLifetime tests connection lifetime enforcement.
func TestPool_MaxLifetime(t *testing.T) {
	var connID atomic.Int32
	factory := func() (Conn, error) {
		return newMockConn(int(connID.Add(1))), nil
	}

	pool, err := NewPool(factory, Config{
		MinConns:    1,
		MaxConns:    3,
		MaxLifetime: 200 * time.Millisecond,
		HealthCheck: 100 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Acquire first connection
	conn1, _ := pool.Acquire(ctx)
	firstID := conn1.(*mockConn).id
	pool.Release(conn1)

	// Wait for max lifetime
	time.Sleep(300 * time.Millisecond)

	// Acquire again (should get new connection)
	conn2, _ := pool.Acquire(ctx)
	secondID := conn2.(*mockConn).id
	pool.Release(conn2)

	if firstID == secondID {
		t.Error("Expected new connection after max lifetime")
	}
}

// TestPool_CircuitBreaker tests circuit breaker functionality.
func TestPool_CircuitBreaker(t *testing.T) {
	var attempts atomic.Int32
	factory := func() (Conn, error) {
		attempts.Add(1)
		return nil, errors.New("factory always fails")
	}

	pool, err := NewPool(factory, Config{
		MinConns: 0, // Allow 0 min to avoid initial failure
		MaxConns: 3,
	})
	if err == nil {
		defer pool.Close()
		t.Fatal("Expected pool creation to fail with 0 min conns and failing factory")
	}

	// Create pool that will fail on acquire
	factory2 := func() (Conn, error) {
		if attempts.Load() < 5 {
			attempts.Add(1)
			return nil, errors.New("factory fails")
		}
		return newMockConn(1), nil
	}

	pool2, _ := NewPool(func() (Conn, error) { return newMockConn(1), nil }, Config{
		MinConns: 1,
		MaxConns: 3,
	})
	defer pool2.Close()

	// Replace factory (hack for testing)
	pool2.factory = factory2
	attempts.Store(0)

	ctx := context.Background()

	// Try multiple acquires (should trip circuit breaker)
	for i := 0; i < 6; i++ {
		pool2.createConnection()
		time.Sleep(10 * time.Millisecond)
	}

	// Circuit should be open now
	metrics := pool2.Metrics()
	if !metrics.CircuitOpen {
		t.Error("Expected circuit breaker to be open")
	}

	_, err = pool2.Acquire(ctx)
	if err == nil || err.Error() != "circuit breaker open" {
		t.Errorf("Expected circuit breaker error, got %v", err)
	}

	t.Logf("Circuit breaker tripped after %d attempts", attempts.Load())
}

// TestPool_Concurrent tests concurrent access.
func TestPool_Concurrent(t *testing.T) {
	var connID atomic.Int32
	factory := func() (Conn, error) {
		time.Sleep(time.Millisecond) // Simulate latency
		return newMockConn(int(connID.Add(1))), nil
	}

	pool, err := NewPool(factory, Config{
		MinConns: 2,
		MaxConns: 10,
	})
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Spawn 100 goroutines
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < 10; j++ {
				conn, err := pool.Acquire(ctx)
				if err != nil {
					errors <- err
					return
				}

				// Simulate work
				time.Sleep(time.Millisecond)

				if err := conn.Query("SELECT 1"); err != nil {
					errors <- err
				}

				if err := pool.Release(conn); err != nil {
					errors <- err
				}
			}
		}()
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent error: %v", err)
	}

	// Verify metrics
	metrics := pool.Metrics()
	if metrics.AcquireCount != 1000 {
		t.Errorf("Expected 1000 acquires, got %d", metrics.AcquireCount)
	}
	if metrics.ReleaseCount != 1000 {
		t.Errorf("Expected 1000 releases, got %d", metrics.ReleaseCount)
	}

	t.Logf("Metrics: %+v", metrics)
}

// TestPool_GracefulShutdown tests graceful shutdown.
func TestPool_GracefulShutdown(t *testing.T) {
	var connID atomic.Int32
	factory := func() (Conn, error) {
		return newMockConn(int(connID.Add(1))), nil
	}

	pool, err := NewPool(factory, Config{
		MinConns: 2,
		MaxConns: 5,
	})
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	ctx := context.Background()

	// Acquire some connections
	conn1, _ := pool.Acquire(ctx)
	conn2, _ := pool.Acquire(ctx)

	// Start long operation
	done := make(chan struct{})
	go func() {
		time.Sleep(100 * time.Millisecond)
		pool.Release(conn1)
		pool.Release(conn2)
		close(done)
	}()

	// Close pool (should wait)
	start := time.Now()
	pool.Close()
	elapsed := time.Since(start)

	// Should have waited for operations
	if elapsed < 100*time.Millisecond {
		t.Errorf("Close() didn't wait for in-flight operations: %v", elapsed)
	}

	<-done

	// Further acquires should fail
	_, err = pool.Acquire(ctx)
	if err == nil {
		t.Error("Expected error after close")
	}
}

// BenchmarkPool_Acquire benchmarks acquire/release performance.
func BenchmarkPool_Acquire(b *testing.B) {
	factory := func() (Conn, error) {
		return newMockConn(1), nil
	}

	pool, _ := NewPool(factory, Config{
		MinConns: 10,
		MaxConns: 50,
	})
	defer pool.Close()

	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		conn, err := pool.Acquire(ctx)
		if err != nil {
			b.Fatal(err)
		}
		pool.Release(conn)
	}
}

// BenchmarkPool_Parallel benchmarks parallel acquire/release.
func BenchmarkPool_Parallel(b *testing.B) {
	factory := func() (Conn, error) {
		return newMockConn(1), nil
	}

	pool, _ := NewPool(factory, Config{
		MinConns: 10,
		MaxConns: 100,
	})
	defer pool.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		ctx := context.Background()
		for pb.Next() {
			conn, err := pool.Acquire(ctx)
			if err != nil {
				b.Fatal(err)
			}
			conn.Query("SELECT 1")
			pool.Release(conn)
		}
	})
}

// Example demonstrates pool usage.
func ExamplePool() {
	// Create pool
	pool, err := NewPool(
		func() (Conn, error) {
			// Your connection factory
			return newMockConn(1), nil
		},
		Config{
			MinConns:    5,
			MaxConns:    20,
			IdleTimeout: 5 * time.Minute,
			MaxLifetime: 30 * time.Minute,
		},
	)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	// Acquire connection
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		panic(err)
	}
	defer pool.Release(conn)

	// Use connection
	if err := conn.Query("SELECT * FROM users"); err != nil {
		panic(err)
	}

	// Check metrics
	metrics := pool.Metrics()
	fmt.Printf("Pool: %d total, %d in use, %d idle\n",
		metrics.TotalConns, metrics.InUseConns, metrics.IdleConns)
}
