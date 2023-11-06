package cache

import (
	"sync"
	"time"
)

const defaultCacheTTL = 86400

type item[V any] struct {
	value     V
	createdAt int64
}

type MemoryCache[T Value] struct {
	cache map[Key]*item[T]
	sync.RWMutex
}

func NewMemoryCache[T Value]() *MemoryCache[T] {
	c := &MemoryCache[T]{cache: make(map[Key]*item[T])}
	go c.setTtlTimer()

	return c
}

func (c *MemoryCache[T]) setTtlTimer() {
	for {
		c.Lock()
		for k, v := range c.cache {
			if time.Now().Unix()-v.createdAt > defaultCacheTTL {
				delete(c.cache, k)
			}
		}
		c.Unlock()

		<-time.After(time.Second)
	}
}

func (c *MemoryCache[T]) Set(key Key, value T) error {
	c.Lock()
	c.cache[key] = &item[T]{
		value:     value,
		createdAt: time.Now().Unix(),
	}
	c.Unlock()

	return nil
}

func (c *MemoryCache[T]) Get(key Key) (T, bool) {
	c.RLock()
	item, ex := c.cache[key]
	c.RUnlock()
	if ex {
		return item.value, true
	}

	var value T
	return value, false
}
