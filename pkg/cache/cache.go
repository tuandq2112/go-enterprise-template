package cache

import (
	"sync"
	"time"
)

// CacheItem represents a cache item
type CacheItem struct {
	Value      interface{}
	Expiration time.Time
}

// IsExpired checks if the cache item is expired
func (ci *CacheItem) IsExpired() bool {
	return time.Now().After(ci.Expiration)
}

// MemoryCache represents an in-memory cache
type MemoryCache struct {
	items map[string]*CacheItem
	mutex sync.RWMutex
}

// NewMemoryCache creates a new memory cache
func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		items: make(map[string]*CacheItem),
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

// Set sets a value in cache with expiration
func (c *MemoryCache) Set(key string, value interface{}, expiration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = &CacheItem{
		Value:      value,
		Expiration: time.Now().Add(expiration),
	}
}

// Get gets a value from cache
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if item.IsExpired() {
		// Remove expired item
		c.mutex.RUnlock()
		c.mutex.Lock()
		delete(c.items, key)
		c.mutex.Unlock()
		c.mutex.RLock()
		return nil, false
	}

	return item.Value, true
}

// Delete deletes a key from cache
func (c *MemoryCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.items, key)
}

// Clear clears all items from cache
func (c *MemoryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]*CacheItem)
}

// Size returns the number of items in cache
func (c *MemoryCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.items)
}

// Keys returns all keys in cache
func (c *MemoryCache) Keys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}

	return keys
}

// cleanup removes expired items periodically
func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		for key, item := range c.items {
			if item.IsExpired() {
				delete(c.items, key)
			}
		}
		c.mutex.Unlock()
	}
}

// Cache interface defines cache operations
type Cache interface {
	Set(key string, value interface{}, expiration time.Duration)
	Get(key string) (interface{}, bool)
	Delete(key string)
	Clear()
	Size() int
	Keys() []string
}

// CacheManager manages multiple caches
type CacheManager struct {
	caches map[string]Cache
	mutex  sync.RWMutex
}

// NewCacheManager creates a new cache manager
func NewCacheManager() *CacheManager {
	return &CacheManager{
		caches: make(map[string]Cache),
	}
}

// GetCache gets or creates a cache with the given name
func (cm *CacheManager) GetCache(name string) Cache {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cache, exists := cm.caches[name]; exists {
		return cache
	}

	cache := NewMemoryCache()
	cm.caches[name] = cache
	return cache
}

// DeleteCache deletes a cache
func (cm *CacheManager) DeleteCache(name string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.caches, name)
}

// ClearAll clears all caches
func (cm *CacheManager) ClearAll() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for _, cache := range cm.caches {
		cache.Clear()
	}
}

// GetCacheNames returns all cache names
func (cm *CacheManager) GetCacheNames() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	names := make([]string, 0, len(cm.caches))
	for name := range cm.caches {
		names = append(names, name)
	}

	return names
}
