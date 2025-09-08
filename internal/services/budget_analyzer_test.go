package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

func TestBudgetAnalyzer_AnalyzeBudget_HealthySurplus_ReturnsPositiveAnalysis(t *testing.T) {
	// Arrange
	summary := &domain.FinanceSummary{
		UserID:              "user-123",
		MonthlyIncome:       6000.00,
		MonthlyExpenses:     3500.00,
		MonthlyLoanPayments: 1500.00,
		DisposableIncome:    1000.00,
		DebtToIncomeRatio:   0.25, // 25% - healthy
		SavingsRate:         0.17, // 17% - decent
		FinancialHealth:     "Good",
		BudgetRemaining:     1000.00,
	}

	analyzer := NewBudgetAnalyzer()

	// Act
	analysis := analyzer.AnalyzeBudget(summary)

	// Assert
	assert.NotNil(t, analysis)
	assert.Equal(t, "user-123", analysis.UserID)
	assert.Equal(t, "Surplus", analysis.Status)
	assert.Equal(t, 1000.00, analysis.Amount)
	assert.Equal(t, "Good", analysis.FinancialHealth)
	assert.True(t, analysis.CanAffordPurchases)
	assert.NotEmpty(t, analysis.Recommendations)
	assert.Contains(t, analysis.Recommendations, "emergency fund")
}

func TestBudgetAnalyzer_AnalyzeBudget_BudgetDeficit_ReturnsDeficitAnalysis(t *testing.T) {
	// Arrange
	summary := &domain.FinanceSummary{
		UserID:              "user-456",
		MonthlyIncome:       4000.00,
		MonthlyExpenses:     3200.00,
		MonthlyLoanPayments: 1200.00,
		DisposableIncome:    -400.00,
		DebtToIncomeRatio:   0.30,
		SavingsRate:         -0.10, // Negative savings
		FinancialHealth:     "Poor",
		BudgetRemaining:     -400.00,
	}

	analyzer := NewBudgetAnalyzer()

	// Act
	analysis := analyzer.AnalyzeBudget(summary)

	// Assert
	assert.NotNil(t, analysis)
	assert.Equal(t, "user-456", analysis.UserID)
	assert.Equal(t, "Deficit", analysis.Status)
	assert.Equal(t, -400.00, analysis.Amount)
	assert.Equal(t, "Poor", analysis.FinancialHealth)
	assert.False(t, analysis.CanAffordPurchases)
	assert.NotEmpty(t, analysis.Recommendations)
	assert.Contains(t, analysis.Recommendations, "reduce expenses")
	assert.Contains(t, analysis.Recommendations, "increase income")
}

func TestBudgetAnalyzer_AnalyzeBudget_BreakEven_ReturnsBreakEvenAnalysis(t *testing.T) {
	// Arrange
	summary := &domain.FinanceSummary{
		UserID:              "user-789",
		MonthlyIncome:       5000.00,
		MonthlyExpenses:     3500.00,
		MonthlyLoanPayments: 1500.00,
		DisposableIncome:    0.00,
		DebtToIncomeRatio:   0.30,
		SavingsRate:         0.00,
		FinancialHealth:     "Fair",
		BudgetRemaining:     0.00,
	}

	analyzer := NewBudgetAnalyzer()

	// Act
	analysis := analyzer.AnalyzeBudget(summary)

	// Assert
	assert.NotNil(t, analysis)
	assert.Equal(t, "user-789", analysis.UserID)
	assert.Equal(t, "Break Even", analysis.Status)
	assert.Equal(t, 0.00, analysis.Amount)
	assert.Equal(t, "Fair", analysis.FinancialHealth)
	assert.False(t, analysis.CanAffordPurchases)
	assert.NotEmpty(t, analysis.Recommendations)
	assert.Contains(t, analysis.Recommendations, "find ways to increase income")
}

func TestBudgetAnalyzer_GetSpendingInsights_CategorizedExpenses_ReturnsCorrectInsights(t *testing.T) {
	// Arrange
	expenses := []domain.Expense{
		{Category: "housing", Amount: 1500.00, Frequency: "monthly", Priority: 1},
		{Category: "housing", Amount: 300.00, Frequency: "monthly", Priority: 1}, // Total housing: 1800
		{Category: "food", Amount: 150.00, Frequency: "weekly", Priority: 1},     // 650 monthly
		{Category: "food", Amount: 200.00, Frequency: "monthly", Priority: 2},    // Total food: 850
		{Category: "transport", Amount: 50.00, Frequency: "daily", Priority: 1},  // 1520.83 monthly
		{Category: "entertainment", Amount: 200.00, Frequency: "monthly", Priority: 3},
		{Category: "utilities", Amount: 150.00, Frequency: "monthly", Priority: 1},
	}

	analyzer := NewBudgetAnalyzer()

	// Act
	insights := analyzer.GetSpendingInsights(expenses)

	// Assert
	assert.NotEmpty(t, insights)

	// Find housing insight
	var housingInsight *SpendingInsight
	for _, insight := range insights {
		if insight.Category == "housing" {
			housingInsight = &insight
			break
		}
	}

	assert.NotNil(t, housingInsight)
	assert.Equal(t, 1800.00, housingInsight.MonthlyAmount)
	assert.InDelta(t, 0.35, housingInsight.PercentageOfTotal, 0.02) // Should be about 35% of total

	// Find transport insight - should be highest category
	var transportInsight *SpendingInsight
	for _, insight := range insights {
		if insight.Category == "transport" {
			transportInsight = &insight
			break
		}
	}

	assert.NotNil(t, transportInsight)
	assert.InDelta(t, 1520.83, transportInsight.MonthlyAmount, 0.01)
	assert.Equal(t, "High", transportInsight.Level) // Should be flagged as high
}

func TestBudgetAnalyzer_GetSpendingInsights_EmptyExpenses_ReturnsEmptyInsights(t *testing.T) {
	// Arrange
	expenses := []domain.Expense{}
	analyzer := NewBudgetAnalyzer()

	// Act
	insights := analyzer.GetSpendingInsights(expenses)

	// Assert
	assert.Empty(t, insights)
}

func TestBudgetAnalyzer_GetSpendingInsights_SingleCategory_ReturnsCorrectInsight(t *testing.T) {
	// Arrange
	expenses := []domain.Expense{
		{Category: "food", Amount: 400.00, Frequency: "monthly", Priority: 1},
	}
	analyzer := NewBudgetAnalyzer()

	// Act
	insights := analyzer.GetSpendingInsights(expenses)

	// Assert
	assert.Len(t, insights, 1)
	insight := insights[0]
	assert.Equal(t, "food", insight.Category)
	assert.Equal(t, 400.00, insight.MonthlyAmount)
	assert.Equal(t, 100.0, insight.PercentageOfTotal) // 100% since it's the only category
	assert.Equal(t, "Normal", insight.Level)          // Single category should be normal
}

func TestBudgetAnalyzer_RecommendSavings_ExcellentHealth_ReturnsGrowthRecommendations(t *testing.T) {
	// Arrange
	summary := &domain.FinanceSummary{
		UserID:              "user-123",
		MonthlyIncome:       8000.00,
		MonthlyExpenses:     4000.00,
		MonthlyLoanPayments: 1600.00, // DTI: 20%
		DisposableIncome:    2400.00,
		DebtToIncomeRatio:   0.20,
		SavingsRate:         0.30, // 30% savings rate
		FinancialHealth:     "Excellent",
		BudgetRemaining:     2400.00,
	}

	analyzer := NewBudgetAnalyzer()

	// Act
	recommendations := analyzer.RecommendSavings(summary)

	// Assert
	assert.NotEmpty(t, recommendations)

	// Should have investment recommendations for excellent health
	var hasInvestmentRec bool
	var hasEmergencyRec bool

	for _, rec := range recommendations {
		if rec.Type == "Investment" {
			hasInvestmentRec = true
			assert.Greater(t, rec.Amount, 0.0)
			assert.Equal(t, "High", rec.Priority)
		}
		if rec.Type == "Emergency Fund" {
			hasEmergencyRec = true
		}
	}

	assert.True(t, hasInvestmentRec)
	assert.True(t, hasEmergencyRec)
}

func TestBudgetAnalyzer_RecommendSavings_PoorHealth_ReturnsBasicRecommendations(t *testing.T) {
	// Arrange
	summary := &domain.FinanceSummary{
		UserID:              "user-456",
		MonthlyIncome:       3000.00,
		MonthlyExpenses:     2800.00,
		MonthlyLoanPayments: 1500.00, // DTI: 50%+
		DisposableIncome:    -300.00,
		DebtToIncomeRatio:   0.50,
		SavingsRate:         -0.10,
		FinancialHealth:     "Poor",
		BudgetRemaining:     -300.00,
	}

	analyzer := NewBudgetAnalyzer()

	// Act
	recommendations := analyzer.RecommendSavings(summary)

	// Assert
	assert.NotEmpty(t, recommendations)

	// Should focus on debt reduction and basic emergency fund for poor health
	var hasDebtRec bool
	var hasEmergencyRec bool

	for _, rec := range recommendations {
		if rec.Type == "Debt Reduction" {
			hasDebtRec = true
			assert.Equal(t, "High", rec.Priority)
		}
		if rec.Type == "Emergency Fund" {
			hasEmergencyRec = true
			assert.Equal(t, "Medium", rec.Priority) // Lower priority when finances are poor
		}
	}

	assert.True(t, hasDebtRec)
	assert.True(t, hasEmergencyRec)
}

func TestBudgetAnalyzer_RecommendSavings_GoodHealth_ReturnsBalancedRecommendations(t *testing.T) {
	// Arrange
	summary := &domain.FinanceSummary{
		UserID:              "user-789",
		MonthlyIncome:       6000.00,
		MonthlyExpenses:     3600.00,
		MonthlyLoanPayments: 1800.00, // DTI: 30%
		DisposableIncome:    600.00,
		DebtToIncomeRatio:   0.30,
		SavingsRate:         0.10, // 10% savings rate
		FinancialHealth:     "Good",
		BudgetRemaining:     600.00,
	}

	analyzer := NewBudgetAnalyzer()

	// Act
	recommendations := analyzer.RecommendSavings(summary)

	// Assert
	assert.NotEmpty(t, recommendations)

	// Should have balanced recommendations for good health
	var hasEmergencyRec bool
	var hasSavingsRec bool

	for _, rec := range recommendations {
		if rec.Type == "Emergency Fund" {
			hasEmergencyRec = true
			assert.Equal(t, "High", rec.Priority)
		}
		if rec.Type == "General Savings" {
			hasSavingsRec = true
		}
	}

	assert.True(t, hasEmergencyRec)
	assert.True(t, hasSavingsRec)
}

func TestBudgetAnalyzer_CalculateCategoryThreshold_ReturnsCorrectThresholds(t *testing.T) {
	tests := []struct {
		category          string
		monthlyIncome     float64
		expectedThreshold float64
	}{
		{"housing", 5000.00, 1750.00},      // 35% of income
		{"transport", 4000.00, 600.00},     // 15% of income
		{"food", 6000.00, 900.00},          // 15% of income
		{"entertainment", 3000.00, 180.00}, // 6% of income
		{"utilities", 5000.00, 300.00},     // 6% of income
		{"other", 4000.00, 200.00},         // 5% of income
	}

	analyzer := NewBudgetAnalyzer()

	for _, tt := range tests {
		t.Run("category_"+tt.category, func(t *testing.T) {
			// Act
			threshold := analyzer.CalculateCategoryThreshold(tt.category, tt.monthlyIncome)

			// Assert
			assert.Equal(t, tt.expectedThreshold, threshold)
		})
	}
}

func TestBudgetAnalyzer_IsOverspendingInCategory_ReturnsCorrectResult(t *testing.T) {
	tests := []struct {
		name           string
		category       string
		monthlyAmount  float64
		monthlyIncome  float64
		expectedResult bool
	}{
		{"housing_within_limit", "housing", 1500.00, 5000.00, false},         // 30% < 35% threshold
		{"housing_over_limit", "housing", 2000.00, 5000.00, true},            // 40% > 35% threshold
		{"food_within_limit", "food", 600.00, 4000.00, false},                // 15% = 15% threshold
		{"food_over_limit", "food", 800.00, 4000.00, true},                   // 20% > 15% threshold
		{"entertainment_over_limit", "entertainment", 300.00, 3000.00, true}, // 10% > 6% threshold
	}

	analyzer := NewBudgetAnalyzer()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			isOverspending := analyzer.IsOverspendingInCategory(tt.category, tt.monthlyAmount, tt.monthlyIncome)

			// Assert
			assert.Equal(t, tt.expectedResult, isOverspending)
		})
	}
}

func TestBudgetAnalyzer_GetOptimizationSuggestions_HighSpendingCategories_ReturnsSuggestions(t *testing.T) {
	// Arrange
	summary := &domain.FinanceSummary{
		UserID:              "user-123",
		MonthlyIncome:       5000.00,
		MonthlyExpenses:     4500.00, // High expenses
		MonthlyLoanPayments: 1000.00,
		DisposableIncome:    -500.00, // Overspending
		DebtToIncomeRatio:   0.20,
		SavingsRate:         -0.10,
		FinancialHealth:     "Poor",
		BudgetRemaining:     -500.00,
	}

	expenses := []domain.Expense{
		{Category: "housing", Amount: 2000.00, Frequency: "monthly"},      // 40% - over 35% threshold
		{Category: "food", Amount: 1000.00, Frequency: "monthly"},         // 20% - over 15% threshold
		{Category: "entertainment", Amount: 500.00, Frequency: "monthly"}, // 10% - over 6% threshold
		{Category: "transport", Amount: 600.00, Frequency: "monthly"},     // 12% - within 15% threshold
		{Category: "utilities", Amount: 400.00, Frequency: "monthly"},     // 8% - over 6% threshold
	}

	analyzer := NewBudgetAnalyzer()

	// Act
	suggestions := analyzer.GetOptimizationSuggestions(summary, expenses)

	// Assert
	assert.NotEmpty(t, suggestions)

	// Should suggest reducing overspending categories
	var housingFound, foodFound, entertainmentFound, utilitiesFound bool

	for _, suggestion := range suggestions {
		switch suggestion.Category {
		case "housing":
			housingFound = true
			assert.Contains(t, suggestion.Suggestion, "reduce")
			assert.Greater(t, suggestion.PotentialSavings, 0.0)
		case "food":
			foodFound = true
		case "entertainment":
			entertainmentFound = true
		case "utilities":
			utilitiesFound = true
		}
	}

	assert.True(t, housingFound)
	assert.True(t, foodFound)
	assert.True(t, entertainmentFound)
	assert.True(t, utilitiesFound)

	// Transport should not be suggested since it's within threshold
	for _, suggestion := range suggestions {
		assert.NotEqual(t, "transport", suggestion.Category)
	}
}
