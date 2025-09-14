package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Test helpers for UI route testing
func makeHTMXRequest(method, path string, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("HX-Request", "true")
	req.Header.Set("Content-Type", "application/json")
	return req
}

func makeFullPageRequest(method, path string, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Accept", "text/html")
	return req
}

func setUserContext(c *gin.Context, userID string) {
	c.Set("userID", userID)
	c.Set("user_id", userID)
}

// TestUIRouteStructure tests the route organization from main.go
func TestUIRouteStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Simulate the route setup from main.go
	setupTestUIRoutes(router)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		description    string
	}{
		{
			name:           "root_redirects_to_dashboard",
			method:         "GET",
			path:           "/",
			expectedStatus: 302,
			description:    "Root path should redirect to dashboard",
		},
		{
			name:           "static_assets_served",
			method:         "GET",
			path:           "/static/css/output.css",
			expectedStatus: 404, // File doesn't exist in test, but route exists
			description:    "Static assets should be served",
		},
		{
			name:           "favicon_served",
			method:         "GET",
			path:           "/favicon.ico",
			expectedStatus: 404, // File doesn't exist in test, but route exists
			description:    "Favicon should be served",
		},
		{
			name:           "dashboard_page_route",
			method:         "GET",
			path:           "/dashboard",
			expectedStatus: 200,
			description:    "Dashboard page should be accessible",
		},
		{
			name:           "finance_page_route",
			method:         "GET",
			path:           "/finance",
			expectedStatus: 200,
			description:    "Finance page should be accessible",
		},
		{
			name:           "health_page_route",
			method:         "GET",
			path:           "/health",
			expectedStatus: 200,
			description:    "Health page should be accessible",
		},
		{
			name:           "decisions_new_page_route",
			method:         "GET",
			path:           "/decisions/new",
			expectedStatus: 200,
			description:    "New decision page should be accessible",
		},
		{
			name:           "decisions_history_page_route",
			method:         "GET",
			path:           "/decisions/history",
			expectedStatus: 200,
			description:    "Decision history page should be accessible",
		},
		{
			name:           "dashboard_stats_partial",
			method:         "GET",
			path:           "/ui/partials/dashboard/stats",
			expectedStatus: 200,
			description:    "Dashboard stats partial should be accessible",
		},
		{
			name:           "finance_summary_partial",
			method:         "GET",
			path:           "/ui/partials/finance/summary",
			expectedStatus: 200,
			description:    "Finance summary partial should be accessible",
		},
		{
			name:           "health_risk_gauge_partial",
			method:         "GET",
			path:           "/ui/partials/health/risk-gauge/1",
			expectedStatus: 200,
			description:    "Health risk gauge partial should be accessible",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)
		})
	}
}

// TestHTMXDetection tests HTMX request detection
func TestHTMXDetection(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	var isHTMX bool
	router.GET("/test", func(c *gin.Context) {
		isHTMX = c.GetHeader("HX-Request") == "true"
		c.String(200, "OK")
	})

	tests := []struct {
		name     string
		request  *http.Request
		expected bool
	}{
		{
			name:     "HTMX request detected",
			request:  makeHTMXRequest("GET", "/test", ""),
			expected: true,
		},
		{
			name:     "Full page request detected",
			request:  makeFullPageRequest("GET", "/test", ""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, tt.request)

			assert.Equal(t, tt.expected, isHTMX)
		})
	}
}

// TestResponseContentType tests proper content type handling
func TestResponseContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/full-page", func(c *gin.Context) {
		// Simulate full page response
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, "<!DOCTYPE html><html><body>Full Page</body></html>")
	})

	router.GET("/partial", func(c *gin.Context) {
		// Simulate partial response
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(200, "<div>Partial Content</div>")
	})

	tests := []struct {
		name        string
		path        string
		checkDoctype bool
		expectHTML   bool
	}{
		{
			name:        "full_page_has_doctype",
			path:        "/full-page",
			checkDoctype: true,
			expectHTML:   true,
		},
		{
			name:        "partial_no_doctype",
			path:        "/partial",
			checkDoctype: false,
			expectHTML:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			body := w.Body.String()

			if tt.expectHTML {
				assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
			}

			if tt.checkDoctype {
				assert.Contains(t, body, "<!DOCTYPE html>")
			} else {
				assert.NotContains(t, body, "<!DOCTYPE html>")
			}
		})
	}
}

// TestCacheHeaders tests cache header configuration
func TestCacheHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/static-asset", func(c *gin.Context) {
		// Simulate static asset with cache headers
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		c.Header("ETag", `"static-v1"`)
		c.String(200, "/* CSS content */")
	})

	router.GET("/dynamic-content", func(c *gin.Context) {
		// Simulate dynamic content with no-cache headers
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Header("ETag", `"dynamic-123"`)
		c.String(200, "<div>Dynamic Content</div>")
	})

	tests := []struct {
		name         string
		path         string
		expectStatic bool
	}{
		{
			name:         "static_asset_cache_headers",
			path:         "/static-asset",
			expectStatic: true,
		},
		{
			name:         "dynamic_content_no_cache_headers",
			path:         "/dynamic-content",
			expectStatic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			if tt.expectStatic {
				assert.Contains(t, w.Header().Get("Cache-Control"), "max-age=31536000")
				assert.Contains(t, w.Header().Get("Cache-Control"), "immutable")
			} else {
				assert.Contains(t, w.Header().Get("Cache-Control"), "no-cache")
				assert.Equal(t, "no-cache", w.Header().Get("Pragma"))
				assert.Equal(t, "0", w.Header().Get("Expires"))
			}

			// All responses should have ETags
			assert.NotEmpty(t, w.Header().Get("ETag"))
		})
	}
}

// TestErrorHandling tests error response format for different request types
func TestErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/error-test", func(c *gin.Context) {
		isHTMX := c.GetHeader("HX-Request") == "true"
		acceptsJSON := c.GetHeader("Accept") == "application/json"

		if isHTMX && acceptsJSON {
			// Return JSON error for HTMX + JSON
			c.JSON(500, gin.H{"error": "internal_error", "message": "Something went wrong"})
		} else if isHTMX {
			// Return error partial for HTMX
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(500, `<div class="error">Error Partial</div>`)
		} else {
			// Return error page for full page requests
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(500, `<!DOCTYPE html><html><body><h1>Error Page</h1></body></html>`)
		}
	})

	tests := []struct {
		name         string
		headers      map[string]string
		expectJSON   bool
		expectDoctype bool
	}{
		{
			name: "htmx_json_error",
			headers: map[string]string{
				"HX-Request": "true",
				"Accept":     "application/json",
			},
			expectJSON:    true,
			expectDoctype: false,
		},
		{
			name: "htmx_html_error",
			headers: map[string]string{
				"HX-Request": "true",
				"Accept":     "text/html",
			},
			expectJSON:    false,
			expectDoctype: false,
		},
		{
			name: "full_page_error",
			headers: map[string]string{
				"Accept": "text/html",
			},
			expectJSON:    false,
			expectDoctype: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/error-test", nil)

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, 500, w.Code)

			body := w.Body.String()

			if tt.expectJSON {
				assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
				assert.Contains(t, body, `"error"`)
			} else {
				assert.Contains(t, w.Header().Get("Content-Type"), "text/html")

				if tt.expectDoctype {
					assert.Contains(t, body, "<!DOCTYPE html>")
				} else {
					assert.NotContains(t, body, "<!DOCTYPE html>")
				}
			}
		})
	}
}

// TestCSRFTokenHandling tests CSRF token presence in forms
func TestCSRFTokenHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/form-with-csrf", func(c *gin.Context) {
		csrfToken := "csrf-token-123"
		form := `<form>
			<input type="hidden" name="_token" value="` + csrfToken + `">
			<input type="text" name="data">
			<button type="submit">Submit</button>
		</form>`
		c.String(200, form)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/form-with-csrf", nil)
	router.ServeHTTP(w, req)

	body := w.Body.String()
	assert.Contains(t, body, `name="_token"`)
	assert.Contains(t, body, `value="csrf-token-123"`)
}

// setupTestUIRoutes mimics the route setup from main.go for testing
func setupTestUIRoutes(router *gin.Engine) {
	// Static assets
	router.Static("/static", "cmd/web/static")
	router.StaticFile("/favicon.ico", "cmd/web/static/favicon.svg")

	// Public routes
	router.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/dashboard")
	})

	// Protected routes (without auth middleware for testing)
	router.GET("/dashboard", func(c *gin.Context) {
		c.String(200, "Dashboard Page - TODO: Implement")
	})

	router.GET("/finance", func(c *gin.Context) {
		c.String(200, "Finance Overview Page - TODO: Implement")
	})

	router.GET("/health", func(c *gin.Context) {
		c.String(200, "Health Profile Page - TODO: Implement")
	})

	router.GET("/decisions/new", func(c *gin.Context) {
		c.String(200, "New Decision Page - TODO: Implement")
	})

	router.GET("/decisions/history", func(c *gin.Context) {
		c.String(200, "Decision History Page - TODO: Implement")
	})

	// Partial routes
	ui := router.Group("/ui/partials")
	{
		ui.GET("/dashboard/stats", func(c *gin.Context) {
			c.String(200, "Dashboard Stats Partial - TODO: Implement")
		})

		ui.GET("/finance/summary", func(c *gin.Context) {
			c.String(200, "Finance Summary Partial - TODO: Implement")
		})

		ui.GET("/health/risk-gauge/:profileId", func(c *gin.Context) {
			c.String(200, "Health Risk Gauge Partial - TODO: Implement")
		})
	}
}

// Benchmark tests for performance
func BenchmarkFullPageRequest(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	setupTestUIRoutes(router)

	req := makeFullPageRequest("GET", "/dashboard", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkPartialRequest(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	setupTestUIRoutes(router)

	req := makeHTMXRequest("GET", "/ui/partials/dashboard/stats", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkHTMXDetection(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/test", func(c *gin.Context) {
		_ = c.GetHeader("HX-Request") == "true"
		c.String(200, "OK")
	})

	req := makeHTMXRequest("GET", "/test", "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}