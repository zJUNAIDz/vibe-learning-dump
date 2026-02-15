package final

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestEnqueue(t *testing.T) {
	jq, err := NewJobQueue(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer jq.Close()

	ctx := context.Background()
	job := &Job{
		ID:       "job1",
		Priority: MediumPriority,
		Payload:  []byte("test"),
		Handler:  func(payload []byte) error { return nil },
	}

	if err := jq.Enqueue(ctx, job); err != nil {
		t.Errorf("Enqueue failed: %v", err)
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	if status := jq.Status("job1"); status != StatusCompleted {
		t.Errorf("Expected completed, got %s", status)
	}
}

func TestPriority(t *testing.T) {
	jq, err := NewJobQueue(Config{
		NumWorkers: 1, // Single worker to test priority order
		QueueSize:  100,
		MaxRetries: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer jq.Close()

	var mu sync.Mutex
	var order []string

	handler := func(id string) func([]byte) error {
		return func(payload []byte) error {
			mu.Lock()
			order = append(order, id)
			mu.Unlock()
			time.Sleep(10 * time.Millisecond) // Slow processing
			return nil
		}
	}

	ctx := context.Background()

	// Enqueue in reverse priority order
	jobs := []*Job{
		{ID: "low1", Priority: LowPriority, Handler: handler("low1")},
		{ID: "med1", Priority: MediumPriority, Handler: handler("med1")},
		{ID: "high1", Priority: HighPriority, Handler: handler("high1")},
	}

	for _, job := range jobs {
		jq.Enqueue(ctx, job)
	}

	// Wait for all to complete
	time.Sleep(200 * time.Millisecond)

	// High priority should be processed first
	if len(order) < 3 {
		t.Fatalf("Expected 3 jobs processed, got %d", len(order))
	}

	if order[0] != "high1" {
		t.Errorf("Expected high priority first, got %s", order[0])
	}
}

func TestRetry(t *testing.T) {
	jq, err := NewJobQueue(Config{
		NumWorkers: 1,
		QueueSize:  10,
		MaxRetries: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer jq.Close()

	attempts := 0
	var mu sync.Mutex

	job := &Job{
		ID:       "retry-job",
		Priority: HighPriority,
		Handler: func(payload []byte) error {
			mu.Lock()
			attempts++
			mu.Unlock()
			return fmt.Errorf("intentional failure")
		},
	}

	ctx := context.Background()
	jq.Enqueue(ctx, job)

	// Wait for retries
	time.Sleep(10 * time.Second) // Backoff: 1s + 2s + 4s

	mu.Lock()
	finalAttempts := attempts
	mu.Unlock()

	// Should try initial + 2 retries = 3 total
	if finalAttempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", finalAttempts)
	}

	if status := jq.Status("retry-job"); status != StatusFailed {
		t.Errorf("Expected failed status, got %s", status)
	}
}

func TestGracefulShutdown(t *testing.T) {
	jq, err := NewJobQueue(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	completed := make(map[string]bool)
	var mu sync.Mutex

	// Enqueue long-running jobs
	for i := 0; i < 10; i++ {
		jobID := fmt.Sprintf("job%d", i)
		job := &Job{
			ID:       jobID,
			Priority: MediumPriority,
			Handler: func(payload []byte) error {
				time.Sleep(100 * time.Millisecond)
				mu.Lock()
				completed[string(payload)] = true
				mu.Unlock()
				return nil
			},
			Payload: []byte(jobID),
		}
		jq.Enqueue(ctx, job)
	}

	// Give jobs time to start
	time.Sleep(50 * time.Millisecond)

	// Graceful close should wait for in-flight jobs
	if err := jq.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// All jobs should complete
	mu.Lock()
	count := len(completed)
	mu.Unlock()

	if count != 10 {
		t.Errorf("Expected 10 completed jobs, got %d", count)
	}
}

func TestMetrics(t *testing.T) {
	jq, err := NewJobQueue(DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	defer jq.Close()

	ctx := context.Background()

	// Enqueue successful jobs
	for i := 0; i < 5; i++ {
		job := &Job{
			ID:       fmt.Sprintf("success%d", i),
			Priority: MediumPriority,
			Handler:  func([]byte) error { return nil },
		}
		jq.Enqueue(ctx, job)
	}

	// Enqueue failing jobs
	for i := 0; i < 3; i++ {
		job := &Job{
			ID:       fmt.Sprintf("fail%d", i),
			Priority: MediumPriority,
			Handler:  func([]byte) error { return fmt.Errorf("fail") },
		}
		jq.Enqueue(ctx, job)
	}

	// Wait for processing (with retries)
	time.Sleep(15 * time.Second)

	m := jq.Metrics()

	if m.Queued != 8 {
		t.Errorf("Expected 8 queued, got %d", m.Queued)
	}
	if m.Completed != 5 {
		t.Errorf("Expected 5 completed, got %d", m.Completed)
	}
	if m.Failed != 3 {
		t.Errorf("Expected 3 failed, got %d", m.Failed)
	}
}

func TestConcurrent(t *testing.T) {
	jq, err := NewJobQueue(Config{
		NumWorkers: 10,
		QueueSize:  1000,
		MaxRetries: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer jq.Close()

	var wg sync.WaitGroup
	numJobs := 1000

	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			job := &Job{
				ID:       fmt.Sprintf("job%d", id),
				Priority: Priority(id % 3),
				Handler:  func([]byte) error { return nil },
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			jq.Enqueue(ctx, job)
		}(i)
	}

	wg.Wait()

	// Wait for completion
	time.Sleep(2 * time.Second)

	m := jq.Metrics()
	if m.Queued != uint64(numJobs) {
		t.Errorf("Expected %d queued, got %d", numJobs, m.Queued)
	}
}

func BenchmarkEnqueue(b *testing.B) {
	jq, _ := NewJobQueue(DefaultConfig())
	defer jq.Close()

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		job := &Job{
			ID:       fmt.Sprintf("job%d", i),
			Priority: MediumPriority,
			Handler:  func([]byte) error { return nil },
		}
		jq.Enqueue(ctx, job)
	}
}

func BenchmarkProcessing(b *testing.B) {
	jq, _ := NewJobQueue(Config{
		NumWorkers: 8,
		QueueSize:  10000,
		MaxRetries: 0,
	})
	defer jq.Close()

	ctx := context.Background()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			job := &Job{
				ID:       fmt.Sprintf("job%d", i),
				Priority: Priority(i % 3),
				Handler:  func([]byte) error { return nil },
			}
			jq.Enqueue(ctx, job)
			i++
		}
	})
}
