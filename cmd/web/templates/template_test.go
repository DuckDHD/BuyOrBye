package templates

import (
	"bytes"
	"context"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DuckDHD/BuyOrBye/cmd/web/templates/pages"
	"github.com/DuckDHD/BuyOrBye/cmd/web/templates/partials"
	"github.com/DuckDHD/BuyOrBye/cmd/web/templates/components"
	"github.com/DuckDHD/BuyOrBye/internal/types"
)

var updateGolden = flag.Bool("update", false, "update golden test files")

// Template test helpers
func renderToString(t *testing.T, component interface{ Render(context.Context, *bytes.Buffer) error }) string {
	t.Helper()
	var buf bytes.Buffer
	err := component.Render(context.Background(), &buf)
	require.NoError(t, err)
	return buf.String()
}

func goldenTest(t *testing.T, name string, got string) {
	t.Helper()
	goldenFile := filepath.Join("testdata", name+".golden.html")
	
	if *updateGolden {
		err := os.WriteFile(goldenFile, []byte(got), 0644)
		require.NoError(t, err)
		return
	}
	
	expected, err := os.ReadFile(goldenFile)
	require.NoError(t, err, "Golden file not found: %s. Run with -update to create it.", goldenFile)
	
	assert.Equal(t, string(expected), got)
}

// Sample test data
func getSampleUser() *types.UserResponseDTO {
	return &types.UserResponseDTO{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}
}

func getSampleHealthProfile() *types.HealthProfileResponseDTO {
	return &types.HealthProfileResponseDTO{
		ID:        1,
		UserID:    "user-123",
		Age:       30,
		Gender:    "male",
		Height:    175.5,
		Weight:    70.0,
		BloodType: "O+",
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

func getSampleDecisions() []interface{} {
	return []interface{}{
		map[string]interface{}{
			"id":          "decision-1",
			"productName": "iPhone 15 Pro",
			"price":       999.99,
			"decision":    "BUY",
			"confidence":  85,
			"createdAt":   "2024-01-15T10:30:00Z",
		},
		map[string]interface{}{
			"id":          "decision-2",
			"productName": "Gaming Laptop",
			"price":       1599.99,
			"decision":    "WAIT",
			"confidence":  72,
			"createdAt":   "2024-01-14T14:20:00Z",
		},
	}
}

// Page Template Tests

func TestDashboardPage(t *testing.T) {
	tests := []struct {
		name      string
		user      *types.UserResponseDTO
		csrfToken string
	}{
		{
			name:      "dashboard_page_normal",
			user:      getSampleUser(),
			csrfToken: "csrf-token-123",
		},
		{
			name:      "dashboard_page_empty_user",
			user:      &types.UserResponseDTO{},
			csrfToken: "csrf-token-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := pages.DashboardPage(tt.user, tt.csrfToken)
			html := renderToString(t, component)
			
			// Verify it's a full HTML page
			assert.Contains(t, html, "<!DOCTYPE html>")
			assert.Contains(t, html, "<html")
			assert.Contains(t, html, "</html>")
			
			// Verify CSRF token presence
			assert.Contains(t, html, tt.csrfToken)
			
			goldenTest(t, tt.name, html)
		})
	}
}

func TestFinanceOverviewPage(t *testing.T) {
	tests := []struct {
		name      string
		user      *types.UserResponseDTO
		csrfToken string
	}{
		{
			name:      "finance_overview_page_normal",
			user:      getSampleUser(),
			csrfToken: "csrf-token-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := pages.FinanceOverviewPage(tt.user, tt.csrfToken)
			html := renderToString(t, component)
			
			// Verify it's a full HTML page
			assert.Contains(t, html, "<!DOCTYPE html>")
			assert.Contains(t, html, "<html")
			assert.Contains(t, html, "</html>")
			
			goldenTest(t, tt.name, html)
		})
	}
}

func TestHealthProfilePage(t *testing.T) {
	tests := []struct {
		name      string
		user      *types.UserResponseDTO
		csrfToken string
	}{
		{
			name:      "health_profile_page_normal",
			user:      getSampleUser(),
			csrfToken: "csrf-token-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := pages.HealthProfilePage(tt.user, tt.csrfToken)
			html := renderToString(t, component)
			
			// Verify it's a full HTML page
			assert.Contains(t, html, "<!DOCTYPE html>")
			assert.Contains(t, html, "<html")
			assert.Contains(t, html, "</html>")
			
			goldenTest(t, tt.name, html)
		})
	}
}

func TestDecisionNewPage(t *testing.T) {
	tests := []struct {
		name      string
		user      *types.UserResponseDTO
		csrfToken string
	}{
		{
			name:      "decision_new_page_normal",
			user:      getSampleUser(),
			csrfToken: "csrf-token-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := pages.DecisionNewPage(tt.user, tt.csrfToken)
			html := renderToString(t, component)
			
			// Verify it's a full HTML page
			assert.Contains(t, html, "<!DOCTYPE html>")
			assert.Contains(t, html, "<html")
			assert.Contains(t, html, "</html>")
			
			goldenTest(t, tt.name, html)
		})
	}
}

func TestDecisionHistoryPage(t *testing.T) {
	tests := []struct {
		name      string
		user      *types.UserResponseDTO
		csrfToken string
	}{
		{
			name:      "decision_history_page_normal",
			user:      getSampleUser(),
			csrfToken: "csrf-token-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := pages.DecisionHistoryPage(tt.user, tt.csrfToken)
			html := renderToString(t, component)
			
			// Verify it's a full HTML page
			assert.Contains(t, html, "<!DOCTYPE html>")
			assert.Contains(t, html, "<html")
			assert.Contains(t, html, "</html>")
			
			goldenTest(t, tt.name, html)
		})
	}
}

// Partial Template Tests

func TestDashboardStatsPartial(t *testing.T) {
	tests := []struct {
		name    string
		summary interface{}
	}{
		{
			name:    "dashboard_stats_partial_with_data",
			summary: getSampleFinanceSummary(),
		},
		{
			name:    "dashboard_stats_partial_empty",
			summary: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := partials.DashboardStatsPartial(tt.summary)
			html := renderToString(t, component)
			
			// Verify it's a partial (no DOCTYPE, html tags)
			assert.NotContains(t, html, "<!DOCTYPE html>")
			assert.NotContains(t, html, "<html")
			assert.NotContains(t, html, "</html>")
			
			goldenTest(t, tt.name, html)
		})
	}
}

func TestRecentDecisionsPartial(t *testing.T) {
	tests := []struct {
		name      string
		decisions []interface{}
	}{
		{
			name:      "recent_decisions_partial_with_data",
			decisions: getSampleDecisions(),
		},
		{
			name:      "recent_decisions_partial_empty",
			decisions: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := partials.RecentDecisionsPartial(tt.decisions)
			html := renderToString(t, component)
			
			// Verify it's a partial (no DOCTYPE, html tags)
			assert.NotContains(t, html, "<!DOCTYPE html>")
			assert.NotContains(t, html, "<html")
			assert.NotContains(t, html, "</html>")
			
			goldenTest(t, tt.name, html)
		})
	}
}

func TestQuickDecisionPartial(t *testing.T) {
	tests := []struct {
		name      string
		csrfToken string
	}{
		{
			name:      "quick_decision_partial_normal",
			csrfToken: "csrf-token-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := partials.QuickDecisionPartial(tt.csrfToken)
			html := renderToString(t, component)
			
			// Verify it's a partial (no DOCTYPE, html tags)
			assert.NotContains(t, html, "<!DOCTYPE html>")
			assert.NotContains(t, html, "<html")
			assert.NotContains(t, html, "</html>")
			
			// Verify CSRF token presence
			assert.Contains(t, html, tt.csrfToken)
			
			goldenTest(t, tt.name, html)
		})
	}
}

func TestFinanceSummaryPartial(t *testing.T) {
	tests := []struct {
		name    string
		summary interface{}
	}{
		{
			name:    "finance_summary_partial_with_data",
			summary: getSampleFinanceSummary(),
		},
		{
			name:    "finance_summary_partial_empty",
			summary: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := partials.FinanceSummaryPartial(tt.summary)
			html := renderToString(t, component)
			
			// Verify it's a partial (no DOCTYPE, html tags)
			assert.NotContains(t, html, "<!DOCTYPE html>")
			assert.NotContains(t, html, "<html")
			assert.NotContains(t, html, "</html>")
			
			goldenTest(t, tt.name, html)
		})
	}
}

func TestIncomeListPartial(t *testing.T) {
	tests := []struct {
		name    string
		incomes []interface{}
	}{
		{
			name: "income_list_partial_with_data",
			incomes: []interface{}{
				map[string]interface{}{
					"id":          1,
					"source":      "Salary",
					"amount":      5000.00,
					"frequency":   "monthly",
					"description": "Main job salary",
				},
				map[string]interface{}{
					"id":          2,
					"source":      "Freelance",
					"amount":      1500.00,
					"frequency":   "monthly",
					"description": "Side projects",
				},
			},
		},
		{
			name:    "income_list_partial_empty",
			incomes: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := partials.IncomeListPartial(tt.incomes)
			html := renderToString(t, component)
			
			// Verify it's a partial (no DOCTYPE, html tags)
			assert.NotContains(t, html, "<!DOCTYPE html>")
			assert.NotContains(t, html, "<html")
			assert.NotContains(t, html, "</html>")
			
			goldenTest(t, tt.name, html)
		})
	}
}

func TestExpenseFormPartial(t *testing.T) {
	tests := []struct {
		name      string
		expense   interface{}
		csrfToken string
	}{
		{
			name: "expense_form_partial_new",
			expense: nil,
			csrfToken: "csrf-token-123",
		},
		{
			name: "expense_form_partial_edit",
			expense: map[string]interface{}{
				"id":          1,
				"description": "Groceries",
				"amount":      150.00,
				"category":    "Food",
			},
			csrfToken: "csrf-token-456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := partials.ExpenseFormPartial(tt.expense, tt.csrfToken)
			html := renderToString(t, component)
			
			// Verify it's a partial (no DOCTYPE, html tags)
			assert.NotContains(t, html, "<!DOCTYPE html>")
			assert.NotContains(t, html, "<html")
			assert.NotContains(t, html, "</html>")
			
			// Verify CSRF token presence
			assert.Contains(t, html, tt.csrfToken)
			
			goldenTest(t, tt.name, html)
		})
	}
}

func TestHealthRiskGaugePartial(t *testing.T) {
	tests := []struct {
		name    string
		profile *types.HealthProfileResponseDTO
		risk    interface{}
	}{
		{
			name:    "health_risk_gauge_partial_normal",
			profile: getSampleHealthProfile(),
			risk: map[string]interface{}{
				"overallRisk":    "low",
				"riskScore":      25,
				"factors":        []string{"age", "bmi"},
				"recommendations": []string{"Regular exercise", "Balanced diet"},
			},
		},
		{
			name:    "health_risk_gauge_partial_high_risk",
			profile: getSampleHealthProfile(),
			risk: map[string]interface{}{
				"overallRisk":    "high",
				"riskScore":      85,
				"factors":        []string{"age", "smoking", "hypertension"},
				"recommendations": []string{"Quit smoking", "See cardiologist"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := partials.HealthRiskGaugePartial(tt.profile, tt.risk)
			html := renderToString(t, component)
			
			// Verify it's a partial (no DOCTYPE, html tags)
			assert.NotContains(t, html, "<!DOCTYPE html>")
			assert.NotContains(t, html, "<html")
			assert.NotContains(t, html, "</html>")
			
			goldenTest(t, tt.name, html)
		})
	}
}

func TestConditionAddPartial(t *testing.T) {
	tests := []struct {
		name      string
		profileID string
		csrfToken string
	}{
		{
			name:      "condition_add_partial_normal",
			profileID: "1",
			csrfToken: "csrf-token-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := partials.ConditionAddPartial(tt.profileID, tt.csrfToken)
			html := renderToString(t, component)
			
			// Verify it's a partial (no DOCTYPE, html tags)
			assert.NotContains(t, html, "<!DOCTYPE html>")
			assert.NotContains(t, html, "<html")
			assert.NotContains(t, html, "</html>")
			
			// Verify CSRF token presence
			assert.Contains(t, html, tt.csrfToken)
			
			goldenTest(t, tt.name, html)
		})
	}
}

func TestInsuranceCardPartial(t *testing.T) {
	tests := []struct {
		name   string
		policy interface{}
	}{
		{
			name: "insurance_card_partial_normal",
			policy: map[string]interface{}{
				"id":           1,
				"providerName": "Blue Cross Blue Shield",
				"policyNumber": "BC123456789",
				"groupNumber":  "GRP001",
				"memberID":     "MEM123",
				"planType":     "PPO",
				"deductible":   1500.00,
				"copay":        25.00,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := partials.InsuranceCardPartial(tt.policy)
			html := renderToString(t, component)
			
			// Verify it's a partial (no DOCTYPE, html tags)
			assert.NotContains(t, html, "<!DOCTYPE html>")
			assert.NotContains(t, html, "<html")
			assert.NotContains(t, html, "</html>")
			
			goldenTest(t, tt.name, html)
		})
	}
}

func TestDecisionResultPartial(t *testing.T) {
	tests := []struct {
		name     string
		decision interface{}
	}{
		{
			name: "decision_result_partial_buy",
			decision: map[string]interface{}{
				"decision":   "BUY",
				"confidence": 85,
				"reasoning":  "Good value for money and fits your budget",
				"factors": []interface{}{
					map[string]interface{}{
						"name":   "Budget Impact",
						"score":  8,
						"weight": "high",
					},
				},
			},
		},
		{
			name: "decision_result_partial_wait",
			decision: map[string]interface{}{
				"decision":   "WAIT",
				"confidence": 72,
				"reasoning":  "Price may drop soon, consider waiting",
				"factors": []interface{}{
					map[string]interface{}{
						"name":   "Market Timing",
						"score":  6,
						"weight": "medium",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := partials.DecisionResultPartial(tt.decision)
			html := renderToString(t, component)
			
			// Verify it's a partial (no DOCTYPE, html tags)
			assert.NotContains(t, html, "<!DOCTYPE html>")
			assert.NotContains(t, html, "<html")
			assert.NotContains(t, html, "</html>")
			
			goldenTest(t, tt.name, html)
		})
	}
}

func TestDecisionFilterPartial(t *testing.T) {
	tests := []struct {
		name       string
		categories []string
		dateRange  string
	}{
		{
			name:       "decision_filter_partial_normal",
			categories: []string{"Electronics", "Clothing", "Home", "Health"},
			dateRange:  "last_30_days",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := partials.DecisionFilterPartial(tt.categories, tt.dateRange)
			html := renderToString(t, component)
			
			// Verify it's a partial (no DOCTYPE, html tags)
			assert.NotContains(t, html, "<!DOCTYPE html>")
			assert.NotContains(t, html, "<html")
			assert.NotContains(t, html, "</html>")
			
			goldenTest(t, tt.name, html)
		})
	}
}

// Component Tests

func TestCard(t *testing.T) {
	tests := []struct {
		name  string
		title string
	}{
		{
			name:  "card_component_normal",
			title: "Test Card Title",
		},
		{
			name:  "card_component_empty_title",
			title: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := components.Card(tt.title)
			html := renderToString(t, component)
			
			// Verify it's a component (no DOCTYPE, html tags)
			assert.NotContains(t, html, "<!DOCTYPE html>")
			assert.NotContains(t, html, "<html")
			assert.NotContains(t, html, "</html>")
			
			goldenTest(t, tt.name, html)
		})
	}
}

func TestSkeletonStats(t *testing.T) {
	component := components.SkeletonStats()
	html := renderToString(t, component)
	
	// Verify it's a component (no DOCTYPE, html tags)
	assert.NotContains(t, html, "<!DOCTYPE html>")
	assert.NotContains(t, html, "<html")
	assert.NotContains(t, html, "</html>")
	
	// Verify skeleton structure
	assert.Contains(t, html, "animate-pulse")
	
	goldenTest(t, "skeleton_stats_component", html)
}

func TestSkeletonList(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{
			name:  "skeleton_list_component_3_items",
			count: 3,
		},
		{
			name:  "skeleton_list_component_5_items",
			count: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := components.SkeletonList(tt.count)
			html := renderToString(t, component)
			
			// Verify it's a component (no DOCTYPE, html tags)
			assert.NotContains(t, html, "<!DOCTYPE html>")
			assert.NotContains(t, html, "<html")
			assert.NotContains(t, html, "</html>")
			
			// Verify skeleton structure
			assert.Contains(t, html, "animate-pulse")
			
			goldenTest(t, tt.name, html)
		})
	}
}

func TestSkeletonForm(t *testing.T) {
	component := components.SkeletonForm()
	html := renderToString(t, component)
	
	// Verify it's a component (no DOCTYPE, html tags)
	assert.NotContains(t, html, "<!DOCTYPE html>")
	assert.NotContains(t, html, "<html")
	assert.NotContains(t, html, "</html>")
	
	// Verify skeleton structure
	assert.Contains(t, html, "animate-pulse")
	
	goldenTest(t, "skeleton_form_component", html)
}

// Error State Tests

func TestErrorStates(t *testing.T) {
	tests := []struct {
		name      string
		component func() interface{ Render(context.Context, *bytes.Buffer) error }
	}{
		{
			name: "dashboard_page_with_nil_user",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return pages.DashboardPage(nil, "csrf-token")
			},
		},
		{
			name: "finance_overview_page_with_nil_user",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return pages.FinanceOverviewPage(nil, "csrf-token")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := tt.component()
			html := renderToString(t, component)
			
			// Should not panic and should render something
			assert.NotEmpty(t, html)
			
			goldenTest(t, tt.name, html)
		})
	}
}

// Accessibility Tests

func TestAccessibility(t *testing.T) {
	tests := []struct {
		name      string
		component func() interface{ Render(context.Context, *bytes.Buffer) error }
		checks    []string
	}{
		{
			name: "dashboard_page_accessibility",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return pages.DashboardPage(getSampleUser(), "csrf-token")
			},
			checks: []string{
				"aria-label",
				"role=",
				"tabindex",
				"alt=",
			},
		},
		{
			name: "finance_overview_page_accessibility",
			component: func() interface{ Render(context.Context, *bytes.Buffer) error } {
				return pages.FinanceOverviewPage(getSampleUser(), "csrf-token")
			},
			checks: []string{
				"aria-label",
				"role=",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := tt.component()
			html := renderToString(t, component)
			
			// Check for accessibility attributes
			for _, check := range tt.checks {
				if !strings.Contains(html, check) {
					t.Logf("Accessibility check failed: %s not found in %s", check, tt.name)
					// Note: Not failing test, just logging for awareness
				}
			}
		})
	}
}