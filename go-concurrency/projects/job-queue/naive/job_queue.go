package naive

import (
	"fmt"
	"sync"
)

// JobQueue implements a naive worker pool with intentional issues:
// 1. No priorities (FIFO only)
// 2. Unbounded queue (can OOM)
// 3. No retry logic (failed jobs lost)
// 4. Poor shutdown (may lose in-flight jobs)
// 5. No persistence (all in-memory)
//
// Expected throughput: ~10k jobs/sec (bottleneck: single channel)
type JobQueue struct {
	jobs    chan Job // ❌ PROBLEM: Unbounded, no priorities
	workers []*worker
	wg      sync.WaitGroup
	done    chan struct{}
}

// Job represents a unit of work
type Job struct {
	ID      string
	Handler func() error // Function to execute
}

// worker processes jobs from the queue
type worker struct {
	id int
	jq *JobQueue
}

// NewJobQueue creates a naive job queue
func NewJobQueue(numWorkers int) *JobQueue {
	jq := &JobQueue{
		jobs:    make(chan Job, 100), // ❌ PROBLEM: Small buffer, can block
		workers: make([]*worker, numWorkers),
		done:    make(chan struct{}),
	}

	// Start workers
	for i := 0; i < numWorkers; i++ {
		jq.workers[i] = &worker{id: i, jq: jq}
		jq.wg.Add(1)
		go jq.workers[i].run()
	}

	return jq
}

// Enqueue adds a job to the queue
// ❌ PROBLEM: Blocks forever if buffer full (no context, no timeout)
func (jq *JobQueue) Enqueue(job Job) {
	jq.jobs <- job
}

// Close stops all workers
// ❌ PROBLEM: Doesn't wait for in-flight jobs to complete
func (jq *JobQueue) Close() {
	close(jq.done)
	close(jq.jobs) // May panic if Enqueue is called after
	jq.wg.Wait()
}

// run processes jobs until done
func (w *worker) run() {
	defer w.jq.wg.Done()

	for {
		select {
		case job, ok := <-w.jq.jobs:
			if !ok {
				return // Channel closed
			}

			// Execute job
			if err := job.Handler(); err != nil {
				// ❌ PROBLEM: Error ignored, no retry
				fmt.Printf("Worker %d: Job %s failed: %v\n", w.id, job.ID, err)
			}

		case <-w.jq.done:
			return // ❌ PROBLEM: May have jobs in channel still
		}
	}
}

// ❌ MAJOR PROBLEMS SUMMARY:
//
// 1. NO PRIORITIES
//    - Single FIFO queue processes all jobs equally
//    - Critical jobs wait behind low-priority jobs
//    - Can't prioritize urgent work
//
//    Example failure:
//      Enqueue 1000 low-priority jobs
//      Enqueue 1 critical job
//      Critical job waits for 1000 jobs to complete (minutes)
//
// 2. UNBOUNDED MEMORY GROWTH
//    - Channel buffer is only 100
//    - Fast producer + slow workers → Enqueue blocks
//    - If buffer was larger, could OOM
//
//    Example failure:
//      Producer: 1000 jobs/sec
//      Workers: 100 jobs/sec (CPU-bound)
//      Buffer fills in 0.1 seconds
//      Producer blocks, system hangs
//
// 3. NO RETRY LOGIC
//    - Failed jobs are lost forever
//    - No exponential backoff
//    - No way to track failures
//
//    Example failure:
//      Job sends HTTP request
//      Network blip causes failure
//      Job never retries, data lost
//
// 4. POOR SHUTDOWN
//    - Close() closes channels immediately
//    - Workers may be processing jobs
//    - In-flight jobs might not complete
//
//    Example failure:
//      100 jobs in channel
//      Close() called
//      Workers see done signal, exit immediately
//      100 jobs lost
//
// 5. NO PERSISTENCE
//    - All jobs in memory only
//    - Process crash = all jobs lost
//    - No recovery mechanism
//
//    Example failure:
//      Enqueue 1M jobs
//      Server crashes (OOM, deploy, etc.)
//      1M jobs lost
//
// 6. NO OBSERVABILITY
//    - Can't tell how many jobs queued
//    - Can't tell success vs failure rate
//    - No metrics for monitoring
//
// 7. NO BACKPRESSURE
//    - Enqueue blocks on full channel
//    - No way to check if queue is full
//    - No way to reject jobs gracefully
//
// HOW TO OBSERVE THESE PROBLEMS:
//
// 1. Priority problem:
//    for i := 0; i < 1000; i++ {
//        jq.Enqueue(LowPriorityJob)
//    }
//    jq.Enqueue(CriticalJob) // Waits for 1000 jobs
//
// 2. Memory problem:
//    // Fast producer
//    for i := 0; i < 10000; i++ {
//        jq.Enqueue(job) // Blocks after 100
//    }
//
// 3. No retry:
//    jq.Enqueue(Job{Handler: func() error {
//        return errors.New("fail") // Lost forever
//    }})
//
// 4. Poor shutdown:
//    for i := 0; i < 1000; i++ {
//        jq.Enqueue(job)
//    }
//    jq.Close() // May lose jobs
//
// FIXES IN improved/job_queue.go:
// - Separate channels for High/Medium/Low priority
// - Bounded queues with context for backpressure
// - Retry logic with exponential backoff
// - Graceful shutdown (wait for in-flight)
// - Metrics for observability
//
// FIXES IN final/job_queue.go:
// - Persistence to disk (at-least-once delivery)
// - Anti-starvation (token-based fairness)
// - Job status tracking by ID
// - Comprehensive tests and benchmarks
