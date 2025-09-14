package services

import (
	"context"
	"strings"
	"testing"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDecisionRepository mocks the decision repository interface
type MockDecisionRepository struct {
	mock.Mock
}

// Implement LegacyDecisionRepository interface
func (m *MockDecisionRepository) Save(ctx context.Context, decision domain.DecisionOutcome) error {
	args := m.Called(ctx, decision)
	return args.Error(0)
}

func (m *MockDecisionRepository) GetByIntentID(ctx context.Context, intentID string) (*domain.DecisionOutcome, error) {
	args := m.Called(ctx, intentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DecisionOutcome), args.Error(1)
}

func (m *MockDecisionRepository) GetRecentDecisions(ctx context.Context, userID string, days int) ([]domain.PastDecision, error) {
	args := m.Called(ctx, userID, days)
	return args.Get(0).([]domain.PastDecision), args.Error(1)
}

// MockFinanceService mocks the finance service interface  
type MockFinanceService struct {
	mock.Mock
}

func (m *MockFinanceService) GetFinancialSnapshot(ctx context.Context, userID string) (*domain.FinancialSnapshot, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FinancialSnapshot), args.Error(1)
}

// MockHealthService mocks the health service interface
type MockHealthService struct {
	mock.Mock
}

func (m *MockHealthService) GetHealthSnapshot(ctx context.Context, userID string) (*domain.HealthSnapshot, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.HealthSnapshot), args.Error(1)
}

// MockAIClient mocks the AI client interface
type MockAIClient struct {
	mock.Mock
}

// Implement AIClient interface
func (m *MockAIClient) GenerateDecision(ctx context.Context, prompt domain.AIPrompt) (*domain.AIResponse, error) {
	args := m.Called(ctx, prompt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AIResponse), args.Error(1)
}

func (m *MockAIClient) GetModel() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAIClient) GetProvider() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAIClient) EstimateTokens(prompt string) int {
	args := m.Called(prompt)
	return args.Int(0)
}

func TestDecisionService_MakeDecision_LowEmergencyFund_ReturnsBye(t *testing.T) {
	// Arrange
	mockRepo := &MockDecisionRepository{}
	mockFinanceService := &MockFinanceService{}
	mockHealthService := &MockHealthService{}
	mockAIClient := &MockAIClient{}

	promptBuilder := NewPromptBuilder()
	interpreter := NewDecisionInterpreter()
	service := NewDecisionService(mockRepo, mockFinanceService, mockHealthService, mockAIClient, promptBuilder, interpreter)

	intent := &domain.PurchaseIntent{
		ID:        "intent-123",
		UserID:    "user-456",
		ItemName:  "Gaming Console",
		ItemCost:  500.0,
		Category:  "entertainment",
		Urgency:   "low",
		Frequency: "one_time",
	}

	financialSnapshot := &domain.FinancialSnapshot{
		MonthlyIncome:        4000.0,
		MonthlyExpenses:      3200.0,
		DisposableIncome:     800.0,
		DebtToIncomeRatio:    0.25,
		EmergencyFundMonths:  2.0, // Less than 3 months - should trigger BYE
		SavingsRate:          0.10,
		FinancialHealth:      "poor",
		BudgetRemaining:      300.0,
	}

	healthSnapshot := &domain.HealthSnapshot{
		HealthRiskScore:          25,
		MonthlyHealthCosts:       100.0,
		InsuranceCoverage:        0.8,
		FinancialVulnerability:   "medium",
		HasCriticalConditions:    false,
		EmergencyFundNeeded:      2000.0,
	}

	mockFinanceService.On("GetFinancialSnapshot", mock.Anything, "user-456").Return(financialSnapshot, nil)
	mockHealthService.On("GetHealthSnapshot", mock.Anything, "user-456").Return(healthSnapshot, nil)
	mockRepo.On("GetRecentDecisions", mock.Anything, "user-456", 30).Return([]domain.PastDecision{}, nil)
	
	expectedOutcome := &domain.DecisionOutcome{
		ID:            "", // Will be generated
		UserID:        "user-456",
		IntentID:      "intent-123",
		Decision:      "BYE",
		Confidence:    0.9,
		PrimaryReason: "Insufficient emergency fund (less than 3 months of expenses). Focus on building emergency fund first.",
	}
	
	mockRepo.On("Save", mock.Anything, mock.MatchedBy(func(outcome domain.DecisionOutcome) bool {
		return outcome.Decision == "BYE" && strings.Contains(outcome.PrimaryReason, "emergency fund")
	})).Return(nil)

	// Act
	ctx := context.Background()
	result, err := service.MakeDecision(ctx, *intent)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "BYE", result.Decision)
	assert.Equal(t, expectedOutcome.PrimaryReason, result.PrimaryReason)
	assert.Equal(t, expectedOutcome.Confidence, result.Confidence)
	mockRepo.AssertExpectations(t)
	mockFinanceService.AssertExpectations(t)
	mockHealthService.AssertExpectations(t)
}

func TestDecisionService_MakeDecision_HighDebtRatio_ReturnsWait(t *testing.T) {
	// Arrange
	mockRepo := &MockDecisionRepository{}
	mockFinanceService := &MockFinanceService{}
	mockHealthService := &MockHealthService{}
	mockAIClient := &MockAIClient{}

	promptBuilder := NewPromptBuilder()
	interpreter := NewDecisionInterpreter()
	service := NewDecisionService(mockRepo, mockFinanceService, mockHealthService, mockAIClient, promptBuilder, interpreter)

	intent := &domain.PurchaseIntent{
		ID:        "intent-123",
		UserID:    "user-456",
		ItemName:  "New TV",
		ItemCost:  800.0,
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "one_time",
	}

	financialSnapshot := &domain.FinancialSnapshot{
		MonthlyIncome:        4000.0,
		MonthlyExpenses:      3000.0,
		DisposableIncome:     1000.0,
		DebtToIncomeRatio:    0.6, // Over 50% - should trigger WAIT
		EmergencyFundMonths:  6.0,
		SavingsRate:          0.15,
		FinancialHealth:      "fair",
		BudgetRemaining:      500.0,
	}

	healthSnapshot := &domain.HealthSnapshot{
		HealthRiskScore:          30,
		MonthlyHealthCosts:       150.0,
		InsuranceCoverage:        0.8,
		FinancialVulnerability:   "medium",
		HasCriticalConditions:    false,
		EmergencyFundNeeded:      3000.0,
	}

	mockFinanceService.On("GetFinancialSnapshot", mock.Anything, "user-456").Return(financialSnapshot, nil)
	mockHealthService.On("GetHealthSnapshot", mock.Anything, "user-456").Return(healthSnapshot, nil)
	mockRepo.On("GetRecentDecisions", mock.Anything, "user-456", 30).Return([]domain.PastDecision{}, nil)
	
	mockRepo.On("Save", mock.Anything, mock.MatchedBy(func(outcome domain.DecisionOutcome) bool {
		return outcome.Decision == "WAIT" && outcome.WaitPeriod == 90
	})).Return(nil)

	// Act
	ctx := context.Background()
	result, err := service.MakeDecision(ctx, *intent)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "WAIT", result.Decision)
	assert.Equal(t, "High debt-to-income ratio indicates financial stress. Wait and focus on debt reduction.", result.PrimaryReason)
	assert.Equal(t, 90, result.WaitPeriod)
	assert.Equal(t, 0.85, result.Confidence)
	mockRepo.AssertExpectations(t)
	mockFinanceService.AssertExpectations(t)
	mockHealthService.AssertExpectations(t)
}

func TestDecisionService_MakeDecision_HealthCritical_ReturnsBuy(t *testing.T) {
	// Arrange
	mockRepo := &MockDecisionRepository{}
	mockFinanceService := &MockFinanceService{}
	mockHealthService := &MockHealthService{}
	mockAIClient := &MockAIClient{}

	promptBuilder := NewPromptBuilder()
	interpreter := NewDecisionInterpreter()
	service := NewDecisionService(mockRepo, mockFinanceService, mockHealthService, mockAIClient, promptBuilder, interpreter)

	intent := &domain.PurchaseIntent{
		ID:        "intent-123",
		UserID:    "user-456",
		ItemName:  "Medical Equipment",
		ItemCost:  1000.0,
		Category:  "health",
		Urgency:   "critical",
		Frequency: "one_time",
	}

	financialSnapshot := &domain.FinancialSnapshot{
		MonthlyIncome:        3000.0,
		MonthlyExpenses:      2800.0,
		DisposableIncome:     200.0,
		DebtToIncomeRatio:    0.7, // High debt ratio
		EmergencyFundMonths:  1.0, // Low emergency fund
		SavingsRate:          0.05,
		FinancialHealth:      "poor",
		BudgetRemaining:      100.0,
	}

	healthSnapshot := &domain.HealthSnapshot{
		HealthRiskScore:          85,
		MonthlyHealthCosts:       400.0,
		InsuranceCoverage:        0.6,
		FinancialVulnerability:   "high",
		HasCriticalConditions:    true, // Critical health condition - should override financial concerns
		EmergencyFundNeeded:      5000.0,
	}

	mockFinanceService.On("GetFinancialSnapshot", mock.Anything, "user-456").Return(financialSnapshot, nil)
	mockHealthService.On("GetHealthSnapshot", mock.Anything, "user-456").Return(healthSnapshot, nil)
	mockRepo.On("GetRecentDecisions", mock.Anything, "user-456", 30).Return([]domain.PastDecision{}, nil)
	
	mockRepo.On("Save", mock.Anything, mock.MatchedBy(func(outcome domain.DecisionOutcome) bool {
		return outcome.Decision == "BUY" && strings.Contains(outcome.PrimaryReason, "Critical health")
	})).Return(nil)

	// Act
	ctx := context.Background()
	result, err := service.MakeDecision(ctx, *intent)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "BUY", result.Decision)
	assert.Equal(t, "Critical health expenses take priority over financial constraints.", result.PrimaryReason)
	assert.Equal(t, 0.95, result.Confidence)
	mockRepo.AssertExpectations(t)
	mockFinanceService.AssertExpectations(t)
	mockHealthService.AssertExpectations(t)
}

func TestDecisionService_MakeDecision_ExceedsDisposableIncome_ReturnsBye(t *testing.T) {
	// Arrange
	mockRepo := &MockDecisionRepository{}
	mockFinanceService := &MockFinanceService{}
	mockHealthService := &MockHealthService{}
	mockAIClient := &MockAIClient{}

	promptBuilder := NewPromptBuilder()
	interpreter := NewDecisionInterpreter()
	service := NewDecisionService(mockRepo, mockFinanceService, mockHealthService, mockAIClient, promptBuilder, interpreter)

	intent := &domain.PurchaseIntent{
		ID:        "intent-123",
		UserID:    "user-456",
		ItemName:  "Expensive Gadget",
		ItemCost:  1500.0, // Exceeds 30% of disposable income
		Category:  "electronics",
		Urgency:   "low",
		Frequency: "one_time",
	}

	financialSnapshot := &domain.FinancialSnapshot{
		MonthlyIncome:        6000.0,
		MonthlyExpenses:      4500.0,
		DisposableIncome:     1500.0, // 30% = $450, purchase is $1500 (exceeds limit)
		DebtToIncomeRatio:    0.2,
		EmergencyFundMonths:  6.0,
		SavingsRate:          0.20,
		FinancialHealth:      "excellent",
		BudgetRemaining:      1000.0,
	}

	healthSnapshot := &domain.HealthSnapshot{
		HealthRiskScore:          15,
		MonthlyHealthCosts:       100.0,
		InsuranceCoverage:        0.9,
		FinancialVulnerability:   "low",
		HasCriticalConditions:    false,
		EmergencyFundNeeded:      2000.0,
	}

	mockFinanceService.On("GetFinancialSnapshot", mock.Anything, "user-456").Return(financialSnapshot, nil)
	mockHealthService.On("GetHealthSnapshot", mock.Anything, "user-456").Return(healthSnapshot, nil)
	mockRepo.On("GetRecentDecisions", mock.Anything, "user-456", 30).Return([]domain.PastDecision{}, nil)
	
	mockRepo.On("Save", mock.Anything, mock.MatchedBy(func(outcome domain.DecisionOutcome) bool {
		return outcome.Decision == "BYE" && 
			   outcome.PrimaryReason == "Exceeds 30% of disposable income" &&
			   outcome.MaxBudget == 450.0 // 30% of $1500
	})).Return(nil)

	// Act
	ctx := context.Background()
	result, err := service.MakeDecision(ctx, *intent)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "BYE", result.Decision)
	assert.Equal(t, "Exceeds 30% of disposable income", result.PrimaryReason)
	assert.Equal(t, 450.0, result.MaxBudget)
	assert.Equal(t, 0.8, result.Confidence)
	mockRepo.AssertExpectations(t)
	mockFinanceService.AssertExpectations(t)
	mockHealthService.AssertExpectations(t)
}

func TestDecisionService_MakeDecision_AffordablePurchase_ReturnsBuy(t *testing.T) {
	// Arrange
	mockRepo := &MockDecisionRepository{}
	mockFinanceService := &MockFinanceService{}
	mockHealthService := &MockHealthService{}
	mockAIClient := &MockAIClient{}

	promptBuilder := NewPromptBuilder()
	interpreter := NewDecisionInterpreter()
	service := NewDecisionService(mockRepo, mockFinanceService, mockHealthService, mockAIClient, promptBuilder, interpreter)

	intent := &domain.PurchaseIntent{
		ID:        "intent-123",
		UserID:    "user-456",
		ItemName:  "Work Laptop",
		ItemCost:  400.0, // Within 30% of disposable income
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "one_time",
	}

	financialSnapshot := &domain.FinancialSnapshot{
		MonthlyIncome:        5000.0,
		MonthlyExpenses:      3500.0,
		DisposableIncome:     1500.0, // 30% = $450, purchase is $400 (within limit)
		DebtToIncomeRatio:    0.2,
		EmergencyFundMonths:  6.0,
		SavingsRate:          0.20,
		FinancialHealth:      "excellent",
		BudgetRemaining:      1000.0,
	}

	healthSnapshot := &domain.HealthSnapshot{
		HealthRiskScore:          20,
		MonthlyHealthCosts:       120.0,
		InsuranceCoverage:        0.8,
		FinancialVulnerability:   "low",
		HasCriticalConditions:    false,
		EmergencyFundNeeded:      2500.0,
	}

	mockFinanceService.On("GetFinancialSnapshot", mock.Anything, "user-456").Return(financialSnapshot, nil)
	mockHealthService.On("GetHealthSnapshot", mock.Anything, "user-456").Return(healthSnapshot, nil)
	mockRepo.On("GetRecentDecisions", mock.Anything, "user-456", 30).Return([]domain.PastDecision{}, nil)
	
	mockRepo.On("Save", mock.Anything, mock.MatchedBy(func(outcome domain.DecisionOutcome) bool {
		return outcome.Decision == "BUY" && 
			   outcome.PrimaryReason == "Purchase fits within budget"
	})).Return(nil)

	// Act
	ctx := context.Background()
	result, err := service.MakeDecision(ctx, *intent)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "BUY", result.Decision)
	assert.Equal(t, "Purchase fits within budget", result.PrimaryReason)
	assert.Equal(t, 0.75, result.Confidence)
	mockRepo.AssertExpectations(t)
	mockFinanceService.AssertExpectations(t)
	mockHealthService.AssertExpectations(t)
}

func TestDecisionService_MakeDecision_WithAIFallback_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockDecisionRepository{}
	mockFinanceService := &MockFinanceService{}
	mockHealthService := &MockHealthService{}
	mockAIClient := &MockAIClient{}

	promptBuilder := NewPromptBuilder()
	interpreter := NewDecisionInterpreter()
	service := NewDecisionService(mockRepo, mockFinanceService, mockHealthService, mockAIClient, promptBuilder, interpreter)

	intent := &domain.PurchaseIntent{
		ID:        "intent-123",
		UserID:    "user-456",
		ItemName:  "Borderline Purchase",
		ItemCost:  300.0,
		Category:  "clothing",
		Urgency:   "medium",
		Frequency: "one_time",
	}

	// Edge case financial situation - no clear business rule applies
	financialSnapshot := &domain.FinancialSnapshot{
		MonthlyIncome:        4000.0,
		MonthlyExpenses:      3000.0,
		DisposableIncome:     1000.0,
		DebtToIncomeRatio:    0.4, // Moderate debt
		EmergencyFundMonths:  4.0, // Moderate emergency fund
		SavingsRate:          0.15,
		FinancialHealth:      "fair",
		BudgetRemaining:      600.0,
	}

	healthSnapshot := &domain.HealthSnapshot{
		HealthRiskScore:          40,
		MonthlyHealthCosts:       150.0,
		InsuranceCoverage:        0.8,
		FinancialVulnerability:   "medium",
		HasCriticalConditions:    false,
		EmergencyFundNeeded:      3000.0,
	}

	aiResponse := &domain.AIResponse{
		RawResponse: `{"decision": "WAIT", "confidence": 0.72, "reasoning": "Moderate financial situation suggests waiting for better timing"}`,
		Decision:    "WAIT",
		Confidence:  0.72,
		Reasoning:   "Moderate financial situation suggests waiting for better timing",
		Factors:     []string{"moderate debt", "fair emergency fund"},
		Suggestions: []string{"Consider waiting 30 days", "Look for sales"},
		TokensUsed:  150,
	}

	mockFinanceService.On("GetFinancialSnapshot", mock.Anything, "user-456").Return(financialSnapshot, nil)
	mockHealthService.On("GetHealthSnapshot", mock.Anything, "user-456").Return(healthSnapshot, nil)
	mockRepo.On("GetRecentDecisions", mock.Anything, "user-456", 30).Return([]domain.PastDecision{}, nil)
	mockAIClient.On("GenerateDecision", mock.Anything, mock.AnythingOfType("domain.AIPrompt")).Return(aiResponse, nil)
	
	mockRepo.On("Save", mock.Anything, mock.MatchedBy(func(outcome domain.DecisionOutcome) bool {
		return outcome.Decision == "WAIT" && outcome.Confidence == 0.72
	})).Return(nil)

	// Act
	ctx := context.Background()
	result, err := service.MakeDecision(ctx, *intent)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "WAIT", result.Decision)
	assert.Equal(t, 0.72, result.Confidence)
	assert.Equal(t, "Moderate financial situation suggests waiting for better timing", result.PrimaryReason)
	mockRepo.AssertExpectations(t)
	mockFinanceService.AssertExpectations(t)
	mockHealthService.AssertExpectations(t)
	mockAIClient.AssertExpectations(t)
}

func TestDecisionService_GetDecisionHistory_Success(t *testing.T) {
	// Arrange
	mockRepo := &MockDecisionRepository{}
	mockFinanceService := &MockFinanceService{}
	mockHealthService := &MockHealthService{}
	mockAIClient := &MockAIClient{}

	promptBuilder := NewPromptBuilder()
	interpreter := NewDecisionInterpreter()
	service := NewDecisionService(mockRepo, mockFinanceService, mockHealthService, mockAIClient, promptBuilder, interpreter)

	expectedDecisions := []domain.PastDecision{
		{
			ItemName: "Gaming Console",
			ItemCost: 500.0,
			Decision: "BUY",
			DaysAgo:  1,
			Category: "entertainment",
		},
		{
			ItemName: "Laptop",
			ItemCost: 1200.0,
			Decision: "WAIT",
			DaysAgo:  7,
			Category: "electronics",
		},
	}

	mockRepo.On("GetRecentDecisions", mock.Anything, "user-456", 10).Return(expectedDecisions, nil)

	// Act
	ctx := context.Background()
	result, err := service.GetDecisionHistory(ctx, "user-456", 10)

	// Assert
	require.NoError(t, err)
	assert.Len(t, result, 2)
	// Compare fields that exist on PastDecision
	assert.Equal(t, "BUY", result[0].Decision)
	assert.Equal(t, "WAIT", result[1].Decision)
	mockRepo.AssertExpectations(t)
}

func TestDecisionService_MakeDecision_InvalidIntent_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := &MockDecisionRepository{}
	mockFinanceService := &MockFinanceService{}
	mockHealthService := &MockHealthService{}
	mockAIClient := &MockAIClient{}

	promptBuilder := NewPromptBuilder()
	interpreter := NewDecisionInterpreter()
	service := NewDecisionService(mockRepo, mockFinanceService, mockHealthService, mockAIClient, promptBuilder, interpreter)

	intent := &domain.PurchaseIntent{
		ID:        "intent-123",
		UserID:    "user-456",
		ItemName:  "", // Invalid - empty item name
		ItemCost:  400.0,
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "one_time",
	}

	// Act
	ctx := context.Background()
	result, err := service.MakeDecision(ctx, *intent)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestDecisionService_MakeDecision_FinanceServiceError_ReturnsError(t *testing.T) {
	// Arrange
	mockRepo := &MockDecisionRepository{}
	mockFinanceService := &MockFinanceService{}
	mockHealthService := &MockHealthService{}
	mockAIClient := &MockAIClient{}

	promptBuilder := NewPromptBuilder()
	interpreter := NewDecisionInterpreter()
	service := NewDecisionService(mockRepo, mockFinanceService, mockHealthService, mockAIClient, promptBuilder, interpreter)

	intent := &domain.PurchaseIntent{
		ID:        "intent-123",
		UserID:    "user-456",
		ItemName:  "Test Item",
		ItemCost:  400.0,
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "one_time",
	}

	mockFinanceService.On("GetFinancialSnapshot", mock.Anything, "user-456").Return(nil, assert.AnError)

	// Act
	ctx := context.Background()
	result, err := service.MakeDecision(ctx, *intent)

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get financial snapshot")
	mockFinanceService.AssertExpectations(t)
}