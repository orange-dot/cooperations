// File: main.go
package main

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
)

var (
	ErrKeyNotFound = errors.New("lru: key not found")
)

// Cache is a generic, threadsafe LRU cache.
type Cache[K comparable, V any] struct {
	mu       sync.Mutex
	capacity int
	ll       *list.List
	items    map[K]*list.Element
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

// New creates an LRU cache with the given capacity.
// capacity must be > 0.
func New[K comparable, V any](capacity int) (*Cache[K, V], error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("lru: capacity must be > 0 (got %d)", capacity)
	}
	return &Cache[K, V]{
		capacity: capacity,
		ll:       list.New(),
		items:    make(map[K]*list.Element, capacity),
	}, nil
}

// Cap returns the configured cache capacity.
func (c *Cache[K, V]) Cap() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.capacity
}

// Len returns the number of items currently in the cache.
func (c *Cache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ll.Len()
}

// Get returns the value for key and whether it was found.
// If found, the entry becomes most-recently-used.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.items[key]; ok {
		c.ll.MoveToFront(ele)
		ent := ele.Value.(entry[K, V])
		return ent.value, true
	}

	var zero V
	return zero, false
}

// MustGet returns the value for key or ErrKeyNotFound if absent.
func (c *Cache[K, V]) MustGet(key K) (V, error) {
	v, ok := c.Get(key)
	if !ok {
		var zero V
		return zero, ErrKeyNotFound
	}
	return v, nil
}

// Put inserts or updates key with value.
// If inserting causes capacity overflow, the least-recently-used item is evicted.
func (c *Cache[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.items[key]; ok {
		ele.Value = entry[K, V]{key: key, value: value}
		c.ll.MoveToFront(ele)
		return
	}

	ele := c.ll.PushFront(entry[K, V]{key: key, value: value})
	c.items[key] = ele

	if c.ll.Len() > c.capacity {
		c.removeOldest()
	}
}

// Remove deletes key from the cache and returns whether it existed.
func (c *Cache[K, V]) Remove(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	ele, ok := c.items[key]
	if !ok {
		return false
	}
	c.removeElement(ele)
	return true
}

// Peek returns the value without updating recency.
func (c *Cache[K, V]) Peek(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, ok := c.items[key]; ok {
		ent := ele.Value.(entry[K, V])
		return ent.value, true
	}
	var zero V
	return zero, false
}

// Purge clears the cache.
func (c *Cache[K, V]) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ll.Init()
	clear(c.items)
}

// Keys returns keys from most-recently-used to least-recently-used.
func (c *Cache[K, V]) Keys() []K {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := make([]K, 0, c.ll.Len())
	for e := c.ll.Front(); e != nil; e = e.Next() {
		ent := e.Value.(entry[K, V])
		keys = append(keys, ent.key)
	}
	return keys
}

func (c *Cache[K, V]) removeOldest() {
	ele := c.ll.Back()
	if ele == nil {
		return
	}
	c.removeElement(ele)
}

func (c *Cache[K, V]) removeElement(ele *list.Element) {
	ent := ele.Value.(entry[K, V])
	delete(c.items, ent.key)
	c.ll.Remove(ele)
}

func main() {
	cache, err := New[string, int](2)
	if err != nil {
		panic(err)
	}

	cache.Put("a", 1)
	cache.Put("b", 2)
	_, _ = cache.Get("a") // "a" becomes MRU
	cache.Put("c", 3)     // evicts "b"

	if _, ok := cache.Get("b"); !ok {
		fmt.Println("b evicted")
	}
	if v, ok := cache.Get("a"); ok {
		fmt.Println("a =", v)
	}
	if v, ok := cache.Get("c"); ok {
		fmt.Println("c =", v)
	}

	fmt.Println("keys MRU->LRU:", cache.Keys())
}