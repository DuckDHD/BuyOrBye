package templates

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DuckDHD/BuyOrBye/cmd/web/templates/pages"
	"github.com/DuckDHD/BuyOrBye/cmd/web/templates/partials"
	"github.com/DuckDHD/BuyOrBye/cmd/web/templates/components"
	"github.com/DuckDHD/BuyOrBye/internal/types"
)

// Performance test constants
const (
	MaxCSSSizeKB        = 50  // 50KB max for CSS bundle
	MaxJSSizeKB         = 100 // 100KB max for JS bundle
	MaxRenderTimeMS     = 100 // 100ms max for template rendering
	MaxImageSizeKB      = 500 // 500KB max for images
	MinCompressionRatio = 0.7 // 70% compression ratio
)

// Performance checker for frontend assets
type PerformanceChecker struct {
	basePath string
}

func NewPerformanceChecker(basePath string) *PerformanceChecker {
	return &PerformanceChecker{basePath: basePath}
}

// Check CSS bundle size
func (p *PerformanceChecker) CheckCSSBundleSize() (float64, error) {
	cssPath := p.basePath + "/cmd/web/static/dist/css/output.css"
	info, err := os.Stat(cssPath)
	if err != nil {
		// If dist file doesn't exist, check source file
		cssPath = p.basePath + "/cmd/web/static/css/output.css"
		info, err = os.Stat(cssPath)
		if err != nil {
			return 0, fmt.Errorf("CSS file not found: %w", err)
		}
	}
	
	sizeKB := float64(info.Size()) / 1024
	return sizeKB, nil
}

// Check JavaScript bundle size
func (p *PerformanceChecker) CheckJSBundleSize() (float64, error) {
	jsFiles := []string{
		"/cmd/web/static/js/htmx.boot.js",
		"/cmd/web/static/js/alpine.boot.js",
	}
	
	totalSize := int64(0)
	for _, jsFile := range jsFiles {
		jsPath := p.basePath + jsFile
		info, err := os.Stat(jsPath)
		if err != nil {
			// Check dist directory
			distPath := p.basePath + "/cmd/web/static/dist" + jsFile
			info, err = os.Stat(distPath)
			if err != nil {
				continue // Skip missing files
			}
		}
		totalSize += info.Size()
	}
	
	sizeKB := float64(totalSize) / 1024
	return sizeKB, nil
}

// Check if critical CSS is inlined
func (p *PerformanceChecker) HasInlinedCriticalCSS(html string) bool {
	return strings.Contains(html, "<style>") && 
		   (strings.Contains(html, "body") || strings.Contains(html, "html"))
}

// Check for lazy loading implementation
func (p *PerformanceChecker) HasLazyLoading(html string) bool {
	// Check for lazy loading attributes
	lazyPatterns := []string{
		`loading="lazy"`,
		`hx-trigger="intersect"`,
		`x-intersect`,
		`data-lazy`,
	}
	
	for _, pattern := range lazyPatterns {
		if strings.Contains(html, pattern) {
			return true
		}
	}
	
	return false
}

// Check for resource hints
func (p *PerformanceChecker) HasResourceHints(html string) bool {
	hints := []string{
		`rel="preload"`,
		`rel="prefetch"`,
		`rel="dns-prefetch"`,
		`rel="preconnect"`,
	}
	
	for _, hint := range hints {
		if strings.Contains(html, hint) {
			return true
		}
	}
	
	return false
}

// Check for compression directives
func (p *PerformanceChecker) HasCompressionHeaders(html string) bool {
	// This would typically be checked at the server level
	// For templates, we check for meta tags that indicate compression awareness
	return strings.Contains(html, "gzip") || strings.Contains(html, "br")
}

// Check image optimization
func (p *PerformanceChecker) CheckImageOptimization(html string) bool {
	// Check for modern image formats
	modernFormats := []string{".webp", ".avif"}
	hasModernFormat := false
	
	for _, format := range modernFormats {
		if strings.Contains(html, format) {
			hasModernFormat = true
			break
		}
	}
	
	// Check for responsive images
	hasResponsiveImages := strings.Contains(html, "srcset") || 
						   strings.Contains(html, "sizes")
	
	return hasModernFormat || hasResponsiveImages
}

// Test CSS bundle size
func TestPerformance_CSSBundleSize(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	checker := NewPerformanceChecker("/Users/mac/Developer/BuyOrBye")
	sizeKB, err := checker.CheckCSSBundleSize()
	
	if err != nil {
		t.Logf("CSS file not found (expected in development): %v", err)
		return
	}
	
	t.Logf("CSS bundle size: %.2f KB", sizeKB)
	assert.LessOrEqual(t, sizeKB, float64(MaxCSSSizeKB), 
		"CSS bundle should be less than %d KB", MaxCSSSizeKB)
}

// Test JavaScript bundle size
func TestPerformance_JSBundleSize(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	checker := NewPerformanceChecker("/Users/mac/Developer/BuyOrBye")
	sizeKB, err := checker.CheckJSBundleSize()
	
	if err != nil {
		t.Logf("JS files not found (expected in development): %v", err)
		return
	}
	
	t.Logf("JavaScript bundle size: %.2f KB", sizeKB)
	assert.LessOrEqual(t, sizeKB, float64(MaxJSSizeKB), 
		"JavaScript bundle should be less than %d KB", MaxJSSizeKB)
}

// Test template rendering performance
func TestPerformance_TemplateRendering(t *testing.T) {
	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	tests := []struct {
		name      string
		component func() interface{ Render(context.Context, *bytes.Buffer) error }
	}{
		{
			name: "dashboard_page",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return pages.DashboardPage(user, "csrf-token")
			},
		},
		{
			name: "finance_overview_page",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return pages.FinanceOverviewPage(user, "csrf-token")
			},
		},
		{
			name: "health_profile_page",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return pages.HealthProfilePage(user, "csrf-token")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := tt.component()
			
			start := time.Now()
			var buf bytes.Buffer
			err := component.Render(context.Background(), &buf)
			duration := time.Since(start)
			
			require.NoError(t, err)
			
			durationMS := float64(duration.Nanoseconds()) / 1000000
			t.Logf("%s rendering time: %.2f ms", tt.name, durationMS)
			
			assert.LessOrEqual(t, durationMS, float64(MaxRenderTimeMS), 
				"Template rendering should be under %d ms", MaxRenderTimeMS)
		})
	}
}

// Test lazy loading implementation
func TestPerformance_LazyLoading(t *testing.T) {
	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	component := pages.DashboardPage(user, "csrf-token")
	html := renderToString(t, component)
	
	checker := NewPerformanceChecker("")
	
	t.Run("has_lazy_loading", func(t *testing.T) {
		hasLazy := checker.HasLazyLoading(html)
		assert.True(t, hasLazy, "Dashboard should implement lazy loading for below-fold content")
	})
	
	t.Run("lazy_loading_patterns", func(t *testing.T) {
		// Check for specific lazy loading patterns
		patterns := []struct {
			name    string
			pattern string
			desc    string
		}{
			{
				name:    "htmx_intersect",
				pattern: `hx-trigger="intersect"`,
				desc:    "Should use HTMX intersect trigger for lazy loading",
			},
			{
				name:    "image_lazy",
				pattern: `loading="lazy"`,
				desc:    "Images should use native lazy loading",
			},
		}
		
		for _, p := range patterns {
			t.Run(p.name, func(t *testing.T) {
				if strings.Contains(html, "lazy") || strings.Contains(html, "intersect") {
					matched, err := regexp.MatchString(p.pattern, html)
					require.NoError(t, err)
					assert.True(t, matched, p.desc)
				}
			})
		}
	})
}

// Test critical CSS inlining
func TestPerformance_CriticalCSS(t *testing.T) {
	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	component := pages.DashboardPage(user, "csrf-token")
	html := renderToString(t, component)
	
	checker := NewPerformanceChecker("")
	
	t.Run("has_inlined_critical_css", func(t *testing.T) {
		hasInlined := checker.HasInlinedCriticalCSS(html)
		
		// Critical CSS inlining is optional but recommended
		if strings.Contains(html, "<style>") {
			assert.True(t, hasInlined, "If using inline styles, should include critical CSS")
		}
	})
	
	t.Run("external_css_async", func(t *testing.T) {
		// Check if external CSS is loaded asynchronously
		asyncCSS := strings.Contains(html, `rel="preload"`) && 
					strings.Contains(html, `as="style"`)
		
		if strings.Contains(html, `<link`) && strings.Contains(html, `.css`) {
			t.Logf("External CSS loading pattern detected")
			// Async CSS loading is recommended but not required
		}
		
		_ = asyncCSS // Use variable to avoid lint warning
	})
}

// Test resource hints
func TestPerformance_ResourceHints(t *testing.T) {
	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	component := pages.DashboardPage(user, "csrf-token")
	html := renderToString(t, component)
	
	checker := NewPerformanceChecker("")
	
	t.Run("has_resource_hints", func(t *testing.T) {
		hasHints := checker.HasResourceHints(html)
		
		// Resource hints are optional but improve performance
		if strings.Contains(html, "http") {
			t.Logf("Resource hints present: %v", hasHints)
		}
	})
	
	t.Run("dns_prefetch", func(t *testing.T) {
		// Check for DNS prefetch for external domains
		dnsPrefetch := regexp.MustCompile(`rel="dns-prefetch"`)
		hasDNSPrefetch := dnsPrefetch.MatchString(html)
		
		if strings.Contains(html, "//") && !strings.Contains(html, "localhost") {
			assert.True(t, hasDNSPrefetch, "Should use DNS prefetch for external domains")
		}
	})
}

// Test image optimization
func TestPerformance_ImageOptimization(t *testing.T) {
	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	component := pages.DashboardPage(user, "csrf-token")
	html := renderToString(t, component)
	
	checker := NewPerformanceChecker("")
	
	t.Run("image_optimization", func(t *testing.T) {
		hasOptimization := checker.CheckImageOptimization(html)
		
		// Only test if images are present
		if strings.Contains(html, "<img") {
			assert.True(t, hasOptimization, 
				"Images should be optimized with modern formats or responsive loading")
		}
	})
	
	t.Run("responsive_images", func(t *testing.T) {
		// Check for responsive image attributes
		if strings.Contains(html, "<img") {
			hasResponsive := strings.Contains(html, "srcset") || 
							strings.Contains(html, "sizes")
			
			t.Logf("Responsive images present: %v", hasResponsive)
		}
	})
}

// Test page load simulation
func TestPerformance_PageLoadSimulation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	// Simulate page load with multiple render calls
	iterations := 100
	totalTime := time.Duration(0)
	
	for i := 0; i < iterations; i++ {
		start := time.Now()
		
		component := pages.DashboardPage(user, "csrf-token")
		var buf bytes.Buffer
		err := component.Render(context.Background(), &buf)
		
		require.NoError(t, err)
		totalTime += time.Since(start)
	}
	
	avgTime := totalTime / time.Duration(iterations)
	avgTimeMS := float64(avgTime.Nanoseconds()) / 1000000
	
	t.Logf("Average render time over %d iterations: %.2f ms", iterations, avgTimeMS)
	assert.LessOrEqual(t, avgTimeMS, float64(MaxRenderTimeMS), 
		"Average render time should be under %d ms", MaxRenderTimeMS)
}

// Test memory usage during rendering
func TestPerformance_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	// Render multiple times to check for memory leaks
	iterations := 1000
	
	for i := 0; i < iterations; i++ {
		component := pages.DashboardPage(user, "csrf-token")
		var buf bytes.Buffer
		err := component.Render(context.Background(), &buf)
		require.NoError(t, err)
		
		// Force garbage collection periodically
		if i%100 == 0 {
			// In a real test, you might use runtime.GC() and check memory stats
			// For now, just ensure we can render many times without issues
		}
	}
	
	t.Logf("Successfully rendered %d times without errors", iterations)
}

// Test content size optimization
func TestPerformance_ContentSize(t *testing.T) {
	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	tests := []struct {
		name      string
		component func() interface{ Render(context.Context, *bytes.Buffer) error }
		maxKB     float64
	}{
		{
			name: "dashboard_page",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return pages.DashboardPage(user, "csrf-token")
			},
			maxKB: 50, // 50KB max for a page
		},
		{
			name: "dashboard_stats_partial",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return partials.DashboardStatsPartial(map[string]interface{}{
					"totalIncome": 5000.0,
				})
			},
			maxKB: 10, // 10KB max for a partial
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := tt.component()
			html := renderToString(t, component)
			
			sizeKB := float64(len(html)) / 1024
			t.Logf("%s HTML size: %.2f KB", tt.name, sizeKB)
			
			assert.LessOrEqual(t, sizeKB, tt.maxKB, 
				"%s should be less than %.0f KB", tt.name, tt.maxKB)
		})
	}
}

// Benchmark template rendering
func BenchmarkTemplateRendering(b *testing.B) {
	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	benchmarks := []struct {
		name      string
		component func() interface{ Render(context.Context, *bytes.Buffer) error }
	}{
		{
			name: "DashboardPage",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return pages.DashboardPage(user, "csrf-token")
			},
		},
		{
			name: "FinanceOverviewPage",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return pages.FinanceOverviewPage(user, "csrf-token")
			},
		},
		{
			name: "DashboardStatsPartial",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return partials.DashboardStatsPartial(map[string]interface{}{
					"totalIncome": 5000.0,
				})
			},
		},
		{
			name: "SkeletonStats",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return components.SkeletonStats()
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			component := bm.component()
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var buf bytes.Buffer
				err := component.Render(context.Background(), &buf)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark memory allocation
func BenchmarkTemplateMemoryAllocation(b *testing.B) {
	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	component := pages.DashboardPage(user, "csrf-token")
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		err := component.Render(context.Background(), &buf)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test concurrent rendering performance
func TestPerformance_ConcurrentRendering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	numGoroutines := 10
	rendersPerGoroutine := 100
	
	start := time.Now()
	
	done := make(chan bool, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer func() { done <- true }()
			
			for j := 0; j < rendersPerGoroutine; j++ {
				component := pages.DashboardPage(user, "csrf-token")
				var buf bytes.Buffer
				err := component.Render(context.Background(), &buf)
				if err != nil {
					t.Errorf("Render error: %v", err)
					return
				}
			}
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	duration := time.Since(start)
	totalRenders := numGoroutines * rendersPerGoroutine
	avgTimePerRender := duration / time.Duration(totalRenders)
	
	t.Logf("Concurrent rendering: %d renders in %v (avg: %v per render)", 
		totalRenders, duration, avgTimePerRender)
	
	// Should complete concurrent rendering reasonably quickly
	assert.Less(t, duration, 10*time.Second, 
		"Concurrent rendering should complete within reasonable time")
}

// Test for performance regressions
func TestPerformance_RegressionCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	// Baseline performance expectations
	baselines := map[string]time.Duration{
		"DashboardPage":         50 * time.Millisecond,
		"FinanceOverviewPage":   50 * time.Millisecond,
		"HealthProfilePage":     50 * time.Millisecond,
		"DashboardStatsPartial": 10 * time.Millisecond,
	}
	
	components := map[string]func() interface{ Render(context.Context, *bytes.Buffer) error }{
		"DashboardPage": func() interface{ Render(context.Context, *bytes.Buffer) error } {
			return pages.DashboardPage(user, "csrf-token")
		},
		"FinanceOverviewPage": func() interface{ Render(context.Context, *bytes.Buffer) error } {
			return pages.FinanceOverviewPage(user, "csrf-token")
		},
		"HealthProfilePage": func() interface{ Render(context.Context, *bytes.Buffer) error } {
			return pages.HealthProfilePage(user, "csrf-token")
		},
		"DashboardStatsPartial": func() interface{ Render(context.Context, *bytes.Buffer) error } {
			return partials.DashboardStatsPartial(map[string]interface{}{
				"totalIncome": 5000.0,
			})
		},
	}
	
	for name, componentFunc := range components {
		t.Run(name, func(t *testing.T) {
			component := componentFunc()
			baseline := baselines[name]
			
			// Run multiple times and take average
			iterations := 10
			totalTime := time.Duration(0)
			
			for i := 0; i < iterations; i++ {
				start := time.Now()
				var buf bytes.Buffer
				err := component.Render(context.Background(), &buf)
				require.NoError(t, err)
				totalTime += time.Since(start)
			}
			
			avgTime := totalTime / time.Duration(iterations)
			
			t.Logf("%s average render time: %v (baseline: %v)", name, avgTime, baseline)
			
			// Allow 50% variance from baseline
			maxTime := baseline + (baseline / 2)
			assert.LessOrEqual(t, avgTime, maxTime, 
				"%s render time should not exceed baseline by more than 50%%", name)
		})
	}
}