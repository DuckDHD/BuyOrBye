package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFinanceSummary_CalculateHealth_ExcellentHealth_ReturnsExcellent(t *testing.T) {
	// Arrange
	summary := FinanceSummary{
		UserID:              "user-123",
		MonthlyIncome:       8000.00,
		MonthlyExpenses:     4000.00,
		MonthlyLoanPayments: 1600.00, // DTI = 20%
		DisposableIncome:    2400.00,
		DebtToIncomeRatio:   0.20,    // 20% - excellent
		SavingsRate:         0.30,    // 30% savings rate - excellent
		BudgetRemaining:     2400.00,
		UpdatedAt:           time.Now(),
	}

	// Act
	health := summary.CalculateHealth()

	// Assert
	assert.Equal(t, "Excellent", health)
}

func TestFinanceSummary_CalculateHealth_GoodHealth_ReturnsGood(t *testing.T) {
	// Arrange
	summary := FinanceSummary{
		UserID:              "user-123",
		MonthlyIncome:       6000.00,
		MonthlyExpenses:     3500.00,
		MonthlyLoanPayments: 1800.00, // DTI = 30%
		DisposableIncome:    700.00,
		DebtToIncomeRatio:   0.30,   // 30% - good range
		SavingsRate:         0.12,   // 12% savings rate - decent
		BudgetRemaining:     700.00,
		UpdatedAt:           time.Now(),
	}

	// Act
	health := summary.CalculateHealth()

	// Assert
	assert.Equal(t, "Good", health)
}

func TestFinanceSummary_CalculateHealth_FairHealth_HighDebtRatio_ReturnsFair(t *testing.T) {
	// Arrange
	summary := FinanceSummary{
		UserID:              "user-123",
		MonthlyIncome:       5000.00,
		MonthlyExpenses:     2500.00,
		MonthlyLoanPayments: 2000.00, // DTI = 40%
		DisposableIncome:    500.00,
		DebtToIncomeRatio:   0.40,   // 40% - fair range
		SavingsRate:         0.10,   // 10% savings rate
		BudgetRemaining:     500.00,
		UpdatedAt:           time.Now(),
	}

	// Act
	health := summary.CalculateHealth()

	// Assert
	assert.Equal(t, "Fair", health)
}

func TestFinanceSummary_CalculateHealth_FairHealth_LowSavings_ReturnsFair(t *testing.T) {
	// Arrange
	summary := FinanceSummary{
		UserID:              "user-123",
		MonthlyIncome:       4000.00,
		MonthlyExpenses:     3200.00,
		MonthlyLoanPayments: 400.00, // DTI = 10% - very good
		DisposableIncome:    400.00,
		DebtToIncomeRatio:   0.10,  // Low debt
		SavingsRate:         0.05,  // Only 5% savings rate - low
		BudgetRemaining:     400.00,
		UpdatedAt:           time.Now(),
	}

	// Act
	health := summary.CalculateHealth()

	// Assert
	assert.Equal(t, "Fair", health) // Low savings rate brings it to Fair
}

func TestFinanceSummary_CalculateHealth_PoorHealth_HighDebtRatio_ReturnsPoor(t *testing.T) {
	// Arrange
	summary := FinanceSummary{
		UserID:              "user-123",
		MonthlyIncome:       4000.00,
		MonthlyExpenses:     2200.00,
		MonthlyLoanPayments: 2200.00, // DTI = 55%
		DisposableIncome:    -400.00, // Negative disposable income
		DebtToIncomeRatio:   0.55,    // 55% - poor
		SavingsRate:         -0.10,   // Negative savings
		BudgetRemaining:     -400.00,
		UpdatedAt:           time.Now(),
	}

	// Act
	health := summary.CalculateHealth()

	// Assert
	assert.Equal(t, "Poor", health)
}

func TestFinanceSummary_CalculateHealth_PoorHealth_NegativeDisposableIncome_ReturnsPoor(t *testing.T) {
	// Arrange
	summary := FinanceSummary{
		UserID:              "user-123",
		MonthlyIncome:       3000.00,
		MonthlyExpenses:     2500.00,
		MonthlyLoanPayments: 800.00,
		DisposableIncome:    -300.00, // Negative - very bad
		DebtToIncomeRatio:   0.27,    // Decent DTI ratio
		SavingsRate:         -0.10,   // Can't save when overspending
		BudgetRemaining:     -300.00,
		UpdatedAt:           time.Now(),
	}

	// Act
	health := summary.CalculateHealth()

	// Assert
	assert.Equal(t, "Poor", health)
}

func TestFinanceSummary_CalculateHealth_EdgeCases_HandlesCorrectly(t *testing.T) {
	tests := []struct {
		name                string
		debtToIncomeRatio   float64
		savingsRate         float64
		disposableIncome    float64
		expectedHealth      string
	}{
		{"boundary_excellent", 0.28, 0.20, 1000.00, "Excellent"},  // Exactly at excellent boundaries
		{"boundary_good_dti", 0.36, 0.15, 500.00, "Good"},         // DTI at 36% boundary
		{"boundary_fair_dti", 0.50, 0.12, 200.00, "Fair"},         // DTI at 50% boundary  
		{"boundary_poor_dti", 0.51, 0.10, 100.00, "Poor"},         // Just over 50% DTI
		{"zero_income", 0.00, 0.00, 0.00, "Poor"},                 // Edge case: no income
		{"perfect_scenario", 0.15, 0.40, 3000.00, "Excellent"},    // Very healthy finances
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := FinanceSummary{
				DebtToIncomeRatio: tt.debtToIncomeRatio,
				SavingsRate:       tt.savingsRate,
				DisposableIncome:  tt.disposableIncome,
			}

			// Act
			health := summary.CalculateHealth()

			// Assert
			assert.Equal(t, tt.expectedHealth, health)
		})
	}
}

func TestFinanceSummary_GetHealthScore_ReturnsCorrectScores(t *testing.T) {
	tests := []struct {
		name           string
		health         string
		expectedScore  int
	}{
		{"excellent_health", "Excellent", 4},
		{"good_health", "Good", 3},
		{"fair_health", "Fair", 2},
		{"poor_health", "Poor", 1},
		{"unknown_health", "Unknown", 0},
		{"invalid_health", "Invalid", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := FinanceSummary{
				FinancialHealth: tt.health,
			}

			// Act
			score := summary.GetHealthScore()

			// Assert
			assert.Equal(t, tt.expectedScore, score)
		})
	}
}

func TestFinanceSummary_CalculateAffordability_ReturnsCorrectAmount(t *testing.T) {
	// Arrange
	tests := []struct {
		name                    string
		disposableIncome        float64
		debtToIncomeRatio       float64
		expectedAffordability   float64
	}{
		{"healthy_finances", 2000.00, 0.25, 6000.00},      // 3x disposable income for healthy DTI
		{"moderate_debt", 1500.00, 0.40, 3000.00},         // 2x disposable income for moderate DTI
		{"high_debt", 1000.00, 0.60, 500.00},              // 0.5x disposable income for high DTI
		{"negative_disposable", -500.00, 0.30, 0.00},      // No affordability if overspending
		{"zero_disposable", 0.00, 0.35, 0.00},             // No affordability if no surplus
		{"excellent_finances", 3000.00, 0.15, 9000.00},    // 3x for excellent DTI
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := FinanceSummary{
				DisposableIncome:  tt.disposableIncome,
				DebtToIncomeRatio: tt.debtToIncomeRatio,
			}

			// Act
			affordability := summary.CalculateAffordability()

			// Assert
			assert.Equal(t, tt.expectedAffordability, affordability)
		})
	}
}

func TestFinanceSummary_GetBudgetStatus_ReturnsCorrectStatus(t *testing.T) {
	tests := []struct {
		name            string
		budgetRemaining float64
		expectedStatus  string
	}{
		{"surplus", 1000.00, "Surplus"},
		{"break_even", 0.00, "Break Even"},
		{"small_deficit", -100.00, "Deficit"},
		{"large_deficit", -1000.00, "Deficit"},
		{"large_surplus", 5000.00, "Surplus"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := FinanceSummary{
				BudgetRemaining: tt.budgetRemaining,
			}

			// Act
			status := summary.GetBudgetStatus()

			// Assert
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func TestFinanceSummary_IsOverspending_ReturnsCorrectResult(t *testing.T) {
	tests := []struct {
		name              string
		monthlyIncome     float64
		monthlyExpenses   float64
		monthlyLoanPayments float64
		expectedResult    bool
	}{
		{"healthy_spending", 5000.00, 3000.00, 1000.00, false}, // Total: 4000, Income: 5000
		{"exact_break_even", 4000.00, 2500.00, 1500.00, false}, // Total: 4000, Income: 4000
		{"overspending", 3000.00, 2500.00, 1000.00, true},      // Total: 3500, Income: 3000
		{"severe_overspending", 2000.00, 3000.00, 1000.00, true}, // Total: 4000, Income: 2000
		{"no_loans", 3000.00, 3500.00, 0.00, true},             // Expenses exceed income
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := FinanceSummary{
				MonthlyIncome:       tt.monthlyIncome,
				MonthlyExpenses:     tt.monthlyExpenses,
				MonthlyLoanPayments: tt.monthlyLoanPayments,
			}

			// Act
			isOverspending := summary.IsOverspending()

			// Assert
			assert.Equal(t, tt.expectedResult, isOverspending)
		})
	}
}

func TestFinanceSummary_GetRecommendation_ReturnsCorrectRecommendations(t *testing.T) {
	tests := []struct {
		name             string
		financialHealth  string
		budgetRemaining  float64
		debtToIncomeRatio float64
		expectedContains []string // Partial matches for recommendations
	}{
		{
			"excellent_health",
			"Excellent",
			2000.00,
			0.20,
			[]string{"continue", "consider investing"},
		},
		{
			"poor_health_high_debt",
			"Poor",
			-500.00,
			0.60,
			[]string{"reduce expenses", "debt consolidation"},
		},
		{
			"fair_health_low_savings",
			"Fair",
			300.00,
			0.40,
			[]string{"increase savings", "reduce debt"},
		},
		{
			"good_health_moderate",
			"Good",
			800.00,
			0.30,
			[]string{"emergency fund", "optimize"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := FinanceSummary{
				FinancialHealth:   tt.financialHealth,
				BudgetRemaining:   tt.budgetRemaining,
				DebtToIncomeRatio: tt.debtToIncomeRatio,
			}

			// Act
			recommendation := summary.GetRecommendation()

			// Assert
			assert.NotEmpty(t, recommendation)
			for _, expectedText := range tt.expectedContains {
				assert.Contains(t, recommendation, expectedText)
			}
		})
	}
}

func TestFinanceSummary_Validate_AllFieldsValid_ReturnsNil(t *testing.T) {
	// Arrange
	summary := FinanceSummary{
		UserID:              "user-123",
		MonthlyIncome:       5000.00,
		MonthlyExpenses:     3000.00,
		MonthlyLoanPayments: 1000.00,
		DisposableIncome:    1000.00,
		DebtToIncomeRatio:   0.20,
		SavingsRate:         0.20,
		FinancialHealth:     "Good",
		BudgetRemaining:     1000.00,
		UpdatedAt:           time.Now(),
	}

	// Act
	err := summary.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestFinanceSummary_Validate_EmptyUserID_ReturnsError(t *testing.T) {
	// Arrange
	summary := FinanceSummary{
		UserID:              "", // Empty user ID
		MonthlyIncome:       5000.00,
		MonthlyExpenses:     3000.00,
		MonthlyLoanPayments: 1000.00,
		DisposableIncome:    1000.00,
		DebtToIncomeRatio:   0.20,
		SavingsRate:         0.20,
		FinancialHealth:     "Good",
		BudgetRemaining:     1000.00,
		UpdatedAt:           time.Now(),
	}

	// Act
	err := summary.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID is required")
}

func TestFinanceSummary_Validate_NegativeIncome_ReturnsError(t *testing.T) {
	// Arrange
	summary := FinanceSummary{
		UserID:              "user-123",
		MonthlyIncome:       -1000.00, // Invalid negative income
		MonthlyExpenses:     3000.00,
		MonthlyLoanPayments: 1000.00,
		DisposableIncome:    1000.00,
		DebtToIncomeRatio:   0.20,
		SavingsRate:         0.20,
		FinancialHealth:     "Good",
		BudgetRemaining:     1000.00,
		UpdatedAt:           time.Now(),
	}

	// Act
	err := summary.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "monthly income cannot be negative")
}

func TestFinanceSummary_Validate_InvalidHealthValue_ReturnsError(t *testing.T) {
	// Arrange
	invalidHealthValues := []string{"Great", "Bad", "Terrible", "Amazing", ""}

	for _, health := range invalidHealthValues {
		t.Run("health_"+health, func(t *testing.T) {
			summary := FinanceSummary{
				UserID:              "user-123",
				MonthlyIncome:       5000.00,
				MonthlyExpenses:     3000.00,
				MonthlyLoanPayments: 1000.00,
				DisposableIncome:    1000.00,
				DebtToIncomeRatio:   0.20,
				SavingsRate:         0.20,
				FinancialHealth:     health,
				BudgetRemaining:     1000.00,
				UpdatedAt:           time.Now(),
			}

			// Act
			err := summary.Validate()

			// Assert
			assert.Error(t, err)
			if health == "" {
				assert.Contains(t, err.Error(), "financial health is required")
			} else {
				assert.Contains(t, err.Error(), "financial health must be one of: Excellent, Good, Fair, Poor")
			}
		})
	}
}