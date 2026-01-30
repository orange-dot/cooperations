package main

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
)

// LRUCache struct with added thread safety via a RWMutex
type LRUCache struct {
	capacity int
	cache    map[int]*list.Element
	list     *list.List
	mu       sync.Mutex // RWMutex replaced with Mutex for correct write lock handling in Get
}

type entry struct {
	key   int
	value int
}

// NewLRUCache constructor with capacity validation returning an error instead of panic
func NewLRUCache(capacity int) (*LRUCache, error) {
	if capacity <= 0 {
		return nil, errors.New("capacity must be positive")
	}
	return &LRUCache{
		capacity: capacity,
		cache:    make(map[int]*list.Element),
		list:     list.New(),
	}, nil
}

// Get method retrieves the value for a key and marks it as recently used.
// Updated to use Lock instead of RLock to fix race condition
func (c *LRUCache) Get(key int) (value int, ok bool) {
	c.mu.Lock() 
	defer c.mu.Unlock()

	if elem, found := c.cache[key]; found {
		c.list.MoveToFront(elem)
		return c.getEntry(elem).value, true
	}
	return 0, false
}

// Put method updates or inserts a value by key into the cache.
func (c *LRUCache) Put(key int, value int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, found := c.cache[key]; found {
		c.list.MoveToFront(elem)
		elem.Value.(*entry).value = value // Safe here as elem is assured to have correct type
		return
	}

	if c.list.Len() == c.capacity {
		c.removeOldest()
	}

	elem := c.list.PushFront(&entry{key, value})
	c.cache[key] = elem
}

// Utility function to safely retrieve entry from a list element
// Added as suggested for safer type assertions
func (c *LRUCache) getEntry(elem *list.Element) *entry {
	if e, ok := elem.Value.(*entry); ok {
		return e
	}
	return nil // Though this should theoretically never happen
}

// removeOldest removes the least recently used (oldest) item from the cache
func (c *LRUCache) removeOldest() {
	oldest := c.list.Back()
	if oldest != nil {
		c.list.Remove(oldest)
		delete(c.cache, c.getEntry(oldest).key) // Updated to use safe type assertion via getEntry
	}
}

// Len returns the current number of items in the cache
func (c *LRUCache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.list.Len()
}

// Clear resets the cache to an empty state
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.list.Init()
	c.cache = make(map[int]*list.Element)
}

func main() {
	lruCache, err := NewLRUCache(2)
	if err != nil {
		panic(err) // For simplicity in this example, panic on error
	}
	lruCache.Put(1, 1) // Cache is {1=1}
	lruCache.Put(2, 2) // Cache is {1=1, 2=2}
	value, ok := lruCache.Get(1) // returns 1, true
	fmt.Println(value, ok)

	lruCache.Put(3, 3)    // Evicts key 2, cache is now {1=1, 3=3}
	value, ok = lruCache.Get(2) // returns 0, false (since 2 was evicted)
	fmt.Println(value, ok)

	lruCache.Put(4, 4)    // Evicts key 1, cache is now {4=4, 3=3}
	value, ok = lruCache.Get(1) // returns 0, false (since 1 was evicted)
	fmt.Println(value, ok)
	value, ok = lruCache.Get(3) // returns 3, true
	fmt.Println(value, ok)
	value, ok = lruCache.Get(4) // returns 4, true
	fmt.Println(value, ok)
}

// Implemented changes based on review feedback:
// - Fixed a high-severity race condition in the `Get` method.
// - Safer type assertions via the new helper method `getEntry`.
// - Adjusted `NewLRUCache` to return an error instead of panic for capacity check.
