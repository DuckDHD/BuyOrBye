package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/DuckDHD/BuyOrBye/internal/types"
)

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	// Requests allowed per window
	Requests int
	// Time window duration
	Window time.Duration
	// Key function to identify clients (default uses IP)
	KeyFunc func(*gin.Context) string
	// Skip function to bypass rate limiting for certain requests
	Skip func(*gin.Context) bool
}

// ClientRecord tracks request count and window start time for a client
type ClientRecord struct {
	Count     int
	WindowStart time.Time
	mutex     sync.Mutex
}

// InMemoryRateLimiter implements rate limiting using in-memory storage
type InMemoryRateLimiter struct {
	config  RateLimitConfig
	clients map[string]*ClientRecord
	mutex   sync.RWMutex
	
	// Cleanup ticker
	cleanupTicker *time.Ticker
	stopCleanup   chan bool
}

// DefaultRateLimitConfig returns default configuration for login rate limiting
// 5 requests per 15 minutes per IP address
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		Requests: 5,
		Window:   15 * time.Minute,
		KeyFunc: func(c *gin.Context) string {
			// Use client IP as the key
			return c.ClientIP()
		},
		Skip: nil, // No skip function by default
	}
}

// NewInMemoryRateLimiter creates a new in-memory rate limiter
func NewInMemoryRateLimiter(config RateLimitConfig) *InMemoryRateLimiter {
	limiter := &InMemoryRateLimiter{
		config:        config,
		clients:       make(map[string]*ClientRecord),
		stopCleanup:   make(chan bool),
	}
	
	// Start cleanup goroutine to remove expired entries
	limiter.cleanupTicker = time.NewTicker(config.Window)
	go limiter.cleanup()
	
	return limiter
}

// RateLimit returns a Gin middleware function that enforces rate limiting
func (rl *InMemoryRateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if this request should skip rate limiting
		if rl.config.Skip != nil && rl.config.Skip(c) {
			c.Next()
			return
		}
		
		// Get client key (typically IP address)
		clientKey := rl.config.KeyFunc(c)
		
		// Check if client is rate limited
		if rl.isRateLimited(clientKey) {
			c.JSON(http.StatusTooManyRequests, types.NewErrorResponse(
				http.StatusTooManyRequests,
				"rate_limit_exceeded",
				"Too many requests. Please try again later.",
			))
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// isRateLimited checks if a client has exceeded the rate limit
func (rl *InMemoryRateLimiter) isRateLimited(clientKey string) bool {
	rl.mutex.RLock()
	record, exists := rl.clients[clientKey]
	rl.mutex.RUnlock()
	
	now := time.Now()
	
	if !exists {
		// First request from this client
		rl.mutex.Lock()
		rl.clients[clientKey] = &ClientRecord{
			Count:       1,
			WindowStart: now,
		}
		rl.mutex.Unlock()
		return false
	}
	
	// Lock the specific client record
	record.mutex.Lock()
	defer record.mutex.Unlock()
	
	// Check if we need to reset the window
	if now.Sub(record.WindowStart) >= rl.config.Window {
		record.Count = 1
		record.WindowStart = now
		return false
	}
	
	// Check if client has exceeded the limit
	if record.Count >= rl.config.Requests {
		return true
	}
	
	// Increment count
	record.Count++
	return false
}

// GetRemainingRequests returns the number of requests remaining for a client
func (rl *InMemoryRateLimiter) GetRemainingRequests(clientKey string) int {
	rl.mutex.RLock()
	record, exists := rl.clients[clientKey]
	rl.mutex.RUnlock()
	
	if !exists {
		return rl.config.Requests
	}
	
	record.mutex.Lock()
	defer record.mutex.Unlock()
	
	now := time.Now()
	
	// Check if window has reset
	if now.Sub(record.WindowStart) >= rl.config.Window {
		return rl.config.Requests
	}
	
	remaining := rl.config.Requests - record.Count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetTimeUntilReset returns the duration until the rate limit window resets
func (rl *InMemoryRateLimiter) GetTimeUntilReset(clientKey string) time.Duration {
	rl.mutex.RLock()
	record, exists := rl.clients[clientKey]
	rl.mutex.RUnlock()
	
	if !exists {
		return 0
	}
	
	record.mutex.Lock()
	defer record.mutex.Unlock()
	
	elapsed := time.Since(record.WindowStart)
	if elapsed >= rl.config.Window {
		return 0
	}
	
	return rl.config.Window - elapsed
}

// cleanup removes expired client records to prevent memory leaks
func (rl *InMemoryRateLimiter) cleanup() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.performCleanup()
		case <-rl.stopCleanup:
			rl.cleanupTicker.Stop()
			return
		}
	}
}

// performCleanup removes expired entries from the clients map
func (rl *InMemoryRateLimiter) performCleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	now := time.Now()
	
	for clientKey, record := range rl.clients {
		record.mutex.Lock()
		if now.Sub(record.WindowStart) >= rl.config.Window*2 { // Keep for 2 windows for safety
			delete(rl.clients, clientKey)
		}
		record.mutex.Unlock()
	}
}

// Close stops the cleanup goroutine and releases resources
func (rl *InMemoryRateLimiter) Close() {
	if rl.stopCleanup != nil {
		close(rl.stopCleanup)
		rl.stopCleanup = nil
	}
}

// NewLoginRateLimiter creates a rate limiter specifically for login endpoints
// Limits to 5 login attempts per 15 minutes per IP address
func NewLoginRateLimiter() *InMemoryRateLimiter {
	config := DefaultRateLimitConfig()
	config.KeyFunc = func(c *gin.Context) string {
		// For login attempts, we might want to combine IP and email for more granular limiting
		// But for simplicity, we'll stick with IP-based limiting
		return c.ClientIP()
	}
	
	return NewInMemoryRateLimiter(config)
}

// NewAPIRateLimiter creates a more permissive rate limiter for general API usage
// Limits to 100 requests per minute per IP address
func NewAPIRateLimiter() *InMemoryRateLimiter {
	config := RateLimitConfig{
		Requests: 100,
		Window:   1 * time.Minute,
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
		Skip: func(c *gin.Context) bool {
			// Skip rate limiting for health check endpoints
			return c.Request.URL.Path == "/health" || c.Request.URL.Path == "/ping"
		},
	}
	
	return NewInMemoryRateLimiter(config)
}

// WithRateLimitHeaders adds rate limit information to response headers
func (rl *InMemoryRateLimiter) WithRateLimitHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientKey := rl.config.KeyFunc(c)
		
		remaining := rl.GetRemainingRequests(clientKey)
		resetTime := rl.GetTimeUntilReset(clientKey)
		
		// Add standard rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.config.Requests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", int64(resetTime.Seconds())))
		
		c.Next()
	}
}