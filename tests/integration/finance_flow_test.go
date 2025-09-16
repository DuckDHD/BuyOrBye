//go:build integration
// +build integration

package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/DuckDHD/BuyOrBye/internal/types"
	"github.com/DuckDHD/BuyOrBye/tests/testutils"
)

// FinanceFlowTestSuite represents the comprehensive finance flow test suite
type FinanceFlowTestSuite struct {
	suite.Suite
	server *testutils.TestServer
	client *testutils.HTTPClient
}

// SetupSuite runs before all tests in the suite
func (s *FinanceFlowTestSuite) SetupSuite() {
	testutils.SetupIntegrationTest()
	s.server = testutils.NewTestServer(s.T())
	s.client = testutils.NewHTTPClient(s.server.BaseURL)
}

// TearDownSuite runs after all tests in the suite
func (s *FinanceFlowTestSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}
	testutils.TeardownIntegrationTest()
}

// SetupTest runs before each test
func (s *FinanceFlowTestSuite) SetupTest() {
	// Reset database for clean state
	s.server.ResetDatabase(s.T())
	// Clear any existing token
	s.client.SetAccessToken("")
}

// TestCompleteFinanceFlow tests the complete user journey from registration to financial analysis
func (s *FinanceFlowTestSuite) TestCompleteFinanceFlow() {
	t := s.T()
	
	// Test: User Registration
	user := testutils.NewTestUser("john.doe@example.com", "John Doe", "password123")
	user.Register(t, s.client)
	
	// Verify user can access protected endpoint
	resp, _ := s.client.GET(t, "/api/v1/finance/summary")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Test: Add Income Sources
	incomeData := []types.AddIncomeDTO{
		{
			Source:    "Software Engineer Salary",
			Amount:    8000.00,
			Frequency: "monthly",
		},
		{
			Source:    "Freelance Work",
			Amount:    300.00, // Weekly
			Frequency: "weekly",
		},
		{
			Source:    "Investment Dividends",
			Amount:    45.00, // Daily
			Frequency: "daily",
		},
	}
	
	for _, income := range incomeData {
		resp, body := s.client.POST(t, "/api/v1/finance/income", income)
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to add income: %s", string(body))
	}
	
	// Test: Add Expenses
	expenseData := []types.AddExpenseDTO{
		{
			Category:  "housing",
			Name:      "Monthly Rent",
			Amount:    2500.00,
			Frequency: "monthly",
			IsFixed:   true,
			Priority:  1,
		},
		{
			Category:  "food",
			Name:      "Weekly Groceries",
			Amount:    150.00, // Weekly
			Frequency: "weekly",
			IsFixed:   false,
			Priority:  1,
		},
		{
			Category:  "transport",
			Name:      "Daily Coffee",
			Amount:    5.50, // Daily
			Frequency: "daily",
			IsFixed:   false,
			Priority:  3,
		},
		{
			Category:  "utilities",
			Name:      "Electricity",
			Amount:    120.00,
			Frequency: "monthly",
			IsFixed:   true,
			Priority:  2,
		},
	}
	
	for _, expense := range expenseData {
		resp, body := s.client.POST(t, "/api/v1/finance/expense", expense)
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to add expense: %s", string(body))
	}
	
	// Test: Add Loans
	loanData := []types.AddLoanDTO{
		{
			Lender:           "Chase Bank",
			Type:             "mortgage",
			PrincipalAmount:  450000.00,
			RemainingBalance: 420000.00,
			MonthlyPayment:   2200.00,
			InterestRate:     4.5,
			EndDate:          time.Date(2054, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}
	
	for _, loan := range loanData {
		resp, body := s.client.POST(t, "/api/v1/finance/loan", loan)
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to add loan: %s", string(body))
	}
	
	// Test: Get Financial Summary
	summary := s.client.GetFinanceSummary(t)
	
	// Expected calculations:
	// Monthly Income: 8000 + (300 * 4.33) + (45 * 30) = 8000 + 1299 + 1350 = 10649
	// Monthly Expenses: 2500 + (150 * 4.33) + (5.50 * 30) + 120 = 2500 + 649.5 + 165 + 120 = 3434.5
	// Monthly Loan Payments: 2200
	// Disposable Income: 10649 - 3434.5 - 2200 = 5014.5
	// Debt-to-Income Ratio: 2200 / 10649 = ~0.206 (20.6%)
	
	assert.Greater(t, summary.MonthlyIncome, 10600.0, "Monthly income should be correctly normalized")
	assert.Greater(t, summary.MonthlyExpenses, 3400.0, "Monthly expenses should be correctly normalized")
	assert.Equal(t, 2200.0, summary.MonthlyLoanPayments, "Monthly loan payments should match")
	assert.Greater(t, summary.DisposableIncome, 5000.0, "Should have significant disposable income")
	assert.Less(t, summary.DebtToIncomeRatio, 0.25, "DTI should be good (<25%)")
	assert.Equal(t, "Good", summary.FinancialHealth, "Should have Good financial health")
	
	// Test: Affordability Calculation
	affordability := s.client.GetAffordability(t)
	
	// With good DTI, should be 3x disposable income
	expectedAffordability := summary.DisposableIncome * 3.0
	assert.InDelta(t, expectedAffordability, affordability, 10.0, "Affordability should be calculated correctly")
}

// TestFinancialHealthTransitions tests transitions between different health states
func (s *FinanceFlowTestSuite) TestFinancialHealthTransitions() {
	t := s.T()
	
	// Register user
	user := testutils.NewTestUser("health.test@example.com", "Health Test", "password123")
	user.Register(t, s.client)
	
	// Test 1: Excellent Financial Health
	excellentData := testutils.NewExcellentFinanceData()
	excellentData.AddFinanceData(t, s.client)
	
	summary := s.client.GetFinanceSummary(t)
	assert.Equal(t, "Excellent", summary.FinancialHealth, "Should have Excellent health with low DTI and high savings")
	assert.Greater(t, summary.SavingsRate, 0.20, "Should have >20% savings rate for excellent health")
	assert.Less(t, summary.DebtToIncomeRatio, 0.28, "Should have <28% DTI for excellent health")
	
	// Clear data for next test
	s.server.ResetDatabase(t)
	user.Register(t, s.client)
	
	// Test 2: Good Financial Health
	goodData := testutils.NewBasicFinanceData()
	goodData.AddFinanceData(t, s.client)
	
	summary = s.client.GetFinanceSummary(t)
	assert.Equal(t, "Good", summary.FinancialHealth, "Should have Good health with moderate metrics")
	
	// Clear data for next test
	s.server.ResetDatabase(t)
	user.Register(t, s.client)
	
	// Test 3: Fair Financial Health (High DTI)
	fairData := &testutils.FinanceTestData{
		Incomes: []types.AddIncomeDTO{
			{
				Source:    "Job",
				Amount:    5000.00,
				Frequency: "monthly",
			},
		},
		Expenses: []types.AddExpenseDTO{
			{
				Category:  "housing",
				Name:      "Rent",
				Amount:    1800.00,
				Frequency: "monthly",
				IsFixed:   true,
				Priority:  1,
			},
		},
		Loans: []types.AddLoanDTO{
			{
				Lender:           "Bank",
				Type:             "personal",
				PrincipalAmount:  50000.00,
				RemainingBalance: 45000.00,
				MonthlyPayment:   2000.00, // High loan payment for DTI ~40%
				InterestRate:     8.0,
				EndDate:          time.Date(2030, 1, 15, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	fairData.AddFinanceData(t, s.client)
	
	summary = s.client.GetFinanceSummary(t)
	assert.Equal(t, "Fair", summary.FinancialHealth, "Should have Fair health with high DTI")
	assert.Greater(t, summary.DebtToIncomeRatio, 0.36, "DTI should be >36% for fair health")
	
	// Clear data for next test
	s.server.ResetDatabase(t)
	user.Register(t, s.client)
	
	// Test 4: Poor Financial Health (Overspending)
	poorData := testutils.NewPoorFinanceData()
	poorData.AddFinanceData(t, s.client)
	
	summary = s.client.GetFinanceSummary(t)
	assert.Equal(t, "Poor", summary.FinancialHealth, "Should have Poor health when overspending")
	assert.True(t, summary.DebtToIncomeRatio > 0.50 || summary.DisposableIncome < 0, "Should have high DTI or negative disposable income")
}

// TestBudgetExceededWarnings tests budget exceeded scenarios with realistic data
func (s *FinanceFlowTestSuite) TestBudgetExceededWarnings() {
	t := s.T()
	
	// Register user
	user := testutils.NewTestUser("budget.test@example.com", "Budget Test", "password123")
	user.Register(t, s.client)
	
	// Scenario 1: High-earning professional with lifestyle inflation
	highEarnerData := &testutils.FinanceTestData{
		Incomes: []types.AddIncomeDTO{
			{
				Source:    "Tech Executive Salary",
				Amount:    15000.00,
				Frequency: "monthly",
			},
		},
		Expenses: []types.AddExpenseDTO{
			{
				Category:  "housing",
				Name:      "Luxury Apartment",
				Amount:    5000.00,
				Frequency: "monthly",
				IsFixed:   true,
				Priority:  1,
			},
			{
				Category:  "food",
				Name:      "Dining Out",
				Amount:    2000.00,
				Frequency: "monthly",
				IsFixed:   false,
				Priority:  2,
			},
			{
				Category:  "transport",
				Name:      "Luxury Car Payment",
				Amount:    1200.00,
				Frequency: "monthly",
				IsFixed:   true,
				Priority:  2,
			},
			{
				Category:  "entertainment",
				Name:      "Entertainment",
				Amount:    1500.00,
				Frequency: "monthly",
				IsFixed:   false,
				Priority:  3,
			},
		},
		Loans: []types.AddLoanDTO{
			{
				Lender:           "Private Bank",
				Type:             "mortgage",
				PrincipalAmount:  800000.00,
				RemainingBalance: 750000.00,
				MonthlyPayment:   4500.00,
				InterestRate:     3.8,
				EndDate:          time.Date(2055, 1, 15, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	
	highEarnerData.AddFinanceData(t, s.client)
	summary := s.client.GetFinanceSummary(t)
	
	// Even with high income, high expenses can lead to poor financial health
	// Total expenses: 5000 + 2000 + 1200 + 1500 = 9700
	// Loan payments: 4500
	// Total outgoing: 14200
	// Income: 15000
	// Disposable: 800 (very low for high income)
	
	assert.Less(t, summary.DisposableIncome, 1000.0, "High earner should have low disposable income due to lifestyle inflation")
	assert.Less(t, summary.SavingsRate, 0.10, "Savings rate should be poor despite high income")
	
	// Scenario 2: Young professional overspending on non-essentials
	s.server.ResetDatabase(t)
	user.Register(t, s.client)
	
	youngProfData := &testutils.FinanceTestData{
		Incomes: []types.AddIncomeDTO{
			{
				Source:    "Junior Developer",
				Amount:    4500.00,
				Frequency: "monthly",
			},
		},
		Expenses: []types.AddExpenseDTO{
			{
				Category:  "housing",
				Name:      "Shared Apartment",
				Amount:    1400.00,
				Frequency: "monthly",
				IsFixed:   true,
				Priority:  1,
			},
			{
				Category:  "food",
				Name:      "Food Delivery",
				Amount:    800.00,
				Frequency: "monthly",
				IsFixed:   false,
				Priority:  1,
			},
			{
				Category:  "entertainment",
				Name:      "Subscriptions & Going Out",
				Amount:    600.00,
				Frequency: "monthly",
				IsFixed:   false,
				Priority:  3,
			},
			{
				Category:  "transport",
				Name:      "Uber & Transport",
				Amount:    400.00,
				Frequency: "monthly",
				IsFixed:   false,
				Priority:  2,
			},
		},
		Loans: []types.AddLoanDTO{
			{
				Lender:           "Student Loan Corp",
				Type:             "student",
				PrincipalAmount:  65000.00,
				RemainingBalance: 58000.00,
				MonthlyPayment:   650.00,
				InterestRate:     5.8,
				EndDate:          time.Date(2035, 8, 15, 0, 0, 0, 0, time.UTC),
			},
		},
	}
	
	youngProfData.AddFinanceData(t, s.client)
	summary = s.client.GetFinanceSummary(t)
	
	// Total expenses: 1400 + 800 + 600 + 400 = 3200
	// Loan payments: 650
	// Total outgoing: 3850
	// Income: 4500
	// Disposable: 650
	
	assert.Less(t, summary.DisposableIncome, 700.0, "Young professional should have limited disposable income")
	assert.Greater(t, summary.DebtToIncomeRatio, 0.14, "Should have significant student debt ratio")
	assert.Contains(t, []string{"Fair", "Poor"}, summary.FinancialHealth, "Should have fair or poor health due to overspending")
}

// TestAffordabilityCalculations tests affordability with various real-world scenarios
func (s *FinanceFlowTestSuite) TestAffordabilityCalculations() {
	t := s.T()
	
	scenarios := []struct {
		name                  string
		data                  *testutils.FinanceTestData
		expectedHealthLevel   string
		minAffordability      float64
		maxAffordability      float64
	}{
		{
			name:                "Conservative Saver",
			data:                testutils.NewExcellentFinanceData(),
			expectedHealthLevel: "Excellent",
			minAffordability:    5000.0, // Should have high affordability
			maxAffordability:    30000.0,
		},
		{
			name:                "Average Professional",
			data:                testutils.NewBasicFinanceData(),
			expectedHealthLevel: "Good",
			minAffordability:    2000.0,
			maxAffordability:    15000.0,
		},
		{
			name:                "Struggling Graduate",
			data:                testutils.NewPoorFinanceData(),
			expectedHealthLevel: "Poor",
			minAffordability:    0.0, // May have very low or no affordability
			maxAffordability:    1000.0,
		},
	}
	
	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Fresh database for each scenario
			s.server.ResetDatabase(t)
			
			user := testutils.NewTestUser("afford.test@example.com", "Afford Test", "password123")
			user.Register(t, s.client)
			
			scenario.data.AddFinanceData(t, s.client)
			
			summary := s.client.GetFinanceSummary(t)
			affordability := s.client.GetAffordability(t)
			
			assert.Equal(t, scenario.expectedHealthLevel, summary.FinancialHealth, "Health level should match expected")
			assert.GreaterOrEqual(t, affordability, scenario.minAffordability, "Affordability should be within expected range")
			assert.LessOrEqual(t, affordability, scenario.maxAffordability, "Affordability should be within expected range")
			
			// Affordability should be based on disposable income and health
			if summary.DisposableIncome > 0 {
				multiplier := affordability / summary.DisposableIncome
				assert.Greater(t, multiplier, 0.0, "Affordability multiplier should be positive")
				assert.LessOrEqual(t, multiplier, 3.5, "Affordability multiplier should be reasonable")
			}
		})
	}
}

// TestFrequencyConversions tests various frequency normalizations
func (s *FinanceFlowTestSuite) TestFrequencyConversions() {
	t := s.T()
	
	// Register user
	user := testutils.NewTestUser("freq.test@example.com", "Frequency Test", "password123")
	user.Register(t, s.client)
	
	// Test different income frequencies
	frequencyTestData := []struct {
		amount    float64
		frequency string
		expected  float64 // Expected monthly equivalent
	}{
		{1000.0, "monthly", 1000.0},
		{250.0, "weekly", 1082.5},    // 250 * 4.33
		{50.0, "daily", 1500.0},      // 50 * 30
		{3000.0, "quarterly", 1000.0}, // 3000 / 3
		{6000.0, "semiannual", 1000.0}, // 6000 / 6
		{12000.0, "annual", 1000.0},   // 12000 / 12
	}
	
	for i, testCase := range frequencyTestData {
		income := types.AddIncomeDTO{
			Source:    fmt.Sprintf("Test Income %d", i),
			Amount:    testCase.amount,
			Frequency: testCase.frequency,
		}
		
		resp, body := s.client.POST(t, "/api/v1/finance/income", income)
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to add income: %s", string(body))
	}
	
	summary := s.client.GetFinanceSummary(t)
	
	// All should normalize to approximately the same monthly amount
	expectedTotal := 6000.0 // 6 * 1000 (roughly)
	assert.InDelta(t, expectedTotal, summary.MonthlyIncome, 200.0, "Frequency conversions should normalize correctly")
}

// TestAuthenticationFlow tests authentication and authorization
func (s *FinanceFlowTestSuite) TestAuthenticationFlow() {
	t := s.T()
	
	// Test: Access protected endpoint without token
	resp, body := s.client.GET(t, "/api/v1/finance/summary")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should require authentication")
	
	// Test: Register and login flow
	user := testutils.NewTestUser("auth.test@example.com", "Auth Test", "password123")
	user.Register(t, s.client)
	
	// Test: Access with valid token
	resp, _ = s.client.GET(t, "/api/v1/finance/summary")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Should allow access with valid token")
	
	// Test: User isolation - create second user
	s.client.SetAccessToken("") // Clear token
	user2 := testutils.NewTestUser("auth2.test@example.com", "Auth Test 2", "password123")
	user2.Register(t, s.client)
	
	// Add data for user 2
	income := types.AddIncomeDTO{
		Source:    "User 2 Income",
		Amount:    5000.00,
		Frequency: "monthly",
	}
	resp, body = s.client.POST(t, "/api/v1/finance/income", income)
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "User 2 should be able to add income")
	
	// Switch back to user 1
	s.client.SetAccessToken(user.Token)
	
	// User 1 should not see user 2's data
	resp, body = s.client.GET(t, "/api/v1/finance/income")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	var incomes []types.IncomeResponseDTO
	err := json.Unmarshal(body, &incomes)
	require.NoError(t, err)
	
	// User 1 should have no income records (only user 2 added income)
	assert.Empty(t, incomes, "User 1 should not see user 2's income data")
}

// TestValidationAndErrorHandling tests input validation and error responses
func (s *FinanceFlowTestSuite) TestValidationAndErrorHandling() {
	t := s.T()
	
	// Register user
	user := testutils.NewTestUser("validation.test@example.com", "Validation Test", "password123")
	user.Register(t, s.client)
	
	// Test: Invalid income data
	invalidIncome := map[string]interface{}{
		"source":    "", // Empty source
		"amount":    -100.0, // Negative amount
		"frequency": "invalid", // Invalid frequency
	}
	
	resp, body := s.client.POST(t, "/api/v1/finance/income", invalidIncome)
	testutils.AssertValidationError(t, resp, body, "source")
	
	// Test: Invalid expense data
	invalidExpense := map[string]interface{}{
		"category":  "invalid_category",
		"name":      "x", // Too short
		"amount":    0.0, // Zero amount
		"frequency": "monthly",
		"priority":  5, // Invalid priority
	}
	
	resp, body = s.client.POST(t, "/api/v1/finance/expense", invalidExpense)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	
	// Test: Valid data should work
	validIncome := types.AddIncomeDTO{
		Source:    "Valid Income",
		Amount:    5000.00,
		Frequency: "monthly",
	}
	
	resp, body = s.client.POST(t, "/api/v1/finance/income", validIncome)
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Valid income should be accepted")
	
	// Test: Malformed JSON - create invalid request manually
	invalidJSON := `{"invalid": json}`
	resp, body = s.client.POST(t, "/api/v1/finance/income", json.RawMessage(invalidJSON))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestBusinessRuleValidation tests domain business rules
func (s *FinanceFlowTestSuite) TestBusinessRuleValidation() {
	t := s.T()
	
	// Register user
	user := testutils.NewTestUser("business.test@example.com", "Business Test", "password123")
	user.Register(t, s.client)
	
	// Test 50/30/20 Rule Analysis
	// Add income that should support good budgeting
	income := types.AddIncomeDTO{
		Source:    "Good Salary",
		Amount:    6000.00,
		Frequency: "monthly",
	}
	resp, _ := s.client.POST(t, "/api/v1/finance/income", income)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	
	// Add expenses following 50/30/20 rule
	// 50% for needs: 3000
	// 30% for wants: 1800
	// 20% for savings: 1200 (should result in good health)
	
	expenses := []types.AddExpenseDTO{
		{
			Category:  "housing",
			Name:      "Rent",
			Amount:    2000.00, // Need
			Frequency: "monthly",
			IsFixed:   true,
			Priority:  1,
		},
		{
			Category:  "food",
			Name:      "Groceries",
			Amount:    500.00, // Need
			Frequency: "monthly",
			IsFixed:   false,
			Priority:  1,
		},
		{
			Category:  "utilities",
			Name:      "Utilities",
			Amount:    200.00, // Need
			Frequency: "monthly",
			IsFixed:   true,
			Priority:  1,
		},
		{
			Category:  "entertainment",
			Name:      "Entertainment",
			Amount:    400.00, // Want
			Frequency: "monthly",
			IsFixed:   false,
			Priority:  3,
		},
	}
	
	for _, expense := range expenses {
		resp, body := s.client.POST(t, "/api/v1/finance/expense", expense)
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to add expense: %s", string(body))
	}
	
	summary := s.client.GetFinanceSummary(t)
	
	// Total expenses: 3100 (just over 50% of 6000)
	// Remaining: 2900 (good savings rate)
	assert.Greater(t, summary.SavingsRate, 0.40, "Should have excellent savings rate following 50/30/20 rule")
	assert.Equal(t, "Excellent", summary.FinancialHealth, "Should have excellent health with good budgeting")
}

// TestRealWorldScenarios tests comprehensive real-world financial scenarios
func (s *FinanceFlowTestSuite) TestRealWorldScenarios() {
	t := s.T()
	
	// Scenario 1: New Graduate with Student Loans
	s.Run("New Graduate Scenario", func() {
		s.server.ResetDatabase(t)
		user := testutils.NewTestUser("graduate@example.com", "New Graduate", "password123")
		user.Register(t, s.client)
		
		// Entry-level salary
		income := types.AddIncomeDTO{
			Source:    "Junior Software Developer",
			Amount:    4200.00,
			Frequency: "monthly",
		}
		resp, _ := s.client.POST(t, "/api/v1/finance/income", income)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		
		// Typical graduate expenses
		expenses := []types.AddExpenseDTO{
			{Category: "housing", Name: "Rent", Amount: 1300.00, Frequency: "monthly", IsFixed: true, Priority: 1},
			{Category: "food", Name: "Food", Amount: 400.00, Frequency: "monthly", IsFixed: false, Priority: 1},
			{Category: "transport", Name: "Car Insurance", Amount: 150.00, Frequency: "monthly", IsFixed: true, Priority: 2},
			{Category: "utilities", Name: "Phone", Amount: 80.00, Frequency: "monthly", IsFixed: true, Priority: 2},
		}
		
		for _, expense := range expenses {
			resp, _ := s.client.POST(t, "/api/v1/finance/expense", expense)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		}
		
		// Student loan
		loan := types.AddLoanDTO{
			Lender:           "Federal Student Aid",
			Type:             "student",
			PrincipalAmount:  35000.00,
			RemainingBalance: 33000.00,
			MonthlyPayment:   380.00,
			InterestRate:     5.5,
			EndDate:          time.Date(2034, 5, 15, 0, 0, 0, 0, time.UTC),
		}
		resp, _ = s.client.POST(t, "/api/v1/finance/loan", loan)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		
		summary := s.client.GetFinanceSummary(t)
		affordability := s.client.GetAffordability(t)
		
		// Validate realistic expectations for new graduate
		assert.InDelta(t, 4200.00, summary.MonthlyIncome, 5.0)
		assert.Greater(t, summary.MonthlyExpenses, 1900.0)
		assert.Equal(t, 380.0, summary.MonthlyLoanPayments)
		assert.Greater(t, summary.DisposableIncome, 500.0) // Should have some disposable income
		assert.Less(t, summary.DisposableIncome, 2000.0)   // But not too much
		assert.Contains(t, []string{"Fair", "Good"}, summary.FinancialHealth)
		assert.Greater(t, affordability, 1000.0) // Should afford some purchases
		assert.Less(t, affordability, 5000.0)    // But not expensive items
	})
	
	// Scenario 2: Mid-Career Professional with Family
	s.Run("Mid-Career Professional Scenario", func() {
		s.server.ResetDatabase(t)
		user := testutils.NewTestUser("professional@example.com", "Mid Career Pro", "password123")
		user.Register(t, s.client)
		
		// Dual income household
		incomes := []types.AddIncomeDTO{
			{Source: "Senior Developer", Amount: 9500.00, Frequency: "monthly"},
			{Source: "Partner Income", Amount: 6500.00, Frequency: "monthly"},
			{Source: "Bonus", Amount: 15000.00, Frequency: "annual"},
		}
		
		for _, income := range incomes {
			resp, _ := s.client.POST(t, "/api/v1/finance/income", income)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		}
		
		// Family expenses
		expenses := []types.AddExpenseDTO{
			{Category: "housing", Name: "Mortgage", Amount: 3200.00, Frequency: "monthly", IsFixed: true, Priority: 1},
			{Category: "food", Name: "Family Groceries", Amount: 900.00, Frequency: "monthly", IsFixed: false, Priority: 1},
			{Category: "transport", Name: "Two Car Payments", Amount: 800.00, Frequency: "monthly", IsFixed: true, Priority: 2},
			{Category: "utilities", Name: "All Utilities", Amount: 350.00, Frequency: "monthly", IsFixed: true, Priority: 1},
			{Category: "other", Name: "Childcare", Amount: 1500.00, Frequency: "monthly", IsFixed: true, Priority: 1},
			{Category: "entertainment", Name: "Family Activities", Amount: 400.00, Frequency: "monthly", IsFixed: false, Priority: 3},
		}
		
		for _, expense := range expenses {
			resp, _ := s.client.POST(t, "/api/v1/finance/expense", expense)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		}
		
		// Mortgage
		loan := types.AddLoanDTO{
			Lender:           "Wells Fargo",
			Type:             "mortgage",
			PrincipalAmount:  520000.00,
			RemainingBalance: 480000.00,
			MonthlyPayment:   2400.00, // Part of housing cost but separate loan payment
			InterestRate:     4.2,
			EndDate:          time.Date(2052, 3, 15, 0, 0, 0, 0, time.UTC),
		}
		resp, _ := s.client.POST(t, "/api/v1/finance/loan", loan)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		
		summary := s.client.GetFinanceSummary(t)
		affordability := s.client.GetAffordability(t)
		
		// Mid-career professional should have good financial health
		expectedIncome := 9500.00 + 6500.00 + (15000.00 / 12) // ~17250/month
		assert.InDelta(t, expectedIncome, summary.MonthlyIncome, 100.0)
		assert.Greater(t, summary.DisposableIncome, 3000.0) // Should have good disposable income
		assert.Contains(t, []string{"Good", "Excellent"}, summary.FinancialHealth)
		assert.Greater(t, affordability, 8000.0) // Should afford significant purchases
	})
}

// Run the test suite
func TestFinanceFlowTestSuite(t *testing.T) {
	suite.Run(t, new(FinanceFlowTestSuite))
}