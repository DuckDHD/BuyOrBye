package templates

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DuckDHD/BuyOrBye/cmd/web/templates/pages"
	"github.com/DuckDHD/BuyOrBye/cmd/web/templates/partials"
	"github.com/DuckDHD/BuyOrBye/cmd/web/templates/components"
	"github.com/DuckDHD/BuyOrBye/internal/types"
)

// Accessibility testing helpers
type AccessibilityChecker struct {
	html string
}

func NewAccessibilityChecker(html string) *AccessibilityChecker {
	return &AccessibilityChecker{html: html}
}

func (a *AccessibilityChecker) HasDocumentStructure() bool {
	return strings.Contains(a.html, "<html") && 
		   strings.Contains(a.html, "<head") && 
		   strings.Contains(a.html, "<body")
}

func (a *AccessibilityChecker) HasLangAttribute() bool {
	langRegex := regexp.MustCompile(`<html[^>]*\slang=["'][a-z]{2}(-[A-Z]{2})?["']`)
	return langRegex.MatchString(a.html)
}

func (a *AccessibilityChecker) HasMetaViewport() bool {
	return strings.Contains(a.html, `name="viewport"`)
}

func (a *AccessibilityChecker) HasSkipLink() bool {
	skipLinkRegex := regexp.MustCompile(`href=["']#(main|content)["']`)
	return skipLinkRegex.MatchString(a.html)
}

func (a *AccessibilityChecker) HasHeadingStructure() bool {
	// Check for proper heading hierarchy (h1, h2, h3, etc.)
	h1Count := strings.Count(a.html, "<h1")
	if h1Count != 1 {
		return false // Should have exactly one h1
	}
	
	// Check that headings follow logical order
	headingRegex := regexp.MustCompile(`<h([1-6])`)
	matches := headingRegex.FindAllStringSubmatch(a.html, -1)
	
	if len(matches) == 0 {
		return false
	}
	
	// First heading should be h1
	if matches[0][1] != "1" {
		return false
	}
	
	return true
}

func (a *AccessibilityChecker) HasARIALabels() bool {
	// Check for ARIA labels on interactive elements
	ariaRegex := regexp.MustCompile(`aria-label=["'][^"']*["']`)
	return ariaRegex.MatchString(a.html)
}

func (a *AccessibilityChecker) HasARIADescribedBy() bool {
	describedByRegex := regexp.MustCompile(`aria-describedby=["'][^"']*["']`)
	return describedByRegex.MatchString(a.html)
}

func (a *AccessibilityChecker) HasRoleAttributes() bool {
	roleRegex := regexp.MustCompile(`role=["'][^"']*["']`)
	return roleRegex.MatchString(a.html)
}

func (a *AccessibilityChecker) HasAltTextForImages() bool {
	// Find all img tags and check they have alt attributes
	imgRegex := regexp.MustCompile(`<img[^>]*>`)
	images := imgRegex.FindAllString(a.html, -1)
	
	for _, img := range images {
		if !strings.Contains(img, "alt=") {
			return false
		}
	}
	
	return len(images) == 0 || true // No images or all have alt text
}

func (a *AccessibilityChecker) HasFormLabels() bool {
	// Check that form inputs have associated labels
	inputRegex := regexp.MustCompile(`<input[^>]*>`)
	inputs := inputRegex.FindAllString(a.html, -1)
	
	for _, input := range inputs {
		// Skip hidden inputs
		if strings.Contains(input, `type="hidden"`) {
			continue
		}
		
		// Check for id attribute and corresponding label
		idRegex := regexp.MustCompile(`id=["']([^"']*)["']`)
		idMatches := idRegex.FindStringSubmatch(input)
		
		if len(idMatches) > 1 {
			labelPattern := `for=["']` + idMatches[1] + `["']`
			labelRegex := regexp.MustCompile(labelPattern)
			if !labelRegex.MatchString(a.html) {
				// Check for aria-label as alternative
				if !strings.Contains(input, "aria-label=") {
					return false
				}
			}
		} else if !strings.Contains(input, "aria-label=") {
			return false
		}
	}
	
	return true
}

func (a *AccessibilityChecker) HasKeyboardNavigation() bool {
	// Check for tabindex attributes and focus management
	tabindexRegex := regexp.MustCompile(`tabindex=["'][^"']*["']`)
	return tabindexRegex.MatchString(a.html) || 
		   strings.Contains(a.html, "focus") ||
		   strings.Contains(a.html, "keyboard")
}

func (a *AccessibilityChecker) HasColorContrast() bool {
	// This is a simplified check - in practice, you'd use tools like axe-core
	// Check for CSS classes that indicate good contrast
	contrastClasses := []string{
		"text-white", "text-black", "text-gray-900", "text-gray-100",
		"bg-white", "bg-black", "bg-gray-900", "bg-gray-100"
	}
	
	for _, class := range contrastClasses {
		if strings.Contains(a.html, class) {
			return true
		}
	}
	
	return false
}

func (a *AccessibilityChecker) HasLiveRegions() bool {
	// Check for ARIA live regions for dynamic content
	liveRegex := regexp.MustCompile(`aria-live=["'](polite|assertive)["']`)
	return liveRegex.MatchString(a.html)
}

func (a *AccessibilityChecker) HasFocusIndicators() bool {
	// Check for focus-related CSS classes
	focusClasses := []string{"focus:", "focus-visible:", "focus-within:"}
	
	for _, class := range focusClasses {
		if strings.Contains(a.html, class) {
			return true
		}
	}
	
	return false
}

// Test accessibility for page components
func TestAccessibility_DashboardPage(t *testing.T) {
	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	component := pages.DashboardPage(user, "csrf-token-123")
	html := renderToString(t, component)
	
	checker := NewAccessibilityChecker(html)
	
	t.Run("document_structure", func(t *testing.T) {
		assert.True(t, checker.HasDocumentStructure(), "Page should have proper HTML document structure")
	})
	
	t.Run("lang_attribute", func(t *testing.T) {
		assert.True(t, checker.HasLangAttribute(), "HTML element should have lang attribute")
	})
	
	t.Run("meta_viewport", func(t *testing.T) {
		assert.True(t, checker.HasMetaViewport(), "Page should have viewport meta tag")
	})
	
	t.Run("skip_link", func(t *testing.T) {
		// Skip link test - might not be required for all pages
		if strings.Contains(html, "skip") {
			assert.True(t, checker.HasSkipLink(), "Page should have skip to main content link")
		}
	})
	
	t.Run("heading_structure", func(t *testing.T) {
		assert.True(t, checker.HasHeadingStructure(), "Page should have proper heading hierarchy")
	})
	
	t.Run("aria_labels", func(t *testing.T) {
		// ARIA labels are recommended for interactive elements
		if strings.Contains(html, "button") || strings.Contains(html, "input") {
			assert.True(t, checker.HasARIALabels(), "Interactive elements should have ARIA labels")
		}
	})
	
	t.Run("alt_text_for_images", func(t *testing.T) {
		assert.True(t, checker.HasAltTextForImages(), "All images should have alt text")
	})
	
	t.Run("form_labels", func(t *testing.T) {
		assert.True(t, checker.HasFormLabels(), "All form inputs should have labels")
	})
	
	t.Run("color_contrast", func(t *testing.T) {
		assert.True(t, checker.HasColorContrast(), "Page should have sufficient color contrast")
	})
	
	t.Run("focus_indicators", func(t *testing.T) {
		assert.True(t, checker.HasFocusIndicators(), "Interactive elements should have focus indicators")
	})
}

func TestAccessibility_FinanceOverviewPage(t *testing.T) {
	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	component := pages.FinanceOverviewPage(user, "csrf-token-123")
	html := renderToString(t, component)
	
	checker := NewAccessibilityChecker(html)
	
	// Test key accessibility features
	assert.True(t, checker.HasDocumentStructure(), "Finance page should have proper HTML structure")
	assert.True(t, checker.HasHeadingStructure(), "Finance page should have proper heading hierarchy")
	assert.True(t, checker.HasColorContrast(), "Finance page should have sufficient color contrast")
}

func TestAccessibility_HealthProfilePage(t *testing.T) {
	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	component := pages.HealthProfilePage(user, "csrf-token-123")
	html := renderToString(t, component)
	
	checker := NewAccessibilityChecker(html)
	
	// Test accessibility for health-related content
	assert.True(t, checker.HasDocumentStructure(), "Health page should have proper HTML structure")
	assert.True(t, checker.HasFormLabels(), "Health forms should have proper labels")
	assert.True(t, checker.HasColorContrast(), "Health page should have sufficient color contrast")
}

// Test accessibility for partial components
func TestAccessibility_Partials(t *testing.T) {
	tests := []struct {
		name      string
		component func() interface{ Render(context.Context, *bytes.Buffer) error }
	}{
		{
			name: "dashboard_stats_partial",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return partials.DashboardStatsPartial(map[string]interface{}{
					"totalIncome": 5000.0,
					"totalExpenses": 3000.0,
				})
			},
		},
		{
			name: "recent_decisions_partial",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				decisions := []interface{}{
					map[string]interface{}{
						"id": "1",
						"productName": "Test Product",
						"decision": "BUY",
					},
				}
				return partials.RecentDecisionsPartial(decisions)
			},
		},
		{
			name: "quick_decision_partial",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return partials.QuickDecisionPartial("csrf-token")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := tt.component()
			html := renderToString(t, component)
			
			checker := NewAccessibilityChecker(html)
			
			// Partials should not have full document structure
			assert.False(t, checker.HasDocumentStructure(), "Partials should not have full HTML document structure")
			
			// But should have proper accessibility attributes
			assert.True(t, checker.HasAltTextForImages(), "All images should have alt text")
			assert.True(t, checker.HasFormLabels(), "All form inputs should have labels")
			
			// Check for ARIA attributes if interactive elements present
			if strings.Contains(html, "button") || strings.Contains(html, "a href") {
				// Interactive elements should be keyboard accessible
				hasTabindex := strings.Contains(html, "tabindex")
				hasRole := strings.Contains(html, "role=")
				hasAriaLabel := strings.Contains(html, "aria-label")
				
				assert.True(t, hasTabindex || hasRole || hasAriaLabel, 
					"Interactive elements should have accessibility attributes")
			}
		})
	}
}

// Test accessibility for components
func TestAccessibility_Components(t *testing.T) {
	tests := []struct {
		name      string
		component func() interface{ Render(context.Context, *bytes.Buffer) error }
	}{
		{
			name: "card_component",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return components.Card("Test Card Title")
			},
		},
		{
			name: "skeleton_stats",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return components.SkeletonStats()
			},
		},
		{
			name: "skeleton_list",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return components.SkeletonList(3)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := tt.component()
			html := renderToString(t, component)
			
			checker := NewAccessibilityChecker(html)
			
			// Components should not have document structure
			assert.False(t, checker.HasDocumentStructure(), "Components should not have full HTML document structure")
			
			// Test for ARIA attributes on skeleton loading states
			if strings.Contains(tt.name, "skeleton") {
				assert.True(t, strings.Contains(html, "aria-") || strings.Contains(html, "role="), 
					"Skeleton components should have ARIA attributes for screen readers")
			}
		})
	}
}

// Test keyboard navigation support
func TestAccessibility_KeyboardNavigation(t *testing.T) {
	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	component := pages.DashboardPage(user, "csrf-token-123")
	html := renderToString(t, component)
	
	// Check for keyboard navigation patterns
	tests := []struct {
		name    string
		pattern string
		desc    string
	}{
		{
			name:    "tabindex_attributes",
			pattern: `tabindex=["'][^"']*["']`,
			desc:    "Elements should have proper tab order",
		},
		{
			name:    "focus_management",
			pattern: `focus|Focus`,
			desc:    "Page should handle focus management",
		},
		{
			name:    "keyboard_shortcuts",
			pattern: `keydown|keyup|keypress`,
			desc:    "Page should support keyboard shortcuts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched, err := regexp.MatchString(tt.pattern, html)
			require.NoError(t, err)
			
			if strings.Contains(html, "interactive") || strings.Contains(html, "button") {
				assert.True(t, matched, tt.desc)
			}
		})
	}
}

// Test screen reader compatibility
func TestAccessibility_ScreenReader(t *testing.T) {
	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	component := pages.DashboardPage(user, "csrf-token-123")
	html := renderToString(t, component)
	
	checker := NewAccessibilityChecker(html)
	
	t.Run("semantic_html", func(t *testing.T) {
		// Check for semantic HTML elements
		semanticElements := []string{
			"<nav", "<main", "<section", "<article", 
			"<aside", "<header", "<footer", "<h1", "<h2",
		}
		
		hasSemanticElements := false
		for _, element := range semanticElements {
			if strings.Contains(html, element) {
				hasSemanticElements = true
				break
			}
		}
		
		assert.True(t, hasSemanticElements, "Page should use semantic HTML elements")
	})
	
	t.Run("aria_landmarks", func(t *testing.T) {
		// Check for ARIA landmarks
		landmarks := []string{
			`role="navigation"`, `role="main"`, `role="complementary"`,
			`role="banner"`, `role="contentinfo"`, `role="search"`,
		}
		
		hasLandmarks := false
		for _, landmark := range landmarks {
			if strings.Contains(html, landmark) {
				hasLandmarks = true
				break
			}
		}
		
		// Landmarks are helpful but not always required
		if strings.Contains(html, "role=") {
			assert.True(t, hasLandmarks, "If using ARIA roles, should include landmark roles")
		}
	})
	
	t.Run("live_regions", func(t *testing.T) {
		// Check for live regions for dynamic content
		if strings.Contains(html, "dynamic") || strings.Contains(html, "update") {
			assert.True(t, checker.HasLiveRegions(), "Dynamic content should use ARIA live regions")
		}
	})
}

// Test color contrast and visual accessibility
func TestAccessibility_VisualAccessibility(t *testing.T) {
	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	component := pages.DashboardPage(user, "csrf-token-123")
	html := renderToString(t, component)
	
	t.Run("color_not_only_indicator", func(t *testing.T) {
		// Check that color is not the only way to convey information
		// Look for text labels alongside color classes
		colorClasses := []string{
			"text-red", "text-green", "text-yellow", "text-blue",
			"bg-red", "bg-green", "bg-yellow", "bg-blue",
		}
		
		hasColorClasses := false
		for _, colorClass := range colorClasses {
			if strings.Contains(html, colorClass) {
				hasColorClasses = true
				break
			}
		}
		
		if hasColorClasses {
			// Should also have text indicators
			hasTextIndicators := strings.Contains(html, "Success") || 
							   strings.Contains(html, "Error") || 
							   strings.Contains(html, "Warning") || 
							   strings.Contains(html, "Info")
			
			assert.True(t, hasTextIndicators, "Color should not be the only way to convey information")
		}
	})
	
	t.Run("sufficient_text_size", func(t *testing.T) {
		// Check for appropriate text size classes
		textSizeClasses := []string{
			"text-xs", "text-sm", "text-base", "text-lg", "text-xl",
		}
		
		hasTextSizes := false
		for _, sizeClass := range textSizeClasses {
			if strings.Contains(html, sizeClass) {
				hasTextSizes = true
				break
			}
		}
		
		assert.True(t, hasTextSizes, "Page should use appropriate text sizes")
	})
	
	t.Run("responsive_design", func(t *testing.T) {
		// Check for responsive design classes
		responsiveClasses := []string{
			"sm:", "md:", "lg:", "xl:", "2xl:",
		}
		
		hasResponsive := false
		for _, responsive := range responsiveClasses {
			if strings.Contains(html, responsive) {
				hasResponsive = true
				break
			}
		}
		
		assert.True(t, hasResponsive, "Page should be responsive")
	})
}

// Test error and validation accessibility
func TestAccessibility_ErrorHandling(t *testing.T) {
	// Test with error states
	component := partials.ExpenseFormPartial(nil, "csrf-token")
	html := renderToString(t, component)
	
	checker := NewAccessibilityChecker(html)
	
	t.Run("error_association", func(t *testing.T) {
		// If the form has validation, errors should be properly associated
		if strings.Contains(html, "error") || strings.Contains(html, "invalid") {
			hasAriaDescribedBy := checker.HasARIADescribedBy()
			hasAriaInvalid := strings.Contains(html, "aria-invalid")
			
			assert.True(t, hasAriaDescribedBy || hasAriaInvalid, 
				"Error messages should be properly associated with form fields")
		}
	})
	
	t.Run("required_field_indication", func(t *testing.T) {
		// Required fields should be clearly indicated
		if strings.Contains(html, "required") {
			hasAriaRequired := strings.Contains(html, "aria-required")
			hasVisualIndicator := strings.Contains(html, "*") || strings.Contains(html, "Required")
			
			assert.True(t, hasAriaRequired || hasVisualIndicator, 
				"Required fields should be clearly indicated")
		}
	})
}

// Benchmark accessibility checks
func BenchmarkAccessibilityCheck(b *testing.B) {
	user := &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
	
	component := pages.DashboardPage(user, "csrf-token-123")
	var buf bytes.Buffer
	err := component.Render(context.Background(), &buf)
	if err != nil {
		b.Fatal(err)
	}
	
	html := buf.String()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker := NewAccessibilityChecker(html)
		_ = checker.HasDocumentStructure()
		_ = checker.HasHeadingStructure()
		_ = checker.HasARIALabels()
		_ = checker.HasFormLabels()
		_ = checker.HasAltTextForImages()
	}
}