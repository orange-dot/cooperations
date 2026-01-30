// File: main.go
package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

var (
	ErrInvalidConfig = errors.New("invalid token bucket config")
	ErrTimeout       = errors.New("rate limiter wait timed out")
)

// TokenBucket is a thread-safe token bucket rate limiter.
// Tokens are replenished continuously at rate tokensPerSecond up to capacity.
// Calls to Allow/Wait consume tokens.
type TokenBucket struct {
	mu sync.Mutex

	capacity        float64
	tokens          float64
	tokensPerSecond float64

	lastRefill time.Time
	nowFn      func() time.Time
}

// TokenBucketConfig defines the limiter behavior.
type TokenBucketConfig struct {
	Capacity        int              // maximum burst size in tokens; must be > 0
	TokensPerSecond float64          // refill rate; must be > 0 and finite
	InitialTokens   *int             // optional; defaults to Capacity
	Now             func() time.Time // optional; defaults to time.Now
}

// NewTokenBucket constructs a new token bucket limiter.
func NewTokenBucket(cfg TokenBucketConfig) (*TokenBucket, error) {
	if cfg.Capacity <= 0 {
		return nil, fmt.Errorf("%w: capacity must be > 0", ErrInvalidConfig)
	}
	if cfg.TokensPerSecond <= 0 || math.IsNaN(cfg.TokensPerSecond) || math.IsInf(cfg.TokensPerSecond, 0) {
		return nil, fmt.Errorf("%w: tokensPerSecond must be a finite positive number", ErrInvalidConfig)
	}

	nowFn := cfg.Now
	if nowFn == nil {
		nowFn = time.Now
	}

	initial := cfg.Capacity
	if cfg.InitialTokens != nil {
		initial = *cfg.InitialTokens
		if initial < 0 {
			return nil, fmt.Errorf("%w: initialTokens must be >= 0", ErrInvalidConfig)
		}
		if initial > cfg.Capacity {
			initial = cfg.Capacity
		}
	}

	now := nowFn()
	return &TokenBucket{
		capacity:        float64(cfg.Capacity),
		tokens:          float64(initial),
		tokensPerSecond: cfg.TokensPerSecond,
		lastRefill:      now,
		nowFn:           nowFn,
	}, nil
}

// Allow tries to take n tokens immediately. It returns true if allowed.
func (tb *TokenBucket) Allow(n int) bool {
	if n <= 0 {
		return true
	}

	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refillLocked(tb.nowFn())

	need := float64(n)
	if need > tb.capacity {
		return false
	}
	if tb.tokens >= need {
		tb.tokens -= need
		return true
	}
	return false
}

// Wait blocks until n tokens are available or context is canceled.
// It consumes n tokens if successful.
func (tb *TokenBucket) Wait(ctx context.Context, n int) error {
	if ctx == nil {
		return errors.New("nil context")
	}
	if n <= 0 {
		return nil
	}
	if float64(n) > tb.Capacity() {
		return fmt.Errorf("%w: requested tokens (%d) exceeds capacity", ErrInvalidConfig, n)
	}

	for {
		wait, ok := tb.tryConsumeOrComputeWait(n, tb.nowFn())
		if ok {
			return nil
		}

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
}

// WaitTimeout blocks until n tokens are available or the timeout elapses.
func (tb *TokenBucket) WaitTimeout(n int, timeout time.Duration) error {
	if timeout < 0 {
		return ErrTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	err := tb.Wait(ctx, n)
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrTimeout
	}
	return err
}

// Capacity returns the configured bucket capacity.
func (tb *TokenBucket) Capacity() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.capacity
}

// Rate returns the configured refill rate in tokens per second.
func (tb *TokenBucket) Rate() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.tokensPerSecond
}

// Available returns the current number of available tokens (refilled to "now").
func (tb *TokenBucket) Available() float64 {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refillLocked(tb.nowFn())
	return tb.tokens
}

// tryConsumeOrComputeWait attempts to consume immediately; if not possible,
// returns the duration to wait until enough tokens will be available.
func (tb *TokenBucket) tryConsumeOrComputeWait(n int, now time.Time) (time.Duration, bool) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refillLocked(now)

	need := float64(n)
	if tb.tokens >= need {
		tb.tokens -= need
		return 0, true
	}

	missing := need - tb.tokens
	seconds := missing / tb.tokensPerSecond
	if seconds < 0 {
		seconds = 0
	}

	wait := time.Duration(math.Ceil(seconds * float64(time.Second)))
	if wait <= 0 {
		wait = time.Nanosecond
	}
	return wait, false
}

func (tb *TokenBucket) refillLocked(now time.Time) {
	if now.Before(tb.lastRefill) {
		tb.lastRefill = now
		return
	}

	elapsed := now.Sub(tb.lastRefill).Seconds()
	if elapsed == 0 {
		return
	}

	tb.tokens += elapsed * tb.tokensPerSecond
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = now
}

func demoRateLimiter() {
	limiter, err := NewTokenBucket(TokenBucketConfig{
		Capacity:        10,
		TokensPerSecond: 5,
	})
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		fmt.Println("allow 1:", limiter.Allow(1))
	}
	fmt.Println("allow after burst:", limiter.Allow(1))

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := limiter.Wait(ctx, 1); err != nil {
		fmt.Println("wait error:", err)
	} else {
		fmt.Println("wait ok; available:", limiter.Available())
	}
}
