package middleware

import (
	"testing"

	"github.com/gorilla/csrf"
	"github.com/stretchr/testify/assert"
)

// Simplified CSRF tests for core functionality
func TestDefaultCSRFConfig_Basic(t *testing.T) {
	config := DefaultCSRFConfig()
	
	assert.Equal(t, "_gorilla_csrf", config.CookieName)
	assert.Equal(t, "/", config.Path)
	assert.Equal(t, 12*60*60, config.MaxAge)
	assert.Equal(t, true, config.Secure)
	assert.Equal(t, true, config.HttpOnly)
	assert.Equal(t, csrf.SameSiteStrictMode, config.SameSite)
}

func TestNewCSRFMiddleware_WithSecret_Works(t *testing.T) {
	config := DefaultCSRFConfig()
	config.Secret = []byte("12345678901234567890123456789012")
	
	assert.NotPanics(t, func() {
		NewCSRFMiddleware(config)
	})
}

func TestNewCSRFMiddleware_ShortSecret_Panics(t *testing.T) {
	config := DefaultCSRFConfig()
	config.Secret = []byte("short")
	
	assert.Panics(t, func() {
		NewCSRFMiddleware(config)
	})
}