package final

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	cache := NewCache(DefaultConfig())
	defer cache.Close()

	cache.Set("key1", "value1")
	val, ok := cache.Get("key1")
	if !ok || val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}
}

func TestTTL(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DefaultTTL = 100 * time.Millisecond
	cache := NewCache(cfg)
	defer cache.Close()

	cache.Set("key1", "value1")
	time.Sleep(200 * time.Millisecond)

	if _, ok := cache.Get("key1"); ok {
		t.Error("Expected key to expire")
	}
}

func TestEviction(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxSize = 10
	cache := NewCache(cfg)
	defer cache.Close()

	// Fill cache
	for i := 0; i < 15; i++ {
		cache.Set(fmt.Sprintf("key%d", i), i)
	}

	// First keys should be evicted
	if _, ok := cache.Get("key0"); ok {
		t.Error("Expected key0 to be evicted")
	}
}

func TestConcurrent(t *testing.T) {
	cache := NewCache(DefaultConfig())
	defer cache.Close()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", id%10)
			for j := 0; j < 1000; j++ {
				cache.Set(key, id)
				cache.Get(key)
			}
		}(i)
	}
	wg.Wait()
}

func BenchmarkSet(b *testing.B) {
	cache := NewCache(DefaultConfig())
	defer cache.Close()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Set(fmt.Sprintf("key%d", i), i)
			i++
		}
	})
}

func BenchmarkGet(b *testing.B) {
	cache := NewCache(DefaultConfig())
	defer cache.Close()

	// Prepopulate
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			cache.Get(fmt.Sprintf("key%d", i%1000))
			i++
		}
	})
}
