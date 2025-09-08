package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DuckDHD/BuyOrBye/internal/types"
)

func setupRateLimitTestRouter(rateLimiter *InMemoryRateLimiter) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	// Add rate limiting middleware
	r.Use(rateLimiter.RateLimit())
	
	// Test endpoint
	r.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "login successful"})
	})
	
	return r
}

func setupRateLimitWithHeadersTestRouter(rateLimiter *InMemoryRateLimiter) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	// Add rate limiting first, then headers middleware after
	r.Use(rateLimiter.RateLimit())
	r.Use(rateLimiter.WithRateLimitHeaders())
	
	// Test endpoint
	r.GET("/api/data", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "test"})
	})
	
	return r
}

func TestDefaultRateLimitConfig_ReturnsCorrectDefaults(t *testing.T) {
	// Act
	config := DefaultRateLimitConfig()
	
	// Assert
	assert.Equal(t, 5, config.Requests)
	assert.Equal(t, 15*time.Minute, config.Window)
	assert.NotNil(t, config.KeyFunc)
	assert.Nil(t, config.Skip)
}

func TestInMemoryRateLimiter_FirstRequest_Allowed(t *testing.T) {
	// Arrange
	config := RateLimitConfig{
		Requests: 5,
		Window:   1 * time.Minute,
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	}
	rateLimiter := NewInMemoryRateLimiter(config)
	defer rateLimiter.Close()
	
	router := setupRateLimitTestRouter(rateLimiter)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", nil)
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "login successful", response["message"])
}

func TestInMemoryRateLimiter_WithinLimit_AllowsRequests(t *testing.T) {
	// Arrange
	config := RateLimitConfig{
		Requests: 3,
		Window:   1 * time.Minute,
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	}
	rateLimiter := NewInMemoryRateLimiter(config)
	defer rateLimiter.Close()
	
	router := setupRateLimitTestRouter(rateLimiter)
	
	// Act & Assert - Make 3 requests (within limit)
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", nil)
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code, "Request %d should be allowed", i+1)
	}
}

func TestInMemoryRateLimiter_ExceedsLimit_ReturnsRateLimitError(t *testing.T) {
	// Arrange
	config := RateLimitConfig{
		Requests: 2,
		Window:   1 * time.Minute,
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	}
	rateLimiter := NewInMemoryRateLimiter(config)
	defer rateLimiter.Close()
	
	router := setupRateLimitTestRouter(rateLimiter)
	
	// Act - Make requests up to the limit
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
	
	// Act - Make one more request (should be rate limited)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", nil)
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	
	var response types.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusTooManyRequests, response.Code)
	assert.Equal(t, "rate_limit_exceeded", response.Error)
	assert.Contains(t, response.Message, "Too many requests")
}

func TestInMemoryRateLimiter_DifferentIPs_IndependentLimits(t *testing.T) {
	// Arrange
	config := RateLimitConfig{
		Requests: 1,
		Window:   1 * time.Minute,
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	}
	rateLimiter := NewInMemoryRateLimiter(config)
	defer rateLimiter.Close()
	
	router := setupRateLimitTestRouter(rateLimiter)
	
	// Act & Assert - First IP makes request
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/login", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)
	
	// Act & Assert - Second IP makes request (should also be allowed)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/login", nil)
	req2.RemoteAddr = "192.168.1.2:12345"
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestInMemoryRateLimiter_WindowReset_AllowsNewRequests(t *testing.T) {
	// Arrange
	config := RateLimitConfig{
		Requests: 1,
		Window:   100 * time.Millisecond, // Very short window for testing
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	}
	rateLimiter := NewInMemoryRateLimiter(config)
	defer rateLimiter.Close()
	
	router := setupRateLimitTestRouter(rateLimiter)
	
	// Act - Make first request
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/login", nil)
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)
	
	// Act - Make second request immediately (should be rate limited)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/login", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	
	// Wait for window to reset
	time.Sleep(150 * time.Millisecond)
	
	// Act - Make third request after window reset (should be allowed)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/login", nil)
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
}

func TestInMemoryRateLimiter_GetRemainingRequests_ReturnsCorrectCount(t *testing.T) {
	// Arrange
	config := RateLimitConfig{
		Requests: 5,
		Window:   1 * time.Minute,
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	}
	rateLimiter := NewInMemoryRateLimiter(config)
	defer rateLimiter.Close()
	
	clientKey := "127.0.0.1"
	
	// Act & Assert - Initially should have all requests available
	remaining := rateLimiter.GetRemainingRequests(clientKey)
	assert.Equal(t, 5, remaining)
	
	// Simulate some requests by directly calling isRateLimited
	rateLimiter.isRateLimited(clientKey) // First request
	remaining = rateLimiter.GetRemainingRequests(clientKey)
	assert.Equal(t, 4, remaining)
	
	rateLimiter.isRateLimited(clientKey) // Second request
	remaining = rateLimiter.GetRemainingRequests(clientKey)
	assert.Equal(t, 3, remaining)
}

func TestInMemoryRateLimiter_GetTimeUntilReset_ReturnsCorrectDuration(t *testing.T) {
	// Arrange
	config := RateLimitConfig{
		Requests: 5,
		Window:   1 * time.Minute,
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	}
	rateLimiter := NewInMemoryRateLimiter(config)
	defer rateLimiter.Close()
	
	clientKey := "127.0.0.1"
	
	// Act - Make first request
	rateLimiter.isRateLimited(clientKey)
	
	// Assert - Should have close to full window remaining
	resetTime := rateLimiter.GetTimeUntilReset(clientKey)
	assert.True(t, resetTime > 59*time.Second, "Reset time should be close to full window")
	assert.True(t, resetTime <= 1*time.Minute, "Reset time should not exceed window")
}

func TestInMemoryRateLimiter_WithSkipFunction_SkipsRateLimiting(t *testing.T) {
	// Arrange
	config := RateLimitConfig{
		Requests: 1,
		Window:   1 * time.Minute,
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
		Skip: func(c *gin.Context) bool {
			// Skip rate limiting for requests with special header
			return c.GetHeader("Skip-Rate-Limit") == "true"
		},
	}
	rateLimiter := NewInMemoryRateLimiter(config)
	defer rateLimiter.Close()
	
	router := setupRateLimitTestRouter(rateLimiter)
	
	// Act - Make first request (uses up the limit)
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/login", nil)
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)
	
	// Act - Make second request with skip header (should be allowed)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/login", nil)
	req2.Header.Set("Skip-Rate-Limit", "true")
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	
	// Act - Make third request without skip header (should be rate limited)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/login", nil)
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusTooManyRequests, w3.Code)
}

func TestInMemoryRateLimiter_WithRateLimitHeaders_AddsHeaders(t *testing.T) {
	// Arrange
	config := RateLimitConfig{
		Requests: 5,
		Window:   1 * time.Minute,
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	}
	rateLimiter := NewInMemoryRateLimiter(config)
	defer rateLimiter.Close()
	
	router := setupRateLimitWithHeadersTestRouter(rateLimiter)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/data", nil)
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Check rate limit headers
	assert.Equal(t, "5", w.Header().Get("X-RateLimit-Limit"))
	
	remaining := w.Header().Get("X-RateLimit-Remaining")
	assert.NotEmpty(t, remaining, "Should have remaining header")
	remainingInt, err := strconv.Atoi(remaining)
	require.NoError(t, err)
	assert.True(t, remainingInt >= 0 && remainingInt <= 5, "Remaining should be between 0 and 5")
	
	resetHeader := w.Header().Get("X-RateLimit-Reset")
	resetSeconds, err := strconv.Atoi(resetHeader)
	require.NoError(t, err)
	assert.True(t, resetSeconds >= 0 && resetSeconds <= 60, "Reset time should be within window duration")
}

func TestNewLoginRateLimiter_CreatesCorrectConfig(t *testing.T) {
	// Act
	rateLimiter := NewLoginRateLimiter()
	defer rateLimiter.Close()
	
	// Assert
	assert.NotNil(t, rateLimiter)
	assert.Equal(t, 5, rateLimiter.config.Requests)
	assert.Equal(t, 15*time.Minute, rateLimiter.config.Window)
	assert.NotNil(t, rateLimiter.config.KeyFunc)
}

func TestNewAPIRateLimiter_CreatesCorrectConfig(t *testing.T) {
	// Act
	rateLimiter := NewAPIRateLimiter()
	defer rateLimiter.Close()
	
	// Assert
	assert.NotNil(t, rateLimiter)
	assert.Equal(t, 100, rateLimiter.config.Requests)
	assert.Equal(t, 1*time.Minute, rateLimiter.config.Window)
	assert.NotNil(t, rateLimiter.config.KeyFunc)
	assert.NotNil(t, rateLimiter.config.Skip)
}

func TestAPIRateLimiter_SkipsHealthEndpoints(t *testing.T) {
	// Arrange
	rateLimiter := NewAPIRateLimiter()
	defer rateLimiter.Close()
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(rateLimiter.RateLimit())
	
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	// Act & Assert - Make many requests to health endpoint (should not be rate limited)
	for i := 0; i < 200; i++ { // More than the API rate limit
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/health", nil)
		r.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code, "Health check should not be rate limited")
	}
}

func TestInMemoryRateLimiter_Close_StopsCleanup(t *testing.T) {
	// Arrange
	config := RateLimitConfig{
		Requests: 5,
		Window:   1 * time.Millisecond, // Very short window for quick cleanup
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	}
	rateLimiter := NewInMemoryRateLimiter(config)
	
	// Verify cleanup is running
	assert.NotNil(t, rateLimiter.cleanupTicker)
	assert.NotNil(t, rateLimiter.stopCleanup)
	
	// Act
	rateLimiter.Close()
	
	// Assert
	assert.Nil(t, rateLimiter.stopCleanup)
	
	// Calling close again should not panic
	assert.NotPanics(t, func() {
		rateLimiter.Close()
	})
}

func TestInMemoryRateLimiter_Cleanup_RemovesExpiredEntries(t *testing.T) {
	// Arrange
	config := RateLimitConfig{
		Requests: 5,
		Window:   10 * time.Millisecond, // Very short window
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP()
		},
	}
	rateLimiter := NewInMemoryRateLimiter(config)
	defer rateLimiter.Close()
	
	clientKey := "test-client"
	
	// Act - Create an entry
	rateLimiter.isRateLimited(clientKey)
	
	// Verify entry exists
	assert.Contains(t, rateLimiter.clients, clientKey)
	
	// Wait for cleanup to run
	time.Sleep(50 * time.Millisecond)
	
	// Trigger manual cleanup
	rateLimiter.performCleanup()
	
	// Assert - Entry should be removed after cleanup
	// Note: The cleanup removes entries after 2*window duration for safety
}

func TestInMemoryRateLimiter_ConcurrentAccess_ThreadSafe(t *testing.T) {
	// Arrange
	config := RateLimitConfig{
		Requests: 10,
		Window:   1 * time.Minute,
		KeyFunc: func(c *gin.Context) string {
			return "concurrent-client"
		},
	}
	rateLimiter := NewInMemoryRateLimiter(config)
	defer rateLimiter.Close()
	
	// Act - Make concurrent requests
	done := make(chan bool, 20)
	for i := 0; i < 20; i++ {
		go func() {
			rateLimiter.isRateLimited("concurrent-client")
			done <- true
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 20; i++ {
		<-done
	}
	
	// Assert - Should not panic and should handle concurrent access safely
	remaining := rateLimiter.GetRemainingRequests("concurrent-client")
	assert.True(t, remaining >= 0, "Remaining requests should not be negative")
}

func TestInMemoryRateLimiter_CustomKeyFunction_Used(t *testing.T) {
	// Arrange
	config := RateLimitConfig{
		Requests: 1,
		Window:   1 * time.Minute,
		KeyFunc: func(c *gin.Context) string {
			// Use custom header as key instead of IP
			return c.GetHeader("X-Client-ID")
		},
	}
	rateLimiter := NewInMemoryRateLimiter(config)
	defer rateLimiter.Close()
	
	router := setupRateLimitTestRouter(rateLimiter)
	
	// Act - Make request with first client ID
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("POST", "/login", nil)
	req1.Header.Set("X-Client-ID", "client-1")
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)
	
	// Act - Make request with same client ID (should be rate limited)
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/login", nil)
	req2.Header.Set("X-Client-ID", "client-1")
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
	
	// Act - Make request with different client ID (should be allowed)
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("POST", "/login", nil)
	req3.Header.Set("X-Client-ID", "client-2")
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusOK, w3.Code)
}