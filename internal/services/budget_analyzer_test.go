package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock for BudgetAnalyzer dependencies
type MockBudgetAnalyzerFinanceService struct {
	mock.Mock
}

func (m *MockBudgetAnalyzerFinanceService) CalculateFinanceSummary(ctx context.Context, userID string) (domain.FinanceSummary, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(domain.FinanceSummary), args.Error(1)
}

func (m *MockBudgetAnalyzerFinanceService) GetUserIncomes(ctx context.Context, userID string) ([]domain.Income, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Income), args.Error(1)
}

func (m *MockBudgetAnalyzerFinanceService) GetUserExpenses(ctx context.Context, userID string) ([]domain.Expense, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Expense), args.Error(1)
}

func (m *MockBudgetAnalyzerFinanceService) NormalizeToMonthly(amount float64, frequency string) (float64, error) {
	args := m.Called(amount, frequency)
	return args.Get(0).(float64), args.Error(1)
}

// Implement all other required methods (not used by BudgetAnalyzer but needed for interface)
func (m *MockBudgetAnalyzerFinanceService) AddIncome(ctx context.Context, income domain.Income) error { return nil }
func (m *MockBudgetAnalyzerFinanceService) UpdateIncome(ctx context.Context, income domain.Income) error { return nil }
func (m *MockBudgetAnalyzerFinanceService) DeleteIncome(ctx context.Context, userID, incomeID string) error { return nil }
func (m *MockBudgetAnalyzerFinanceService) GetActiveUserIncomes(ctx context.Context, userID string) ([]domain.Income, error) { return nil, nil }
func (m *MockBudgetAnalyzerFinanceService) AddExpense(ctx context.Context, expense domain.Expense) error { return nil }
func (m *MockBudgetAnalyzerFinanceService) UpdateExpense(ctx context.Context, expense domain.Expense) error { return nil }
func (m *MockBudgetAnalyzerFinanceService) DeleteExpense(ctx context.Context, userID, expenseID string) error { return nil }
func (m *MockBudgetAnalyzerFinanceService) GetUserExpensesByCategory(ctx context.Context, userID, category string) ([]domain.Expense, error) { return nil, nil }
func (m *MockBudgetAnalyzerFinanceService) AddLoan(ctx context.Context, loan domain.Loan) error { return nil }
func (m *MockBudgetAnalyzerFinanceService) UpdateLoan(ctx context.Context, loan domain.Loan) error { return nil }
func (m *MockBudgetAnalyzerFinanceService) GetUserLoans(ctx context.Context, userID string) ([]domain.Loan, error) { return nil, nil }
func (m *MockBudgetAnalyzerFinanceService) UpdateLoanBalance(ctx context.Context, userID, loanID string, newBalance float64) error { return nil }
func (m *MockBudgetAnalyzerFinanceService) CalculateDisposableIncome(ctx context.Context, userID string) (float64, error) { return 0, nil }
func (m *MockBudgetAnalyzerFinanceService) CalculateDebtToIncomeRatio(ctx context.Context, userID string) (float64, error) { return 0, nil }
func (m *MockBudgetAnalyzerFinanceService) EvaluateFinancialHealth(ctx context.Context, userID string) (string, error) { return "", nil }
func (m *MockBudgetAnalyzerFinanceService) GetMaxAffordableAmount(ctx context.Context, userID string) (float64, error) { return 0, nil }

func setupBudgetAnalyzer() (*budgetAnalyzer, *MockBudgetAnalyzerFinanceService) {
	mockFinanceService := &MockBudgetAnalyzerFinanceService{}
	analyzer := &budgetAnalyzer{
		financeService: mockFinanceService,
	}
	return analyzer, mockFinanceService
}

func createTestFinanceSummary(userID string, monthlyIncome, monthlyExpenses, monthlyLoans, disposable float64) domain.FinanceSummary {
	return domain.FinanceSummary{
		UserID:              userID,
		MonthlyIncome:       monthlyIncome,
		MonthlyExpenses:     monthlyExpenses,
		MonthlyLoanPayments: monthlyLoans,
		DisposableIncome:    disposable,
		DebtToIncomeRatio:   monthlyLoans / monthlyIncome,
		SavingsRate:         disposable / monthlyIncome,
		UpdatedAt:          time.Now(),
	}
}

func TestBudgetAnalyzer_AnalyzeBudget_BalancedBudget(t *testing.T) {
	analyzer, mockFinanceService := setupBudgetAnalyzer()
	ctx := context.Background()

	// Balanced budget scenario
	summary := createTestFinanceSummary("user1", 5000.0, 3500.0, 1000.0, 500.0)
	
	expenses := []domain.Expense{
		createTestExpense("exp1", "user1", "housing", "Rent", 1500.0, "monthly", true, 1),
		createTestExpense("exp2", "user1", "food", "Groceries", 500.0, "monthly", false, 1),
		createTestExpense("exp3", "user1", "transportation", "Gas", 200.0, "monthly", false, 2),
		createTestExpense("exp4", "user1", "entertainment", "Movies", 100.0, "monthly", false, 3),
	}

	mockFinanceService.On("CalculateFinanceSummary", ctx, "user1").Return(summary, nil)
	mockFinanceService.On("GetUserExpenses", ctx, "user1").Return(expenses, nil)
	// Mock NormalizeToMonthly calls for each expense
	mockFinanceService.On("NormalizeToMonthly", 1500.0, "monthly").Return(1500.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 500.0, "monthly").Return(500.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 200.0, "monthly").Return(200.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 100.0, "monthly").Return(100.0, nil)

	analysis, err := analyzer.AnalyzeBudget(ctx, "user1")

	assert.NoError(t, err)
	assert.Equal(t, "user1", analysis.UserID)
	assert.Equal(t, 5000.0, analysis.TotalMonthlyIncome)
	assert.Equal(t, 3500.0, analysis.TotalMonthlyExpenses)
	assert.Equal(t, 1000.0, analysis.MonthlyLoanPayments)
	assert.Equal(t, "Surplus", analysis.BudgetStatus) // Has positive disposable income
	assert.Equal(t, 0.0, analysis.OverspendingAmount)
	assert.Greater(t, analysis.BudgetHealthScore, 5) // Should be good score for balanced budget
	
	mockFinanceService.AssertExpectations(t)
}

func TestBudgetAnalyzer_AnalyzeBudget_DeficitBudget(t *testing.T) {
	analyzer, mockFinanceService := setupBudgetAnalyzer()
	ctx := context.Background()

	// Deficit budget scenario - overspending
	summary := createTestFinanceSummary("user1", 4000.0, 3500.0, 1200.0, -700.0)
	
	expenses := []domain.Expense{
		createTestExpense("exp1", "user1", "housing", "Rent", 2000.0, "monthly", true, 1), // Over 30% of income
		createTestExpense("exp2", "user1", "food", "Groceries", 800.0, "monthly", false, 1), // High grocery spending
		createTestExpense("exp3", "user1", "entertainment", "Entertainment", 700.0, "monthly", false, 3), // High entertainment
	}

	mockFinanceService.On("CalculateFinanceSummary", ctx, "user1").Return(summary, nil)
	mockFinanceService.On("GetUserExpenses", ctx, "user1").Return(expenses, nil)
	// Mock NormalizeToMonthly calls for each expense
	mockFinanceService.On("NormalizeToMonthly", 2000.0, "monthly").Return(2000.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 800.0, "monthly").Return(800.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 700.0, "monthly").Return(700.0, nil)

	analysis, err := analyzer.AnalyzeBudget(ctx, "user1")

	assert.NoError(t, err)
	assert.Equal(t, "user1", analysis.UserID)
	assert.Equal(t, "Deficit", analysis.BudgetStatus)
	assert.Equal(t, 700.0, analysis.OverspendingAmount)
	assert.NotEmpty(t, analysis.OverspendingCategories)
	assert.NotEmpty(t, analysis.RecommendedActions)
	assert.Less(t, analysis.BudgetHealthScore, 5) // Should be low score for deficit
	
	mockFinanceService.AssertExpectations(t)
}

func TestBudgetAnalyzer_AnalyzeBudget_FinanceServiceError(t *testing.T) {
	analyzer, mockFinanceService := setupBudgetAnalyzer()
	ctx := context.Background()

	mockFinanceService.On("CalculateFinanceSummary", ctx, "user1").Return(domain.FinanceSummary{}, errors.New("database error"))

	_, err := analyzer.AnalyzeBudget(ctx, "user1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get financial summary")
	mockFinanceService.AssertExpectations(t)
}

func TestBudgetAnalyzer_GetSpendingInsights_Success(t *testing.T) {
	analyzer, mockFinanceService := setupBudgetAnalyzer()
	ctx := context.Background()

	expenses := []domain.Expense{
		createTestExpense("exp1", "user1", "housing", "Rent", 2000.0, "monthly", true, 1),
		createTestExpense("exp2", "user1", "housing", "Utilities", 200.0, "monthly", true, 1),
		createTestExpense("exp3", "user1", "food", "Groceries", 600.0, "monthly", false, 1),
		createTestExpense("exp4", "user1", "food", "Restaurants", 400.0, "monthly", false, 2),
		createTestExpense("exp5", "user1", "transportation", "Car Payment", 300.0, "monthly", true, 1),
	}

	summary := createTestFinanceSummary("user1", 6000.0, 3500.0, 0.0, 2500.0)

	mockFinanceService.On("CalculateFinanceSummary", ctx, "user1").Return(summary, nil)
	mockFinanceService.On("GetUserExpenses", ctx, "user1").Return(expenses, nil)
	// Mock NormalizeToMonthly calls for each expense
	mockFinanceService.On("NormalizeToMonthly", 2000.0, "monthly").Return(2000.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 200.0, "monthly").Return(200.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 600.0, "monthly").Return(600.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 400.0, "monthly").Return(400.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 300.0, "monthly").Return(300.0, nil)

	insights, err := analyzer.GetSpendingInsights(ctx, "user1")

	assert.NoError(t, err)
	assert.Equal(t, "user1", insights.UserID)
	assert.Equal(t, 3500.0, insights.TotalMonthlySpending)
	assert.NotEmpty(t, insights.CategoryBreakdown)
	
	// Verify category breakdown
	housingFound := false
	foodFound := false
	for _, category := range insights.CategoryBreakdown {
		if category.Category == "housing" {
			housingFound = true
			assert.Equal(t, 2200.0, category.MonthlyAmount) // Rent + Utilities
			assert.InDelta(t, 36.67, category.PercentageOfIncome, 0.1) // 2200/6000 * 100
		}
		if category.Category == "food" {
			foodFound = true
			assert.Equal(t, 1000.0, category.MonthlyAmount) // Groceries + Restaurants
		}
	}
	
	assert.True(t, housingFound)
	assert.True(t, foodFound)
	assert.NotEmpty(t, insights.HighestCategory.Category)
	assert.NotEmpty(t, insights.LowestCategory.Category)
	
	mockFinanceService.AssertExpectations(t)
}

func TestBudgetAnalyzer_GetSpendingInsights_NoExpenses(t *testing.T) {
	analyzer, mockFinanceService := setupBudgetAnalyzer()
	ctx := context.Background()

	mockFinanceService.On("GetUserExpenses", ctx, "user1").Return([]domain.Expense{}, nil)
	mockFinanceService.On("GetUserIncomes", ctx, "user1").Return([]domain.Income{}, nil)

	insights, err := analyzer.GetSpendingInsights(ctx, "user1")

	assert.NoError(t, err)
	assert.Equal(t, "user1", insights.UserID)
	assert.Equal(t, 0.0, insights.TotalMonthlySpending)
	assert.Empty(t, insights.CategoryBreakdown)
	
	mockFinanceService.AssertExpectations(t)
}

func TestBudgetAnalyzer_RecommendSavings_HealthyBudget(t *testing.T) {
	analyzer, mockFinanceService := setupBudgetAnalyzer()
	ctx := context.Background()

	// User following 50/30/20 rule well
	summary := createTestFinanceSummary("user1", 6000.0, 3000.0, 600.0, 2400.0) // 50% expenses, 10% debt, 40% savings
	
	expenses := []domain.Expense{
		createTestExpense("exp1", "user1", "housing", "Rent", 1800.0, "monthly", true, 1),  // Needs
		createTestExpense("exp2", "user1", "food", "Groceries", 600.0, "monthly", true, 1), // Needs
		createTestExpense("exp3", "user1", "utilities", "Utilities", 200.0, "monthly", true, 1), // Needs
		createTestExpense("exp4", "user1", "entertainment", "Movies", 200.0, "monthly", false, 3), // Wants
		createTestExpense("exp5", "user1", "dining", "Restaurants", 200.0, "monthly", false, 2), // Wants
	}

	mockFinanceService.On("CalculateFinanceSummary", ctx, "user1").Return(summary, nil)
	mockFinanceService.On("GetUserExpenses", ctx, "user1").Return(expenses, nil)
	// Mock NormalizeToMonthly calls for each expense
	mockFinanceService.On("NormalizeToMonthly", 1800.0, "monthly").Return(1800.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 600.0, "monthly").Return(600.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 200.0, "monthly").Return(200.0, nil).Times(3) // Three 200.0 expenses

	recommendation, err := analyzer.RecommendSavings(ctx, "user1")

	assert.NoError(t, err)
	assert.Equal(t, "user1", recommendation.UserID)
	assert.Equal(t, 6000.0, recommendation.MonthlyIncome)
	
	// Verify 50/30/20 rule calculations
	assert.Equal(t, 3000.0, recommendation.Rule5030020.Needs)   // 50% of 6000
	assert.Equal(t, 1800.0, recommendation.Rule5030020.Wants)   // 30% of 6000
	assert.Equal(t, 1200.0, recommendation.Rule5030020.Savings) // 20% of 6000
	
	// User is saving more than 20%, so savings gap should be positive (they're doing well)
	assert.GreaterOrEqual(t, recommendation.SavingsGap, 0.0) // Actually saving 40%, so gap should be positive
	assert.Greater(t, recommendation.AchievabilityScore, 7) // High achievability since already doing well
	
	mockFinanceService.AssertExpectations(t)
}

func TestBudgetAnalyzer_RecommendSavings_PoorSavings(t *testing.T) {
	analyzer, mockFinanceService := setupBudgetAnalyzer()
	ctx := context.Background()

	// User not saving enough
	summary := createTestFinanceSummary("user1", 5000.0, 4200.0, 900.0, -100.0) // Overspending
	
	expenses := []domain.Expense{
		createTestExpense("exp1", "user1", "housing", "Rent", 2500.0, "monthly", true, 1),    // 50% - high housing
		createTestExpense("exp2", "user1", "food", "Groceries", 800.0, "monthly", true, 1),   // Needs
		createTestExpense("exp3", "user1", "entertainment", "Entertainment", 900.0, "monthly", false, 3), // High wants
	}

	mockFinanceService.On("CalculateFinanceSummary", ctx, "user1").Return(summary, nil)
	mockFinanceService.On("GetUserExpenses", ctx, "user1").Return(expenses, nil)
	// Mock NormalizeToMonthly calls for each expense
	mockFinanceService.On("NormalizeToMonthly", 2500.0, "monthly").Return(2500.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 800.0, "monthly").Return(800.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 900.0, "monthly").Return(900.0, nil)

	recommendation, err := analyzer.RecommendSavings(ctx, "user1")

	assert.NoError(t, err)
	assert.Equal(t, "user1", recommendation.UserID)
	
	// Should have negative savings gap (not saving enough)
	assert.Less(t, recommendation.SavingsGap, 0.0) 
	assert.NotEmpty(t, recommendation.RecommendedActions)
	assert.Less(t, recommendation.AchievabilityScore, 5) // Low achievability due to overspending
	
	mockFinanceService.AssertExpectations(t)
}

func TestBudgetAnalyzer_IdentifyUnnecessaryExpenses_Success(t *testing.T) {
	analyzer, mockFinanceService := setupBudgetAnalyzer()
	ctx := context.Background()

	// Mix of necessary and unnecessary expenses
	expenses := []domain.Expense{
		createTestExpense("exp1", "user1", "housing", "Rent", 1500.0, "monthly", true, 1),     // Necessary
		createTestExpense("exp2", "user1", "food", "Groceries", 400.0, "monthly", true, 1),    // Necessary
		createTestExpense("exp3", "user1", "entertainment", "Streaming", 50.0, "monthly", false, 3), // Could optimize
		createTestExpense("exp4", "user1", "dining", "Restaurants", 800.0, "monthly", false, 2), // High discretionary
		createTestExpense("exp5", "user1", "shopping", "Clothes", 300.0, "monthly", false, 3),  // High discretionary
		createTestExpense("exp6", "user1", "transportation", "Uber", 200.0, "monthly", false, 2), // Could reduce
	}

	incomes := []domain.Income{
		createTestIncome("inc1", "user1", "Salary", 5000.0, "monthly", true),
	}

	mockFinanceService.On("GetUserExpenses", ctx, "user1").Return(expenses, nil)
	mockFinanceService.On("GetUserIncomes", ctx, "user1").Return(incomes, nil)
	// Mock NormalizeToMonthly calls for each expense
	mockFinanceService.On("NormalizeToMonthly", 1500.0, "monthly").Return(1500.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 400.0, "monthly").Return(400.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 50.0, "monthly").Return(50.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 800.0, "monthly").Return(800.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 300.0, "monthly").Return(300.0, nil)
	mockFinanceService.On("NormalizeToMonthly", 200.0, "monthly").Return(200.0, nil)

	optimizations, err := analyzer.IdentifyUnnecessaryExpenses(ctx, "user1")

	assert.NoError(t, err)
	assert.NotEmpty(t, optimizations)
	
	// Should identify high discretionary spending as optimization opportunities
	foundDining := false
	foundShopping := false
	
	for _, opt := range optimizations {
		assert.Greater(t, opt.PotentialSavings, 0.0)
		assert.NotEmpty(t, opt.Reasoning)
		assert.Greater(t, opt.Priority, 0)
		assert.LessOrEqual(t, opt.Priority, 5)
		
		if opt.Category == "dining" {
			foundDining = true
			assert.Equal(t, "Reduce", opt.OptimizationType)
		}
		if opt.Category == "shopping" {
			foundShopping = true
		}
	}
	
	assert.True(t, foundDining)
	assert.True(t, foundShopping)
	
	mockFinanceService.AssertExpectations(t)
}

func TestBudgetAnalyzer_GetSpendingInsights_ExpenseError(t *testing.T) {
	analyzer, mockFinanceService := setupBudgetAnalyzer()
	ctx := context.Background()

	mockFinanceService.On("GetUserExpenses", ctx, "user1").Return([]domain.Expense{}, errors.New("database error"))

	_, err := analyzer.GetSpendingInsights(ctx, "user1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user expenses")
	mockFinanceService.AssertExpectations(t)
}

func TestBudgetAnalyzer_RecommendSavings_ExpenseError(t *testing.T) {
	analyzer, mockFinanceService := setupBudgetAnalyzer()
	ctx := context.Background()

	summary := createTestFinanceSummary("user1", 5000.0, 3000.0, 600.0, 1400.0)
	mockFinanceService.On("CalculateFinanceSummary", ctx, "user1").Return(summary, nil)
	mockFinanceService.On("GetUserExpenses", ctx, "user1").Return([]domain.Expense{}, errors.New("database error"))

	_, err := analyzer.RecommendSavings(ctx, "user1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get expenses")
	mockFinanceService.AssertExpectations(t)
}

func TestBudgetAnalyzer_IdentifyUnnecessaryExpenses_IncomeError(t *testing.T) {
	analyzer, mockFinanceService := setupBudgetAnalyzer()
	ctx := context.Background()

	expenses := []domain.Expense{
		createTestExpense("exp1", "user1", "food", "Groceries", 400.0, "monthly", true, 1),
	}

	mockFinanceService.On("GetUserExpenses", ctx, "user1").Return(expenses, nil)
	mockFinanceService.On("GetUserIncomes", ctx, "user1").Return([]domain.Income{}, errors.New("database error"))

	_, err := analyzer.IdentifyUnnecessaryExpenses(ctx, "user1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user income")
	mockFinanceService.AssertExpectations(t)
}