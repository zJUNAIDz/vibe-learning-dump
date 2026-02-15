package final

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestCrawler_Basic tests basic crawling functionality.
func TestCrawler_Basic(t *testing.T) {
	// Create test server
	visited := make(map[string]bool)
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		visited[r.URL.Path] = true
		mu.Unlock()

		w.Header().Set("Content-Type", "text/html")

		switch r.URL.Path {
		case "/":
			fmt.Fprintln(w, `<html><body>
				<a href="/page1">Page 1</a>
				<a href="/page2">Page 2</a>
			</body></html>`)
		case "/page1":
			fmt.Fprintln(w, `<html><body>
				<a href="/page3">Page 3</a>
			</body></html>`)
		case "/page2", "/page3":
			fmt.Fprintln(w, `<html><body>End</body></html>`)
		}
	}))
	defer server.Close()

	// Create crawler
	crawler := NewCrawler(Config{
		MaxWorkers: 2,
		MaxDepth:   2,
		Timeout:    time.Second,
	})

	// Crawl
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := crawler.Crawl(ctx, server.URL)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Crawl failed: %v", err)
	}

	// Verify results
	metrics := crawler.Metrics()
	if metrics.Visited < 3 {
		t.Errorf("Expected at least 3 visited URLs, got %d", metrics.Visited)
	}

	if metrics.Fetched < 3 {
		t.Errorf("Expected at least 3 fetched pages, got %d", metrics.Fetched)
	}
}

// TestCrawler_DepthLimit tests depth limiting.
func TestCrawler_DepthLimit(t *testing.T) {
	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "text/html")

		// Each page links to next
		page := strings.TrimPrefix(r.URL.Path, "/page")
		if page == "" || page == "/" {
			page = "0"
		}

		nextPage := "/page" + page + "a"
		fmt.Fprintf(w, `<html><body><a href="%s">Next</a></body></html>`, nextPage)
	}))
	defer server.Close()

	// Crawl with max depth 2
	crawler := NewCrawler(Config{
		MaxWorkers: 1,
		MaxDepth:   2,
		Timeout:    time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	crawler.Crawl(ctx, server.URL)

	// Should visit: / (depth 0), /page0a (depth 1)
	// Should NOT visit: /page0aa (depth 2, at limit)
	metrics := crawler.Metrics()
	if metrics.Visited > 3 {
		t.Errorf("Depth limit not enforced: visited %d URLs (expected ≤3)", metrics.Visited)
	}
}

// TestCrawler_RateLimiting tests per-domain rate limiting.
func TestCrawler_RateLimiting(t *testing.T) {
	var requestTimes []time.Time
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requestTimes = append(requestTimes, time.Now())
		mu.Unlock()

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, `<html><body>Page</body></html>`)
	}))
	defer server.Close()

	// Create crawler with 500ms rate limit
	crawler := NewCrawler(Config{
		MaxWorkers: 10, // Many workers, but rate limit should control
		MaxDepth:   0,  // Only root
		RateLimit:  500 * time.Millisecond,
		Timeout:    time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	crawler.Crawl(ctx, server.URL)
	elapsed := time.Since(start)

	mu.Lock()
	count := len(requestTimes)

	// Check spacing between requests
	for i := 1; i < len(requestTimes); i++ {
		gap := requestTimes[i].Sub(requestTimes[i-1])
		if gap < 400*time.Millisecond { // Allow 100ms slack
			t.Errorf("Requests too close: %v (expected ≥500ms)", gap)
		}
	}
	mu.Unlock()

	t.Logf("Crawled %d pages in %v (rate limited)", count, elapsed)
}

// TestCrawler_RobotsTxt tests robots.txt parsing and respect.
func TestCrawler_RobotsTxt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		switch r.URL.Path {
		case "/robots.txt":
			fmt.Fprintln(w, "User-agent: *")
			fmt.Fprintln(w, "Disallow: /admin")
			fmt.Fprintln(w, "Disallow: /private")
		case "/":
			fmt.Fprintln(w, `<html><body>
				<a href="/public">Public</a>
				<a href="/admin">Admin</a>
				<a href="/private">Private</a>
			</body></html>`)
		case "/public":
			fmt.Fprintln(w, `<html><body>Public page</body></html>`)
		case "/admin", "/private":
			t.Errorf("Crawler visited disallowed path: %s", r.URL.Path)
			fmt.Fprintln(w, `<html><body>Should not visit</body></html>`)
		}
	}))
	defer server.Close()

	// Create crawler with robots.txt respect
	crawler := NewCrawler(Config{
		MaxWorkers:    2,
		MaxDepth:      2,
		RespectRobots: true,
		Timeout:       time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	crawler.Crawl(ctx, server.URL)

	metrics := crawler.Metrics()
	if metrics.Blocked == 0 {
		t.Error("Expected some URLs to be blocked by robots.txt")
	}

	t.Logf("Blocked %d URLs by robots.txt", metrics.Blocked)
}

// TestCrawler_Cancellation tests context cancellation.
func TestCrawler_Cancellation(t *testing.T) {
	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		// Slow response
		time.Sleep(100 * time.Millisecond)

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, `<html><body>
			<a href="/page1">1</a>
			<a href="/page2">2</a>
			<a href="/page3">3</a>
		</body></html>`)
	}))
	defer server.Close()

	crawler := NewCrawler(Config{
		MaxWorkers: 2,
		MaxDepth:   3,
		Timeout:    time.Second,
	})

	// Cancel after 500ms
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	start := time.Now()
	crawler.Crawl(ctx, server.URL)
	elapsed := time.Since(start)

	// Should stop quickly after cancellation
	if elapsed > time.Second {
		t.Errorf("Cancellation took too long: %v", elapsed)
	}

	count := requests.Load()
	t.Logf("Processed %d requests before cancellation", count)

	// Verify metrics accessible after cancellation
	metrics := crawler.Metrics()
	if metrics.Visited == 0 {
		t.Error("Expected some visited URLs")
	}
}

// TestCrawler_Metrics tests metrics tracking.
func TestCrawler_Metrics(t *testing.T) {
	var fetchCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetchCount.Add(1)
		w.Header().Set("Content-Type", "text/html")

		switch r.URL.Path {
		case "/":
			fmt.Fprintln(w, `<html><body>
				<a href="/ok">OK</a>
				<a href="/external">External</a>
			</body></html>`)
		case "/ok":
			fmt.Fprintln(w, `<html><body>OK</body></html>`)
		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	crawler := NewCrawler(Config{
		MaxWorkers: 2,
		MaxDepth:   1,
		Timeout:    time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	crawler.Crawl(ctx, server.URL)

	metrics := crawler.Metrics()

	// Should have visited at least root and /ok
	if metrics.Visited < 2 {
		t.Errorf("Expected at least 2 visited, got %d", metrics.Visited)
	}

	// Should have fetched some pages
	if metrics.Fetched == 0 {
		t.Error("Expected some fetched pages")
	}

	// Rate limiting should have occurred
	if metrics.RateLimited == 0 {
		t.Error("Expected some rate limiting")
	}

	t.Logf("Metrics: %+v", metrics)
}

// TestCrawler_Concurrent tests concurrent crawling.
func TestCrawler_Concurrent(t *testing.T) {
	var activeRequests atomic.Int32
	var maxConcurrent atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Track concurrent requests
		active := activeRequests.Add(1)
		defer activeRequests.Add(-1)

		// Update max
		for {
			current := maxConcurrent.Load()
			if active <= current || maxConcurrent.CompareAndSwap(current, active) {
				break
			}
		}

		// Simulate work
		time.Sleep(50 * time.Millisecond)

		w.Header().Set("Content-Type", "text/html")

		// Generate 5 links
		var links strings.Builder
		for i := 0; i < 5; i++ {
			fmt.Fprintf(&links, `<a href="/page%d">Page %d</a>`, i, i)
		}

		fmt.Fprintf(w, `<html><body>%s</body></html>`, links.String())
	}))
	defer server.Close()

	// Create crawler with 5 workers
	crawler := NewCrawler(Config{
		MaxWorkers: 5,
		MaxDepth:   2,
		Timeout:    time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	crawler.Crawl(ctx, server.URL)

	max := maxConcurrent.Load()
	if max > 5 {
		t.Errorf("Max concurrent requests (%d) exceeded worker limit (5)", max)
	}

	if max < 2 {
		t.Error("Expected at least 2 concurrent requests")
	}

	t.Logf("Max concurrent requests: %d", max)
}

// TestCrawler_CircuitBreaker tests circuit breaker for failing domains.
func TestCrawler_CircuitBreaker(t *testing.T) {
	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := requests.Add(1)

		// Always fail
		w.WriteHeader(500)

		// After some failures, should trip circuit breaker
		if count > 15 {
			t.Error("Circuit breaker did not trip (too many requests to failing domain)")
		}
	}))
	defer server.Close()

	crawler := NewCrawler(Config{
		MaxWorkers: 5,
		MaxDepth:   2,
		Timeout:    time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	crawler.Crawl(ctx, server.URL)

	count := requests.Load()
	t.Logf("Total requests before circuit break: %d", count)

	// Should have stopped after ~10 errors
	if count > 20 {
		t.Error("Circuit breaker did not prevent excessive retries")
	}
}

// BenchmarkCrawler benchmarks crawling performance.
func BenchmarkCrawler(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, `<html><body>
			<a href="/page1">1</a>
			<a href="/page2">2</a>
		</body></html>`)
	}))
	defer server.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		crawler := NewCrawler(Config{
			MaxWorkers: 10,
			MaxDepth:   1,
			Timeout:    time.Second,
		})

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		crawler.Crawl(ctx, server.URL)
		cancel()
	}
}

// BenchmarkCrawler_Parallel benchmarks parallel crawling.
func BenchmarkCrawler_Parallel(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, `<html><body>Test</body></html>`)
	}))
	defer server.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			crawler := NewCrawler(Config{
				MaxWorkers: 5,
				MaxDepth:   0,
				Timeout:    time.Second,
			})

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			crawler.Crawl(ctx, server.URL)
			cancel()
		}
	})
}

// Example demonstrates crawler usage.
func ExampleCrawler() {
	crawler := NewCrawler(Config{
		MaxWorkers:    10,
		MaxDepth:      3,
		MaxVisited:    1000,
		Timeout:       10 * time.Second,
		RespectRobots: true,
		RateLimit:     time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Start crawling
	go func() {
		if err := crawler.Crawl(ctx, "https://example.com"); err != nil {
			fmt.Printf("Crawl error: %v\n", err)
		}
	}()

	// Monitor progress
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			metrics := crawler.Metrics()
			fmt.Printf("Progress: visited=%d fetched=%d failed=%d blocked=%d queue=%d workers=%d\n",
				metrics.Visited, metrics.Fetched, metrics.Failed,
				metrics.Blocked, metrics.QueueSize, metrics.ActiveWorkers)
		case <-ctx.Done():
			return
		}
	}
}
