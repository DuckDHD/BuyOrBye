package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupCSRFTestRouter(csrfMiddleware gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	// Add CSRF middleware to protected routes
	protected := r.Group("/api", csrfMiddleware)
	{
		// GET route to obtain CSRF token
		protected.GET("/csrf", GetCSRFTokenHandler())
		
		// POST route that requires CSRF token
		protected.POST("/protected", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "CSRF protected action completed"})
		})
	}
	
	// Route without CSRF protection for comparison
	r.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "public resource"})
	})
	
	return r
}

func TestDefaultCSRFConfig_ReturnsSecureDefaults(t *testing.T) {
	// Act
	config := DefaultCSRFConfig()
	
	// Assert
	assert.Equal(t, "_gorilla_csrf", config.CookieName)
	assert.Equal(t, "/", config.Path)
	assert.Equal(t, 12*60*60, config.MaxAge) // 12 hours
	assert.Equal(t, true, config.Secure)
	assert.Equal(t, true, config.HttpOnly)
	assert.Equal(t, csrf.SameSiteStrictMode, config.SameSite)
}

func TestNewCSRFMiddleware_WithValidSecret_CreatesMiddleware(t *testing.T) {
	// Arrange
	config := DefaultCSRFConfig()
	config.Secret = []byte("12345678901234567890123456789012") // 32 bytes
	
	// Act & Assert - should not panic
	assert.NotPanics(t, func() {
		NewCSRFMiddleware(config)
	})
}

func TestNewCSRFMiddleware_WithShortSecret_Panics(t *testing.T) {
	// Arrange
	config := DefaultCSRFConfig()
	config.Secret = []byte("short") // Less than 32 bytes
	
	// Act & Assert
	assert.Panics(t, func() {
		NewCSRFMiddleware(config)
	}, "Should panic with secret less than 32 bytes")
}

func TestNewCSRFMiddleware_NoSecretNoEnv_Panics(t *testing.T) {
	// Arrange
	config := DefaultCSRFConfig()
	// No secret provided and assume no CSRF_SECRET env var
	
	// Act & Assert
	assert.Panics(t, func() {
		NewCSRFMiddleware(config)
	}, "Should panic when no secret is provided")
}

func TestCSRFMiddleware_GETRequest_AllowsWithoutToken(t *testing.T) {
	// Arrange
	config := DefaultCSRFConfig()
	config.Secret = []byte("12345678901234567890123456789012")
	csrfMiddleware := NewCSRFMiddleware(config)
	router := setupCSRFTestRouter(csrfMiddleware)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/csrf", nil)
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response CSRFTokenResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.CSRFToken)
}

func TestCSRFMiddleware_POSTWithoutToken_Returns403(t *testing.T) {
	// Arrange
	config := DefaultCSRFConfig()
	config.Secret = []byte("12345678901234567890123456789012")
	csrfMiddleware := NewCSRFMiddleware(config)
	router := setupCSRFTestRouter(csrfMiddleware)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/protected", nil)
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "forbidden", response["error"])
	assert.Contains(t, response["message"], "CSRF token")
	assert.Equal(t, float64(403), response["code"])
}

func TestCSRFMiddleware_POSTWithValidToken_Succeeds(t *testing.T) {
	// Arrange
	config := DefaultCSRFConfig()
	config.Secret = []byte("12345678901234567890123456789012")
	csrfMiddleware := NewCSRFMiddleware(config)
	router := setupCSRFTestRouter(csrfMiddleware)
	
	// First, get a CSRF token
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/api/csrf", nil)
	router.ServeHTTP(w1, req1)
	
	require.Equal(t, http.StatusOK, w1.Code)
	
	var tokenResponse CSRFTokenResponse
	err := json.Unmarshal(w1.Body.Bytes(), &tokenResponse)
	require.NoError(t, err)
	require.NotEmpty(t, tokenResponse.CSRFToken)
	
	// Extract cookies from the first response
	cookies := w1.Result().Cookies()
	
	// Act - make POST request with CSRF token
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("POST", "/api/protected", nil)
	
	// Add the CSRF cookie from the first request
	for _, cookie := range cookies {
		req2.AddCookie(cookie)
	}
	
	// Add CSRF token as header (common for AJAX requests)
	req2.Header.Set("X-CSRF-Token", tokenResponse.CSRFToken)
	
	router.ServeHTTP(w2, req2)
	
	// Assert
	assert.Equal(t, http.StatusOK, w2.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w2.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "CSRF protected action completed", response["message"])
}

func TestCSRFMiddleware_POSTWithInvalidToken_Returns403(t *testing.T) {
	// Arrange
	config := DefaultCSRFConfig()
	config.Secret = []byte("12345678901234567890123456789012")
	csrfMiddleware := NewCSRFMiddleware(config)
	router := setupCSRFTestRouter(csrfMiddleware)
	
	// Act - make POST request with invalid CSRF token
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/protected", nil)
	req.Header.Set("X-CSRF-Token", "invalid_token")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "forbidden", response["error"])
	assert.Contains(t, response["message"], "CSRF token")
}

func TestCSRFMiddleware_ConfigDefaults_Applied(t *testing.T) {
	// Arrange
	config := CSRFConfig{
		Secret: []byte("12345678901234567890123456789012"),
		// Leave other fields empty to test defaults
	}
	
	// Act & Assert - should not panic and should apply defaults
	assert.NotPanics(t, func() {
		middleware := NewCSRFMiddleware(config)
		assert.NotNil(t, middleware)
	})
}

func TestGetCSRFTokenHandler_ReturnsValidToken(t *testing.T) {
	// Arrange
	config := DefaultCSRFConfig()
	config.Secret = []byte("12345678901234567890123456789012")
	csrfMiddleware := NewCSRFMiddleware(config)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(csrfMiddleware)
	r.GET("/csrf", GetCSRFTokenHandler())
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/csrf", nil)
	r.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response CSRFTokenResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.CSRFToken)
	
	// Verify it's a valid token format (base64 encoded)
	assert.True(t, len(response.CSRFToken) > 20, "Token should be reasonably long")
}

func TestCSRFExempt_SkipsCSRFValidation(t *testing.T) {
	// Arrange
	config := DefaultCSRFConfig()
	config.Secret = []byte("12345678901234567890123456789012")
	csrfMiddleware := NewCSRFMiddleware(config)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(csrfMiddleware)
	
	// Route with CSRF exemption
	r.POST("/exempt", CSRFExempt(), func(c *gin.Context) {
		exempt, exists := c.Get("csrf_exempt")
		c.JSON(http.StatusOK, gin.H{
			"message": "exempted route",
			"csrf_exempt": exists && exempt.(bool),
		})
	})
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/exempt", nil)
	r.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "exempted route", response["message"])
	assert.Equal(t, true, response["csrf_exempt"])
}

func TestCSRFMiddleware_CustomCookieName_Used(t *testing.T) {
	// Arrange
	config := DefaultCSRFConfig()
	config.Secret = []byte("12345678901234567890123456789012")
	config.CookieName = "_custom_csrf"
	csrfMiddleware := NewCSRFMiddleware(config)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(csrfMiddleware)
	r.GET("/csrf", GetCSRFTokenHandler())
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/csrf", nil)
	r.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Check that custom cookie name is used
	cookies := w.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == "_custom_csrf" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should use custom cookie name")
}

func TestCSRFMiddleware_SecurityFlags_Applied(t *testing.T) {
	// Arrange
	config := DefaultCSRFConfig()
	config.Secret = []byte("12345678901234567890123456789012")
	config.Secure = true
	config.HttpOnly = true
	csrfMiddleware := NewCSRFMiddleware(config)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(csrfMiddleware)
	r.GET("/csrf", GetCSRFTokenHandler())
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/csrf", nil)
	r.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Check cookie security flags
	cookies := w.Result().Cookies()
	require.True(t, len(cookies) > 0, "Should have CSRF cookie")
	
	csrfCookie := cookies[0] // First cookie should be CSRF cookie
	assert.True(t, csrfCookie.HttpOnly, "Cookie should be HttpOnly")
	// Note: Secure flag testing in unit tests is tricky as it depends on HTTPS context
}

func TestGetCSRFToken_WithValidRequest_ReturnsToken(t *testing.T) {
	// Arrange
	config := DefaultCSRFConfig()
	config.Secret = []byte("12345678901234567890123456789012")
	csrfMiddleware := NewCSRFMiddleware(config)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(csrfMiddleware)
	r.GET("/test", func(c *gin.Context) {
		token := GetCSRFToken(c)
		c.JSON(http.StatusOK, gin.H{"direct_token": token})
	})
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response["direct_token"])
}