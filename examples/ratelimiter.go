package main

import (
	"errors"
	"sync"
	"time"
)

// TokenBucket represents a rate limiter using the token bucket algorithm.
// It limits the rate at which an action can be performed.
type TokenBucket struct {
	rate       float64    // tokens per second
	bucketSize int        // maximum number of tokens
	tokens     float64    // current number of tokens
	lastCheck  time.Time  // last time tokens were replenished
	mu         sync.Mutex // ensures atomic access to token bucket
}

// NewTokenBucket creates a new TokenBucket.
// rate specifies the refill rate in tokens per second.
// bucketSize specifies the maximum number of tokens.
func NewTokenBucket(rate float64, bucketSize int) (*TokenBucket, error) {
	if rate <= 0 {
		return nil, errors.New("rate must be positive")
	}
	if bucketSize <= 0 {
		return nil, errors.New("bucket size must be positive")
	}
	return &TokenBucket{
		rate:       rate,
		bucketSize: bucketSize,
		tokens:     float64(bucketSize), // start with full capacity
		lastCheck:  time.Now(),
	}, nil
}

// Allow attempts to take a token from the bucket.
// Returns true if a token was successfully taken, false otherwise.
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Calculate elapsed time since last check and replenish tokens accordingly
	now := time.Now()
	elapsed := now.Sub(tb.lastCheck).Seconds()
	tb.tokens += elapsed * tb.rate // Add new tokens based on elapsed time and rate
	if tb.tokens > float64(tb.bucketSize) {
		tb.tokens = float64(tb.bucketSize) // Ensure bucket doesn't overflow
	}

	tb.lastCheck = now

	// Check if at least one token is available
	if tb.tokens >= 1.0 {
		tb.tokens-- // Take a token
		return true
	}

	return false
}

func demoRateLimiter() {
	// Example usage of the TokenBucket rate limiter
	rate := 1.0 // 1 token per second
	bucketSize := 5
	limiter, err := NewTokenBucket(rate, bucketSize)
	if err != nil {
		println("rate limiter error:", err.Error())
		return
	}

	// Attempt to perform 10 actions as fast as possible
	for i := 0; i < 10; i++ {
		if limiter.Allow() {
			// Action allowed
			println("Action allowed")
		} else {
			// Action not allowed
			println("Action not allowed")
		}
		time.Sleep(200 * time.Millisecond) // simulate work
	}
}

// NEXT: reviewer
