package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecisionContext_Validate_ValidData_Success(t *testing.T) {
	pastDecisions := []PastDecision{
		{
			ItemName:  "Previous laptop",
			ItemCost:  800.0,
			Decision:  "BUY",
			DaysAgo:   30,
			Category:  "electronics",
		},
		{
			ItemName:  "Gaming console",
			ItemCost:  400.0,
			Decision:  "WAIT",
			DaysAgo:   15,
			Category:  "entertainment",
		},
	}

	context := DecisionContext{
		UserID: "user-123",
		FinancialContext: FinancialSnapshot{
			MonthlyIncome:        5000.0,
			MonthlyExpenses:      3500.0,
			DisposableIncome:     1500.0,
			DebtToIncomeRatio:    0.25,
			EmergencyFundMonths:  6.0,
			SavingsRate:          0.20,
			FinancialHealth:      "excellent",
			BudgetRemaining:      800.0,
		},
		HealthContext: HealthSnapshot{
			HealthRiskScore:          25,
			MonthlyHealthCosts:       150.0,
			InsuranceCoverage:        0.8,
			FinancialVulnerability:   "low",
			HasCriticalConditions:    false,
			EmergencyFundNeeded:      3000.0,
		},
		TransportContext: TransportSnapshot{
			HasVehicle:           true,
			MonthlyTransportCost: 350.0,
			PublicTransitAccess:  true,
			CommuteDistance:      15.5,
		},
		PurchaseHistory: pastDecisions,
		CurrentDate:     time.Now(),
	}

	err := context.Validate()
	assert.NoError(t, err)
}

func TestDecisionContext_Validate_EmptyUserID_ReturnsError(t *testing.T) {
	context := DecisionContext{
		UserID: "", // Empty UserID should fail
		FinancialContext: FinancialSnapshot{
			MonthlyIncome:       5000.0,
			MonthlyExpenses:     3500.0,
			DisposableIncome:    1500.0,
			DebtToIncomeRatio:   0.25,
			EmergencyFundMonths: 6.0,
			SavingsRate:         0.20,
			FinancialHealth:     "excellent",
			BudgetRemaining:     800.0,
		},
		HealthContext:    HealthSnapshot{},
		TransportContext: TransportSnapshot{},
		PurchaseHistory:  []PastDecision{},
		CurrentDate:      time.Now(),
	}

	err := context.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user ID is required")
}

func TestDecisionContext_Validate_InvalidFinancialContext_ReturnsError(t *testing.T) {
	context := DecisionContext{
		UserID: "user-123",
		FinancialContext: FinancialSnapshot{
			MonthlyIncome:       -1000.0, // Negative income should fail
			MonthlyExpenses:     3500.0,
			DisposableIncome:    1500.0,
			DebtToIncomeRatio:   0.25,
			EmergencyFundMonths: 6.0,
			SavingsRate:         0.20,
			FinancialHealth:     "excellent",
			BudgetRemaining:     800.0,
		},
		HealthContext:    HealthSnapshot{},
		TransportContext: TransportSnapshot{},
		PurchaseHistory:  []PastDecision{},
		CurrentDate:      time.Now(),
	}

	err := context.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "financial snapshot validation failed")
}

func TestDecisionContext_Validate_InvalidHealthContext_ReturnsError(t *testing.T) {
	context := DecisionContext{
		UserID: "user-123",
		FinancialContext: FinancialSnapshot{
			MonthlyIncome:       5000.0,
			MonthlyExpenses:     3500.0,
			DisposableIncome:    1500.0,
			DebtToIncomeRatio:   0.25,
			EmergencyFundMonths: 6.0,
			SavingsRate:         0.20,
			FinancialHealth:     "excellent",
			BudgetRemaining:     800.0,
		},
		HealthContext: HealthSnapshot{
			HealthRiskScore: -10, // Negative risk score should fail
		},
		TransportContext: TransportSnapshot{},
		PurchaseHistory:  []PastDecision{},
		CurrentDate:      time.Now(),
	}

	err := context.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "health snapshot validation failed")
}

func TestDecisionContext_Validate_InvalidTransportContext_ReturnsError(t *testing.T) {
	context := DecisionContext{
		UserID: "user-123",
		FinancialContext: FinancialSnapshot{
			MonthlyIncome:       5000.0,
			MonthlyExpenses:     3500.0,
			DisposableIncome:    1500.0,
			DebtToIncomeRatio:   0.25,
			EmergencyFundMonths: 6.0,
			SavingsRate:         0.20,
			FinancialHealth:     "excellent",
			BudgetRemaining:     800.0,
		},
		HealthContext: HealthSnapshot{},
		TransportContext: TransportSnapshot{
			MonthlyTransportCost: -100.0, // Negative cost should fail
		},
		PurchaseHistory: []PastDecision{},
		CurrentDate:     time.Now(),
	}

	err := context.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "transport snapshot validation failed")
}

func TestDecisionContext_Validate_InvalidPastDecision_ReturnsError(t *testing.T) {
	invalidPastDecisions := []PastDecision{
		{
			ItemName: "", // Empty item name should fail
			ItemCost: 800.0,
			Decision: "BUY",
			DaysAgo:  30,
			Category: "electronics",
		},
	}

	context := DecisionContext{
		UserID: "user-123",
		FinancialContext: FinancialSnapshot{
			MonthlyIncome:       5000.0,
			MonthlyExpenses:     3500.0,
			DisposableIncome:    1500.0,
			DebtToIncomeRatio:   0.25,
			EmergencyFundMonths: 6.0,
			SavingsRate:         0.20,
			FinancialHealth:     "excellent",
			BudgetRemaining:     800.0,
		},
		HealthContext:    HealthSnapshot{},
		TransportContext: TransportSnapshot{},
		PurchaseHistory:  invalidPastDecisions,
		CurrentDate:      time.Now(),
	}

	err := context.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "past decision validation failed")
}

func TestDecisionContext_GetTotalRecentSpending_ReturnsCorrectSum(t *testing.T) {
	pastDecisions := []PastDecision{
		{
			ItemName: "Recent purchase 1",
			ItemCost: 200.0,
			Decision: "BUY",
			DaysAgo:  10,
			Category: "electronics",
		},
		{
			ItemName: "Recent purchase 2",
			ItemCost: 150.0,
			Decision: "BUY",
			DaysAgo:  5,
			Category: "clothing",
		},
		{
			ItemName: "Old purchase",
			ItemCost: 300.0,
			Decision: "BUY",
			DaysAgo:  40, // Should be excluded (> 30 days)
			Category: "electronics",
		},
		{
			ItemName: "Rejected purchase",
			ItemCost: 100.0,
			Decision: "BYE",
			DaysAgo:  5, // Should be excluded (not BUY)
			Category: "entertainment",
		},
	}

	context := DecisionContext{
		PurchaseHistory: pastDecisions,
	}

	totalSpending := context.GetTotalRecentSpending(30)
	assert.Equal(t, 350.0, totalSpending) // Only recent BUY decisions
}

func TestDecisionContext_GetRecentDecisionsByCategory_ReturnsCorrectDecisions(t *testing.T) {
	pastDecisions := []PastDecision{
		{
			ItemName: "Laptop",
			ItemCost: 800.0,
			Decision: "BUY",
			DaysAgo:  10,
			Category: "electronics",
		},
		{
			ItemName: "Phone",
			ItemCost: 400.0,
			Decision: "BUY",
			DaysAgo:  5,
			Category: "electronics",
		},
		{
			ItemName: "Shirt",
			ItemCost: 50.0,
			Decision: "BUY",
			DaysAgo:  15,
			Category: "clothing",
		},
		{
			ItemName: "Old TV",
			ItemCost: 600.0,
			Decision: "BUY",
			DaysAgo:  40, // Should be excluded (> 30 days)
			Category: "electronics",
		},
	}

	context := DecisionContext{
		PurchaseHistory: pastDecisions,
	}

	electronicsDecisions := context.GetRecentDecisionsByCategory("electronics", 30)
	assert.Len(t, electronicsDecisions, 2)
	assert.Equal(t, "Laptop", electronicsDecisions[0].ItemName)
	assert.Equal(t, "Phone", electronicsDecisions[1].ItemName)
}

func TestDecisionContext_HasRecentPurchaseInCategory_ReturnsTrueIfExists(t *testing.T) {
	pastDecisions := []PastDecision{
		{
			ItemName: "Recent laptop",
			ItemCost: 800.0,
			Decision: "BUY",
			DaysAgo:  10,
			Category: "electronics",
		},
	}

	context := DecisionContext{
		PurchaseHistory: pastDecisions,
	}

	hasRecent := context.HasRecentPurchaseInCategory("electronics", 30)
	assert.True(t, hasRecent)
}

func TestDecisionContext_HasRecentPurchaseInCategory_ReturnsFalseIfNotExists(t *testing.T) {
	pastDecisions := []PastDecision{
		{
			ItemName: "Old laptop",
			ItemCost: 800.0,
			Decision: "BUY",
			DaysAgo:  40, // Too old
			Category: "electronics",
		},
	}

	context := DecisionContext{
		PurchaseHistory: pastDecisions,
	}

	hasRecent := context.HasRecentPurchaseInCategory("electronics", 30)
	assert.False(t, hasRecent)
}

func TestDecisionContext_IsFinanciallyStressed_HighDebtRatio_ReturnsTrue(t *testing.T) {
	context := DecisionContext{
		FinancialContext: FinancialSnapshot{
			DebtToIncomeRatio:   0.6, // High debt ratio
			EmergencyFundMonths: 6.0,
		},
	}

	assert.True(t, context.IsFinanciallyStressed())
}

func TestDecisionContext_IsFinanciallyStressed_LowEmergencyFund_ReturnsTrue(t *testing.T) {
	context := DecisionContext{
		FinancialContext: FinancialSnapshot{
			DebtToIncomeRatio:   0.2,
			EmergencyFundMonths: 2.0, // Low emergency fund
		},
	}

	assert.True(t, context.IsFinanciallyStressed())
}

func TestDecisionContext_IsFinanciallyStressed_GoodFinances_ReturnsFalse(t *testing.T) {
	context := DecisionContext{
		FinancialContext: FinancialSnapshot{
			DebtToIncomeRatio:   0.2,
			EmergencyFundMonths: 6.0,
		},
	}

	assert.False(t, context.IsFinanciallyStressed())
}

func TestDecisionContext_HasHealthRisks_HighRiskScore_ReturnsTrue(t *testing.T) {
	context := DecisionContext{
		HealthContext: HealthSnapshot{
			HealthRiskScore: 80, // High risk
		},
	}

	assert.True(t, context.HasHealthRisks())
}

func TestDecisionContext_HasHealthRisks_CriticalConditions_ReturnsTrue(t *testing.T) {
	context := DecisionContext{
		HealthContext: HealthSnapshot{
			HealthRiskScore:       30,
			HasCriticalConditions: true,
		},
	}

	assert.True(t, context.HasHealthRisks())
}

func TestDecisionContext_HasHealthRisks_LowRisk_ReturnsFalse(t *testing.T) {
	context := DecisionContext{
		HealthContext: HealthSnapshot{
			HealthRiskScore:       20, // Low risk
			HasCriticalConditions: false,
		},
	}

	assert.False(t, context.HasHealthRisks())
}