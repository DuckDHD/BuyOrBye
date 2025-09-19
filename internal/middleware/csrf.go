package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	CSRFTokenLength = 32
	CSRFCookieName  = "csrf_token"
	CSRFHeaderName  = "X-CSRF-Token"
	CSRFContextKey  = "csrf_token"
)

func generateCSRFToken() (string, error) {
	bytes := make([]byte, CSRFTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func ensureCSRFCookie(c *gin.Context) (string, error) {
	if cur, err := c.Cookie(CSRFCookieName); err == nil && cur != "" {
		c.Set(CSRFContextKey, cur) // so templates can render it
		return cur, nil
	}
	tok, err := generateCSRFToken()
	if err != nil {
		return "", err
	}
	secure := c.Request.TLS != nil // use true behind HTTPS/terminating proxy
	httpOnly := true
	c.SetCookie(CSRFCookieName, tok, 3600, "/", "", secure, httpOnly)
	c.Set(CSRFContextKey, tok)
	return tok, nil
}

func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip CSRF on static assets (prevents accidental rotations)
		if strings.HasPrefix(c.Request.URL.Path, "/static/") ||
			strings.HasPrefix(c.Request.URL.Path, "/favicon") {
			c.Next()
			return
		}

		// Make sure there is a cookie token (but don't rotate if present)
		if _, err := ensureCSRFCookie(c); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Enforce only on state-changing requests
		if isStateChangingMethod(c.Request.Method) {
			if !validateCSRFToken(c) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error": "CSRF token missing or invalid",
				})
				return
			}
		}

		c.Next()
	}
}

func isStateChangingMethod(method string) bool {
	switch strings.ToUpper(method) {
	case "POST", "PUT", "DELETE", "PATCH":
		return true
	default:
		return false
	}
}

func validateCSRFToken(c *gin.Context) bool {
	// Get the token from cookie (this is our reference)
	cookieToken, err := c.Cookie(CSRFCookieName)
	if err != nil {
		return false // No CSRF cookie found
	}

	// Check header first (HTMX requests)
	headerToken := c.GetHeader(CSRFHeaderName)
	if headerToken != "" {
		return headerToken == cookieToken
	}

	// Check form field (HTML form submissions)
	formToken := c.PostForm("csrf_token")
	if formToken != "" {
		return formToken == cookieToken
	}

	return false // No token provided in header or form
}
