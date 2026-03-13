// Package lib provides rate limiting functionality using a moving window approach
package lib

import (
	"sync"
	"time"
)

// RequestRecord represents a single request timestamp
type RequestRecord struct {
	Timestamp time.Time
}

// MovingWindowLimiter implements a sliding window rate limiter optimized for memory efficiency
type MovingWindowLimiter struct {
	requests       map[string][]RequestRecord // IP -> list of request timestamps
	maxRequests    int                        // Maximum requests per window
	windowDuration time.Duration              // Window size
	mu             sync.RWMutex               // Read-write mutex for concurrent access
	cleanupTicker  *time.Ticker               // Ticker for periodic cleanup
	stopCleanup    chan bool                  // Channel to stop cleanup goroutine
}

// NewMovingWindowLimiter creates a new rate limiter with specified parameters
func NewMovingWindowLimiter(maxRequests int, windowSeconds int, cleanupIntervalMs int) *MovingWindowLimiter {
	limiter := &MovingWindowLimiter{
		requests:       make(map[string][]RequestRecord),
		maxRequests:    maxRequests,
		windowDuration: time.Duration(windowSeconds) * time.Second,
		stopCleanup:    make(chan bool),
	}

	// Start background cleanup routine
	limiter.startCleanup(time.Duration(cleanupIntervalMs) * time.Millisecond)

	return limiter
}

// IsAllowed checks if a request from the given IP is allowed
func (mwl *MovingWindowLimiter) IsAllowed(ip string) bool {
	now := time.Now()

	mwl.mu.Lock()
	defer mwl.mu.Unlock()

	// Get existing requests for this IP
	requests, exists := mwl.requests[ip]
	if !exists {
		// First request from this IP
		mwl.requests[ip] = []RequestRecord{{Timestamp: now}}
		return true
	}

	// Remove requests outside the current window
	validRequests := mwl.filterValidRequests(requests, now)

	// Check if we're within the rate limit
	if len(validRequests) >= mwl.maxRequests {
		// Update the requests slice even if rate limited (remove old entries)
		mwl.requests[ip] = validRequests
		return false
	}

	// Add current request and update
	validRequests = append(validRequests, RequestRecord{Timestamp: now})
	mwl.requests[ip] = validRequests

	return true
}

// filterValidRequests removes requests outside the current window
func (mwl *MovingWindowLimiter) filterValidRequests(requests []RequestRecord, now time.Time) []RequestRecord {
	windowStart := now.Add(-mwl.windowDuration)

	// Find the first valid request (binary search could be used for optimization if needed)
	validIndex := 0
	for i, req := range requests {
		if req.Timestamp.After(windowStart) {
			validIndex = i
			break
		}
		if i == len(requests)-1 {
			// All requests are expired
			return []RequestRecord{}
		}
	}

	// Return slice of valid requests (reuse underlying array for memory efficiency)
	return requests[validIndex:]
}

// startCleanup starts a background goroutine to periodically clean up expired entries
func (mwl *MovingWindowLimiter) startCleanup(interval time.Duration) {
	mwl.cleanupTicker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-mwl.cleanupTicker.C:
				mwl.cleanup()
			case <-mwl.stopCleanup:
				mwl.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// cleanup removes expired entries and empty IP records to prevent memory leaks
func (mwl *MovingWindowLimiter) cleanup() {
	now := time.Now()
	windowStart := now.Add(-mwl.windowDuration)

	mwl.mu.Lock()
	defer mwl.mu.Unlock()

	// Track IPs to remove
	ipsToRemove := make([]string, 0)

	for ip, requests := range mwl.requests {
		validRequests := make([]RequestRecord, 0)

		// Keep only valid requests
		for _, req := range requests {
			if req.Timestamp.After(windowStart) {
				validRequests = append(validRequests, req)
			}
		}

		// Remove IP entry if no valid requests remain
		if len(validRequests) == 0 {
			ipsToRemove = append(ipsToRemove, ip)
		} else {
			// Update with cleaned requests (reallocate slice to free memory)
			mwl.requests[ip] = validRequests
		}
	}

	// Remove empty IP entries
	for _, ip := range ipsToRemove {
		delete(mwl.requests, ip)
	}

	Logger.Debug("Rate limiter cleanup completed",
		"active_ips", len(mwl.requests),
		"removed_ips", len(ipsToRemove))
}

// GetStats returns current statistics about the rate limiter
func (mwl *MovingWindowLimiter) GetStats() map[string]interface{} {
	mwl.mu.RLock()
	defer mwl.mu.RUnlock()

	totalRequests := 0
	for _, requests := range mwl.requests {
		totalRequests += len(requests)
	}

	return map[string]interface{}{
		"active_ips":     len(mwl.requests),
		"total_requests": totalRequests,
		"max_requests":   mwl.maxRequests,
		"window_seconds": int(mwl.windowDuration.Seconds()),
	}
}

// Stop gracefully stops the rate limiter and cleanup routines
func (mwl *MovingWindowLimiter) Stop() {
	close(mwl.stopCleanup)
}
