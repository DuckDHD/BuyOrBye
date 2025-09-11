package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"

	"github.com/DuckDHD/BuyOrBye/internal/logging"
)

// CSRFConfig holds configuration for CSRF protection middleware
type CSRFConfig struct {
	// Secret key for CSRF token generation (32 bytes recommended)
	Secret []byte
	// Cookie name for CSRF token (default: "_gorilla_csrf")
	CookieName string
	// Cookie domain (optional)
	Domain string
	// Cookie path (default: "/")
	Path string
	// Cookie max age in seconds (default: 12 hours)
	MaxAge int
	// Secure flag - set to true in production with HTTPS
	Secure bool
	// HttpOnly flag - recommended to be true for security
	HttpOnly bool
	// SameSite policy - recommended to be SameSiteStrictMode for security
	SameSite csrf.SameSiteMode
}

// DefaultCSRFConfig returns a default CSRF configuration
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		CookieName: "_gorilla_csrf",
		Path:       "/",
		MaxAge:     12 * 60 * 60,            // 12 hours
		Secure:     true,                    // Set to true for production HTTPS
		HttpOnly:   true,                    // Prevent XSS attacks
		SameSite:   csrf.SameSiteStrictMode, // Strict CSRF protection
	}
}

// NewCSRFMiddleware creates a new CSRF protection middleware for Gin
// Uses gorilla/csrf with secure defaults: SameSite=Strict, HttpOnly=true, Secure=true
func NewCSRFMiddleware(config CSRFConfig) gin.HandlerFunc {
	// Validate secret key
	if len(config.Secret) == 0 {
		// Try to get from environment variable
		secretEnv := os.Getenv("CSRF_SECRET")
		if secretEnv == "" {
			panic("CSRF secret key is required. Set CSRF_SECRET environment variable or provide in config.")
		}
		config.Secret = []byte(secretEnv)
	}

	// Ensure secret is at least 32 bytes for security
	if len(config.Secret) < 32 {
		panic("CSRF secret key must be at least 32 bytes for security")
	}

	// Set defaults for empty values
	if config.CookieName == "" {
		config.CookieName = "_gorilla_csrf"
	}
	if config.Path == "" {
		config.Path = "/"
	}
	if config.MaxAge == 0 {
		config.MaxAge = 12 * 60 * 60 // 12 hours
	}

	// Create gorilla/csrf middleware with configuration
	csrfMiddleware := csrf.Protect(
		config.Secret,
		csrf.CookieName(config.CookieName),
		csrf.Path(config.Path),
		csrf.MaxAge(config.MaxAge),
		csrf.Secure(config.Secure),
		csrf.HttpOnly(config.HttpOnly),
		csrf.SameSite(config.SameSite),
		csrf.Domain(config.Domain),
		// Custom error handler to return JSON error response
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if reason := csrf.FailureReason(r); reason != nil {
				logger := logging.MiddlewareLogger()
				logger.Warn("CSRF blocked request", logging.WithComponent("csrf"), logging.WithError(reason))
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"error":"forbidden","message":"CSRF token invalid or missing","code":403}`))
		})),
	)

	return func(c *gin.Context) {
		// Check if this route is exempt from CSRF protection
		if exempt, exists := c.Get("csrf_exempt"); exists && exempt.(bool) {
			c.Next()
			return
		}

		// Let gorilla/csrf gatekeep the chain.
		// It will call our inner handler ONLY if the token is valid.
		calledNext := false
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calledNext = true
			c.Request = r // r has the context gorilla/csrf injected
			c.Next()      // continue Gin chain on success
		})

		csrfMiddleware(h).ServeHTTP(c.Writer, c.Request)

		if !calledNext {
			// CSRF failed: gorilla already wrote 403. Stop Gin from running handlers.
			c.Abort()
			return
		}
		// On success, we've already run c.Next() above.
	}
}

// GetCSRFToken extracts the CSRF token for the current request
// This token should be included in forms or AJAX requests
func GetCSRFToken(c *gin.Context) string {
	return csrf.Token(c.Request)
}

// CSRFTokenResponse is a helper struct for returning CSRF tokens in API responses
type CSRFTokenResponse struct {
	CSRFToken string `json:"csrf_token"`
}

// GetCSRFTokenHandler returns a Gin handler that provides CSRF tokens
// Useful for SPA applications that need to fetch CSRF tokens via API
func GetCSRFTokenHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := GetCSRFToken(c)
		c.JSON(http.StatusOK, CSRFTokenResponse{
			CSRFToken: token,
		})
	}
}

// CSRFExempt creates a middleware that exempts certain routes from CSRF protection
// This should be used sparingly and only for routes that don't modify state
func CSRFExempt() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF validation by setting a flag in context
		c.Set("csrf_exempt", true)
		c.Next()
	}
}
