package final

import (
	"container/list"
	"hash/fnv"
	"sync"
	"sync/atomic"
	"time"
)

// Cache is production-ready concurrent LRU cache with:
// ✅ 256 shards (high concurrency)
// ✅ RWMutex (parallel reads)
// ✅ LRU eviction
// ✅ TTL expiration with background cleanup
// ✅ Atomic metrics
type Cache struct {
	shards []*shard
	config Config
	done   chan struct{}
	wg     sync.WaitGroup

	// Metrics
	hits      uint64
	misses    uint64
	evictions uint64
}

type Config struct {
	MaxSize         int
	DefaultTTL      time.Duration
	Shards          int
	CleanupInterval time.Duration
}

type shard struct {
	mu      sync.RWMutex
	items   map[string]*entry
	lru     *list.List
	maxSize int
}

type entry struct {
	key     string
	value   interface{}
	expires time.Time
	element *list.Element
}

func DefaultConfig() Config {
	return Config{
		MaxSize:         10000,
		DefaultTTL:      5 * time.Minute,
		Shards:          256,
		CleanupInterval: 1 * time.Minute,
	}
}

func NewCache(cfg Config) *Cache {
	if cfg.Shards == 0 {
		cfg = DefaultConfig()
	}

	c := &Cache{
		shards: make([]*shard, cfg.Shards),
		config: cfg,
		done:   make(chan struct{}),
	}

	perShardSize := cfg.MaxSize / cfg.Shards
	for i := 0; i < cfg.Shards; i++ {
		c.shards[i] = &shard{
			items:   make(map[string]*entry),
			lru:     list.New(),
			maxSize: perShardSize,
		}
	}

	// Background cleanup
	c.wg.Add(1)
	go c.cleanupLoop()

	return c
}

func (c *Cache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.config.DefaultTTL)
}

func (c *Cache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	s := c.getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()

	if e, ok := s.items[key]; ok {
		e.value = value
		e.expires = time.Now().Add(ttl)
		s.lru.MoveToFront(e.element)
		return
	}

	// Check if need eviction
	if s.lru.Len() >= s.maxSize {
		oldest := s.lru.Back()
		if oldest != nil {
			e := oldest.Value.(*entry)
			delete(s.items, e.key)
			s.lru.Remove(oldest)
			atomic.AddUint64(&c.evictions, 1)
		}
	}

	e := &entry{
		key:     key,
		value:   value,
		expires: time.Now().Add(ttl),
	}
	e.element = s.lru.PushFront(e)
	s.items[key] = e
}

func (c *Cache) Get(key string) (interface{}, bool) {
	s := c.getShard(key)

	// Try read lock first
	s.mu.RLock()
	e, ok := s.items[key]
	if !ok {
		s.mu.RUnlock()
		atomic.AddUint64(&c.misses, 1)
		return nil, false
	}

	// Check expiration
	if time.Now().After(e.expires) {
		s.mu.RUnlock()
		// Need write lock to delete
		s.mu.Lock()
		delete(s.items, key)
		s.lru.Remove(e.element)
		s.mu.Unlock()
		atomic.AddUint64(&c.misses, 1)
		return nil, false
	}

	s.mu.RUnlock()

	// Move to front (need write lock)
	s.mu.Lock()
	s.lru.MoveToFront(e.element)
	value := e.value
	s.mu.Unlock()

	atomic.AddUint64(&c.hits, 1)
	return value, true
}

func (c *Cache) Delete(key string) {
	s := c.getShard(key)
	s.mu.Lock()
	defer s.mu.Unlock()

	if e, ok := s.items[key]; ok {
		delete(s.items, key)
		s.lru.Remove(e.element)
	}
}

func (c *Cache) getShard(key string) *shard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return c.shards[h.Sum32()%uint32(len(c.shards))]
}

func (c *Cache) cleanupLoop() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.done:
			return
		}
	}
}

func (c *Cache) cleanup() {
	now := time.Now()
	for _, s := range c.shards {
		s.mu.Lock()
		for key, e := range s.items {
			if now.After(e.expires) {
				delete(s.items, key)
				s.lru.Remove(e.element)
			}
		}
		s.mu.Unlock()
	}
}

func (c *Cache) Close() {
	close(c.done)
	c.wg.Wait()
}

func (c *Cache) Metrics() (hits, misses, evictions uint64) {
	return atomic.LoadUint64(&c.hits),
		atomic.LoadUint64(&c.misses),
		atomic.LoadUint64(&c.evictions)
}
