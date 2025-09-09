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

// Mock for DebtCalculator dependencies
type MockDebtCalculatorFinanceService struct {
	mock.Mock
}

func (m *MockDebtCalculatorFinanceService) GetUserLoans(ctx context.Context, userID string) ([]domain.Loan, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Loan), args.Error(1)
}

func (m *MockDebtCalculatorFinanceService) CalculateFinanceSummary(ctx context.Context, userID string) (domain.FinanceSummary, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(domain.FinanceSummary), args.Error(1)
}

// Implement all other required methods (not used by DebtCalculator but needed for interface)
func (m *MockDebtCalculatorFinanceService) AddIncome(ctx context.Context, income domain.Income) error { return nil }
func (m *MockDebtCalculatorFinanceService) UpdateIncome(ctx context.Context, income domain.Income) error { return nil }
func (m *MockDebtCalculatorFinanceService) DeleteIncome(ctx context.Context, userID, incomeID string) error { return nil }
func (m *MockDebtCalculatorFinanceService) GetUserIncomes(ctx context.Context, userID string) ([]domain.Income, error) { return nil, nil }
func (m *MockDebtCalculatorFinanceService) GetActiveUserIncomes(ctx context.Context, userID string) ([]domain.Income, error) { return nil, nil }
func (m *MockDebtCalculatorFinanceService) AddExpense(ctx context.Context, expense domain.Expense) error { return nil }
func (m *MockDebtCalculatorFinanceService) UpdateExpense(ctx context.Context, expense domain.Expense) error { return nil }
func (m *MockDebtCalculatorFinanceService) DeleteExpense(ctx context.Context, userID, expenseID string) error { return nil }
func (m *MockDebtCalculatorFinanceService) GetUserExpenses(ctx context.Context, userID string) ([]domain.Expense, error) { return nil, nil }
func (m *MockDebtCalculatorFinanceService) GetUserExpensesByCategory(ctx context.Context, userID, category string) ([]domain.Expense, error) { return nil, nil }
func (m *MockDebtCalculatorFinanceService) AddLoan(ctx context.Context, loan domain.Loan) error { return nil }
func (m *MockDebtCalculatorFinanceService) UpdateLoan(ctx context.Context, loan domain.Loan) error { return nil }
func (m *MockDebtCalculatorFinanceService) UpdateLoanBalance(ctx context.Context, userID, loanID string, newBalance float64) error { return nil }
func (m *MockDebtCalculatorFinanceService) NormalizeToMonthly(amount float64, frequency string) (float64, error) { return amount, nil }
func (m *MockDebtCalculatorFinanceService) CalculateDisposableIncome(ctx context.Context, userID string) (float64, error) { return 0, nil }
func (m *MockDebtCalculatorFinanceService) CalculateDebtToIncomeRatio(ctx context.Context, userID string) (float64, error) { return 0, nil }
func (m *MockDebtCalculatorFinanceService) EvaluateFinancialHealth(ctx context.Context, userID string) (string, error) { return "", nil }
func (m *MockDebtCalculatorFinanceService) GetMaxAffordableAmount(ctx context.Context, userID string) (float64, error) { return 0, nil }

func setupDebtCalculator() (*debtCalculator, *MockDebtCalculatorFinanceService) {
	mockFinanceService := &MockDebtCalculatorFinanceService{}
	calculator := &debtCalculator{
		financeService: mockFinanceService,
	}
	return calculator, mockFinanceService
}

func createTestFinanceSummaryForDebt(userID string, monthlyIncome, monthlyLoans float64) domain.FinanceSummary {
	return domain.FinanceSummary{
		UserID:              userID,
		MonthlyIncome:       monthlyIncome,
		MonthlyLoanPayments: monthlyLoans,
		DebtToIncomeRatio:   monthlyLoans / monthlyIncome,
		UpdatedAt:          time.Now(),
	}
}

func TestDebtCalculator_GetDebtAnalysis_MultipleLoans(t *testing.T) {
	calculator, mockFinanceService := setupDebtCalculator()
	ctx := context.Background()

	loans := []domain.Loan{
		createTestLoan("loan1", "user1", "Bank A", "credit_card", 5000.0, 4500.0, 200.0, 18.99),
		createTestLoan("loan2", "user1", "Bank B", "auto", 25000.0, 20000.0, 450.0, 6.5),
		createTestLoan("loan3", "user1", "Credit Union", "personal", 10000.0, 8000.0, 300.0, 12.0),
	}

	summary := createTestFinanceSummaryForDebt("user1", 6000.0, 950.0)

	mockFinanceService.On("GetUserLoans", ctx, "user1").Return(loans, nil)
	mockFinanceService.On("CalculateFinanceSummary", ctx, "user1").Return(summary, nil)

	analysis, err := calculator.GetDebtAnalysis(ctx, "user1")

	assert.NoError(t, err)
	assert.Equal(t, "user1", analysis.UserID)
	assert.Equal(t, 32500.0, analysis.TotalDebt) // Sum of remaining balances
	assert.Equal(t, 950.0, analysis.TotalMonthlyPayments) // Sum of monthly payments
	assert.Greater(t, analysis.WeightedAverageRate, 0.0)
	assert.Equal(t, "loan1", analysis.HighestRateLoan.ID) // Credit card at 18.99%
	assert.Equal(t, "loan2", analysis.LowestRateLoan.ID)  // Auto loan at 6.5%
	assert.Equal(t, "loan2", analysis.LargestBalanceLoan.ID) // Auto loan with $20k
	assert.Equal(t, "loan3", analysis.SmallestBalanceLoan.ID) // Personal loan with $8k
	assert.NotEmpty(t, analysis.DebtHealthStatus)
	assert.NotEmpty(t, analysis.Recommendations)
	assert.NotEmpty(t, analysis.PayoffProjections)
	
	mockFinanceService.AssertExpectations(t)
}

func TestDebtCalculator_GetDebtAnalysis_SingleLoan(t *testing.T) {
	calculator, mockFinanceService := setupDebtCalculator()
	ctx := context.Background()

	loans := []domain.Loan{
		createTestLoan("loan1", "user1", "Bank", "auto", 25000.0, 15000.0, 400.0, 5.5),
	}

	summary := createTestFinanceSummaryForDebt("user1", 5000.0, 400.0)

	mockFinanceService.On("GetUserLoans", ctx, "user1").Return(loans, nil)
	mockFinanceService.On("CalculateFinanceSummary", ctx, "user1").Return(summary, nil)

	analysis, err := calculator.GetDebtAnalysis(ctx, "user1")

	assert.NoError(t, err)
	assert.Equal(t, "user1", analysis.UserID)
	assert.Equal(t, 15000.0, analysis.TotalDebt)
	assert.Equal(t, 400.0, analysis.TotalMonthlyPayments)
	assert.Equal(t, 5.5, analysis.WeightedAverageRate) // Only one loan
	assert.Equal(t, "loan1", analysis.HighestRateLoan.ID)
	assert.Equal(t, "loan1", analysis.LowestRateLoan.ID)
	assert.Equal(t, "loan1", analysis.LargestBalanceLoan.ID)
	assert.Equal(t, "loan1", analysis.SmallestBalanceLoan.ID)
	
	mockFinanceService.AssertExpectations(t)
}

func TestDebtCalculator_CalculateTotalDebt_NoLoans(t *testing.T) {
	calculator, mockFinanceService := setupDebtCalculator()
	ctx := context.Background()

	mockFinanceService.On("GetUserLoans", ctx, "user1").Return([]domain.Loan{}, nil)

	totalDebt, err := calculator.CalculateTotalDebt(ctx, "user1")

	assert.NoError(t, err)
	assert.Equal(t, 0.0, totalDebt)
	
	mockFinanceService.AssertExpectations(t)
}

func TestDebtCalculator_CalculateTotalDebt_ServiceError(t *testing.T) {
	calculator, mockFinanceService := setupDebtCalculator()
	ctx := context.Background()

	mockFinanceService.On("GetUserLoans", ctx, "user1").Return([]domain.Loan{}, errors.New("database error"))

	_, err := calculator.CalculateTotalDebt(ctx, "user1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user loans")
	mockFinanceService.AssertExpectations(t)
}

func TestDebtCalculator_SuggestPaymentStrategy_AvalancheVsSnowball(t *testing.T) {
	calculator, mockFinanceService := setupDebtCalculator()
	ctx := context.Background()

	// Multiple loans with different rates and balances
	loans := []domain.Loan{
		createTestLoan("loan1", "user1", "Bank A", "credit_card", 3000.0, 2800.0, 100.0, 22.99),   // High rate, low balance
		createTestLoan("loan2", "user1", "Bank B", "auto", 25000.0, 18000.0, 400.0, 6.5),         // Low rate, high balance
		createTestLoan("loan3", "user1", "Credit Union", "personal", 8000.0, 5000.0, 200.0, 15.0), // Medium rate, medium balance
	}

	summary := createTestFinanceSummaryForDebt("user1", 6000.0, 700.0)

	mockFinanceService.On("GetUserLoans", ctx, "user1").Return(loans, nil)
	mockFinanceService.On("CalculateFinanceSummary", ctx, "user1").Return(summary, nil)

	strategy, err := calculator.SuggestPaymentStrategy(ctx, "user1", 300.0)

	assert.NoError(t, err)
	assert.Equal(t, "user1", strategy.UserID)
	assert.Equal(t, 300.0, strategy.ExtraPaymentAmount)
	assert.Contains(t, []string{"Avalanche", "Snowball", "Custom"}, strategy.StrategyType)
	assert.NotEmpty(t, strategy.PrioritizedLoans)
	assert.Len(t, strategy.PrioritizedLoans, 3) // Should have all loans
	assert.Greater(t, strategy.TotalInterestSaved, 0.0) // Should save interest with extra payments
	assert.Greater(t, strategy.MonthsSaved, 0) // Should save time
	assert.NotEmpty(t, strategy.RecommendedReason)
	assert.Greater(t, strategy.MonthlyPaymentPlan, 700.0) // Original + extra
	
	mockFinanceService.AssertExpectations(t)
}

func TestDebtCalculator_SuggestPaymentStrategy_NoExtraPayment(t *testing.T) {
	calculator, mockFinanceService := setupDebtCalculator()
	ctx := context.Background()

	loans := []domain.Loan{
		createTestLoan("loan1", "user1", "Bank", "auto", 20000.0, 15000.0, 350.0, 7.5),
	}

	summary := createTestFinanceSummaryForDebt("user1", 5000.0, 350.0)

	mockFinanceService.On("GetUserLoans", ctx, "user1").Return(loans, nil)
	mockFinanceService.On("CalculateFinanceSummary", ctx, "user1").Return(summary, nil)

	strategy, err := calculator.SuggestPaymentStrategy(ctx, "user1", 0.0)

	assert.NoError(t, err)
	assert.Equal(t, "user1", strategy.UserID)
	assert.Equal(t, 0.0, strategy.ExtraPaymentAmount)
	assert.Equal(t, 0.0, strategy.TotalInterestSaved) // No extra payment = no savings
	assert.Equal(t, 0, strategy.MonthsSaved)
	assert.Equal(t, 350.0, strategy.MonthlyPaymentPlan) // Just minimum payments
	
	mockFinanceService.AssertExpectations(t)
}

func TestDebtCalculator_CalculateInterestSavings_WithExtraPayment(t *testing.T) {
	calculator, mockFinanceService := setupDebtCalculator()
	ctx := context.Background()

	loans := []domain.Loan{
		createTestLoan("loan1", "user1", "Bank", "personal", 10000.0, 8000.0, 300.0, 12.0),
	}

	mockFinanceService.On("GetUserLoans", ctx, "user1").Return(loans, nil)

	savings, err := calculator.CalculateInterestSavings(ctx, "user1", 200.0)

	assert.NoError(t, err)
	assert.Equal(t, "user1", savings.UserID)
	assert.Equal(t, 200.0, savings.ExtraPaymentAmount)
	assert.Greater(t, savings.InterestSaved, 0.0) // Should save interest
	assert.Greater(t, savings.MonthsSaved, 0) // Should save time
	assert.Greater(t, savings.CurrentTotalInterest, savings.NewTotalInterest) // New should be less
	assert.True(t, savings.NewDebtFreeDate.Before(savings.CurrentDebtFreeDate)) // Earlier payoff
	assert.Greater(t, savings.BreakEvenMonths, 0) // Should have break-even period
	
	mockFinanceService.AssertExpectations(t)
}

func TestDebtCalculator_CalculateInterestSavings_NoLoans(t *testing.T) {
	calculator, mockFinanceService := setupDebtCalculator()
	ctx := context.Background()

	mockFinanceService.On("GetUserLoans", ctx, "user1").Return([]domain.Loan{}, nil)

	savings, err := calculator.CalculateInterestSavings(ctx, "user1", 100.0)

	assert.NoError(t, err)
	assert.Equal(t, "user1", savings.UserID)
	assert.Equal(t, 0.0, savings.InterestSaved) // No loans = no savings
	assert.Equal(t, 0, savings.MonthsSaved)
	assert.Equal(t, 0.0, savings.CurrentTotalInterest)
	assert.Equal(t, 0.0, savings.NewTotalInterest)
	
	mockFinanceService.AssertExpectations(t)
}

func TestDebtCalculator_ProjectDebtFreeDate_MultipleLoans(t *testing.T) {
	calculator, mockFinanceService := setupDebtCalculator()
	ctx := context.Background()

	// Loans with reasonable payment schedules
	loans := []domain.Loan{
		createTestLoan("loan1", "user1", "Bank A", "credit_card", 5000.0, 3000.0, 150.0, 24.99),
		createTestLoan("loan2", "user1", "Bank B", "auto", 20000.0, 15000.0, 350.0, 6.5),
	}

	mockFinanceService.On("GetUserLoans", ctx, "user1").Return(loans, nil)

	debtFreeDate, err := calculator.ProjectDebtFreeDate(ctx, "user1")

	assert.NoError(t, err)
	assert.True(t, debtFreeDate.After(time.Now())) // Should be in the future
	
	mockFinanceService.AssertExpectations(t)
}

func TestDebtCalculator_ProjectDebtFreeDate_NoLoans(t *testing.T) {
	calculator, mockFinanceService := setupDebtCalculator()
	ctx := context.Background()

	mockFinanceService.On("GetUserLoans", ctx, "user1").Return([]domain.Loan{}, nil)

	debtFreeDate, err := calculator.ProjectDebtFreeDate(ctx, "user1")

	assert.NoError(t, err)
	// Should be current time when no loans exist
	assert.True(t, time.Since(debtFreeDate) < time.Hour) // Within an hour of now
	
	mockFinanceService.AssertExpectations(t)
}

func TestDebtCalculator_CalculateTotalDebt_WithLoans(t *testing.T) {
	calculator, mockFinanceService := setupDebtCalculator()
	ctx := context.Background()

	loans := []domain.Loan{
		createTestLoan("loan1", "user1", "Bank A", "auto", 20000.0, 15000.0, 350.0, 7.5),
		createTestLoan("loan2", "user1", "Bank B", "credit_card", 5000.0, 3000.0, 150.0, 18.0),
	}

	mockFinanceService.On("GetUserLoans", ctx, "user1").Return(loans, nil)

	totalDebt, err := calculator.CalculateTotalDebt(ctx, "user1")

	assert.NoError(t, err)
	assert.Equal(t, 18000.0, totalDebt) // 15000 + 3000
	mockFinanceService.AssertExpectations(t)
}

func TestDebtCalculator_SuggestPaymentStrategy_HighDebtToIncomeRatio(t *testing.T) {
	calculator, mockFinanceService := setupDebtCalculator()
	ctx := context.Background()

	loans := []domain.Loan{
		createTestLoan("loan1", "user1", "Bank", "credit_card", 10000.0, 8000.0, 400.0, 20.0),
		createTestLoan("loan2", "user1", "Auto", "auto", 30000.0, 25000.0, 600.0, 8.0),
	}

	// High DTI ratio scenario (40% DTI = $1000 payments on $2500 income)
	summary := createTestFinanceSummaryForDebt("user1", 2500.0, 1000.0)

	mockFinanceService.On("GetUserLoans", ctx, "user1").Return(loans, nil)
	mockFinanceService.On("CalculateFinanceSummary", ctx, "user1").Return(summary, nil)

	strategy, err := calculator.SuggestPaymentStrategy(ctx, "user1", 100.0)

	assert.NoError(t, err)
	assert.Equal(t, "user1", strategy.UserID)
	
	// Should recommend avalanche for high DTI to minimize interest
	assert.Equal(t, "Avalanche", strategy.StrategyType)
	assert.Contains(t, strategy.RecommendedReason, "high debt-to-income ratio")
	
	// First loan should be the credit card (higher interest)
	assert.Equal(t, "loan1", strategy.PrioritizedLoans[0].LoanID)
	assert.Equal(t, 1, strategy.PrioritizedLoans[0].PayoffOrder)
	
	mockFinanceService.AssertExpectations(t)
}

func TestDebtCalculator_CalculateInterestSavings_LoanError(t *testing.T) {
	calculator, mockFinanceService := setupDebtCalculator()
	ctx := context.Background()

	mockFinanceService.On("GetUserLoans", ctx, "user1").Return([]domain.Loan{}, errors.New("database error"))

	_, err := calculator.CalculateInterestSavings(ctx, "user1", 200.0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user loans")
	mockFinanceService.AssertExpectations(t)
}

func TestDebtCalculator_ProjectDebtFreeDate_ServiceError(t *testing.T) {
	calculator, mockFinanceService := setupDebtCalculator()
	ctx := context.Background()

	mockFinanceService.On("GetUserLoans", ctx, "user1").Return([]domain.Loan{}, errors.New("database error"))

	_, err := calculator.ProjectDebtFreeDate(ctx, "user1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user loans")
	mockFinanceService.AssertExpectations(t)
}