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

func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if state-changing method first
		if isStateChangingMethod(c.Request.Method) {
			if !validateCSRFToken(c) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		// Generate token for all requests
		token, err := generateCSRFToken()
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Set(CSRFContextKey, token)

		secure := c.Request.TLS != nil
		c.SetCookie(CSRFCookieName, token, 3600, "/", "", secure, true)

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
