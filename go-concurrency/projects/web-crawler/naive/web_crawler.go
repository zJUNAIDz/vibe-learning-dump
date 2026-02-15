// Package naive implements a web crawler with INTENTIONAL PROBLEMS.
//
// This implementation demonstrates common mistakes when building concurrent crawlers:
// 1. Unbounded Goroutines - No limit on concurrent fetches (can spawn thousands)
// 2. Race Condition - Concurrent map writes cause panic
// 3. No Rate Limiting - Hammers servers (can get banned)
// 4. No Robots.txt - Ignores server crawling rules
// 5. Poor Cancellation - Can't stop crawl gracefully
// 6. Memory Leak - Keeps all URLs in memory forever
//
// Expected behavior:
// - Small sites: Works okay (lucky!)
// - Medium sites: Panics with "concurrent map write"
// - Large sites: Out of memory or too many goroutines
//
// HOW TO OBSERVE THESE PROBLEMS:
//
//  1. Unbounded Goroutines (causes resource exhaustion):
//     crawler := NewCrawler()
//     crawler.Crawl("https://news.ycombinator.com")
//     // Watch goroutine count explode: use runtime.NumGoroutine()
//     // On large sites: thousands of goroutines, system slows down
//
//  2. Race Condition (causes panic):
//     // Run with: go run -race web_crawler.go
//     // Or crawl any site with multiple pages
//     // Result: "fatal error: concurrent map writes"
//
//  3. No Rate Limiting (gets you banned):
//     crawler.Crawl("https://example.com")
//     // Check server logs: hundreds of requests per second
//     // Result: 429 Too Many Requests or IP ban
//
//  4. No Robots.txt (unethical crawling):
//     crawler.Crawl("https://site-with-robots.txt")
//     // Crawls /admin, /private, etc. that robots.txt forbids
//
//  5. Poor Cancellation (wastes resources):
//     go crawler.Crawl("huge-site.com")
//     // No way to stop it
//     // Continues running forever consuming CPU/network
//
//  6. Memory Leak (eventually OOM):
//     for i := 0; i < 100; i++ {
//     crawler.Crawl(fmt.Sprintf("site-%d.com", i))
//     }
//     // visited map grows without bound
//     // Eventually: out of memory
package naive

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Crawler represents a naive web crawler.
//
// PROBLEM: No configuration, no resource limits, no graceful shutdown.
type Crawler struct {
	visited map[string]bool // RACE: concurrent writes without mutex
	results []string        // RACE: concurrent appends without mutex
}

// NewCrawler creates a new naive crawler.
func NewCrawler() *Crawler {
	return &Crawler{
		visited: make(map[string]bool),
		results: make([]string, 0),
	}
}

// Crawl starts crawling from the given URL.
//
// PROBLEMS:
// - Spawns unlimited goroutines (one per link)
// - No rate limiting per domain
// - Concurrent map access causes panic
// - No robots.txt checking
// - No cancellation support
func (c *Crawler) Crawl(startURL string) []string {
	// PROBLEM: No validation of startURL
	c.crawlURL(startURL)

	// PROBLEM: Returns immediately, but goroutines still running
	// Race condition: results slice being modified by goroutines
	return c.results
}

// crawlURL fetches and processes a single URL.
func (c *Crawler) crawlURL(urlStr string) {
	// PROBLEM: Concurrent map read without mutex
	if c.visited[urlStr] {
		return
	}

	// PROBLEM: Concurrent map write - WILL PANIC!
	c.visited[urlStr] = true

	fmt.Printf("Crawling: %s\n", urlStr)

	// PROBLEM: No timeout, can hang forever
	resp, err := http.Get(urlStr)
	if err != nil {
		fmt.Printf("Error fetching %s: %v\n", urlStr, err)
		return
	}
	defer resp.Body.Close()

	// PROBLEM: No size limit, can download gigabytes
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading %s: %v\n", urlStr, err)
		return
	}

	// PROBLEM: Concurrent append to slice - RACE!
	c.results = append(c.results, urlStr)

	// Extract links (very naive parsing)
	links := c.extractLinks(string(body), urlStr)

	// PROBLEM: Spawn goroutine for EVERY link
	// No limit! Can easily spawn 10,000+ goroutines
	for _, link := range links {
		// PROBLEM: Each link spawns another goroutine
		// Creates exponential goroutine explosion
		go c.crawlURL(link)
	}

	// PROBLEM: Function returns immediately
	// Parent has no way to know when children finish
}

// extractLinks extracts links from HTML (very naive).
//
// PROBLEM: Doesn't parse HTML properly, just looks for href=
// Will miss many links and include false positives
func (c *Crawler) extractLinks(body, baseURL string) []string {
	var links []string

	// PROBLEM: Terrible parsing - just string search
	parts := strings.Split(body, "href=\"")
	for i := 1; i < len(parts); i++ {
		end := strings.Index(parts[i], "\"")
		if end == -1 {
			continue
		}

		link := parts[i][:end]

		// Skip anchors and javascript
		if strings.HasPrefix(link, "#") || strings.HasPrefix(link, "javascript:") {
			continue
		}

		// Resolve relative URLs
		absoluteURL := c.resolveURL(link, baseURL)
		if absoluteURL != "" {
			links = append(links, absoluteURL)
		}
	}

	return links
}

// resolveURL converts relative URLs to absolute.
func (c *Crawler) resolveURL(link, base string) string {
	// Absolute URL
	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		return link
	}

	baseURL, err := url.Parse(base)
	if err != nil {
		return ""
	}

	linkURL, err := url.Parse(link)
	if err != nil {
		return ""
	}

	return baseURL.ResolveReference(linkURL).String()
}

// GetResults returns crawled URLs.
//
// PROBLEM: No mutex, concurrent access while goroutines modify results
func (c *Crawler) GetResults() []string {
	// RACE: results being modified by goroutines
	return c.results
}

// GetStats returns crawl statistics.
//
// PROBLEM: No mutex, race condition on visited map
func (c *Crawler) GetStats() (int, int) {
	// RACE: visited and results being modified concurrently
	return len(c.visited), len(c.results)
}

// Example usage (will panic on most sites):
//
//	func main() {
//		crawler := naive.NewCrawler()
//
//		// This will panic with "concurrent map writes"
//		results := crawler.Crawl("https://example.com")
//
//		// This will show wrong counts (race condition)
//		visited, found := crawler.GetStats()
//		fmt.Printf("Visited: %d, Found: %d\n", visited, found)
//
//		// This will show incomplete results
//		// (goroutines still running)
//		fmt.Printf("URLs: %v\n", results)
//	}
//
// WHY THIS IS BAD:
//
// 1. Crashes on any site with multiple links (concurrent map panic)
// 2. Exhausts system resources (thousands of goroutines)
// 3. Gets banned by servers (no rate limiting)
// 4. Violates robots.txt (unethical)
// 5. Can't be stopped gracefully
// 6. Memory grows without bound
// 7. Race detector shows dozens of races
//
// PERFORMANCE:
// - Small sites (<10 pages): ~100 pages/sec (before panic)
// - Medium sites: Crashes within seconds
// - Large sites: Never completes
//
// FIX STRATEGY (see improved/):
// 1. Add sync.Mutex for visited map
// 2. Limit goroutines with worker pool (e.g., 10 workers)
// 3. Add per-domain rate limiting (1 req/sec)
// 4. Use sync.WaitGroup to wait for completion
// 5. Add context.Context for cancellation
// 6. Implement proper HTML parsing
