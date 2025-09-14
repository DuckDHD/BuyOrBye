# Frontend Testing Suite

Comprehensive testing suite for the BuyOrBye frontend components, covering templates, JavaScript, accessibility, and performance.

## Overview

This testing suite ensures the frontend meets high standards for:
- **Template Rendering**: Golden file testing with sample data
- **JavaScript Functionality**: HTMX and Alpine.js behavior
- **Accessibility**: WCAG compliance and screen reader support
- **Performance**: Bundle sizes, render times, and optimization

## Quick Start

```bash
# Run all frontend tests
./scripts/test-frontend.sh

# Update golden files
./scripts/test-frontend.sh --update-golden

# Run specific test types
go test ./cmd/web/templates -run TestTemplate -v           # Golden tests
go test ./cmd/web/templates -run TestAccessibility -v     # Accessibility tests
go test ./cmd/web/templates -run TestPerformance -v       # Performance tests
npm test                                                   # JavaScript tests
```

## Test Structure

### 1. Golden Tests (`template_test.go`)

Tests template rendering with sample data and compares output to golden files.

```go
func TestDashboardPage(t *testing.T) {
    user := getSampleUser()
    component := pages.DashboardPage(user, "csrf-token-123")
    html := renderToString(t, component)
    goldenTest(t, "dashboard_page_normal", html)
}
```

**Features:**
- ✅ Page vs partial separation verification
- ✅ CSRF token presence checks
- ✅ Error and empty state handling
- ✅ DOCTYPE validation for full pages
- ✅ Sample data with realistic values

**Golden Files Location:** `cmd/web/templates/testdata/*.golden.html`

### 2. UI Route Tests (`ui_routes_test.go`)

Tests the route organization and HTMX detection logic.

```go
func TestHTMXDetection(t *testing.T) {
    req := makeHTMXRequest("GET", "/test", "")
    // Verify HX-Request header detection
}
```

**Features:**
- ✅ Route structure validation
- ✅ HTMX vs full page request handling
- ✅ Cache header configuration
- ✅ Error response formats
- ✅ CSRF token handling

### 3. JavaScript Tests (`static/js/test/`)

Jest-based tests for HTMX and Alpine.js functionality.

```javascript
describe('HTMX Boot Configuration', () => {
  test('should configure CSRF token injection', () => {
    // Test CSRF header injection
  });
});
```

**HTMX Tests (`htmx.test.js`):**
- ✅ CSRF protection and token injection
- ✅ Loading states and indicators
- ✅ Retry logic for failed requests
- ✅ Global error management
- ✅ Request logging and debugging
- ✅ Form submission handling
- ✅ Modal interactions
- ✅ Performance optimizations (debouncing, caching)

**Alpine Tests (`alpine.test.js`):**
- ✅ UI store management (theme, sidebar, modals, toasts)
- ✅ Magic helpers ($currency, $date, $validate, $clipboard)
- ✅ Utility functions (debounce, throttle, formatting)
- ✅ Decision helpers (colors, confidence levels)
- ✅ Component integration
- ✅ Event handling and keyboard shortcuts

### 4. Accessibility Tests (`accessibility_test.go`)

Comprehensive accessibility compliance testing.

```go
func TestAccessibility_DashboardPage(t *testing.T) {
    checker := NewAccessibilityChecker(html)
    assert.True(t, checker.HasDocumentStructure())
    assert.True(t, checker.HasHeadingStructure())
    assert.True(t, checker.HasARIALabels())
}
```

**Accessibility Checks:**
- ✅ Document structure (DOCTYPE, html, head, body)
- ✅ Language attribute on html element
- ✅ Viewport meta tag
- ✅ Skip links for navigation
- ✅ Proper heading hierarchy (single h1, logical order)
- ✅ ARIA labels and roles
- ✅ Alt text for all images
- ✅ Form labels and associations
- ✅ Keyboard navigation support
- ✅ Color contrast considerations
- ✅ Live regions for dynamic content
- ✅ Focus indicators
- ✅ Screen reader compatibility
- ✅ Semantic HTML elements
- ✅ Error message associations

### 5. Performance Tests (`performance_test.go`)

Performance benchmarks and optimization checks.

```go
func TestPerformance_CSSBundleSize(t *testing.T) {
    sizeKB, err := checker.CheckCSSBundleSize()
    assert.LessOrEqual(t, sizeKB, float64(MaxCSSSizeKB))
}
```

**Performance Targets:**
- ✅ CSS Bundle: < 50KB gzipped
- ✅ JavaScript Bundle: < 100KB gzipped
- ✅ Template Rendering: < 100ms per render
- ✅ Image Optimization: WebP with JPEG fallback
- ✅ Lazy Loading: Below-fold content
- ✅ Critical CSS: Inlined for above-fold content
- ✅ Resource Hints: DNS prefetch, preload
- ✅ Memory Usage: No leaks during rendering
- ✅ Concurrent Rendering: Thread-safe templates

## Running Tests

### Individual Test Suites

```bash
# Golden template tests
go test ./cmd/web/templates -run TestTemplate -v

# Update golden files
go test ./cmd/web/templates -run TestTemplate -update -v

# Accessibility tests
go test ./cmd/web/templates -run TestAccessibility -v

# Performance tests
go test ./cmd/web/templates -run TestPerformance -v

# JavaScript tests
cd cmd/web && npm test

# JavaScript with coverage
cd cmd/web && npm run test:coverage

# Lint checks
cd cmd/web && npm run lint:js && npm run lint:css
```

### Comprehensive Test Runner

```bash
# Run all tests
./scripts/test-frontend.sh

# Verbose output
./scripts/test-frontend.sh --verbose

# Skip specific test types
./scripts/test-frontend.sh --skip-performance --skip-js

# Update golden files
./scripts/test-frontend.sh --update-golden
```

### Test Configuration

**Environment Variables:**
```bash
export UPDATE_GOLDEN=true      # Update golden files
export VERBOSE=true            # Verbose test output
export RUN_GOLDEN=false        # Skip golden tests
export RUN_JAVASCRIPT=false    # Skip JS tests
export RUN_ACCESSIBILITY=false # Skip accessibility tests
export RUN_PERFORMANCE=false   # Skip performance tests
```

## Test Data

### Sample Data Generators

```go
func getSampleUser() *types.UserResponseDTO {
    return &types.UserResponseDTO{
        ID:    "user-123",
        Email: "test@example.com",
        Name:  "Test User",
    }
}

func getSampleFinanceSummary() interface{} {
    return map[string]interface{}{
        "totalIncome":        5000.00,
        "totalExpenses":      3500.00,
        "disposableIncome":   1500.00,
        "savingsRate":        30.0,
        "debtToIncomeRatio":  0.2,
        "financialHealth":    "good",
    }
}
```

### Mock Services

JavaScript tests include comprehensive mocking:
- Global objects (localStorage, sessionStorage, fetch)
- DOM APIs (IntersectionObserver, ResizeObserver)
- Browser APIs (matchMedia, requestAnimationFrame)
- Console methods for testing logging

## Continuous Integration

### Pre-commit Hooks

```bash
# Add to .git/hooks/pre-commit
#!/bin/bash
./scripts/test-frontend.sh --skip-performance
```

### GitHub Actions

```yaml
name: Frontend Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v3
      - uses: actions/setup-node@v3
      - run: ./scripts/test-frontend.sh
      - uses: actions/upload-artifact@v3
        with:
          name: test-coverage
          path: coverage/
```

## Debugging Tests

### Failed Golden Tests

```bash
# Update specific golden file
go test ./cmd/web/templates -run TestDashboardPage -update -v

# Compare differences
diff cmd/web/templates/testdata/dashboard_page_normal.golden.html actual_output.html
```

### JavaScript Test Debugging

```bash
# Run specific test file
npm test -- htmx.test.js

# Run with debugging
npm test -- --verbose --no-coverage

# Watch mode for development
npm run test:watch
```

### Performance Test Analysis

```bash
# Run benchmarks
go test ./cmd/web/templates -bench=BenchmarkTemplate -benchmem

# Profile memory usage
go test ./cmd/web/templates -run TestPerformance_MemoryUsage -memprofile=mem.prof

# Analyze profile
go tool pprof mem.prof
```

## Coverage Reports

### JavaScript Coverage

Generated in `cmd/web/coverage/` with:
- Line coverage
- Branch coverage  
- Function coverage
- Statement coverage

### Go Test Coverage

```bash
go test ./cmd/web/templates -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Best Practices

### Golden Tests

1. **Use realistic data**: Sample data should represent actual application usage
2. **Test edge cases**: Empty states, error conditions, boundary values
3. **Keep files organized**: Use descriptive names for golden files
4. **Regular updates**: Review and update golden files when templates change

### Accessibility Tests

1. **Follow WCAG guidelines**: Test against Web Content Accessibility Guidelines
2. **Use semantic HTML**: Verify proper use of semantic elements
3. **Test with screen readers**: Consider actual screen reader compatibility
4. **Keyboard navigation**: Ensure all functionality is keyboard accessible

### Performance Tests

1. **Set realistic targets**: Base targets on actual application requirements
2. **Test on CI**: Include performance tests in continuous integration
3. **Monitor trends**: Track performance metrics over time
4. **Profile bottlenecks**: Use profiling tools to identify optimization opportunities

### JavaScript Tests

1. **Mock external dependencies**: Isolate units under test
2. **Test user interactions**: Focus on actual user workflows
3. **Cover error scenarios**: Test error handling and edge cases
4. **Keep tests fast**: Use efficient mocks and avoid unnecessary complexity

## Troubleshooting

### Common Issues

**Templates not found:**
```bash
# Generate templates first
templ generate
```

**Node modules missing:**
```bash
cd cmd/web && npm install
```

**Golden files outdated:**
```bash
./scripts/test-frontend.sh --update-golden
```

**Performance tests failing:**
```bash
# Skip performance tests in development
go test ./cmd/web/templates -short
```

### Debug Mode

```bash
# Enable verbose logging
VERBOSE=true ./scripts/test-frontend.sh

# Run individual test with output
go test ./cmd/web/templates -run TestSpecificTest -v
```

## Contributing

When adding new templates or components:

1. **Add golden tests** for new templates
2. **Include accessibility checks** for interactive elements  
3. **Write JavaScript tests** for new functionality
4. **Update performance targets** if adding significant assets
5. **Run full test suite** before submitting PRs

### Test Naming Convention

```go
func TestTemplateName_Scenario_ExpectedResult(t *testing.T)
func TestAccessibility_ComponentName(t *testing.T)
func TestPerformance_MetricName(t *testing.T)
func BenchmarkTemplateName(b *testing.B)
```

This comprehensive testing suite ensures the BuyOrBye frontend maintains high quality, accessibility, and performance standards throughout development.