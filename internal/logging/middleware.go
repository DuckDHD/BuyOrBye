package logging

import (
	"bytes"
	"io"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HTTPLoggingConfig holds configuration for HTTP logging middleware
type HTTPLoggingConfig struct {
	// SkipPaths contains paths to skip logging (e.g., health checks)
	SkipPaths []string
	// LogRequestBody enables request body logging (be careful with sensitive data)
	LogRequestBody bool
	// LogResponseBody enables response body logging (be careful with large responses)
	LogResponseBody bool
	// MaxBodySize limits the size of body to log (in bytes)
	MaxBodySize int64
}

// DefaultHTTPLoggingConfig returns a sensible default configuration
func DefaultHTTPLoggingConfig() HTTPLoggingConfig {
	return HTTPLoggingConfig{
		SkipPaths: []string{
			"/health",
			"/healthz",
			"/ping",
			"/metrics",
		},
		LogRequestBody:  false, // Disabled by default for security
		LogResponseBody: false, // Disabled by default for performance
		MaxBodySize:     1024,  // 1KB limit
	}
}

// HTTPLoggingMiddleware returns a Gin middleware for structured HTTP request logging
func HTTPLoggingMiddleware(config HTTPLoggingConfig) gin.HandlerFunc {
	logger := MiddlewareLogger()
	skipMap := make(map[string]bool, len(config.SkipPaths))
	for _, path := range config.SkipPaths {
		skipMap[path] = true
	}

	return func(c *gin.Context) {
		// Skip logging for specified paths
		if skipMap[c.Request.URL.Path] {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Generate request ID for tracing
		requestID := generateRequestID()
		c.Set("request_id", requestID)

		// Log request body if configured
		var requestBody string
		if config.LogRequestBody && c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, config.MaxBodySize))
			if err == nil {
				requestBody = string(bodyBytes)
				// Restore the body for the actual handler
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Capture response body if configured
		var responseBody string
		if config.LogResponseBody {
			blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
			c.Writer = blw
		}

		// Log request start
		fields := []zap.Field{
			WithRequestID(requestID),
			WithMethod(c.Request.Method),
			WithPath(path),
			WithIP(c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		if raw != "" {
			fields = append(fields, zap.String("query", raw))
		}

		if requestBody != "" {
			fields = append(fields, zap.String("request_body", requestBody))
		}

		logger.Info("HTTP request started", fields...)

		// Process request
		c.Next()

		// Calculate request latency
		latency := time.Since(start)

		// Build response fields
		responseFields := []zap.Field{
			WithRequestID(requestID),
			WithMethod(c.Request.Method),
			WithPath(path),
			WithIP(c.ClientIP()),
			WithHTTPStatus(c.Writer.Status()),
			WithLatency(latency),
			zap.Int("response_size", c.Writer.Size()),
		}

		// Add response body if captured
		if config.LogResponseBody {
			if blw, ok := c.Writer.(*bodyLogWriter); ok {
				responseBody = blw.body.String()
				if len(responseBody) > int(config.MaxBodySize) {
					responseBody = responseBody[:config.MaxBodySize] + "... (truncated)"
				}
				responseFields = append(responseFields, zap.String("response_body", responseBody))
			}
		}

		// Add error information if any
		if len(c.Errors) > 0 {
			responseFields = append(responseFields, zap.String("errors", c.Errors.String()))
		}

		// Determine log level based on status code
		status := c.Writer.Status()
		switch {
		case status >= 500:
			logger.Error("HTTP request completed with server error", responseFields...)
		case status >= 400:
			logger.Warn("HTTP request completed with client error", responseFields...)
		case status >= 300:
			logger.Info("HTTP request completed with redirect", responseFields...)
		default:
			logger.Info("HTTP request completed successfully", responseFields...)
		}
	}
}

// bodyLogWriter wraps gin.ResponseWriter to capture response body
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// generateRequestID generates a simple request ID
// In production, you might want to use a more sophisticated ID generator
func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// RequestIDMiddleware adds request ID to context
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// GetRequestID retrieves request ID from Gin context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// ErrorLoggingMiddleware logs panics and errors
func ErrorLoggingMiddleware() gin.HandlerFunc {
	logger := MiddlewareLogger()
	
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID := GetRequestID(c)
		
		fields := []zap.Field{
			WithRequestID(requestID),
			WithMethod(c.Request.Method),
			WithPath(c.Request.URL.Path),
			WithIP(c.ClientIP()),
			zap.Any("panic", recovered),
		}

		logger.Error("HTTP request panicked", fields...)
		
		c.AbortWithStatus(500)
	})
}

// ContextLogger returns a logger with request context
func ContextLogger(c *gin.Context) *zap.Logger {
	base := GetLogger()
	if base == nil {
		return nil // Return nil if logger not initialized
	}
	
	requestID := GetRequestID(c)
	if requestID != "" {
		return base.With(WithRequestID(requestID))
	}
	return base
}