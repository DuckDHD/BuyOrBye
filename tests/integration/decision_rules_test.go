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

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/types"
	"github.com/DuckDHD/BuyOrBye/tests/testutils"
)

// DecisionRulesTestSuite tests business rule enforcement across different scenarios
type DecisionRulesTestSuite struct {
	suite.Suite
	server   *testutils.TestServer
	client   *testutils.HTTPClient
	mockAI   *MockAIClient
}

// SetupSuite runs before all tests in the suite
func (s *DecisionRulesTestSuite) SetupSuite() {
	testutils.SetupIntegrationTest()
	s.server = testutils.NewTestServer(s.T())
	s.client = testutils.NewHTTPClient(s.server.BaseURL)
	s.mockAI = NewMockAIClient()
}

// TearDownSuite runs after all tests in the suite
func (s *DecisionRulesTestSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}
	testutils.TeardownIntegrationTest()
}

// SetupTest runs before each test
func (s *DecisionRulesTestSuite) SetupTest() {
	s.server.ResetDatabase(s.T())
	s.client.SetAccessToken("")
	s.mockAI.Reset()
}

// TestEmergencyFundRule tests the emergency fund requirement enforcement
func (s *DecisionRulesTestSuite) TestEmergencyFundRule() {
	t := s.T()

	testCases := []struct {
		name                    string
		monthlyIncome          float64
		monthlyExpenses        float64
		emergencyFundAmount    float64
		itemCost               float64
		category               string
		urgency                string
		expectedDecision       string
		shouldAllowPurchase    bool
	}{
		{
			name:                    "Good Emergency Fund - Allow Purchase",
			monthlyIncome:          6000.0,
			monthlyExpenses:        3000.0,
			emergencyFundAmount:    15000.0, // 5 months of expenses
			itemCost:               1000.0,
			category:               "electronics",
			urgency:                "medium",
			expectedDecision:       "BUY",
			shouldAllowPurchase:    true,
		},
		{
			name:                    "Low Emergency Fund - Reject Non-Critical",
			monthlyIncome:          5000.0,
			monthlyExpenses:        4000.0,
			emergencyFundAmount:    8000.0, // 2 months of expenses (< 3 months)
			itemCost:               800.0,
			category:               "entertainment",
			urgency:                "low",
			expectedDecision:       "BYE",
			shouldAllowPurchase:    false,
		},
		{
			name:                    "Low Emergency Fund - Allow Critical Health",
			monthlyIncome:          4000.0,
			monthlyExpenses:        3500.0,
			emergencyFundAmount:    7000.0, // 2 months of expenses
			itemCost:               600.0,
			category:               "health",
			urgency:                "critical",
			expectedDecision:       "BUY",
			shouldAllowPurchase:    true,
		},
		{
			name:                    "No Emergency Fund - Reject Everything Except Critical Health",
			monthlyIncome:          3500.0,
			monthlyExpenses:        3200.0,
			emergencyFundAmount:    500.0, // Less than 1 month
			itemCost:               400.0,
			category:               "clothing",
			urgency:                "medium",
			expectedDecision:       "BYE",
			shouldAllowPurchase:    false,
		},
		{
			name:                    "Borderline Emergency Fund - Wait for Non-Essential",
			monthlyIncome:          5500.0,
			monthlyExpenses:        3800.0,
			emergencyFundAmount:    11000.0, // Almost 3 months (2.9 months)
			itemCost:               1200.0,
			category:               "electronics",
			urgency:                "medium",
			expectedDecision:       "WAIT",
			shouldAllowPurchase:    false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset for each test case
			s.server.ResetDatabase(t)
			s.mockAI.Reset()

			// Register user
			user := testutils.NewTestUser("emergency.fund@example.com", "Emergency Fund Test", "password123")
			user.Register(t, s.client)

			// Create financial scenario with specific emergency fund level
			financeData := &testutils.FinanceTestData{
				Incomes: []types.AddIncomeDTO{
					{Source: "Primary Job", Amount: tc.monthlyIncome, Frequency: "monthly"},
				},
				Expenses: []types.AddExpenseDTO{
					{Category: "housing", Name: "Rent", Amount: tc.monthlyExpenses * 0.6, Frequency: "monthly", IsFixed: true, Priority: 1},
					{Category: "food", Name: "Food", Amount: tc.monthlyExpenses * 0.4, Frequency: "monthly", IsFixed: false, Priority: 1},
				},
				Loans: []types.AddLoanDTO{},
				// Note: Emergency fund is calculated from savings, so we need to ensure savings amount
			}
			financeData.AddFinanceData(t, s.client)

			// Add basic health data
			healthData := testutils.NewBasicHealthData()
			if tc.category == "health" && tc.urgency == "critical" {
				// Add critical health condition
				healthData = testutils.NewCriticalHealthData()
			}
			healthData.AddHealthData(t, s.client)

			// Configure AI to always recommend BUY (business rules should override if needed)
			s.mockAI.SetNextResponse(domain.AIResponse{
				Decision:   "BUY",
				Confidence: 0.85,
				Reasoning:  "AI recommends purchase based on context",
				Factors: []domain.DecisionFactor{
					{Category: "ai_analysis", Impact: "positive", Weight: 0.8, Description: "AI analysis supports purchase"},
				},
			})

			// Make decision request
			intent := types.PurchaseIntentDTO{
				ItemName:  fmt.Sprintf("Emergency Fund Test - %s", tc.name),
				ItemCost:  tc.itemCost,
				Category:  tc.category,
				Urgency:   tc.urgency,
				Frequency: "one_time",
				Purpose:   "Testing emergency fund business rule",
			}

			resp, body := s.client.POST(t, "/api/v1/decision/evaluate", intent)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var decision types.DecisionResponseDTO
			err := json.Unmarshal(body, &decision)
			require.NoError(t, err)

			// Verify business rule enforcement
			assert.Equal(t, tc.expectedDecision, decision.Decision, "Emergency fund rule should be enforced")

			if !tc.shouldAllowPurchase && tc.expectedDecision == "BYE" {
				assert.Contains(t, decision.PrimaryReason, "emergency fund", "Should mention emergency fund in reasoning")
			}

			if tc.category == "health" && tc.urgency == "critical" && tc.shouldAllowPurchase {
				assert.Contains(t, decision.PrimaryReason, "critical health", "Should mention health priority override")
			}
		})
	}
}

// TestDTIRatioLimits tests debt-to-income ratio enforcement
func (s *DecisionRulesTestSuite) TestDTIRatioLimits() {
	t := s.T()

	testCases := []struct {
		name               string
		monthlyIncome      float64
		monthlyLoanPayment float64
		itemCost           float64
		category           string
		expectedDecision   string
		expectedDTIRatio   float64
	}{
		{
			name:               "Low DTI - Allow Purchase",
			monthlyIncome:      8000.0,
			monthlyLoanPayment: 1500.0, // 18.75% DTI - Good
			itemCost:           1200.0,
			category:           "electronics",
			expectedDecision:   "BUY",
			expectedDTIRatio:   0.1875,
		},
		{
			name:               "Moderate DTI - Allow with Caution",
			monthlyIncome:      6000.0,
			monthlyLoanPayment: 2000.0, // 33.33% DTI - Moderate
			itemCost:           800.0,
			category:           "electronics",
			expectedDecision:   "BUY", // Still acceptable but with warnings
			expectedDTIRatio:   0.333,
		},
		{
			name:               "High DTI - Wait Recommendation",
			monthlyIncome:      5000.0,
			monthlyLoanPayment: 2600.0, // 52% DTI - High
			itemCost:           700.0,
			category:           "entertainment",
			expectedDecision:   "WAIT",
			expectedDTIRatio:   0.52,
		},
		{
			name:               "High DTI - Health Exception",
			monthlyIncome:      4500.0,
			monthlyLoanPayment: 2500.0, // 55.5% DTI - Very High
			itemCost:           500.0,
			category:           "health",
			expectedDecision:   "BUY", // Health needs override DTI concerns
			expectedDTIRatio:   0.555,
		},
		{
			name:               "Very High DTI - Reject Non-Essential",
			monthlyIncome:      4000.0,
			monthlyLoanPayment: 2800.0, // 70% DTI - Extremely High
			itemCost:           600.0,
			category:           "clothing",
			expectedDecision:   "BYE",
			expectedDTIRatio:   0.70,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset for each test case
			s.server.ResetDatabase(t)
			s.mockAI.Reset()

			// Register user
			user := testutils.NewTestUser("dti.test@example.com", "DTI Test", "password123")
			user.Register(t, s.client)

			// Create financial scenario with specific DTI ratio
			financeData := &testutils.FinanceTestData{
				Incomes: []types.AddIncomeDTO{
					{Source: "Primary Job", Amount: tc.monthlyIncome, Frequency: "monthly"},
				},
				Expenses: []types.AddExpenseDTO{
					{Category: "housing", Name: "Rent", Amount: tc.monthlyIncome * 0.25, Frequency: "monthly", IsFixed: true, Priority: 1},
					{Category: "food", Name: "Food", Amount: tc.monthlyIncome * 0.15, Frequency: "monthly", IsFixed: false, Priority: 1},
				},
				Loans: []types.AddLoanDTO{
					{
						Lender:           "Test Bank",
						Type:             "personal",
						PrincipalAmount:  tc.monthlyLoanPayment * 60, // 5 years of payments
						RemainingBalance: tc.monthlyLoanPayment * 50,
						MonthlyPayment:   tc.monthlyLoanPayment,
						InterestRate:     7.5,
						EndDate:          time.Date(2029, 12, 31, 0, 0, 0, 0, time.UTC),
					},
				},
			}
			financeData.AddFinanceData(t, s.client)

			// Verify DTI calculation
			financeSummary := s.client.GetFinanceSummary(t)
			assert.InDelta(t, tc.expectedDTIRatio, financeSummary.DebtToIncomeRatio, 0.02, "DTI ratio should be calculated correctly")

			// Add health data (critical if health category)
			var healthData *testutils.HealthTestData
			if tc.category == "health" {
				healthData = testutils.NewCriticalHealthData()
			} else {
				healthData = testutils.NewBasicHealthData()
			}
			healthData.AddHealthData(t, s.client)

			// Configure AI to recommend BUY
			s.mockAI.SetNextResponse(domain.AIResponse{
				Decision:   "BUY",
				Confidence: 0.80,
				Reasoning:  "AI analysis supports purchase",
			})

			// Make decision request
			intent := types.PurchaseIntentDTO{
				ItemName:  fmt.Sprintf("DTI Test - %s", tc.name),
				ItemCost:  tc.itemCost,
				Category:  tc.category,
				Urgency:   "medium",
				Frequency: "one_time",
			}

			resp, body := s.client.POST(t, "/api/v1/decision/evaluate", intent)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var decision types.DecisionResponseDTO
			err := json.Unmarshal(body, &decision)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedDecision, decision.Decision, "DTI rule should be enforced correctly")

			if tc.expectedDTIRatio > 0.5 && tc.expectedDecision != "BUY" {
				assert.Contains(t, decision.PrimaryReason, "debt", "Should mention debt concerns for high DTI")
			}

			if tc.category == "health" && tc.expectedDecision == "BUY" && tc.expectedDTIRatio > 0.5 {
				assert.Contains(t, decision.PrimaryReason, "health", "Should mention health priority override")
			}
		})
	}
}

// TestAffordabilityCalculations tests the 30% disposable income threshold
func (s *DecisionRulesTestSuite) TestAffordabilityCalculations() {
	t := s.T()

	testCases := []struct {
		name                    string
		monthlyIncome          float64
		monthlyExpenses        float64
		monthlyLoanPayments    float64
		itemCost               float64
		expectedAffordability  float64
		expectedDecision       string
		shouldBeAffordable     bool
	}{
		{
			name:                   "Well Within Affordability",
			monthlyIncome:          8000.0,
			monthlyExpenses:        3000.0,
			monthlyLoanPayments:    1000.0,
			itemCost:               1000.0, // 25% of disposable income (4000)
			expectedAffordability:  12000.0, // 3x disposable income
			expectedDecision:       "BUY",
			shouldBeAffordable:     true,
		},
		{
			name:                   "At Affordability Limit",
			monthlyIncome:          6000.0,
			monthlyExpenses:        3500.0,
			monthlyLoanPayments:    800.0,
			itemCost:               510.0, // 30% of disposable income (1700)
			expectedAffordability:  5100.0, // 3x disposable income
			expectedDecision:       "BUY",
			shouldBeAffordable:     true,
		},
		{
			name:                   "Slightly Over Affordability - Wait",
			monthlyIncome:          5000.0,
			monthlyExpenses:        3000.0,
			monthlyLoanPayments:    500.0,
			itemCost:               600.0, // 40% of disposable income (1500)
			expectedAffordability:  4500.0,
			expectedDecision:       "WAIT",
			shouldBeAffordable:     false,
		},
		{
			name:                   "Way Over Affordability - Reject",
			monthlyIncome:          4000.0,
			monthlyExpenses:        2800.0,
			monthlyLoanPayments:    400.0,
			itemCost:               1000.0, // 125% of disposable income (800)
			expectedAffordability:  2400.0,
			expectedDecision:       "BYE",
			shouldBeAffordable:     false,
		},
		{
			name:                   "Negative Disposable Income - Reject",
			monthlyIncome:          4000.0,
			monthlyExpenses:        3500.0,
			monthlyLoanPayments:    800.0,
			itemCost:               200.0, // Negative disposable income
			expectedAffordability:  0.0,
			expectedDecision:       "BYE",
			shouldBeAffordable:     false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset for each test case
			s.server.ResetDatabase(t)
			s.mockAI.Reset()

			// Register user
			user := testutils.NewTestUser("afford.test@example.com", "Affordability Test", "password123")
			user.Register(t, s.client)

			// Create specific financial scenario
			financeData := &testutils.FinanceTestData{
				Incomes: []types.AddIncomeDTO{
					{Source: "Job", Amount: tc.monthlyIncome, Frequency: "monthly"},
				},
				Expenses: []types.AddExpenseDTO{
					{Category: "housing", Name: "Rent", Amount: tc.monthlyExpenses * 0.6, Frequency: "monthly", IsFixed: true, Priority: 1},
					{Category: "food", Name: "Food", Amount: tc.monthlyExpenses * 0.4, Frequency: "monthly", IsFixed: false, Priority: 1},
				},
				Loans: []types.AddLoanDTO{},
			}

			if tc.monthlyLoanPayments > 0 {
				financeData.Loans = append(financeData.Loans, types.AddLoanDTO{
					Lender:           "Bank",
					Type:             "personal",
					PrincipalAmount:  tc.monthlyLoanPayments * 48, // 4 years
					RemainingBalance: tc.monthlyLoanPayments * 40,
					MonthlyPayment:   tc.monthlyLoanPayments,
					InterestRate:     8.0,
					EndDate:          time.Date(2028, 12, 31, 0, 0, 0, 0, time.UTC),
				})
			}

			financeData.AddFinanceData(t, s.client)

			// Verify financial calculations
			financeSummary := s.client.GetFinanceSummary(t)
			expectedDisposableIncome := tc.monthlyIncome - tc.monthlyExpenses - tc.monthlyLoanPayments
			assert.InDelta(t, expectedDisposableIncome, financeSummary.DisposableIncome, 50.0, "Disposable income should be calculated correctly")

			affordability := s.client.GetAffordability(t)
			if tc.expectedAffordability > 0 {
				assert.InDelta(t, tc.expectedAffordability, affordability, 100.0, "Affordability should be calculated correctly")
			}

			// Add basic health data
			healthData := testutils.NewBasicHealthData()
			healthData.AddHealthData(t, s.client)

			// Configure AI to recommend BUY
			s.mockAI.SetNextResponse(domain.AIResponse{
				Decision:   "BUY",
				Confidence: 0.75,
				Reasoning:  "AI recommends purchase",
			})

			// Make decision request
			intent := types.PurchaseIntentDTO{
				ItemName:  fmt.Sprintf("Affordability Test - %s", tc.name),
				ItemCost:  tc.itemCost,
				Category:  "electronics",
				Urgency:   "medium",
				Frequency: "one_time",
			}

			resp, body := s.client.POST(t, "/api/v1/decision/evaluate", intent)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var decision types.DecisionResponseDTO
			err := json.Unmarshal(body, &decision)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedDecision, decision.Decision, "Affordability rule should be enforced")

			if !tc.shouldBeAffordable {
				assert.Contains(t, decision.PrimaryReason, "afford", "Should mention affordability concerns")
			}
		})
	}
}

// TestHealthCriticalPriorityOverride tests health priority overriding financial rules
func (s *DecisionRulesTestSuite) TestHealthCriticalPriorityOverride() {
	t := s.T()

	testCases := []struct {
		name                   string
		financialHealth       string
		hasHealthCondition    bool
		healthCategory        bool
		urgency               string
		expectedDecision      string
		shouldOverrideFinance bool
	}{
		{
			name:                   "Critical Health - Override Poor Finance",
			financialHealth:       "Poor",
			hasHealthCondition:    true,
			healthCategory:        true,
			urgency:               "critical",
			expectedDecision:      "BUY",
			shouldOverrideFinance: true,
		},
		{
			name:                   "High Health Priority - Override Fair Finance",
			financialHealth:       "Fair",
			hasHealthCondition:    true,
			healthCategory:        true,
			urgency:               "high",
			expectedDecision:      "BUY",
			shouldOverrideFinance: true,
		},
		{
			name:                   "Medium Health Priority - Don't Override Poor Finance",
			financialHealth:       "Poor",
			hasHealthCondition:    false,
			healthCategory:        true,
			urgency:               "medium",
			expectedDecision:      "BYE",
			shouldOverrideFinance: false,
		},
		{
			name:                   "Critical Non-Health - No Override",
			financialHealth:       "Poor",
			hasHealthCondition:    false,
			healthCategory:        false,
			urgency:               "critical",
			expectedDecision:      "BYE",
			shouldOverrideFinance: false,
		},
		{
			name:                   "Good Finance - Health Not Needed",
			financialHealth:       "Good",
			hasHealthCondition:    false,
			healthCategory:        false,
			urgency:               "medium",
			expectedDecision:      "BUY",
			shouldOverrideFinance: false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset for each test case
			s.server.ResetDatabase(t)
			s.mockAI.Reset()

			// Register user
			user := testutils.NewTestUser("health.priority@example.com", "Health Priority Test", "password123")
			user.Register(t, s.client)

			// Set up financial data based on health level
			var financeData *testutils.FinanceTestData
			switch tc.financialHealth {
			case "Poor":
				financeData = testutils.NewPoorFinanceData()
			case "Fair":
				financeData = &testutils.FinanceTestData{
					Incomes: []types.AddIncomeDTO{
						{Source: "Job", Amount: 4500, Frequency: "monthly"},
					},
					Expenses: []types.AddExpenseDTO{
						{Category: "housing", Name: "Rent", Amount: 1800, Frequency: "monthly", IsFixed: true, Priority: 1},
						{Category: "food", Name: "Food", Amount: 600, Frequency: "monthly", IsFixed: false, Priority: 1},
					},
					Loans: []types.AddLoanDTO{
						{Lender: "Bank", Type: "personal", PrincipalAmount: 30000, RemainingBalance: 25000,
							MonthlyPayment: 1500, InterestRate: 9.0, EndDate: time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC)},
					},
				}
			default: // Good
				financeData = testutils.NewBasicFinanceData()
			}
			financeData.AddFinanceData(t, s.client)

			// Set up health data
			var healthData *testutils.HealthTestData
			if tc.hasHealthCondition {
				healthData = testutils.NewCriticalHealthData()
			} else {
				healthData = testutils.NewBasicHealthData()
			}
			healthData.AddHealthData(t, s.client)

			// Configure AI to recommend BUY (business rules may override)
			s.mockAI.SetNextResponse(domain.AIResponse{
				Decision:   "BUY",
				Confidence: 0.85,
				Reasoning:  "AI analysis supports the purchase",
			})

			// Determine category and item based on test case
			category := "electronics"
			itemName := "Regular Electronics"
			if tc.healthCategory {
				category = "health"
				itemName = "Medical Equipment"
			}

			intent := types.PurchaseIntentDTO{
				ItemName:  itemName,
				ItemCost:  800.0,
				Category:  category,
				Urgency:   tc.urgency,
				Frequency: "one_time",
				Purpose:   "Health-related test purchase",
			}

			resp, body := s.client.POST(t, "/api/v1/decision/evaluate", intent)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var decision types.DecisionResponseDTO
			err := json.Unmarshal(body, &decision)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedDecision, decision.Decision, "Health priority rule should work correctly")

			if tc.shouldOverrideFinance {
				assert.Contains(t, decision.PrimaryReason, "health", "Should mention health priority override")
			}

			if tc.expectedDecision == "BYE" && tc.financialHealth == "Poor" {
				assert.Contains(t, decision.PrimaryReason, "financial", "Should mention financial concerns")
			}
		})
	}
}

// TestBusinessRuleCombinations tests combinations of business rules
func (s *DecisionRulesTestSuite) TestBusinessRuleCombinations() {
	t := s.T()

	testCases := []struct {
		name                    string
		setupFinance           func() *testutils.FinanceTestData
		setupHealth            func() *testutils.HealthTestData
		itemCost               float64
		category               string
		urgency                string
		expectedDecision       string
		expectedRuleViolations []string
	}{
		{
			name: "Multiple Rule Violations - Low Emergency Fund + High DTI",
			setupFinance: func() *testutils.FinanceTestData {
				return &testutils.FinanceTestData{
					Incomes: []types.AddIncomeDTO{
						{Source: "Job", Amount: 4000, Frequency: "monthly"},
					},
					Expenses: []types.AddExpenseDTO{
						{Category: "housing", Name: "Rent", Amount: 2000, Frequency: "monthly", IsFixed: true, Priority: 1},
						{Category: "food", Name: "Food", Amount: 800, Frequency: "monthly", IsFixed: false, Priority: 1},
					},
					Loans: []types.AddLoanDTO{
						{Lender: "Bank", Type: "personal", PrincipalAmount: 40000, RemainingBalance: 35000,
							MonthlyPayment: 2200, InterestRate: 10.0, EndDate: time.Date(2028, 12, 31, 0, 0, 0, 0, time.UTC)}, // 55% DTI
					},
				}
			},
			setupHealth: func() *testutils.HealthTestData {
				return testutils.NewBasicHealthData()
			},
			itemCost:               600.0,
			category:               "entertainment",
			urgency:                "low",
			expectedDecision:       "BYE",
			expectedRuleViolations: []string{"emergency fund", "debt"},
		},
		{
			name: "High DTI but Health Override",
			setupFinance: func() *testutils.FinanceTestData {
				return &testutils.FinanceTestData{
					Incomes: []types.AddIncomeDTO{
						{Source: "Job", Amount: 5000, Frequency: "monthly"},
					},
					Expenses: []types.AddExpenseDTO{
						{Category: "housing", Name: "Rent", Amount: 1800, Frequency: "monthly", IsFixed: true, Priority: 1},
					},
					Loans: []types.AddLoanDTO{
						{Lender: "Bank", Type: "car", PrincipalAmount: 25000, RemainingBalance: 20000,
							MonthlyPayment: 2600, InterestRate: 6.5, EndDate: time.Date(2027, 6, 30, 0, 0, 0, 0, time.UTC)}, // 52% DTI
					},
				}
			},
			setupHealth: func() *testutils.HealthTestData {
				return testutils.NewCriticalHealthData()
			},
			itemCost:               400.0,
			category:               "health",
			urgency:                "critical",
			expectedDecision:       "BUY",
			expectedRuleViolations: []string{}, // Health overrides DTI concerns
		},
		{
			name: "All Good Rules - Allow Purchase",
			setupFinance: func() *testutils.FinanceTestData {
				return testutils.NewExcellentFinanceData()
			},
			setupHealth: func() *testutils.HealthTestData {
				return testutils.NewBasicHealthData()
			},
			itemCost:               1500.0,
			category:               "electronics",
			urgency:                "medium",
			expectedDecision:       "BUY",
			expectedRuleViolations: []string{},
		},
		{
			name: "Good Finance but Excessive Cost",
			setupFinance: func() *testutils.FinanceTestData {
				return testutils.NewBasicFinanceData()
			},
			setupHealth: func() *testutils.HealthTestData {
				return testutils.NewBasicHealthData()
			},
			itemCost:               8000.0, // Way over affordability
			category:               "transport",
			urgency:                "medium",
			expectedDecision:       "BYE",
			expectedRuleViolations: []string{"afford"},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset for each test case
			s.server.ResetDatabase(t)
			s.mockAI.Reset()

			// Register user
			user := testutils.NewTestUser("combination.test@example.com", "Combination Test", "password123")
			user.Register(t, s.client)

			// Set up financial and health data
			financeData := tc.setupFinance()
			financeData.AddFinanceData(t, s.client)

			healthData := tc.setupHealth()
			healthData.AddHealthData(t, s.client)

			// Configure AI to recommend BUY (business rules should override if needed)
			s.mockAI.SetNextResponse(domain.AIResponse{
				Decision:   "BUY",
				Confidence: 0.90,
				Reasoning:  "AI strongly recommends this purchase",
			})

			// Make decision request
			intent := types.PurchaseIntentDTO{
				ItemName:  fmt.Sprintf("Rule Combination Test - %s", tc.name),
				ItemCost:  tc.itemCost,
				Category:  tc.category,
				Urgency:   tc.urgency,
				Frequency: "one_time",
			}

			resp, body := s.client.POST(t, "/api/v1/decision/evaluate", intent)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var decision types.DecisionResponseDTO
			err := json.Unmarshal(body, &decision)
			require.NoError(t, err)

			assert.Equal(t, tc.expectedDecision, decision.Decision, "Business rule combination should work correctly")

			// Check that expected rule violations are mentioned in the response
			for _, violation := range tc.expectedRuleViolations {
				assert.Contains(t, decision.PrimaryReason, violation, "Should mention rule violation: %s", violation)
			}

			if len(tc.expectedRuleViolations) == 0 && tc.expectedDecision == "BUY" {
				assert.NotContains(t, decision.PrimaryReason, "business rule", "Should not mention business rule override for valid purchases")
			}
		})
	}
}

// TestEdgeCases tests edge cases and boundary conditions
func (s *DecisionRulesTestSuite) TestEdgeCases() {
	t := s.T()

	// Test zero costs, maximum values, etc.
	edgeCases := []struct {
		name             string
		itemCost         float64
		monthlyIncome    float64
		expectedDecision string
		shouldSucceed    bool
	}{
		{
			name:             "Zero Cost Item",
			itemCost:         0.01, // Minimum valid cost
			monthlyIncome:    3000.0,
			expectedDecision: "BUY",
			shouldSucceed:    true,
		},
		{
			name:             "Very High Cost Item",
			itemCost:         999999.99, // Near maximum
			monthlyIncome:    5000.0,
			expectedDecision: "BYE",
			shouldSucceed:    true,
		},
		{
			name:             "Very Low Income",
			itemCost:         50.0,
			monthlyIncome:    100.0, // Extremely low income
			expectedDecision: "BYE",
			shouldSucceed:    true,
		},
		{
			name:             "Very High Income",
			itemCost:         5000.0,
			monthlyIncome:    50000.0, // Very high income
			expectedDecision: "BUY",
			shouldSucceed:    true,
		},
	}

	for _, tc := range edgeCases {
		s.Run(tc.name, func() {
			// Reset for each test case
			s.server.ResetDatabase(t)
			s.mockAI.Reset()

			// Register user
			user := testutils.NewTestUser("edge.case@example.com", "Edge Case Test", "password123")
			user.Register(t, s.client)

			// Set up financial data with specific income
			financeData := &testutils.FinanceTestData{
				Incomes: []types.AddIncomeDTO{
					{Source: "Income", Amount: tc.monthlyIncome, Frequency: "monthly"},
				},
				Expenses: []types.AddExpenseDTO{
					{Category: "housing", Name: "Basic", Amount: tc.monthlyIncome * 0.3, Frequency: "monthly", IsFixed: true, Priority: 1},
				},
				Loans: []types.AddLoanDTO{},
			}
			financeData.AddFinanceData(t, s.client)

			healthData := testutils.NewBasicHealthData()
			healthData.AddHealthData(t, s.client)

			// Configure AI
			s.mockAI.SetNextResponse(domain.AIResponse{
				Decision:   "BUY",
				Confidence: 0.75,
				Reasoning:  "Edge case test",
			})

			// Make decision request
			intent := types.PurchaseIntentDTO{
				ItemName:  fmt.Sprintf("Edge Case - %s", tc.name),
				ItemCost:  tc.itemCost,
				Category:  "other",
				Urgency:   "medium",
				Frequency: "one_time",
			}

			resp, body := s.client.POST(t, "/api/v1/decision/evaluate", intent)

			if tc.shouldSucceed {
				assert.Equal(t, http.StatusOK, resp.StatusCode)

				var decision types.DecisionResponseDTO
				err := json.Unmarshal(body, &decision)
				require.NoError(t, err)

				assert.Equal(t, tc.expectedDecision, decision.Decision)
				assert.Greater(t, decision.ProcessingTime, int64(0), "Should record processing time")
			} else {
				assert.NotEqual(t, http.StatusOK, resp.StatusCode)
			}
		})
	}
}

// Run the test suite
func TestDecisionRulesTestSuite(t *testing.T) {
	suite.Run(t, new(DecisionRulesTestSuite))
}