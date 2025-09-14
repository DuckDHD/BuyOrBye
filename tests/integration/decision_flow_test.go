//go:build integration
// +build integration

package integration

import (
	"context"
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

// DecisionFlowTestSuite represents comprehensive decision flow integration tests
type DecisionFlowTestSuite struct {
	suite.Suite
	server     *testutils.TestServer
	client     *testutils.HTTPClient
	mockAI     *MockAIClient
	testUsers  []*testutils.TestUser
}

// MockAIClient provides controlled AI responses for testing
type MockAIClient struct {
	responses      []domain.AIResponse
	currentIndex   int
	shouldTimeout  bool
	shouldError    bool
	errorMessage   string
	responseTimes  []time.Duration
}

func NewMockAIClient() *MockAIClient {
	return &MockAIClient{
		responses:     make([]domain.AIResponse, 0),
		responseTimes: make([]time.Duration, 0),
	}
}

func (m *MockAIClient) GenerateDecision(ctx context.Context, prompt domain.AIPrompt) (*domain.AIResponse, error) {
	// Simulate timeout scenario
	if m.shouldTimeout {
		time.Sleep(35 * time.Second) // Exceed 30s timeout
		return nil, fmt.Errorf("context deadline exceeded")
	}

	// Simulate error scenario
	if m.shouldError {
		return nil, fmt.Errorf("openai error: %s", m.errorMessage)
	}

	// Return pre-configured response
	if m.currentIndex >= len(m.responses) {
		return nil, fmt.Errorf("no more mock responses available")
	}

	response := m.responses[m.currentIndex]
	m.currentIndex++

	// Simulate response time
	if m.currentIndex-1 < len(m.responseTimes) {
		time.Sleep(m.responseTimes[m.currentIndex-1])
	}

	return &response, nil
}

func (m *MockAIClient) SetResponses(responses []domain.AIResponse) {
	m.responses = responses
	m.currentIndex = 0
}

func (m *MockAIClient) SetNextResponse(response domain.AIResponse) {
	m.responses = []domain.AIResponse{response}
	m.currentIndex = 0
}

func (m *MockAIClient) SetTimeout(timeout bool) {
	m.shouldTimeout = timeout
}

func (m *MockAIClient) SetError(shouldError bool, message string) {
	m.shouldError = shouldError
	m.errorMessage = message
}

func (m *MockAIClient) SetResponseTimes(times []time.Duration) {
	m.responseTimes = times
}

func (m *MockAIClient) Reset() {
	m.responses = make([]domain.AIResponse, 0)
	m.currentIndex = 0
	m.shouldTimeout = false
	m.shouldError = false
	m.errorMessage = ""
	m.responseTimes = make([]time.Duration, 0)
}

// SetupSuite runs before all tests in the suite
func (s *DecisionFlowTestSuite) SetupSuite() {
	testutils.SetupIntegrationTest()
	s.server = testutils.NewTestServer(s.T())
	s.client = testutils.NewHTTPClient(s.server.BaseURL)
	s.mockAI = NewMockAIClient()
	s.testUsers = make([]*testutils.TestUser, 0)
}

// TearDownSuite runs after all tests in the suite
func (s *DecisionFlowTestSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}
	testutils.TeardownIntegrationTest()
}

// SetupTest runs before each test
func (s *DecisionFlowTestSuite) SetupTest() {
	s.server.ResetDatabase(s.T())
	s.client.SetAccessToken("")
	s.mockAI.Reset()
	s.testUsers = s.testUsers[:0]
}

// TestCompleteDecisionFlow tests the complete user journey: Login → Finance → Health → Decision
func (s *DecisionFlowTestSuite) TestCompleteDecisionFlow() {
	t := s.T()

	// Step 1: Register and authenticate user
	user := testutils.NewTestUser("decision.user@example.com", "Decision User", "password123")
	user.Register(t, s.client)
	s.testUsers = append(s.testUsers, user)

	// Step 2: Add financial data for good financial health
	financeData := testutils.NewBasicFinanceData()
	financeData.AddFinanceData(t, s.client)

	// Verify financial summary
	financeSummary := s.client.GetFinanceSummary(t)
	assert.Equal(t, "Good", financeSummary.FinancialHealth, "Should have good financial health")
	assert.Greater(t, financeSummary.DisposableIncome, 2000.0, "Should have adequate disposable income")

	// Step 3: Add health data for low risk profile
	healthData := testutils.NewBasicHealthData()
	healthData.AddHealthData(t, s.client)

	// Step 4: Configure mock AI for BUY recommendation
	s.mockAI.SetNextResponse(domain.AIResponse{
		Decision:   "BUY",
		Confidence: 0.85,
		Reasoning:  "User has good financial health, adequate disposable income, and low health risks. The laptop is a reasonable purchase for their financial situation.",
		Factors: []domain.DecisionFactor{
			{Category: "financial", Impact: "positive", Weight: 0.8, Description: "Strong disposable income"},
			{Category: "health", Impact: "neutral", Weight: 0.1, Description: "Low health risks"},
			{Category: "necessity", Impact: "positive", Weight: 0.9, Description: "Technology upgrade needed"},
		},
		Recommendations: []string{
			"Consider purchasing during a sale for better value",
			"Ensure you maintain emergency fund after purchase",
		},
	})

	// Step 5: Make decision request
	purchaseIntent := types.PurchaseIntentDTO{
		ItemName:    "MacBook Pro 14-inch",
		ItemCost:    2499.99,
		Category:    "electronics",
		Urgency:     "medium",
		Frequency:   "one_time",
		Purpose:     "Work laptop upgrade for software development",
		Alternative: "Refurbished MacBook or Windows laptop",
	}

	resp, body := s.client.POST(t, "/api/v1/decision/evaluate", purchaseIntent)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Decision evaluation should succeed")

	var decisionResponse types.DecisionResponseDTO
	err := json.Unmarshal(body, &decisionResponse)
	require.NoError(t, err, "Should parse decision response")

	// Step 6: Validate decision outcome
	assert.Equal(t, "BUY", decisionResponse.Decision, "Should recommend BUY")
	assert.Greater(t, decisionResponse.Confidence, 0.8, "Should have high confidence")
	assert.NotEmpty(t, decisionResponse.PrimaryReason, "Should provide reasoning")
	assert.Greater(t, len(decisionResponse.Factors), 0, "Should include decision factors")
	assert.Greater(t, len(decisionResponse.Recommendations), 0, "Should provide recommendations")
	assert.Greater(t, decisionResponse.ProcessingTime, int64(0), "Should record processing time")

	// Step 7: Verify decision is saved in history
	historyResp, historyBody := s.client.GET(t, "/api/v1/decision/history?limit=10&days=1")
	assert.Equal(t, http.StatusOK, historyResp.StatusCode)

	var historyData types.DecisionHistoryDTO
	err = json.Unmarshal(historyBody, &historyData)
	require.NoError(t, err)

	assert.Equal(t, 1, historyData.TotalDecisions, "Should have one decision in history")
	assert.Equal(t, 1, len(historyData.RecentDecisions), "Should show the recent decision")
	assert.Equal(t, "MacBook Pro 14-inch", historyData.RecentDecisions[0].ItemName, "Should match the item name")
}

// TestFinancialScenarios tests decision-making across different financial health levels
func (s *DecisionFlowTestSuite) TestFinancialScenarios() {
	t := s.T()

	scenarios := []struct {
		name               string
		financeData       *testutils.FinanceTestData
		expectedHealth    string
		purchaseCost      float64
		expectedDecision  string
		expectBusinessRule bool
	}{
		{
			name:              "Excellent Financial Health - High Cost Item",
			financeData:       testutils.NewExcellentFinanceData(),
			expectedHealth:    "Excellent",
			purchaseCost:      5000.0,
			expectedDecision:  "BUY",
			expectBusinessRule: false,
		},
		{
			name:              "Good Financial Health - Moderate Cost",
			financeData:       testutils.NewBasicFinanceData(),
			expectedHealth:    "Good",
			purchaseCost:      1500.0,
			expectedDecision:  "BUY",
			expectBusinessRule: false,
		},
		{
			name:              "Poor Financial Health - Any Cost",
			financeData:       testutils.NewPoorFinanceData(),
			expectedHealth:    "Poor",
			purchaseCost:      500.0,
			expectedDecision:  "BYE",
			expectBusinessRule: true, // Business rules should override
		},
	}

	for _, scenario := range scenarios {
		s.Run(scenario.name, func() {
			// Reset for each scenario
			s.server.ResetDatabase(t)
			s.mockAI.Reset()

			// Register user
			user := testutils.NewTestUser(fmt.Sprintf("scenario.%s@example.com", scenario.name), scenario.name, "password123")
			user.Register(t, s.client)

			// Add financial data
			scenario.financeData.AddFinanceData(t, s.client)

			// Verify expected financial health
			summary := s.client.GetFinanceSummary(t)
			assert.Equal(t, scenario.expectedHealth, summary.FinancialHealth, "Financial health should match expected")

			// Add basic health data
			healthData := testutils.NewBasicHealthData()
			healthData.AddHealthData(t, s.client)

			// Configure AI response (BUY recommendation)
			s.mockAI.SetNextResponse(domain.AIResponse{
				Decision:   "BUY",
				Confidence: 0.75,
				Reasoning:  "AI recommends purchase",
			})

			// Make decision
			intent := types.PurchaseIntentDTO{
				ItemName:  "Test Item",
				ItemCost:  scenario.purchaseCost,
				Category:  "electronics",
				Urgency:   "medium",
				Frequency: "one_time",
			}

			resp, body := s.client.POST(t, "/api/v1/decision/evaluate", intent)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var decision types.DecisionResponseDTO
			err := json.Unmarshal(body, &decision)
			require.NoError(t, err)

			assert.Equal(t, scenario.expectedDecision, decision.Decision, "Decision should match expected")

			if scenario.expectBusinessRule {
				assert.Contains(t, decision.PrimaryReason, "business rule", "Should indicate business rule involvement")
			}
		})
	}
}

// TestHealthCriticalPriorityOverride tests that critical health conditions override financial concerns
func (s *DecisionFlowTestSuite) TestHealthCriticalPriorityOverride() {
	t := s.T()

	// Register user with poor financial health
	user := testutils.NewTestUser("health.critical@example.com", "Health Critical User", "password123")
	user.Register(t, s.client)

	// Add poor financial data
	poorFinance := testutils.NewPoorFinanceData()
	poorFinance.AddFinanceData(t, s.client)

	// Add critical health condition
	criticalHealthData := testutils.NewCriticalHealthData()
	criticalHealthData.AddHealthData(t, s.client)

	// Configure AI response
	s.mockAI.SetNextResponse(domain.AIResponse{
		Decision:   "BUY",
		Confidence: 0.90,
		Reasoning:  "Critical health condition overrides financial concerns",
	})

	// Request for critical health item
	intent := types.PurchaseIntentDTO{
		ItemName:  "Emergency Medical Equipment",
		ItemCost:  800.0,
		Category:  "health",
		Urgency:   "critical",
		Frequency: "one_time",
		Purpose:   "Critical medical need for ongoing health condition",
	}

	resp, body := s.client.POST(t, "/api/v1/decision/evaluate", intent)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var decision types.DecisionResponseDTO
	err := json.Unmarshal(body, &decision)
	require.NoError(t, err)

	// Health should override financial concerns for critical items
	assert.Equal(t, "BUY", decision.Decision, "Critical health should override poor finances")
	assert.Contains(t, decision.PrimaryReason, "critical health", "Should mention health priority")
}

// TestAITimeoutFallback tests business rules fallback when AI times out
func (s *DecisionFlowTestSuite) TestAITimeoutFallback() {
	t := s.T()

	// Register user with good financial health
	user := testutils.NewTestUser("ai.timeout@example.com", "AI Timeout Test", "password123")
	user.Register(t, s.client)

	financeData := testutils.NewBasicFinanceData()
	financeData.AddFinanceData(t, s.client)

	healthData := testutils.NewBasicHealthData()
	healthData.AddHealthData(t, s.client)

	// Configure AI to timeout
	s.mockAI.SetTimeout(true)

	intent := types.PurchaseIntentDTO{
		ItemName:  "Test Purchase",
		ItemCost:  1000.0,
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "one_time",
	}

	resp, body := s.client.POST(t, "/api/v1/decision/evaluate", intent)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Should still succeed with business rules fallback")

	var decision types.DecisionResponseDTO
	err := json.Unmarshal(body, &decision)
	require.NoError(t, err)

	assert.Contains(t, decision.PrimaryReason, "business rule", "Should indicate business rules fallback")
	assert.NotEmpty(t, decision.Decision, "Should still provide a decision")
}

// TestCachingFunctionality tests 1-hour TTL caching for identical requests
func (s *DecisionFlowTestSuite) TestCachingFunctionality() {
	t := s.T()

	// Register user
	user := testutils.NewTestUser("cache.test@example.com", "Cache Test", "password123")
	user.Register(t, s.client)

	// Add financial and health data
	financeData := testutils.NewBasicFinanceData()
	financeData.AddFinanceData(t, s.client)

	healthData := testutils.NewBasicHealthData()
	healthData.AddHealthData(t, s.client)

	// Configure AI response
	aiResponse := domain.AIResponse{
		Decision:   "BUY",
		Confidence: 0.80,
		Reasoning:  "First AI call",
	}
	s.mockAI.SetNextResponse(aiResponse)

	// Make first request
	intent := types.PurchaseIntentDTO{
		ItemName:  "Cached Item",
		ItemCost:  1200.0,
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "one_time",
		Purpose:   "Same purpose for caching test",
	}

	start1 := time.Now()
	resp1, body1 := s.client.POST(t, "/api/v1/decision/evaluate", intent)
	duration1 := time.Since(start1)

	assert.Equal(t, http.StatusOK, resp1.StatusCode)

	var decision1 types.DecisionResponseDTO
	err := json.Unmarshal(body1, &decision1)
	require.NoError(t, err)

	// Configure different AI response for second call (should not be used due to cache)
	differentResponse := domain.AIResponse{
		Decision:   "WAIT",
		Confidence: 0.60,
		Reasoning:  "Second AI call - should not appear due to cache",
	}
	s.mockAI.SetNextResponse(differentResponse)

	// Make identical second request (should hit cache)
	start2 := time.Now()
	resp2, body2 := s.client.POST(t, "/api/v1/decision/evaluate", intent)
	duration2 := time.Since(start2)

	assert.Equal(t, http.StatusOK, resp2.StatusCode)

	var decision2 types.DecisionResponseDTO
	err = json.Unmarshal(body2, &decision2)
	require.NoError(t, err)

	// Verify cache hit
	assert.Equal(t, decision1.Decision, decision2.Decision, "Cached response should match")
	assert.Equal(t, decision1.Confidence, decision2.Confidence, "Cached confidence should match")
	assert.Equal(t, decision1.PrimaryReason, decision2.PrimaryReason, "Cached reason should match")
	assert.True(t, duration2 < duration1, "Second request should be faster due to cache")

	// Make request with different parameters (should not hit cache)
	differentIntent := intent
	differentIntent.ItemCost = 1300.0 // Different cost should bypass cache

	resp3, _ := s.client.POST(t, "/api/v1/decision/evaluate", differentIntent)
	assert.Equal(t, http.StatusOK, resp3.StatusCode)
	// This should call AI again since it's a different request
}

// TestBusinessRuleOverrideScenarios tests scenarios where business rules override AI
func (s *DecisionFlowTestSuite) TestBusinessRuleOverrideScenarios() {
	t := s.T()

	testCases := []struct {
		name               string
		financeSetup      func(*testutils.HTTPClient, *testing.T)
		healthSetup       func(*testutils.HTTPClient, *testing.T)
		aiDecision        string
		expectedDecision  string
		expectedOverride  bool
		itemCost          float64
	}{
		{
			name: "Emergency Fund Too Low - AI says BUY but business rules say BYE",
			financeSetup: func(client *testutils.HTTPClient, t *testing.T) {
				// Set up very low emergency fund scenario
				lowEmergencyFund := &testutils.FinanceTestData{
					Incomes: []types.AddIncomeDTO{{Source: "Job", Amount: 3000, Frequency: "monthly"}},
					Expenses: []types.AddExpenseDTO{
						{Category: "housing", Name: "Rent", Amount: 1200, Frequency: "monthly", IsFixed: true, Priority: 1},
						{Category: "food", Name: "Food", Amount: 500, Frequency: "monthly", IsFixed: false, Priority: 1},
					},
					Loans: []types.AddLoanDTO{}, // No emergency fund, high expenses
				}
				lowEmergencyFund.AddFinanceData(t, client)
			},
			healthSetup: func(client *testutils.HTTPClient, t *testing.T) {
				basicHealth := testutils.NewBasicHealthData()
				basicHealth.AddHealthData(t, client)
			},
			aiDecision:       "BUY",
			expectedDecision: "BYE",
			expectedOverride: true,
			itemCost:         800.0,
		},
		{
			name: "High DTI Ratio - AI says BUY but business rules say WAIT",
			financeSetup: func(client *testutils.HTTPClient, t *testing.T) {
				// Set up high debt-to-income scenario
				highDTI := &testutils.FinanceTestData{
					Incomes: []types.AddIncomeDTO{{Source: "Job", Amount: 4000, Frequency: "monthly"}},
					Expenses: []types.AddExpenseDTO{
						{Category: "housing", Name: "Rent", Amount: 1500, Frequency: "monthly", IsFixed: true, Priority: 1},
					},
					Loans: []types.AddLoanDTO{
						{Lender: "Bank", Type: "personal", PrincipalAmount: 50000, RemainingBalance: 45000,
							MonthlyPayment: 2200, InterestRate: 8.0, EndDate: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)},
					},
				}
				highDTI.AddFinanceData(t, client)
			},
			healthSetup: func(client *testutils.HTTPClient, t *testing.T) {
				basicHealth := testutils.NewBasicHealthData()
				basicHealth.AddHealthData(t, client)
			},
			aiDecision:       "BUY",
			expectedDecision: "WAIT",
			expectedOverride: true,
			itemCost:         1200.0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset for each test case
			s.server.ResetDatabase(t)
			s.mockAI.Reset()

			// Register user
			user := testutils.NewTestUser("override.test@example.com", "Override Test", "password123")
			user.Register(t, s.client)

			// Set up financial and health context
			tc.financeSetup(s.client, t)
			tc.healthSetup(s.client, t)

			// Configure AI to recommend the specified decision
			s.mockAI.SetNextResponse(domain.AIResponse{
				Decision:   tc.aiDecision,
				Confidence: 0.85,
				Reasoning:  fmt.Sprintf("AI recommends %s", tc.aiDecision),
			})

			// Make decision request
			intent := types.PurchaseIntentDTO{
				ItemName:  "Override Test Item",
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

			assert.Equal(t, tc.expectedDecision, decision.Decision, "Business rules should override AI when necessary")

			if tc.expectedOverride {
				assert.Contains(t, decision.PrimaryReason, "business rule", "Should mention business rule override")
			}
		})
	}
}

// TestDecisionHistoryAndStats tests decision history retrieval and statistics
func (s *DecisionFlowTestSuite) TestDecisionHistoryAndStats() {
	t := s.T()

	// Register user
	user := testutils.NewTestUser("history.test@example.com", "History Test", "password123")
	user.Register(t, s.client)

	// Add financial and health context
	financeData := testutils.NewBasicFinanceData()
	financeData.AddFinanceData(t, s.client)

	healthData := testutils.NewBasicHealthData()
	healthData.AddHealthData(t, s.client)

	// Make multiple decisions with different outcomes
	decisions := []struct {
		item     string
		cost     float64
		decision string
		category string
	}{
		{"Laptop", 2000.0, "BUY", "electronics"},
		{"Expensive Watch", 5000.0, "BYE", "other"},
		{"Phone", 800.0, "WAIT", "electronics"},
		{"Headphones", 300.0, "BUY", "electronics"},
	}

	for i, decisionTest := range decisions {
		s.mockAI.SetNextResponse(domain.AIResponse{
			Decision:   decisionTest.decision,
			Confidence: 0.80,
			Reasoning:  fmt.Sprintf("Decision %d reasoning", i+1),
		})

		intent := types.PurchaseIntentDTO{
			ItemName:  decisionTest.item,
			ItemCost:  decisionTest.cost,
			Category:  decisionTest.category,
			Urgency:   "medium",
			Frequency: "one_time",
		}

		resp, _ := s.client.POST(t, "/api/v1/decision/evaluate", intent)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// Test history retrieval
	historyResp, historyBody := s.client.GET(t, "/api/v1/decision/history?limit=10&days=1")
	assert.Equal(t, http.StatusOK, historyResp.StatusCode)

	var history types.DecisionHistoryDTO
	err := json.Unmarshal(historyBody, &history)
	require.NoError(t, err)

	assert.Equal(t, 4, history.TotalDecisions, "Should have 4 decisions")
	assert.Equal(t, 4, len(history.RecentDecisions), "Should show all recent decisions")

	// Check decision pattern
	assert.Equal(t, 2, history.DecisionPattern["BUY"], "Should have 2 BUY decisions")
	assert.Equal(t, 1, history.DecisionPattern["WAIT"], "Should have 1 WAIT decision")
	assert.Equal(t, 1, history.DecisionPattern["BYE"], "Should have 1 BYE decision")

	// Test statistics endpoint
	statsResp, statsBody := s.client.GET(t, "/api/v1/decision/stats?days=30")
	assert.Equal(t, http.StatusOK, statsResp.StatusCode)

	var stats map[string]interface{}
	err = json.Unmarshal(statsBody, &stats)
	require.NoError(t, err)

	assert.Equal(t, float64(4), stats["total_decisions"], "Stats should show 4 decisions")
	assert.Greater(t, stats["total_spending"].(float64), 0.0, "Should show spending for BUY decisions")

	// Verify pagination
	paginatedResp, paginatedBody := s.client.GET(t, "/api/v1/decision/history?limit=2&offset=2&days=1")
	assert.Equal(t, http.StatusOK, paginatedResp.StatusCode)

	var paginatedHistory types.DecisionHistoryDTO
	err = json.Unmarshal(paginatedBody, &paginatedHistory)
	require.NoError(t, err)

	assert.Equal(t, 4, paginatedHistory.TotalDecisions, "Total should remain 4")
	assert.Equal(t, 2, len(paginatedHistory.RecentDecisions), "Should show 2 paginated results")
}

// TestAuthenticationAndAuthorization tests security aspects of decision endpoints
func (s *DecisionFlowTestSuite) TestAuthenticationAndAuthorization() {
	t := s.T()

	// Test access without authentication
	intent := types.PurchaseIntentDTO{
		ItemName:  "Unauthorized Item",
		ItemCost:  100.0,
		Category:  "electronics",
		Urgency:   "low",
		Frequency: "one_time",
	}

	resp, _ := s.client.POST(t, "/api/v1/decision/evaluate", intent)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should require authentication")

	// Test history access without authentication
	resp, _ = s.client.GET(t, "/api/v1/decision/history")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "History should require authentication")

	// Test stats access without authentication
	resp, _ = s.client.GET(t, "/api/v1/decision/stats")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Stats should require authentication")
}

// TestValidationAndErrorHandling tests input validation and error scenarios
func (s *DecisionFlowTestSuite) TestValidationAndErrorHandling() {
	t := s.T()

	// Register user
	user := testutils.NewTestUser("validation.test@example.com", "Validation Test", "password123")
	user.Register(t, s.client)

	// Test invalid purchase intent
	invalidIntents := []struct {
		name   string
		intent map[string]interface{}
	}{
		{
			name: "Missing required fields",
			intent: map[string]interface{}{
				"item_cost": 100.0,
				// Missing item_name, category, urgency, frequency
			},
		},
		{
			name: "Invalid cost",
			intent: map[string]interface{}{
				"item_name": "Test Item",
				"item_cost": -100.0, // Negative cost
				"category":  "electronics",
				"urgency":   "medium",
				"frequency": "one_time",
			},
		},
		{
			name: "Invalid category",
			intent: map[string]interface{}{
				"item_name": "Test Item",
				"item_cost": 100.0,
				"category":  "invalid_category",
				"urgency":   "medium",
				"frequency": "one_time",
			},
		},
		{
			name: "Invalid urgency",
			intent: map[string]interface{}{
				"item_name": "Test Item",
				"item_cost": 100.0,
				"category":  "electronics",
				"urgency":   "invalid_urgency",
				"frequency": "one_time",
			},
		},
	}

	for _, tc := range invalidIntents {
		s.Run(tc.name, func() {
			resp, _ := s.client.POST(t, "/api/v1/decision/evaluate", tc.intent)
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should reject invalid input")
		})
	}

	// Test malformed JSON
	resp, _ := s.client.POST(t, "/api/v1/decision/evaluate", json.RawMessage(`{"invalid": json}`))
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should reject malformed JSON")
}

// Run the test suite
func TestDecisionFlowTestSuite(t *testing.T) {
	suite.Run(t, new(DecisionFlowTestSuite))
}