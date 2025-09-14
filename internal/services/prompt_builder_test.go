package services

import (
	"testing"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptBuilder_BuildDecisionPrompt_IncludesAllContext(t *testing.T) {
	// Arrange
	builder := NewPromptBuilder()

	intent := &domain.PurchaseIntent{
		ID:          "intent-123",
		UserID:      "user-456",
		ItemName:    "Gaming Laptop",
		ItemCost:    1200.0,
		Category:    "electronics",
		Urgency:     "medium",
		Frequency:   "one_time",
		Purpose:     "For work and gaming",
		Alternative: "Desktop computer for $800",
		CreatedAt:   time.Now(),
	}

	pastDecisions := []domain.PastDecision{
		{
			ItemName: "Phone",
			ItemCost: 400.0,
			Decision: "BUY",
			DaysAgo:  30,
			Category: "electronics",
		},
	}

	context := &domain.DecisionContext{
		UserID: "user-456",
		FinancialContext: domain.FinancialSnapshot{
			MonthlyIncome:        5000.0,
			MonthlyExpenses:      3500.0,
			DisposableIncome:     1500.0,
			DebtToIncomeRatio:    0.25,
			EmergencyFundMonths:  6.0,
			SavingsRate:          0.20,
			FinancialHealth:      "excellent",
			BudgetRemaining:      800.0,
		},
		HealthContext: domain.HealthSnapshot{
			HealthRiskScore:          25,
			MonthlyHealthCosts:       150.0,
			InsuranceCoverage:        0.8,
			FinancialVulnerability:   "low",
			HasCriticalConditions:    false,
			EmergencyFundNeeded:      3000.0,
		},
		TransportContext: domain.TransportSnapshot{
			HasVehicle:           true,
			MonthlyTransportCost: 350.0,
			PublicTransitAccess:  true,
			CommuteDistance:      15.5,
		},
		PurchaseHistory: pastDecisions,
		CurrentDate:     time.Now(),
	}

	// Act
	prompt, err := builder.BuildDecisionPrompt(*intent, *context)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, prompt)

	// Verify all components are included
	assert.NotEmpty(t, prompt.SystemContext)
	assert.NotEmpty(t, prompt.UserContext)
	assert.NotEmpty(t, prompt.PurchaseDetails)
	assert.NotEmpty(t, prompt.DecisionCriteria)
	assert.NotEmpty(t, prompt.ResponseFormat)

	// Check that purchase details include key information
	assert.Contains(t, prompt.PurchaseDetails, "Gaming Laptop")
	assert.Contains(t, prompt.PurchaseDetails, "1200.0")
	assert.Contains(t, prompt.PurchaseDetails, "electronics")
	assert.Contains(t, prompt.PurchaseDetails, "medium")

	// Check that user context includes financial info
	assert.Contains(t, prompt.UserContext, "5000.0") // Monthly income
	assert.Contains(t, prompt.UserContext, "6.0") // Emergency fund months
	assert.Contains(t, prompt.UserContext, "0.25") // Debt ratio

	// Check that past decisions are included
	assert.Contains(t, prompt.UserContext, "Phone")
	assert.Contains(t, prompt.UserContext, "400.0")

	// Verify default configuration
	assert.Equal(t, 500, prompt.MaxTokens)
	assert.Equal(t, 0.7, prompt.Temperature)
}

func TestPromptBuilder_BuildDecisionPrompt_UnderTokenLimit(t *testing.T) {
	// Arrange
	builder := NewPromptBuilder()

	intent := &domain.PurchaseIntent{
		ID:        "intent-123",
		UserID:    "user-456",
		ItemName:  "Simple item",
		ItemCost:  100.0,
		Category:  "electronics",
		Urgency:   "low",
		Frequency: "one_time",
	}

	context := &domain.DecisionContext{
		UserID: "user-456",
		FinancialContext: domain.FinancialSnapshot{
			MonthlyIncome:        3000.0,
			MonthlyExpenses:      2000.0,
			DisposableIncome:     1000.0,
			DebtToIncomeRatio:    0.3,
			EmergencyFundMonths:  3.0,
			SavingsRate:          0.15,
			FinancialHealth:      "fair",
			BudgetRemaining:      500.0,
		},
		HealthContext: domain.HealthSnapshot{
			HealthRiskScore: 30,
		},
		TransportContext: domain.TransportSnapshot{
			HasVehicle: true,
		},
		PurchaseHistory: []domain.PastDecision{},
		CurrentDate:     time.Now(),
	}

	// Act
	prompt, err := builder.BuildDecisionPrompt(*intent, *context)

	// Assert
	require.NoError(t, err)
	assert.True(t, prompt.IsWithinTokenLimit(), "Prompt should be within token limit")

	estimatedTokens := prompt.EstimateTokens()
	assert.LessOrEqual(t, estimatedTokens, prompt.MaxTokens)
}

func TestPromptBuilder_BuildDecisionPrompt_StructuredFormat(t *testing.T) {
	// Arrange
	builder := NewPromptBuilder()

	intent := &domain.PurchaseIntent{
		ID:        "intent-123",
		UserID:    "user-456",
		ItemName:  "Test Item",
		ItemCost:  500.0,
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "one_time",
	}

	context := &domain.DecisionContext{
		UserID: "user-456",
		FinancialContext: domain.FinancialSnapshot{
			MonthlyIncome:    4000.0,
			MonthlyExpenses:  3000.0,
			DisposableIncome: 1000.0,
		},
		HealthContext:    domain.HealthSnapshot{},
		TransportContext: domain.TransportSnapshot{},
		PurchaseHistory:  []domain.PastDecision{},
		CurrentDate:      time.Now(),
	}

	// Act
	prompt, err := builder.BuildDecisionPrompt(*intent, *context)

	// Assert
	require.NoError(t, err)

	// Verify structured response format is specified
	assert.Contains(t, prompt.ResponseFormat, "JSON")
	assert.Contains(t, prompt.ResponseFormat, "decision")
	assert.Contains(t, prompt.ResponseFormat, "confidence")
	assert.Contains(t, prompt.ResponseFormat, "reasoning")
	assert.Contains(t, prompt.ResponseFormat, "BUY")
	assert.Contains(t, prompt.ResponseFormat, "WAIT")
	assert.Contains(t, prompt.ResponseFormat, "BYE")
}

func TestPromptBuilder_BuildDecisionPrompt_SanitizesPII(t *testing.T) {
	// Arrange
	builder := NewPromptBuilder()

	intent := &domain.PurchaseIntent{
		ID:        "intent-123",
		UserID:    "user-456",
		ItemName:  "Phone with plan SSN 123-45-6789",
		ItemCost:  500.0,
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "monthly",
		Purpose:   "Need phone, credit card 4532-1234-5678-9012",
	}

	context := &domain.DecisionContext{
		UserID: "user-456",
		FinancialContext: domain.FinancialSnapshot{
			MonthlyIncome:    4000.0,
			MonthlyExpenses:  3000.0,
			DisposableIncome: 1000.0,
		},
		HealthContext:    domain.HealthSnapshot{},
		TransportContext: domain.TransportSnapshot{},
		PurchaseHistory:  []domain.PastDecision{},
		CurrentDate:      time.Now(),
	}

	// Act
	prompt, err := builder.BuildDecisionPrompt(*intent, *context)

	// Assert
	require.NoError(t, err)

	// Check that PII is sanitized
	totalContent := prompt.GetTotalContent()
	assert.NotContains(t, totalContent, "123-45-6789") // SSN should be removed
	assert.NotContains(t, totalContent, "4532-1234-5678-9012") // Credit card should be removed
}

func TestPromptBuilder_BuildSystemPrompt_ContainsEssentialElements(t *testing.T) {
	// Arrange
	builder := NewPromptBuilder()

	// Act
	systemPrompt := builder.BuildSystemPrompt()

	// Assert
	assert.NotEmpty(t, systemPrompt)
	assert.Contains(t, systemPrompt, "BuyOrBye")
	assert.Contains(t, systemPrompt, "financial advisor")
	assert.Contains(t, systemPrompt, "BUY")
	assert.Contains(t, systemPrompt, "WAIT")
	assert.Contains(t, systemPrompt, "BYE")
	assert.Contains(t, systemPrompt, "emergency fund")
	assert.Contains(t, systemPrompt, "debt")
	assert.Contains(t, systemPrompt, "health")
}

func TestPromptBuilder_BuildUserContext_IncludesFinancialData(t *testing.T) {
	// Arrange
	builder := NewPromptBuilder()

	context := &domain.DecisionContext{
		UserID: "user-456",
		FinancialContext: domain.FinancialSnapshot{
			MonthlyIncome:        5000.0,
			MonthlyExpenses:      3500.0,
			DisposableIncome:     1500.0,
			DebtToIncomeRatio:    0.25,
			EmergencyFundMonths:  6.0,
			SavingsRate:          0.20,
			FinancialHealth:      "excellent",
			BudgetRemaining:      800.0,
		},
		HealthContext: domain.HealthSnapshot{
			HealthRiskScore: 25,
		},
		TransportContext: domain.TransportSnapshot{
			HasVehicle: true,
		},
		PurchaseHistory: []domain.PastDecision{},
		CurrentDate:     time.Now(),
	}

	// Act
	userContext := builder.BuildUserContext(*context)

	// Assert
	assert.NotEmpty(t, userContext)
	assert.Contains(t, userContext, "5000.0") // Monthly income
	assert.Contains(t, userContext, "1500.0") // Disposable income
	assert.Contains(t, userContext, "0.25") // Debt ratio
	assert.Contains(t, userContext, "6.0") // Emergency fund
	assert.Contains(t, userContext, "excellent") // Financial health
}

func TestPromptBuilder_BuildUserContext_IncludesHealthData(t *testing.T) {
	// Arrange
	builder := NewPromptBuilder()

	context := &domain.DecisionContext{
		UserID:           "user-456",
		FinancialContext: domain.FinancialSnapshot{},
		HealthContext: domain.HealthSnapshot{
			HealthRiskScore:          80,
			MonthlyHealthCosts:       300.0,
			InsuranceCoverage:        0.6,
			FinancialVulnerability:   "high",
			HasCriticalConditions:    true,
			EmergencyFundNeeded:      5000.0,
		},
		TransportContext: domain.TransportSnapshot{},
		PurchaseHistory:  []domain.PastDecision{},
		CurrentDate:      time.Now(),
	}

	// Act
	userContext := builder.BuildUserContext(*context)

	// Assert
	assert.NotEmpty(t, userContext)
	assert.Contains(t, userContext, "80") // Health risk score
	assert.Contains(t, userContext, "300.0") // Monthly health costs
	assert.Contains(t, userContext, "critical") // Critical conditions
	assert.Contains(t, userContext, "high") // Financial vulnerability
}

func TestPromptBuilder_BuildUserContext_IncludesPurchaseHistory(t *testing.T) {
	// Arrange
	builder := NewPromptBuilder()

	pastDecisions := []domain.PastDecision{
		{
			ItemName: "Previous Laptop",
			ItemCost: 800.0,
			Decision: "BUY",
			DaysAgo:  30,
			Category: "electronics",
		},
		{
			ItemName: "Gaming Console",
			ItemCost: 400.0,
			Decision: "WAIT",
			DaysAgo:  15,
			Category: "entertainment",
		},
	}

	context := &domain.DecisionContext{
		UserID:           "user-456",
		FinancialContext: domain.FinancialSnapshot{},
		HealthContext:    domain.HealthSnapshot{},
		TransportContext: domain.TransportSnapshot{},
		PurchaseHistory:  pastDecisions,
		CurrentDate:      time.Now(),
	}

	// Act
	userContext := builder.BuildUserContext(*context)

	// Assert
	assert.NotEmpty(t, userContext)
	assert.Contains(t, userContext, "Previous Laptop")
	assert.Contains(t, userContext, "Gaming Console")
	assert.Contains(t, userContext, "800.0")
	assert.Contains(t, userContext, "BUY")
	assert.Contains(t, userContext, "WAIT")
}

func TestPromptBuilder_BuildDecisionCriteria_ContainsAllFactors(t *testing.T) {
	// Arrange
	builder := NewPromptBuilder()

	// Act
	criteria := builder.BuildDecisionCriteria()

	// Assert
	assert.NotEmpty(t, criteria)
	assert.Contains(t, criteria, "emergency fund")
	assert.Contains(t, criteria, "debt-to-income")
	assert.Contains(t, criteria, "health")
	assert.Contains(t, criteria, "affordability")
	assert.Contains(t, criteria, "savings")
	assert.Contains(t, criteria, "necessity")
}

func TestPromptBuilder_BuildResponseFormat_SpecifiesJSONStructure(t *testing.T) {
	// Arrange
	builder := NewPromptBuilder()

	// Act
	responseFormat := builder.BuildResponseFormat()

	// Assert
	assert.NotEmpty(t, responseFormat)
	assert.Contains(t, responseFormat, "JSON")
	assert.Contains(t, responseFormat, "decision")
	assert.Contains(t, responseFormat, "confidence")
	assert.Contains(t, responseFormat, "reasoning")
	assert.Contains(t, responseFormat, "factors")
	assert.Contains(t, responseFormat, "recommendations")
}

func TestPromptBuilder_BuildDecisionPrompt_WithLargePurchaseHistory_TruncatesAppropriately(t *testing.T) {
	// Arrange
	builder := NewPromptBuilder()

	intent := &domain.PurchaseIntent{
		ID:        "intent-123",
		UserID:    "user-456",
		ItemName:  "Test Item",
		ItemCost:  500.0,
		Category:  "electronics",
		Urgency:   "medium",
		Frequency: "one_time",
	}

	// Create large purchase history
	var largePurchaseHistory []domain.PastDecision
	for i := 0; i < 50; i++ {
		largePurchaseHistory = append(largePurchaseHistory, domain.PastDecision{
			ItemName: "Item " + string(rune(i)),
			ItemCost: float64(i * 100),
			Decision: "BUY",
			DaysAgo:  i,
			Category: "electronics",
		})
	}

	context := &domain.DecisionContext{
		UserID: "user-456",
		FinancialContext: domain.FinancialSnapshot{
			MonthlyIncome:    4000.0,
			MonthlyExpenses:  3000.0,
			DisposableIncome: 1000.0,
		},
		HealthContext:    domain.HealthSnapshot{},
		TransportContext: domain.TransportSnapshot{},
		PurchaseHistory:  largePurchaseHistory,
		CurrentDate:      time.Now(),
	}

	// Act
	prompt, err := builder.BuildDecisionPrompt(*intent, *context)

	// Assert
	require.NoError(t, err)
	assert.True(t, prompt.IsWithinTokenLimit(), "Prompt with large history should still be within token limit")

	estimatedTokens := prompt.EstimateTokens()
	assert.LessOrEqual(t, estimatedTokens, prompt.MaxTokens)
}

func TestPromptBuilder_BuildDecisionPrompt_NilIntent_ReturnsError(t *testing.T) {
	// Arrange
	builder := NewPromptBuilder()

	context := &domain.DecisionContext{
		UserID: "user-456",
	}

	// Act - pass empty struct instead of nil
	prompt, err := builder.BuildDecisionPrompt(domain.PurchaseIntent{}, *context)

	// Assert - empty intent should cause validation error
	require.Error(t, err)
	assert.Nil(t, prompt)
	assert.Contains(t, err.Error(), "required")
}

func TestPromptBuilder_BuildDecisionPrompt_NilContext_ReturnsError(t *testing.T) {
	// Arrange
	builder := NewPromptBuilder()

	intent := &domain.PurchaseIntent{
		ID:       "intent-123",
		UserID:   "user-456",
		ItemName: "Test Item",
		ItemCost: 500.0,
	}

	// Act - pass empty struct instead of nil
	prompt, err := builder.BuildDecisionPrompt(*intent, domain.DecisionContext{})

	// Assert - empty context should cause validation error
	require.Error(t, err)
	assert.Nil(t, prompt)
	assert.Contains(t, err.Error(), "required")
}