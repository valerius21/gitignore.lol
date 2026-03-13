// Package lib provides enhanced rate limiting functionality for handling vulnerability scanners
package lib

import (
	"strings"
	"sync"
	"time"
)

// ResponseType represents different types of responses for rate limiting
type ResponseType int

const (
	ResponseTypeNormal ResponseType = iota
	ResponseType404
	ResponseTypeError
)

// IPState tracks the state of an IP address
type IPState struct {
	NormalRequests []RequestRecord
	ErrorRequests  []RequestRecord
	BlockedUntil   time.Time
	Violations     int
}

// EnhancedRateLimiter provides advanced rate limiting with scanner protection
type EnhancedRateLimiter struct {
	ipStates       map[string]*IPState
	normalLimit    int // Normal requests per window
	errorLimit     int // 404/error requests per window
	windowDuration time.Duration
	blockDuration  time.Duration // How long to block repeat offenders
	maxViolations  int           // Max violations before longer block
	mu             sync.RWMutex
	cleanupTicker  *time.Ticker
	stopCleanup    chan bool
	scannerPaths   []string // Common scanner paths to block immediately
}

// NewEnhancedRateLimiter creates a new enhanced rate limiter for scanner protection
func NewEnhancedRateLimiter(normalLimit, errorLimit int, windowSeconds, blockMinutes, maxViolations int, cleanupIntervalMs int) *EnhancedRateLimiter {
	limiter := &EnhancedRateLimiter{
		ipStates:       make(map[string]*IPState),
		normalLimit:    normalLimit,
		errorLimit:     errorLimit,
		windowDuration: time.Duration(windowSeconds) * time.Second,
		blockDuration:  time.Duration(blockMinutes) * time.Minute,
		maxViolations:  maxViolations,
		stopCleanup:    make(chan bool),
		scannerPaths:   getScannerPaths(),
	}

	limiter.startCleanup(time.Duration(cleanupIntervalMs) * time.Millisecond)
	return limiter
}

// getScannerPaths returns common vulnerability scanner paths to block immediately
func getScannerPaths() []string {
	return []string{
		"/wp-admin", "/wp-login", "/wordpress", "/wp-content",
		"/phpmyadmin", "/phpMyAdmin",
		"/cgi-bin", "/HNAP1", "/wpad.dat",
		"/.env", "/.git", "/.aws", "/.docker", "/.kube",
		"/xmlrpc.php", "/wp-config.php", "/web.config",
		"/vendor", "/node_modules", "/package.json",
		"/solr", "/elasticsearch", "/kibana",
		"/actuator",
		"/server-status", "/server-info",
		"/crossdomain.xml", "/clientaccesspolicy.xml",
	}
}

// IsScannerPath checks if the path matches common scanner patterns
func (erl *EnhancedRateLimiter) IsScannerPath(path string) bool {
	path = strings.ToLower(path)

	for _, scannerPath := range erl.scannerPaths {
		if strings.HasPrefix(path, scannerPath) {
			return true
		}
	}

	// Check for common file extensions targeted by scanners
	suspiciousExtensions := []string{
		".php", ".asp", ".aspx", ".jsp", ".cgi", ".pl",
		".sql", ".db", ".bak", ".old", ".tmp",
		".zip", ".tar", ".rar", ".7z",
	}

	for _, ext := range suspiciousExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}

	return false
}

// IsAllowed checks if a request is allowed based on IP, path, and response type
func (erl *EnhancedRateLimiter) IsAllowed(ip, path string, responseType ResponseType) bool {
	now := time.Now()

	// Block scanner paths immediately
	if erl.IsScannerPath(path) {
		Logger.Warn("Blocking scanner path", "ip", ip, "path", path)
		erl.recordViolation(ip, now)
		return false
	}

	erl.mu.Lock()
	defer erl.mu.Unlock()

	// Get or create IP state
	state, exists := erl.ipStates[ip]
	if !exists {
		state = &IPState{
			NormalRequests: make([]RequestRecord, 0),
			ErrorRequests:  make([]RequestRecord, 0),
		}
		erl.ipStates[ip] = state
	}

	// Check if IP is currently blocked
	if now.Before(state.BlockedUntil) {
		Logger.Debug("Request blocked - IP still in timeout", "ip", ip, "blocked_until", state.BlockedUntil)
		return false
	}

	// Filter valid requests based on response type
	var requestList []RequestRecord
	var limit int

	switch responseType {
	case ResponseType404, ResponseTypeError:
		requestList = erl.filterValidRequests(state.ErrorRequests, now)
		limit = erl.errorLimit
	default:
		requestList = erl.filterValidRequests(state.NormalRequests, now)
		limit = erl.normalLimit
	}

	// Check if limit exceeded
	if len(requestList) >= limit {
		erl.blockIP(ip, state, now)
		Logger.Warn("Rate limit exceeded", "ip", ip, "response_type", responseType, "requests", len(requestList), "limit", limit)
		return false
	}

	// Add current request
	newRequest := RequestRecord{Timestamp: now}
	switch responseType {
	case ResponseType404, ResponseTypeError:
		state.ErrorRequests = append(requestList, newRequest)
	default:
		state.NormalRequests = append(requestList, newRequest)
	}

	return true
}

// recordViolation records a violation for an IP (for scanner paths)
func (erl *EnhancedRateLimiter) recordViolation(ip string, now time.Time) {
	erl.mu.Lock()
	defer erl.mu.Unlock()

	state, exists := erl.ipStates[ip]
	if !exists {
		state = &IPState{}
		erl.ipStates[ip] = state
	}

	state.Violations++

	// Block immediately for scanner paths
	blockDuration := erl.blockDuration
	if state.Violations >= erl.maxViolations {
		// Longer block for repeat offenders
		blockDuration = blockDuration * time.Duration(state.Violations)
	}

	state.BlockedUntil = now.Add(blockDuration)
	Logger.Warn("IP blocked for scanner activity", "ip", ip, "violations", state.Violations, "blocked_until", state.BlockedUntil)
}

// blockIP blocks an IP for rate limit violations
func (erl *EnhancedRateLimiter) blockIP(ip string, state *IPState, now time.Time) {
	state.Violations++
	blockDuration := erl.blockDuration

	// Exponential backoff for repeat offenders
	if state.Violations >= erl.maxViolations {
		blockDuration = blockDuration * time.Duration(state.Violations)
	}

	state.BlockedUntil = now.Add(blockDuration)
}

// filterValidRequests removes requests outside the current window
func (erl *EnhancedRateLimiter) filterValidRequests(requests []RequestRecord, now time.Time) []RequestRecord {
	windowStart := now.Add(-erl.windowDuration)

	validIndex := 0
	for i, req := range requests {
		if req.Timestamp.After(windowStart) {
			validIndex = i
			break
		}
		if i == len(requests)-1 {
			return []RequestRecord{}
		}
	}

	return requests[validIndex:]
}

// startCleanup starts background cleanup
func (erl *EnhancedRateLimiter) startCleanup(interval time.Duration) {
	erl.cleanupTicker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-erl.cleanupTicker.C:
				erl.cleanup()
			case <-erl.stopCleanup:
				erl.cleanupTicker.Stop()
				return
			}
		}
	}()
}

// cleanup removes expired entries
func (erl *EnhancedRateLimiter) cleanup() {
	now := time.Now()

	erl.mu.Lock()
	defer erl.mu.Unlock()

	ipsToRemove := make([]string, 0)

	for ip, state := range erl.ipStates {
		// Clean expired requests
		state.NormalRequests = erl.filterValidRequests(state.NormalRequests, now)
		state.ErrorRequests = erl.filterValidRequests(state.ErrorRequests, now)

		// Reset violations for IPs that haven't been active
		if len(state.NormalRequests) == 0 && len(state.ErrorRequests) == 0 && now.After(state.BlockedUntil) {
			if state.Violations > 0 {
				state.Violations = max(0, state.Violations-1) // Gradually reduce violations
			}
		}

		// Remove inactive IPs
		if len(state.NormalRequests) == 0 && len(state.ErrorRequests) == 0 &&
			now.After(state.BlockedUntil) && state.Violations == 0 {
			ipsToRemove = append(ipsToRemove, ip)
		}
	}

	for _, ip := range ipsToRemove {
		delete(erl.ipStates, ip)
	}

	Logger.Debug("Enhanced rate limiter cleanup completed",
		"active_ips", len(erl.ipStates),
		"removed_ips", len(ipsToRemove))
}

// GetStats returns enhanced statistics
func (erl *EnhancedRateLimiter) GetStats() map[string]interface{} {
	erl.mu.RLock()
	defer erl.mu.RUnlock()

	totalNormalRequests := 0
	totalErrorRequests := 0
	blockedIPs := 0
	now := time.Now()

	for _, state := range erl.ipStates {
		totalNormalRequests += len(state.NormalRequests)
		totalErrorRequests += len(state.ErrorRequests)
		if now.Before(state.BlockedUntil) {
			blockedIPs++
		}
	}

	return map[string]interface{}{
		"active_ips":      len(erl.ipStates),
		"blocked_ips":     blockedIPs,
		"normal_requests": totalNormalRequests,
		"error_requests":  totalErrorRequests,
		"normal_limit":    erl.normalLimit,
		"error_limit":     erl.errorLimit,
		"window_seconds":  int(erl.windowDuration.Seconds()),
		"block_minutes":   int(erl.blockDuration.Minutes()),
	}
}

// Stop gracefully stops the enhanced rate limiter
func (erl *EnhancedRateLimiter) Stop() {
	close(erl.stopCleanup)
}
