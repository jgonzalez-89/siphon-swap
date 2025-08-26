package cache

import (
	"sync"
	"time"
)

// CacheItem representa un elemento en cache
type CacheItem struct {
	Value      any
	Expiration time.Time
}

// Cache es un cache simple en memoria
type Cache struct {
	items map[string]CacheItem
	mu    sync.RWMutex
	ttl   time.Duration
}

// NewCache crea una nueva instancia de cache
func NewCache(defaultTTL time.Duration) *Cache {
	cache := &Cache{
		items: make(map[string]CacheItem),
		ttl:   defaultTTL,
	}

	// Limpieza periódica de elementos expirados
	go cache.cleanup()

	return cache
}

// Set guarda un valor en cache
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiration := time.Now().Add(ttl)
	c.items[key] = CacheItem{
		Value:      value,
		Expiration: expiration,
	}
}

// Get obtiene un valor del cache
func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return item, false
	}

	// Verificar si expiró
	if time.Now().After(item.Expiration) {
		return item.Value, false
	}

	return item.Value, true
}

// Delete elimina un elemento del cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear limpia todo el cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]CacheItem)
}

// cleanup elimina elementos expirados periódicamente
func (c *Cache) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.Expiration) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

// Size retorna el número de elementos en cache
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
