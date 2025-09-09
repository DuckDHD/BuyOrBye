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

// Mock repositories for testing
type MockIncomeRepository struct {
	mock.Mock
}

func (m *MockIncomeRepository) SaveIncome(ctx context.Context, income domain.Income) error {
	args := m.Called(ctx, income)
	return args.Error(0)
}

func (m *MockIncomeRepository) GetIncomeByID(ctx context.Context, id string) (domain.Income, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Income), args.Error(1)
}

func (m *MockIncomeRepository) UpdateIncome(ctx context.Context, income domain.Income) error {
	args := m.Called(ctx, income)
	return args.Error(0)
}

func (m *MockIncomeRepository) DeleteIncome(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockIncomeRepository) GetUserIncomes(ctx context.Context, userID string) ([]domain.Income, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Income), args.Error(1)
}

func (m *MockIncomeRepository) GetActiveIncomes(ctx context.Context, userID string) ([]domain.Income, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Income), args.Error(1)
}

func (m *MockIncomeRepository) GetUserIncomesByFrequency(ctx context.Context, userID, frequency string) ([]domain.Income, error) {
	args := m.Called(ctx, userID, frequency)
	return args.Get(0).([]domain.Income), args.Error(1)
}

func (m *MockIncomeRepository) CalculateUserTotalIncome(ctx context.Context, userID string, activeOnly bool) (float64, error) {
	args := m.Called(ctx, userID, activeOnly)
	return args.Get(0).(float64), args.Error(1)
}

type MockExpenseRepository struct {
	mock.Mock
}

func (m *MockExpenseRepository) SaveExpense(ctx context.Context, expense domain.Expense) error {
	args := m.Called(ctx, expense)
	return args.Error(0)
}

func (m *MockExpenseRepository) GetExpenseByID(ctx context.Context, id string) (domain.Expense, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Expense), args.Error(1)
}

func (m *MockExpenseRepository) UpdateExpense(ctx context.Context, expense domain.Expense) error {
	args := m.Called(ctx, expense)
	return args.Error(0)
}

func (m *MockExpenseRepository) DeleteExpense(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockExpenseRepository) GetUserExpenses(ctx context.Context, userID string) ([]domain.Expense, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Expense), args.Error(1)
}

func (m *MockExpenseRepository) GetExpensesByCategory(ctx context.Context, userID, category string) ([]domain.Expense, error) {
	args := m.Called(ctx, userID, category)
	return args.Get(0).([]domain.Expense), args.Error(1)
}

func (m *MockExpenseRepository) GetExpensesByFrequency(ctx context.Context, userID, frequency string) ([]domain.Expense, error) {
	args := m.Called(ctx, userID, frequency)
	return args.Get(0).([]domain.Expense), args.Error(1)
}

func (m *MockExpenseRepository) GetExpensesByPriority(ctx context.Context, userID string, priority int) ([]domain.Expense, error) {
	args := m.Called(ctx, userID, priority)
	return args.Get(0).([]domain.Expense), args.Error(1)
}

func (m *MockExpenseRepository) GetFixedExpenses(ctx context.Context, userID string) ([]domain.Expense, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Expense), args.Error(1)
}

func (m *MockExpenseRepository) GetVariableExpenses(ctx context.Context, userID string) ([]domain.Expense, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Expense), args.Error(1)
}

func (m *MockExpenseRepository) CalculateUserTotalExpenses(ctx context.Context, userID string) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockExpenseRepository) CalculateTotalByCategory(ctx context.Context, userID, category string) (float64, error) {
	args := m.Called(ctx, userID, category)
	return args.Get(0).(float64), args.Error(1)
}

type MockLoanRepository struct {
	mock.Mock
}

func (m *MockLoanRepository) SaveLoan(ctx context.Context, loan domain.Loan) error {
	args := m.Called(ctx, loan)
	return args.Error(0)
}

func (m *MockLoanRepository) GetLoanByID(ctx context.Context, id string) (domain.Loan, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(domain.Loan), args.Error(1)
}

func (m *MockLoanRepository) UpdateLoan(ctx context.Context, loan domain.Loan) error {
	args := m.Called(ctx, loan)
	return args.Error(0)
}

func (m *MockLoanRepository) DeleteLoan(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockLoanRepository) GetUserLoans(ctx context.Context, userID string) ([]domain.Loan, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Loan), args.Error(1)
}

func (m *MockLoanRepository) GetLoansByType(ctx context.Context, userID, loanType string) ([]domain.Loan, error) {
	args := m.Called(ctx, userID, loanType)
	return args.Get(0).([]domain.Loan), args.Error(1)
}

func (m *MockLoanRepository) GetLoansByInterestRateRange(ctx context.Context, userID string, minRate, maxRate float64) ([]domain.Loan, error) {
	args := m.Called(ctx, userID, minRate, maxRate)
	return args.Get(0).([]domain.Loan), args.Error(1)
}

func (m *MockLoanRepository) UpdateLoanBalance(ctx context.Context, loanID string, newBalance float64) error {
	args := m.Called(ctx, loanID, newBalance)
	return args.Error(0)
}

func (m *MockLoanRepository) GetNearPayoffLoans(ctx context.Context, userID string, threshold float64) ([]domain.Loan, error) {
	args := m.Called(ctx, userID, threshold)
	return args.Get(0).([]domain.Loan), args.Error(1)
}

func (m *MockLoanRepository) CalculateUserTotalDebt(ctx context.Context, userID string) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockLoanRepository) CalculateUserMonthlyPayments(ctx context.Context, userID string) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}

type MockFinanceSummaryRepository struct {
	mock.Mock
}

func (m *MockFinanceSummaryRepository) SaveFinanceSummary(ctx context.Context, summary domain.FinanceSummary) error {
	args := m.Called(ctx, summary)
	return args.Error(0)
}

func (m *MockFinanceSummaryRepository) GetFinanceSummaryByUserID(ctx context.Context, userID string) (domain.FinanceSummary, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(domain.FinanceSummary), args.Error(1)
}

func (m *MockFinanceSummaryRepository) UpdateFinanceSummary(ctx context.Context, summary domain.FinanceSummary) error {
	args := m.Called(ctx, summary)
	return args.Error(0)
}

func (m *MockFinanceSummaryRepository) DeleteFinanceSummary(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockFinanceSummaryRepository) GetFinanceSummariesByHealthStatus(ctx context.Context, healthStatus string) ([]domain.FinanceSummary, error) {
	args := m.Called(ctx, healthStatus)
	return args.Get(0).([]domain.FinanceSummary), args.Error(1)
}

func (m *MockFinanceSummaryRepository) GetUsersWithHighDebtRatio(ctx context.Context, threshold float64) ([]domain.FinanceSummary, error) {
	args := m.Called(ctx, threshold)
	return args.Get(0).([]domain.FinanceSummary), args.Error(1)
}

func (m *MockFinanceSummaryRepository) GetUsersWithLowSavingsRate(ctx context.Context, threshold float64) ([]domain.FinanceSummary, error) {
	args := m.Called(ctx, threshold)
	return args.Get(0).([]domain.FinanceSummary), args.Error(1)
}

// Test helper functions
func createTestIncome(id, userID, source string, amount float64, frequency string, isActive bool) domain.Income {
	return domain.Income{
		ID:        id,
		UserID:    userID,
		Source:    source,
		Amount:    amount,
		Frequency: frequency,
		IsActive:  isActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestExpense(id, userID, category, name string, amount float64, frequency string, isFixed bool, priority int) domain.Expense {
	return domain.Expense{
		ID:        id,
		UserID:    userID,
		Category:  category,
		Name:      name,
		Amount:    amount,
		Frequency: frequency,
		IsFixed:   isFixed,
		Priority:  priority,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestLoan(id, userID, lender, loanType string, principal, remaining, payment, rate float64) domain.Loan {
	return domain.Loan{
		ID:               id,
		UserID:           userID,
		Lender:           lender,
		Type:             loanType,
		PrincipalAmount:  principal,
		RemainingBalance: remaining,
		MonthlyPayment:   payment,
		InterestRate:     rate,
		EndDate:          time.Now().AddDate(5, 0, 0),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func setupFinanceService() (*financeService, *MockIncomeRepository, *MockExpenseRepository, *MockLoanRepository, *MockFinanceSummaryRepository) {
	mockIncomeRepo := &MockIncomeRepository{}
	mockExpenseRepo := &MockExpenseRepository{}
	mockLoanRepo := &MockLoanRepository{}
	mockFinanceSummaryRepo := &MockFinanceSummaryRepository{}

	repos := &FinanceRepositories{
		Income:         mockIncomeRepo,
		Expense:        mockExpenseRepo,
		Loan:           mockLoanRepo,
		FinanceSummary: mockFinanceSummaryRepo,
	}

	service := &financeService{repos: repos}
	return service, mockIncomeRepo, mockExpenseRepo, mockLoanRepo, mockFinanceSummaryRepo
}

// Key service tests
func TestFinanceService_AddIncome_Success(t *testing.T) {
	service, mockIncomeRepo, _, _, _ := setupFinanceService()
	ctx := context.Background()

	income := createTestIncome("income-1", "user-1", "Salary", 5000.0, "monthly", true)
	mockIncomeRepo.On("SaveIncome", ctx, income).Return(nil)

	err := service.AddIncome(ctx, income)

	assert.NoError(t, err)
	mockIncomeRepo.AssertExpectations(t)
}

func TestFinanceService_AddIncome_InvalidData_ReturnsError(t *testing.T) {
	service, _, _, _, _ := setupFinanceService()
	ctx := context.Background()

	invalidIncome := domain.Income{} // Invalid - missing required fields

	err := service.AddIncome(ctx, invalidIncome)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid income data")
}

func TestFinanceService_CalculateFinanceSummary_Success(t *testing.T) {
	service, mockIncomeRepo, mockExpenseRepo, mockLoanRepo, _ := setupFinanceService()
	ctx := context.Background()

	incomes := []domain.Income{
		createTestIncome("income-1", "user-1", "Salary", 5000.0, "monthly", true),
		createTestIncome("income-2", "user-1", "Freelance", 1000.0, "monthly", true),
	}
	
	expenses := []domain.Expense{
		createTestExpense("exp-1", "user-1", "housing", "Rent", 2000.0, "monthly", true, 1),
		createTestExpense("exp-2", "user-1", "food", "Groceries", 500.0, "monthly", false, 1),
	}
	
	loans := []domain.Loan{
		createTestLoan("loan-1", "user-1", "Bank", "auto", 25000.0, 20000.0, 400.0, 5.0),
	}

	mockIncomeRepo.On("GetActiveIncomes", ctx, "user-1").Return(incomes, nil)
	mockExpenseRepo.On("GetUserExpenses", ctx, "user-1").Return(expenses, nil)
	mockLoanRepo.On("GetUserLoans", ctx, "user-1").Return(loans, nil)

	summary, err := service.CalculateFinanceSummary(ctx, "user-1")

	assert.NoError(t, err)
	assert.Equal(t, "user-1", summary.UserID)
	assert.Equal(t, 6000.0, summary.MonthlyIncome)
	assert.Equal(t, 2500.0, summary.MonthlyExpenses)
	assert.Equal(t, 400.0, summary.MonthlyLoanPayments)
	assert.Equal(t, 3100.0, summary.DisposableIncome)
	assert.InDelta(t, 0.0667, summary.DebtToIncomeRatio, 0.001)
	assert.Equal(t, domain.HealthExcellent, summary.FinancialHealth)
	
	mockIncomeRepo.AssertExpectations(t)
	mockExpenseRepo.AssertExpectations(t)
	mockLoanRepo.AssertExpectations(t)
}

func TestFinanceService_CalculateFinanceSummary_RepositoryError(t *testing.T) {
	service, mockIncomeRepo, _, _, _ := setupFinanceService()
	ctx := context.Background()

	mockIncomeRepo.On("GetActiveIncomes", ctx, "user-1").Return([]domain.Income{}, errors.New("db error"))

	_, err := service.CalculateFinanceSummary(ctx, "user-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user incomes")
	mockIncomeRepo.AssertExpectations(t)
}

func TestFinanceService_NormalizeToMonthly_AllFrequencies(t *testing.T) {
	service, _, _, _, _ := setupFinanceService()

	tests := []struct {
		name      string
		amount    float64
		frequency string
		expected  float64
	}{
		{"Daily", 100.0, "daily", 3000.0},
		{"Weekly", 1000.0, "weekly", 4330.0},
		{"Monthly", 5000.0, "monthly", 5000.0},
		{"Annual", 60000.0, "annual", 5000.0},
		{"OneTime", 1200.0, "one-time", 100.0},
		{"Quarterly", 3000.0, "quarterly", 1000.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.NormalizeToMonthly(tt.amount, tt.frequency)
			assert.NoError(t, err)
			assert.InDelta(t, tt.expected, result, 1.0)
		})
	}
}

func TestFinanceService_NormalizeToMonthly_InvalidFrequency(t *testing.T) {
	service, _, _, _, _ := setupFinanceService()

	_, err := service.NormalizeToMonthly(1000.0, "invalid")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported frequency")
}

func TestFinanceService_EvaluateFinancialHealth_PoorHealth(t *testing.T) {
	service, mockIncomeRepo, mockExpenseRepo, mockLoanRepo, _ := setupFinanceService()
	ctx := context.Background()

	// Poor scenario - overspending
	incomes := []domain.Income{
		createTestIncome("income-1", "user-1", "Salary", 3000.0, "monthly", true),
	}
	
	expenses := []domain.Expense{
		createTestExpense("exp-1", "user-1", "housing", "Rent", 2500.0, "monthly", true, 1),
	}
	
	loans := []domain.Loan{
		createTestLoan("loan-1", "user-1", "Bank", "auto", 25000.0, 20000.0, 1000.0, 8.0),
	}

	mockIncomeRepo.On("GetActiveIncomes", ctx, "user-1").Return(incomes, nil)
	mockExpenseRepo.On("GetUserExpenses", ctx, "user-1").Return(expenses, nil)
	mockLoanRepo.On("GetUserLoans", ctx, "user-1").Return(loans, nil)

	health, err := service.EvaluateFinancialHealth(ctx, "user-1")

	assert.NoError(t, err)
	assert.Equal(t, domain.HealthPoor, health) // Overspending
	
	mockIncomeRepo.AssertExpectations(t)
	mockExpenseRepo.AssertExpectations(t)
	mockLoanRepo.AssertExpectations(t)
}

func TestFinanceService_GetMaxAffordableAmount_Success(t *testing.T) {
	service, mockIncomeRepo, mockExpenseRepo, mockLoanRepo, _ := setupFinanceService()
	ctx := context.Background()

	incomes := []domain.Income{
		createTestIncome("income-1", "user-1", "Salary", 6000.0, "monthly", true),
	}
	
	expenses := []domain.Expense{
		createTestExpense("exp-1", "user-1", "housing", "Rent", 2000.0, "monthly", true, 1),
	}
	
	loans := []domain.Loan{
		createTestLoan("loan-1", "user-1", "Bank", "auto", 25000.0, 20000.0, 400.0, 5.0),
	}

	mockIncomeRepo.On("GetActiveIncomes", ctx, "user-1").Return(incomes, nil)
	mockExpenseRepo.On("GetUserExpenses", ctx, "user-1").Return(expenses, nil)
	mockLoanRepo.On("GetUserLoans", ctx, "user-1").Return(loans, nil)

	maxAffordable, err := service.GetMaxAffordableAmount(ctx, "user-1")

	assert.NoError(t, err)
	assert.Greater(t, maxAffordable, 0.0)
	// With good DTI and 3600 disposable, should get 3x multiplier
	assert.Equal(t, 10800.0, maxAffordable)
	
	mockIncomeRepo.AssertExpectations(t)
	mockExpenseRepo.AssertExpectations(t)
	mockLoanRepo.AssertExpectations(t)
}

func TestFinanceService_UpdateIncome_WrongOwner_ReturnsError(t *testing.T) {
	service, mockIncomeRepo, _, _, _ := setupFinanceService()
	ctx := context.Background()

	income := createTestIncome("income-1", "user-1", "Salary", 5500.0, "monthly", true)
	existingIncome := createTestIncome("income-1", "user-2", "Salary", 5000.0, "monthly", true) // Different user
	
	mockIncomeRepo.On("GetIncomeByID", ctx, "income-1").Return(existingIncome, nil)

	err := service.UpdateIncome(ctx, income)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not belong to user")
	mockIncomeRepo.AssertExpectations(t)
}

func TestFinanceService_UpdateLoanBalance_NegativeBalance_ReturnsError(t *testing.T) {
	service, _, _, mockLoanRepo, _ := setupFinanceService()
	ctx := context.Background()

	existingLoan := createTestLoan("loan-1", "user-1", "Bank", "auto", 25000.0, 20000.0, 400.0, 5.0)
	mockLoanRepo.On("GetLoanByID", ctx, "loan-1").Return(existingLoan, nil)

	err := service.UpdateLoanBalance(ctx, "user-1", "loan-1", -1000.0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be negative")
	mockLoanRepo.AssertExpectations(t)
}

// Additional comprehensive tests
func TestFinanceService_AddIncome_ValidationError(t *testing.T) {
	service, mockIncomeRepo, _, _, _ := setupFinanceService()
	ctx := context.Background()

	invalidIncome := domain.Income{
		ID:     "",
		UserID: "", // Invalid - empty UserID
		Amount: -100, // Invalid - negative amount
	}

	err := service.AddIncome(ctx, invalidIncome)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid income data")
	mockIncomeRepo.AssertNotCalled(t, "SaveIncome")
}

func TestFinanceService_AddExpense_Success(t *testing.T) {
	service, _, mockExpenseRepo, _, _ := setupFinanceService()
	ctx := context.Background()

	expense := createTestExpense("exp-1", "user-1", "food", "Groceries", 100.0, "monthly", false, 1)
	mockExpenseRepo.On("SaveExpense", ctx, expense).Return(nil)

	err := service.AddExpense(ctx, expense)

	assert.NoError(t, err)
	mockExpenseRepo.AssertExpectations(t)
}

func TestFinanceService_AddExpense_ValidationError(t *testing.T) {
	service, _, mockExpenseRepo, _, _ := setupFinanceService()
	ctx := context.Background()

	invalidExpense := domain.Expense{
		ID:     "",
		UserID: "", // Invalid - empty UserID
		Amount: -100, // Invalid - negative amount
	}

	err := service.AddExpense(ctx, invalidExpense)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid expense data")
	mockExpenseRepo.AssertNotCalled(t, "SaveExpense")
}

func TestFinanceService_AddLoan_Success(t *testing.T) {
	service, _, _, mockLoanRepo, _ := setupFinanceService()
	ctx := context.Background()

	loan := createTestLoan("loan-1", "user-1", "Bank", "auto", 25000.0, 20000.0, 400.0, 5.0)
	mockLoanRepo.On("SaveLoan", ctx, loan).Return(nil)

	err := service.AddLoan(ctx, loan)

	assert.NoError(t, err)
	mockLoanRepo.AssertExpectations(t)
}

func TestFinanceService_AddLoan_ValidationError(t *testing.T) {
	service, _, _, mockLoanRepo, _ := setupFinanceService()
	ctx := context.Background()

	invalidLoan := domain.Loan{
		ID:               "",
		UserID:           "", // Invalid - empty UserID
		PrincipalAmount:  -100, // Invalid - negative amount
	}

	err := service.AddLoan(ctx, invalidLoan)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid loan data")
	mockLoanRepo.AssertNotCalled(t, "SaveLoan")
}

func TestFinanceService_UpdateIncome_Success(t *testing.T) {
	service, mockIncomeRepo, _, _, _ := setupFinanceService()
	ctx := context.Background()

	income := createTestIncome("income-1", "user-1", "Salary", 5500.0, "monthly", true)
	existingIncome := createTestIncome("income-1", "user-1", "Salary", 5000.0, "monthly", true)
	
	mockIncomeRepo.On("GetIncomeByID", ctx, "income-1").Return(existingIncome, nil)
	mockIncomeRepo.On("UpdateIncome", ctx, income).Return(nil)

	err := service.UpdateIncome(ctx, income)

	assert.NoError(t, err)
	mockIncomeRepo.AssertExpectations(t)
}

func TestFinanceService_UpdateIncome_OwnershipMismatch(t *testing.T) {
	service, mockIncomeRepo, _, _, _ := setupFinanceService()
	ctx := context.Background()

	income := createTestIncome("income-1", "user-1", "Salary", 5000.0, "monthly", true)
	existingIncome := createTestIncome("income-1", "different-user", "Salary", 5000.0, "monthly", true)
	
	mockIncomeRepo.On("GetIncomeByID", ctx, "income-1").Return(existingIncome, nil)

	err := service.UpdateIncome(ctx, income)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "income does not belong to user")
	mockIncomeRepo.AssertNotCalled(t, "UpdateIncome")
}

func TestFinanceService_UpdateIncome_GetByIDError(t *testing.T) {
	service, mockIncomeRepo, _, _, _ := setupFinanceService()
	ctx := context.Background()

	income := createTestIncome("income-1", "user-1", "Salary", 5000.0, "monthly", true)
	
	mockIncomeRepo.On("GetIncomeByID", ctx, "income-1").Return(domain.Income{}, errors.New("not found"))

	err := service.UpdateIncome(ctx, income)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to verify income ownership")
	mockIncomeRepo.AssertNotCalled(t, "UpdateIncome")
}

func TestFinanceService_UpdateExpense_Success(t *testing.T) {
	service, _, mockExpenseRepo, _, _ := setupFinanceService()
	ctx := context.Background()

	expense := createTestExpense("exp-1", "user-1", "food", "Groceries", 150.0, "monthly", false, 1)
	existing := createTestExpense("exp-1", "user-1", "food", "Groceries", 100.0, "monthly", false, 1)
	
	mockExpenseRepo.On("GetExpenseByID", ctx, "exp-1").Return(existing, nil)
	mockExpenseRepo.On("UpdateExpense", ctx, expense).Return(nil)

	err := service.UpdateExpense(ctx, expense)

	assert.NoError(t, err)
	mockExpenseRepo.AssertExpectations(t)
}

func TestFinanceService_UpdateExpense_OwnershipMismatch(t *testing.T) {
	service, _, mockExpenseRepo, _, _ := setupFinanceService()
	ctx := context.Background()

	expense := createTestExpense("exp-1", "user-1", "food", "Groceries", 100.0, "monthly", false, 1)
	existing := createTestExpense("exp-1", "different-user", "food", "Groceries", 100.0, "monthly", false, 1)
	
	mockExpenseRepo.On("GetExpenseByID", ctx, "exp-1").Return(existing, nil)

	err := service.UpdateExpense(ctx, expense)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expense does not belong to user")
	mockExpenseRepo.AssertNotCalled(t, "UpdateExpense")
}

func TestFinanceService_UpdateLoan_Success(t *testing.T) {
	service, _, _, mockLoanRepo, _ := setupFinanceService()
	ctx := context.Background()

	loan := createTestLoan("loan-1", "user-1", "Bank", "auto", 25000.0, 18000.0, 400.0, 5.0)
	existing := createTestLoan("loan-1", "user-1", "Bank", "auto", 25000.0, 20000.0, 400.0, 5.0)
	
	mockLoanRepo.On("GetLoanByID", ctx, "loan-1").Return(existing, nil)
	mockLoanRepo.On("UpdateLoan", ctx, loan).Return(nil)

	err := service.UpdateLoan(ctx, loan)

	assert.NoError(t, err)
	mockLoanRepo.AssertExpectations(t)
}

func TestFinanceService_UpdateLoan_OwnershipMismatch(t *testing.T) {
	service, _, _, mockLoanRepo, _ := setupFinanceService()
	ctx := context.Background()

	loan := createTestLoan("loan-1", "user-1", "Bank", "auto", 25000.0, 20000.0, 400.0, 5.0)
	existing := createTestLoan("loan-1", "different-user", "Bank", "auto", 25000.0, 20000.0, 400.0, 5.0)
	
	mockLoanRepo.On("GetLoanByID", ctx, "loan-1").Return(existing, nil)

	err := service.UpdateLoan(ctx, loan)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "loan does not belong to user")
	mockLoanRepo.AssertNotCalled(t, "UpdateLoan")
}

func TestFinanceService_DeleteIncome_Success(t *testing.T) {
	service, mockIncomeRepo, _, _, _ := setupFinanceService()
	ctx := context.Background()

	existing := createTestIncome("income-1", "user-1", "Salary", 5000.0, "monthly", true)
	mockIncomeRepo.On("GetIncomeByID", ctx, "income-1").Return(existing, nil)
	mockIncomeRepo.On("DeleteIncome", ctx, "income-1").Return(nil)

	err := service.DeleteIncome(ctx, "user-1", "income-1")

	assert.NoError(t, err)
	mockIncomeRepo.AssertExpectations(t)
}

func TestFinanceService_DeleteIncome_NotFound(t *testing.T) {
	service, mockIncomeRepo, _, _, _ := setupFinanceService()
	ctx := context.Background()

	mockIncomeRepo.On("GetIncomeByID", ctx, "income-1").Return(domain.Income{}, errors.New("not found"))

	err := service.DeleteIncome(ctx, "user-1", "income-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "income not found")
	mockIncomeRepo.AssertNotCalled(t, "DeleteIncome")
}

func TestFinanceService_DeleteIncome_OwnershipMismatch(t *testing.T) {
	service, mockIncomeRepo, _, _, _ := setupFinanceService()
	ctx := context.Background()

	existing := createTestIncome("income-1", "different-user", "Salary", 5000.0, "monthly", true)
	mockIncomeRepo.On("GetIncomeByID", ctx, "income-1").Return(existing, nil)

	err := service.DeleteIncome(ctx, "user-1", "income-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "income does not belong to user")
	mockIncomeRepo.AssertNotCalled(t, "DeleteIncome")
}

func TestFinanceService_DeleteExpense_Success(t *testing.T) {
	service, _, mockExpenseRepo, _, _ := setupFinanceService()
	ctx := context.Background()

	existing := createTestExpense("exp-1", "user-1", "food", "Groceries", 100.0, "monthly", false, 1)
	mockExpenseRepo.On("GetExpenseByID", ctx, "exp-1").Return(existing, nil)
	mockExpenseRepo.On("DeleteExpense", ctx, "exp-1").Return(nil)

	err := service.DeleteExpense(ctx, "user-1", "exp-1")

	assert.NoError(t, err)
	mockExpenseRepo.AssertExpectations(t)
}

func TestFinanceService_DeleteExpense_OwnershipMismatch(t *testing.T) {
	service, _, mockExpenseRepo, _, _ := setupFinanceService()
	ctx := context.Background()

	existing := createTestExpense("exp-1", "different-user", "food", "Groceries", 100.0, "monthly", false, 1)
	mockExpenseRepo.On("GetExpenseByID", ctx, "exp-1").Return(existing, nil)

	err := service.DeleteExpense(ctx, "user-1", "exp-1")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expense does not belong to user")
	mockExpenseRepo.AssertNotCalled(t, "DeleteExpense")
}

func TestFinanceService_GetUserIncomes_Success(t *testing.T) {
	service, mockIncomeRepo, _, _, _ := setupFinanceService()
	ctx := context.Background()

	expectedIncomes := []domain.Income{
		createTestIncome("income-1", "user-1", "Salary", 5000.0, "monthly", true),
		createTestIncome("income-2", "user-1", "Freelance", 1000.0, "monthly", true),
	}
	
	mockIncomeRepo.On("GetUserIncomes", ctx, "user-1").Return(expectedIncomes, nil)

	incomes, err := service.GetUserIncomes(ctx, "user-1")

	assert.NoError(t, err)
	assert.Equal(t, expectedIncomes, incomes)
	mockIncomeRepo.AssertExpectations(t)
}

func TestFinanceService_GetActiveUserIncomes_Success(t *testing.T) {
	service, mockIncomeRepo, _, _, _ := setupFinanceService()
	ctx := context.Background()

	expectedIncomes := []domain.Income{
		createTestIncome("income-1", "user-1", "Salary", 5000.0, "monthly", true),
	}
	
	mockIncomeRepo.On("GetActiveIncomes", ctx, "user-1").Return(expectedIncomes, nil)

	incomes, err := service.GetActiveUserIncomes(ctx, "user-1")

	assert.NoError(t, err)
	assert.Equal(t, expectedIncomes, incomes)
	mockIncomeRepo.AssertExpectations(t)
}

func TestFinanceService_GetUserExpenses_Success(t *testing.T) {
	service, _, mockExpenseRepo, _, _ := setupFinanceService()
	ctx := context.Background()

	expectedExpenses := []domain.Expense{
		createTestExpense("exp-1", "user-1", "food", "Groceries", 100.0, "monthly", false, 1),
		createTestExpense("exp-2", "user-1", "housing", "Rent", 2000.0, "monthly", true, 1),
	}
	
	mockExpenseRepo.On("GetUserExpenses", ctx, "user-1").Return(expectedExpenses, nil)

	expenses, err := service.GetUserExpenses(ctx, "user-1")

	assert.NoError(t, err)
	assert.Equal(t, expectedExpenses, expenses)
	mockExpenseRepo.AssertExpectations(t)
}

func TestFinanceService_GetUserExpensesByCategory_Success(t *testing.T) {
	service, _, mockExpenseRepo, _, _ := setupFinanceService()
	ctx := context.Background()

	expectedExpenses := []domain.Expense{
		createTestExpense("exp-1", "user-1", "food", "Groceries", 100.0, "monthly", false, 1),
		createTestExpense("exp-2", "user-1", "food", "Restaurant", 50.0, "weekly", false, 2),
	}
	
	mockExpenseRepo.On("GetExpensesByCategory", ctx, "user-1", "food").Return(expectedExpenses, nil)

	expenses, err := service.GetUserExpensesByCategory(ctx, "user-1", "food")

	assert.NoError(t, err)
	assert.Equal(t, expectedExpenses, expenses)
	mockExpenseRepo.AssertExpectations(t)
}

func TestFinanceService_GetUserLoans_Success(t *testing.T) {
	service, _, _, mockLoanRepo, _ := setupFinanceService()
	ctx := context.Background()

	expectedLoans := []domain.Loan{
		createTestLoan("loan-1", "user-1", "Bank", "auto", 25000.0, 20000.0, 400.0, 5.0),
		createTestLoan("loan-2", "user-1", "Credit Union", "personal", 10000.0, 8000.0, 200.0, 7.0),
	}
	
	mockLoanRepo.On("GetUserLoans", ctx, "user-1").Return(expectedLoans, nil)

	loans, err := service.GetUserLoans(ctx, "user-1")

	assert.NoError(t, err)
	assert.Equal(t, expectedLoans, loans)
	mockLoanRepo.AssertExpectations(t)
}

func TestFinanceService_UpdateLoanBalance_Success(t *testing.T) {
	service, _, _, mockLoanRepo, _ := setupFinanceService()
	ctx := context.Background()

	existing := createTestLoan("loan-1", "user-1", "Bank", "auto", 25000.0, 20000.0, 400.0, 5.0)
	mockLoanRepo.On("GetLoanByID", ctx, "loan-1").Return(existing, nil)
	mockLoanRepo.On("UpdateLoanBalance", ctx, "loan-1", 18000.0).Return(nil)

	err := service.UpdateLoanBalance(ctx, "user-1", "loan-1", 18000.0)

	assert.NoError(t, err)
	mockLoanRepo.AssertExpectations(t)
}

func TestFinanceService_UpdateLoanBalance_OwnershipMismatch(t *testing.T) {
	service, _, _, mockLoanRepo, _ := setupFinanceService()
	ctx := context.Background()

	existing := createTestLoan("loan-1", "different-user", "Bank", "auto", 25000.0, 20000.0, 400.0, 5.0)
	mockLoanRepo.On("GetLoanByID", ctx, "loan-1").Return(existing, nil)

	err := service.UpdateLoanBalance(ctx, "user-1", "loan-1", 18000.0)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "loan does not belong to user")
	mockLoanRepo.AssertNotCalled(t, "UpdateLoanBalance")
}

func TestFinanceService_CalculateDisposableIncome_Success(t *testing.T) {
	service, mockIncomeRepo, mockExpenseRepo, mockLoanRepo, _ := setupFinanceService()
	ctx := context.Background()

	incomes := []domain.Income{
		createTestIncome("income-1", "user-1", "Salary", 6000.0, "monthly", true),
	}
	
	expenses := []domain.Expense{
		createTestExpense("exp-1", "user-1", "housing", "Rent", 2000.0, "monthly", true, 1),
	}
	
	loans := []domain.Loan{
		createTestLoan("loan-1", "user-1", "Bank", "auto", 25000.0, 20000.0, 400.0, 5.0),
	}

	mockIncomeRepo.On("GetActiveIncomes", ctx, "user-1").Return(incomes, nil)
	mockExpenseRepo.On("GetUserExpenses", ctx, "user-1").Return(expenses, nil)
	mockLoanRepo.On("GetUserLoans", ctx, "user-1").Return(loans, nil)

	disposableIncome, err := service.CalculateDisposableIncome(ctx, "user-1")

	assert.NoError(t, err)
	assert.Equal(t, 3600.0, disposableIncome) // 6000 - 2000 - 400
	
	mockIncomeRepo.AssertExpectations(t)
	mockExpenseRepo.AssertExpectations(t)
	mockLoanRepo.AssertExpectations(t)
}

func TestFinanceService_CalculateDebtToIncomeRatio_Success(t *testing.T) {
	service, mockIncomeRepo, mockExpenseRepo, mockLoanRepo, _ := setupFinanceService()
	ctx := context.Background()

	incomes := []domain.Income{
		createTestIncome("income-1", "user-1", "Salary", 6000.0, "monthly", true),
	}
	
	expenses := []domain.Expense{
		createTestExpense("exp-1", "user-1", "housing", "Rent", 2000.0, "monthly", true, 1),
	}
	
	loans := []domain.Loan{
		createTestLoan("loan-1", "user-1", "Bank", "auto", 25000.0, 20000.0, 1200.0, 5.0),
	}

	mockIncomeRepo.On("GetActiveIncomes", ctx, "user-1").Return(incomes, nil)
	mockExpenseRepo.On("GetUserExpenses", ctx, "user-1").Return(expenses, nil)
	mockLoanRepo.On("GetUserLoans", ctx, "user-1").Return(loans, nil)

	dtiRatio, err := service.CalculateDebtToIncomeRatio(ctx, "user-1")

	assert.NoError(t, err)
	assert.Equal(t, 20.0, dtiRatio) // (1200 / 6000) * 100 = 20%
	
	mockIncomeRepo.AssertExpectations(t)
	mockExpenseRepo.AssertExpectations(t)
	mockLoanRepo.AssertExpectations(t)
}

func TestFinanceService_CalculateFinanceSummary_NoData(t *testing.T) {
	service, mockIncomeRepo, mockExpenseRepo, mockLoanRepo, _ := setupFinanceService()
	ctx := context.Background()

	mockIncomeRepo.On("GetActiveIncomes", ctx, "user-1").Return([]domain.Income{}, nil)
	mockExpenseRepo.On("GetUserExpenses", ctx, "user-1").Return([]domain.Expense{}, nil)
	mockLoanRepo.On("GetUserLoans", ctx, "user-1").Return([]domain.Loan{}, nil)

	summary, err := service.CalculateFinanceSummary(ctx, "user-1")

	assert.NoError(t, err)
	assert.Equal(t, "user-1", summary.UserID)
	assert.Equal(t, 0.0, summary.MonthlyIncome)
	assert.Equal(t, 0.0, summary.MonthlyExpenses)
	assert.Equal(t, 0.0, summary.MonthlyLoanPayments)
	assert.Equal(t, 0.0, summary.DisposableIncome)
	assert.Equal(t, 0.0, summary.DebtToIncomeRatio)
	assert.Equal(t, 0.0, summary.SavingsRate)
}