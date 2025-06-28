# Rate Limiting Implementation for gitignore.lol

## Overview

Implemented a high-performance moving window rate limiter optimized for Raspberry Pi (16GB, 10% load) specifications.

## Features

### 🚀 Moving Window Algorithm
- **Sliding Window**: More accurate than fixed windows, counts requests within a rolling time period
- **Memory Efficient**: Automatically cleans up expired entries to prevent memory leaks
- **Concurrent Safe**: Thread-safe implementation using RWMutex for high-performance concurrent access

### 🔧 Configuration Options

New CLI flags added:

```bash
--rate-limit=100          # Maximum requests per window per IP (default: 100)
--rate-window=60          # Rate limiting window in seconds (default: 60)  
--rate-cleanup=30000      # Cleanup interval in milliseconds (default: 30000)
--enable-rate-limit       # Enable/disable rate limiting (default: true)
```

### 📊 Monitoring & Stats

- **Stats Endpoint**: `/stats` - View rate limiter statistics
- **Real-time Metrics**: Active IPs, total requests, configuration details
- **Debug Logging**: Detailed logs for rate limit events and cleanup operations

## Implementation Details

### Files Created/Modified

1. **`pkg/lib/rate_limiter.go`** - Core moving window rate limiter
2. **`pkg/lib/rate_middleware.go`** - Fiber middleware integration  
3. **`pkg/lib/cli.go`** - Added rate limiting configuration options
4. **`pkg/server/server.go`** - Integrated rate limiting middleware
5. **`cmd/main.go`** - Initialize and configure rate limiter
6. **`pkg/lib/rate_limiter_test.go`** - Comprehensive test suite

### Architecture

```
Request → Fiber Router → Rate Limit Middleware → API Handler
                              ↓
                      Moving Window Limiter
                              ↓
                      Background Cleanup
```

### Memory Optimization for Raspberry Pi

1. **Efficient Data Structures**: 
   - Uses slices instead of maps for request timestamps
   - Reuses underlying arrays when possible
   - Periodic cleanup removes expired entries

2. **Background Cleanup**:
   - Configurable cleanup interval (default: 30 seconds)
   - Removes expired IP entries automatically
   - Prevents memory leaks during high traffic

3. **Sliding Window Benefits**:
   - More memory efficient than token bucket
   - Accurate rate limiting without bursts
   - Scales well with concurrent requests

## Usage Examples

### Basic Usage
```bash
# Default settings (100 req/min)
./gitignore-server

# Custom rate limiting  
./gitignore-server --rate-limit 50 --rate-window 30

# Disable rate limiting
./gitignore-server --enable-rate-limit=false
```

### API Behavior

**Normal Request:**
```
GET /api/go,node
→ 200 OK (gitignore content)
```

**Rate Limited Request:**
```
GET /api/go,node  
→ 429 Too Many Requests
{
  "error": "Too Many Requests",
  "message": "Rate limit exceeded. Please try again later."
}
```

**Statistics:**
```
GET /stats
→ {
  "rate_limiting": "enabled",
  "stats": {
    "active_ips": 5,
    "total_requests": 247,
    "max_requests": 100,
    "window_seconds": 60
  }
}
```

## Performance Characteristics

### Raspberry Pi Optimized
- **Low CPU Usage**: Efficient algorithms minimize processing overhead
- **Memory Bounded**: Automatic cleanup prevents unbounded memory growth
- **Concurrent**: RWMutex allows multiple readers, scales with concurrent requests
- **Non-blocking**: Background cleanup doesn't block request processing

### Benchmarks
All tests pass, including:
- ✅ Basic rate limiting functionality
- ✅ Per-IP isolation
- ✅ Sliding window accuracy  
- ✅ Memory cleanup
- ✅ Statistics reporting
- ✅ Concurrent access safety

## Rate Limiting Strategy

1. **API Routes Only**: Rate limiting applied only to `/api/*` endpoints
2. **IP-based**: Each client IP has independent rate limits
3. **Graceful Handling**: Clean error responses with appropriate HTTP status codes
4. **Bypass Options**: Static files and documentation not rate limited

## Integration Points

- **Middleware**: Integrates seamlessly with existing Fiber routes
- **Logging**: Uses existing logger for rate limit events
- **Configuration**: Extends existing CLI argument system
- **Monitoring**: New stats endpoint for operational visibility

This implementation provides enterprise-grade rate limiting optimized for resource-constrained environments while maintaining high performance and reliability. 