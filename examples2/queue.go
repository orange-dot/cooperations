// go.mod
module example.com/queue

go 1.22
```

```go
// queue/queue.go
// Package queue provides a simple generic FIFO queue.
//
// Use SliceQueue when the queue is confined to a single goroutine and you want
// to avoid synchronization overhead. Use SafeQueue when the queue will be used
// concurrently across goroutines.
package queue

import "sync"

// Queue defines the interface for a FIFO queue.
type Queue[T any] interface {
	Enqueue(item T)
	Dequeue() (T, bool)
	Peek() (T, bool)
	Len() int
	IsEmpty() bool
	Clear()
}

// SliceQueue is a non-thread-safe implementation.
type SliceQueue[T any] struct {
	items []T
}

// SafeQueue is a thread-safe implementation.
type SafeQueue[T any] struct {
	items []T
	mu    sync.RWMutex
}

var _ Queue[int] = (*SliceQueue[int])(nil)
var _ Queue[int] = (*SafeQueue[int])(nil)

// New creates a new non-thread-safe queue.
func New[T any]() *SliceQueue[T] {
	return &SliceQueue[T]{items: make([]T, 0)}
}

// NewWithCapacity creates a new non-thread-safe queue with initial capacity.
func NewWithCapacity[T any](capacity int) *SliceQueue[T] {
	if capacity < 0 {
		capacity = 0
	}
	return &SliceQueue[T]{items: make([]T, 0, capacity)}
}

// NewSafe creates a new thread-safe queue.
func NewSafe[T any]() *SafeQueue[T] {
	return &SafeQueue[T]{items: make([]T, 0)}
}

// NewSafeWithCapacity creates a new thread-safe queue with initial capacity.
func NewSafeWithCapacity[T any](capacity int) *SafeQueue[T] {
	if capacity < 0 {
		capacity = 0
	}
	return &SafeQueue[T]{items: make([]T, 0, capacity)}
}
```

```go
// queue/slice_queue.go
package queue

const shrinkMinCapacity = 1024
const shrinkFactor = 4

func (q *SliceQueue[T]) Enqueue(item T) {
	q.items = append(q.items, item)
}

func (q *SliceQueue[T]) Dequeue() (T, bool) {
	if len(q.items) == 0 {
		var zero T
		return zero, false
	}

	item := q.items[0]

	var zero T
	q.items[0] = zero
	q.items = q.items[1:]

	if cap(q.items) > shrinkMinCapacity && len(q.items)*shrinkFactor < cap(q.items) {
		newItems := make([]T, len(q.items))
		copy(newItems, q.items)
		q.items = newItems
	}

	return item, true
}

func (q *SliceQueue[T]) Peek() (T, bool) {
	if len(q.items) == 0 {
		var zero T
		return zero, false
	}
	return q.items[0], true
}

func (q *SliceQueue[T]) Len() int {
	return len(q.items)
}

func (q *SliceQueue[T]) IsEmpty() bool {
	return len(q.items) == 0
}

func (q *SliceQueue[T]) Clear() {
	var zero T
	for i := range q.items {
		q.items[i] = zero
	}
	q.items = q.items[:0]

	// Drop backing array to release memory.
	q.items = nil
}
```

```go
// queue/safe_queue.go
package queue

func (q *SafeQueue[T]) Enqueue(item T) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, item)
}

func (q *SafeQueue[T]) Dequeue() (T, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		var zero T
		return zero, false
	}

	item := q.items[0]

	var zero T
	q.items[0] = zero
	q.items = q.items[1:]

	if cap(q.items) > shrinkMinCapacity && len(q.items)*shrinkFactor < cap(q.items) {
		newItems := make([]T, len(q.items))
		copy(newItems, q.items)
		q.items = newItems
	}

	return item, true
}

func (q *SafeQueue[T]) Peek() (T, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if len(q.items) == 0 {
		var zero T
		return zero, false
	}
	return q.items[0], true
}

func (q *SafeQueue[T]) Len() int {
	q.mu.RLock()
	n := len(q.items)
	q.mu.RUnlock()
	return n
}

func (q *SafeQueue[T]) IsEmpty() bool {
	return q.Len() == 0
}

func (q *SafeQueue[T]) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	var zero T
	for i := range q.items {
		q.items[i] = zero
	}
	q.items = q.items[:0]

	// Drop backing array to release memory.
	q.items = nil
}
```

```go
// queue/queue_test.go
package queue

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestSliceQueueBasicFIFO(t *testing.T) {
	q := New[int]()
	if !q.IsEmpty() {
		t.Fatalf("expected empty queue")
	}

	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	if got := q.Len(); got != 3 {
		t.Fatalf("Len()=%d, want 3", got)
	}

	if v, ok := q.Peek(); !ok || v != 1 {
		t.Fatalf("Peek()=(%v,%v), want (1,true)", v, ok)
	}

	if v, ok := q.Dequeue(); !ok || v != 1 {
		t.Fatalf("Dequeue()=(%v,%v), want (1,true)", v, ok)
	}
	if v, ok := q.Dequeue(); !ok || v != 2 {
		t.Fatalf("Dequeue()=(%v,%v), want (2,true)", v, ok)
	}
	if v, ok := q.Dequeue(); !ok || v != 3 {
		t.Fatalf("Dequeue()=(%v,%v), want (3,true)", v, ok)
	}

	if _, ok := q.Dequeue(); ok {
		t.Fatalf("expected Dequeue on empty to return ok=false")
	}
	if _, ok := q.Peek(); ok {
		t.Fatalf("expected Peek on empty to return ok=false")
	}
}

func TestSliceQueueClear(t *testing.T) {
	q := NewWithCapacity[string](10)
	q.Enqueue("a")
	q.Enqueue("b")

	q.Clear()
	if q.Len() != 0 {
		t.Fatalf("Len()=%d, want 0", q.Len())
	}
	if !q.IsEmpty() {
		t.Fatalf("expected empty after Clear")
	}
	if _, ok := q.Dequeue(); ok {
		t.Fatalf("expected Dequeue on empty after Clear to return ok=false")
	}
}

func TestNewWithCapacityNegative(t *testing.T) {
	q := NewWithCapacity[int](-1)
	if q == nil {
		t.Fatalf("expected non-nil queue")
	}
	if q.Len() != 0 {
		t.Fatalf("Len()=%d, want 0", q.Len())
	}
}

func TestNewSafeWithCapacityNegative(t *testing.T) {
	q := NewSafeWithCapacity[int](-123)
	if q == nil {
		t.Fatalf("expected non-nil queue")
	}
	if q.Len() != 0 {
		t.Fatalf("Len()=%d, want 0", q.Len())
	}
}

func TestSafeQueueConcurrentEnqueue(t *testing.T) {
	q := NewSafe[int]()

	const goroutines = 8
	const perG = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func(base int) {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				q.Enqueue(base*perG + i)
			}
		}(g)
	}
	wg.Wait()

	if got := q.Len(); got != goroutines*perG {
		t.Fatalf("Len()=%d, want %d", got, goroutines*perG)
	}
}

func TestSafeQueueBasicOperations(t *testing.T) {
	q := NewSafe[int]()
	if !q.IsEmpty() {
		t.Fatalf("expected empty")
	}

	q.Enqueue(42)
	if v, ok := q.Peek(); !ok || v != 42 {
		t.Fatalf("Peek()=(%v,%v), want (42,true)", v, ok)
	}

	if v, ok := q.Dequeue(); !ok || v != 42 {
		t.Fatalf("Dequeue()=(%v,%v), want (42,true)", v, ok)
	}
	if _, ok := q.Dequeue(); ok {
		t.Fatalf("expected Dequeue on empty to return ok=false")
	}
}

func TestSafeQueueConcurrentReadWrite(t *testing.T) {
	q := NewSafe[int]()

	const producers = 4
	const perProducer = 5000
	const consumers = 4

	var produced int64
	var consumed int64

	start := make(chan struct{})
	var prodWG sync.WaitGroup
	prodWG.Add(producers)
	for p := 0; p < producers; p++ {
		go func(base int) {
			defer prodWG.Done()
			<-start
			for i := 0; i < perProducer; i++ {
				q.Enqueue(base*perProducer + i)
				atomic.AddInt64(&produced, 1)
			}
		}(p)
	}

	doneProducers := make(chan struct{})
	go func() {
		prodWG.Wait()
		close(doneProducers)
	}()

	var consWG sync.WaitGroup
	consWG.Add(consumers)
	for c := 0; c < consumers; c++ {
		go func() {
			defer consWG.Done()
			<-start
			for {
				if _, ok := q.Dequeue(); ok {
					atomic.AddInt64(&consumed, 1)
					continue
				}

				select {
				case <-doneProducers:
					// Drain any last items.
					for {
						if _, ok := q.Dequeue(); !ok {
							return
						}
						atomic.AddInt64(&consumed, 1)
					}
				default:
					time.Sleep(50 * time.Microsecond)
				}
			}
		}()
	}

	close(start)
	consWG.Wait()

	if got, want := atomic.LoadInt64(&produced), int64(producers*perProducer); got != want {
		t.Fatalf("produced=%d, want %d", got, want)
	}
	if got, want := atomic.LoadInt64(&consumed), atomic.LoadInt64(&produced); got != want {
		t.Fatalf("consumed=%d, want %d", got, want)
	}
	if q.Len() != 0 {
		t.Fatalf("Len()=%d, want 0", q.Len())
	}
}
```

```go
// queue/example_test.go
package queue

import "fmt"

func ExampleSliceQueue() {
	q := New[int]()
	q.Enqueue(10)
	q.Enqueue(20)

	v, _ := q.Dequeue()
	fmt.Println(v)

	v, _ = q.Peek()
	fmt.Println(v)

	// Output:
	// 10
	// 20
}

func ExampleSafeQueue() {
	q := NewSafe[string]()
	q.Enqueue("a")
	q.Enqueue("b")

	v, _ := q.Dequeue()
	fmt.Println(v)

	// Output:
	// a
}