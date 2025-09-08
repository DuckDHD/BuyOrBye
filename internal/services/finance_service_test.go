package services

// import (
// 	"context"
// 	"errors"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"

// 	"github.com/DuckDHD/BuyOrBye/internal/domain"
// )

// // MockIncomeRepository is a mock implementation of IncomeRepository
// type MockIncomeRepository struct {
// 	mock.Mock
// }

// func (m *MockIncomeRepository) Save(ctx context.Context, income *domain.Income) error {
// 	args := m.Called(ctx, income)
// 	return args.Error(0)
// }

// func (m *MockIncomeRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Income, error) {
// 	args := m.Called(ctx, userID)
// 	return args.Get(0).([]domain.Income), args.Error(1)
// }

// func (m *MockIncomeRepository) Update(ctx context.Context, income *domain.Income) error {
// 	args := m.Called(ctx, income)
// 	return args.Error(0)
// }

// func (m *MockIncomeRepository) Delete(ctx context.Context, userID, incomeID string) error {
// 	args := m.Called(ctx, userID, incomeID)
// 	return args.Error(0)
// }

// // MockExpenseRepository is a mock implementation of ExpenseRepository
// type MockExpenseRepository struct {
// 	mock.Mock
// }

// func (m *MockExpenseRepository) Save(ctx context.Context, expense *domain.Expense) error {
// 	args := m.Called(ctx, expense)
// 	return args.Error(0)
// }

// func (m *MockExpenseRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Expense, error) {
// 	args := m.Called(ctx, userID)
// 	return args.Get(0).([]domain.Expense), args.Error(1)
// }

// func (m *MockExpenseRepository) GetByCategory(ctx context.Context, userID, category string) ([]domain.Expense, error) {
// 	args := m.Called(ctx, userID, category)
// 	return args.Get(0).([]domain.Expense), args.Error(1)
// }

// func (m *MockExpenseRepository) Update(ctx context.Context, expense *domain.Expense) error {
// 	args := m.Called(ctx, expense)
// 	return args.Error(0)
// }

// func (m *MockExpenseRepository) Delete(ctx context.Context, userID, expenseID string) error {
// 	args := m.Called(ctx, userID, expenseID)
// 	return args.Error(0)
// }

// // MockLoanRepository is a mock implementation of LoanRepository
// type MockLoanRepository struct {
// 	mock.Mock
// }

// func (m *MockLoanRepository) Save(ctx context.Context, loan *domain.Loan) error {
// 	args := m.Called(ctx, loan)
// 	return args.Error(0)
// }

// func (m *MockLoanRepository) GetByUserID(ctx context.Context, userID string) ([]domain.Loan, error) {
// 	args := m.Called(ctx, userID)
// 	return args.Get(0).([]domain.Loan), args.Error(1)
// }

// func (m *MockLoanRepository) Update(ctx context.Context, loan *domain.Loan) error {
// 	args := m.Called(ctx, loan)
// 	return args.Error(0)
// }

// func (m *MockLoanRepository) Delete(ctx context.Context, userID, loanID string) error {
// 	args := m.Called(ctx, userID, loanID)
// 	return args.Error(0)
// }

// func TestFinanceService_CalculateDisposableIncome_PositiveDisposableIncome_ReturnsCorrectAmount(t *testing.T) {
// 	// Arrange
// 	ctx := context.Background()
// 	userID := "user-123"

// 	incomes := []domain.Income{
// 		{Amount: 5000.00, Frequency: "monthly", IsActive: true},
// 		{Amount: 1000.00, Frequency: "weekly", IsActive: true}, // 4333.33 monthly
// 	}
// 	expenses := []domain.Expense{
// 		{Amount: 1200.00, Frequency: "monthly"},
// 		{Amount: 150.00, Frequency: "weekly"}, // 650 monthly
// 	}
// 	loans := []domain.Loan{
// 		{MonthlyPayment: 800.00},
// 		{MonthlyPayment: 300.00},
// 	}

// 	mockIncomeRepo := &MockIncomeRepository{}
// 	mockExpenseRepo := &MockExpenseRepository{}
// 	mockLoanRepo := &MockLoanRepository{}

// 	mockIncomeRepo.On("GetByUserID", ctx, userID).Return(incomes, nil)
// 	mockExpenseRepo.On("GetByUserID", ctx, userID).Return(expenses, nil)
// 	mockLoanRepo.On("GetByUserID", ctx, userID).Return(loans, nil)

// 	service := NewFinanceService(mockIncomeRepo, mockExpenseRepo, mockLoanRepo)

// 	// Act
// 	disposableIncome, err := service.CalculateDisposableIncome(ctx, userID)

// 	// Assert
// 	assert.NoError(t, err)
// 	// Total monthly income: 5000 + 4333.33 = 9333.33
// 	// Total monthly expenses: 1200 + 650 = 1850
// 	// Total monthly loans: 800 + 300 = 1100
// 	// Disposable: 9333.33 - 1850 - 1100 = 6383.33
// 	assert.InDelta(t, 6383.33, disposableIncome, 0.01)
// 	mockIncomeRepo.AssertExpectations(t)
// 	mockExpenseRepo.AssertExpectations(t)
// 	mockLoanRepo.AssertExpectations(t)
// }

// func TestFinanceService_CalculateDisposableIncome_NegativeDisposableIncome_ReturnsNegativeAmount(t *testing.T) {
// 	// Arrange
// 	ctx := context.Background()
// 	userID := "user-456"

// 	incomes := []domain.Income{
// 		{Amount: 3000.00, Frequency: "monthly", IsActive: true},
// 	}
// 	expenses := []domain.Expense{
// 		{Amount: 2500.00, Frequency: "monthly"},
// 		{Amount: 200.00, Frequency: "weekly"}, // 866.67 monthly
// 	}
// 	loans := []domain.Loan{
// 		{MonthlyPayment: 1000.00},
// 	}

// 	mockIncomeRepo := &MockIncomeRepository{}
// 	mockExpenseRepo := &MockExpenseRepository{}
// 	mockLoanRepo := &MockLoanRepository{}

// 	mockIncomeRepo.On("GetByUserID", ctx, userID).Return(incomes, nil)
// 	mockExpenseRepo.On("GetByUserID", ctx, userID).Return(expenses, nil)
// 	mockLoanRepo.On("GetByUserID", ctx, userID).Return(loans, nil)

// 	service := NewFinanceService(mockIncomeRepo, mockExpenseRepo, mockLoanRepo)

// 	// Act
// 	disposableIncome, err := service.CalculateDisposableIncome(ctx, userID)

// 	// Assert
// 	assert.NoError(t, err)
// 	// Total monthly income: 3000
// 	// Total monthly expenses: 2500 + 866.67 = 3366.67
// 	// Total monthly loans: 1000
// 	// Disposable: 3000 - 3366.67 - 1000 = -1366.67
// 	assert.InDelta(t, -1366.67, disposableIncome, 0.01)
// 	mockIncomeRepo.AssertExpectations(t)
// 	mockExpenseRepo.AssertExpectations(t)
// 	mockLoanRepo.AssertExpectations(t)
// }

// func TestFinanceService_CalculateDisposableIncome_RepositoryError_ReturnsError(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		setupMocks     func(*MockIncomeRepository, *MockExpenseRepository, *MockLoanRepository)
// 		expectedError  string
// 	}{
// 		{
// 			name: "income_repo_error",
// 			setupMocks: func(incomeRepo *MockIncomeRepository, expenseRepo *MockExpenseRepository, loanRepo *MockLoanRepository) {
// 				incomeRepo.On("GetByUserID", mock.Anything, mock.Anything).Return([]domain.Income{}, errors.New("income repo error"))
// 			},
// 			expectedError: "failed to get incomes: income repo error",
// 		},
// 		{
// 			name: "expense_repo_error",
// 			setupMocks: func(incomeRepo *MockIncomeRepository, expenseRepo *MockExpenseRepository, loanRepo *MockLoanRepository) {
// 				incomeRepo.On("GetByUserID", mock.Anything, mock.Anything).Return([]domain.Income{}, nil)
// 				expenseRepo.On("GetByUserID", mock.Anything, mock.Anything).Return([]domain.Expense{}, errors.New("expense repo error"))
// 			},
// 			expectedError: "failed to get expenses: expense repo error",
// 		},
// 		{
// 			name: "loan_repo_error",
// 			setupMocks: func(incomeRepo *MockIncomeRepository, expenseRepo *MockExpenseRepository, loanRepo *MockLoanRepository) {
// 				incomeRepo.On("GetByUserID", mock.Anything, mock.Anything).Return([]domain.Income{}, nil)
// 				expenseRepo.On("GetByUserID", mock.Anything, mock.Anything).Return([]domain.Expense{}, nil)
// 				loanRepo.On("GetByUserID", mock.Anything, mock.Anything).Return([]domain.Loan{}, errors.New("loan repo error"))
// 			},
// 			expectedError: "failed to get loans: loan repo error",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Arrange
// 			ctx := context.Background()
// 			userID := "user-123"

// 			mockIncomeRepo := &MockIncomeRepository{}
// 			mockExpenseRepo := &MockExpenseRepository{}
// 			mockLoanRepo := &MockLoanRepository{}

// 			tt.setupMocks(mockIncomeRepo, mockExpenseRepo, mockLoanRepo)

// 			service := NewFinanceService(mockIncomeRepo, mockExpenseRepo, mockLoanRepo)

// 			// Act
// 			_, err := service.CalculateDisposableIncome(ctx, userID)

// 			// Assert
// 			assert.Error(t, err)
// 			assert.Contains(t, err.Error(), tt.expectedError)
// 		})
// 	}
// }

// func TestFinanceService_CalculateDebtToIncomeRatio_ValidScenarios_ReturnsCorrectRatio(t *testing.T) {
// 	tests := []struct {
// 		name                string
// 		monthlyIncome       float64
// 		totalLoanPayments   float64
// 		expectedRatio       float64
// 	}{
// 		{"healthy_ratio", 5000.00, 1500.00, 0.30},     // 30%
// 		{"high_ratio", 4000.00, 2200.00, 0.55},        // 55%
// 		{"low_ratio", 8000.00, 1200.00, 0.15},         // 15%
// 		{"perfect_ratio", 6000.00, 2160.00, 0.36},     // Exactly 36%
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Arrange
// 			ctx := context.Background()
// 			userID := "user-123"

// 			incomes := []domain.Income{
// 				{Amount: tt.monthlyIncome, Frequency: "monthly", IsActive: true},
// 			}
// 			loans := []domain.Loan{
// 				{MonthlyPayment: tt.totalLoanPayments},
// 			}

// 			mockIncomeRepo := &MockIncomeRepository{}
// 			mockLoanRepo := &MockLoanRepository{}

// 			mockIncomeRepo.On("GetByUserID", ctx, userID).Return(incomes, nil)
// 			mockLoanRepo.On("GetByUserID", ctx, userID).Return(loans, nil)

// 			service := NewFinanceService(mockIncomeRepo, nil, mockLoanRepo)

// 			// Act
// 			ratio, err := service.CalculateDebtToIncomeRatio(ctx, userID)

// 			// Assert
// 			assert.NoError(t, err)
// 			assert.InDelta(t, tt.expectedRatio, ratio, 0.01)
// 			mockIncomeRepo.AssertExpectations(t)
// 			mockLoanRepo.AssertExpectations(t)
// 		})
// 	}
// }

// func TestFinanceService_CalculateDebtToIncomeRatio_ZeroIncome_ReturnsError(t *testing.T) {
// 	// Arrange
// 	ctx := context.Background()
// 	userID := "user-123"

// 	incomes := []domain.Income{} // No income
// 	loans := []domain.Loan{
// 		{MonthlyPayment: 1000.00},
// 	}

// 	mockIncomeRepo := &MockIncomeRepository{}
// 	mockLoanRepo := &MockLoanRepository{}

// 	mockIncomeRepo.On("GetByUserID", ctx, userID).Return(incomes, nil)
// 	mockLoanRepo.On("GetByUserID", ctx, userID).Return(loans, nil)

// 	service := NewFinanceService(mockIncomeRepo, nil, mockLoanRepo)

// 	// Act
// 	_, err := service.CalculateDebtToIncomeRatio(ctx, userID)

// 	// Assert
// 	assert.Error(t, err)
// 	assert.Contains(t, err.Error(), "cannot calculate debt-to-income ratio with zero income")
// 	mockIncomeRepo.AssertExpectations(t)
// 	mockLoanRepo.AssertExpectations(t)
// }

// func TestFinanceService_EvaluateFinancialHealth_AllScenarios_ReturnsCorrectHealth(t *testing.T) {
// 	tests := []struct {
// 		name              string
// 		monthlyIncome     float64
// 		monthlyExpenses   float64
// 		monthlyLoanPayments float64
// 		expectedHealth    string
// 	}{
// 		{
// 			name:              "excellent_finances",
// 			monthlyIncome:     10000.00,
// 			monthlyExpenses:   5000.00,
// 			monthlyLoanPayments: 2000.00, // DTI: 20%, Savings: 30%
// 			expectedHealth:    "Excellent",
// 		},
// 		{
// 			name:              "good_finances",
// 			monthlyIncome:     6000.00,
// 			monthlyExpenses:   4000.00,
// 			monthlyLoanPayments: 1800.00, // DTI: 30%, Savings: 3.33%
// 			expectedHealth:    "Good",
// 		},
// 		{
// 			name:              "fair_finances_high_debt",
// 			monthlyIncome:     5000.00,
// 			monthlyExpenses:   2500.00,
// 			monthlyLoanPayments: 2000.00, // DTI: 40%
// 			expectedHealth:    "Fair",
// 		},
// 		{
// 			name:              "fair_finances_low_savings",
// 			monthlyIncome:     4000.00,
// 			monthlyExpenses:   3600.00,
// 			monthlyLoanPayments: 200.00, // DTI: 5%, but very low savings
// 			expectedHealth:    "Fair",
// 		},
// 		{
// 			name:              "poor_finances_high_debt",
// 			monthlyIncome:     4000.00,
// 			monthlyExpenses:   2000.00,
// 			monthlyLoanPayments: 2200.00, // DTI: 55%
// 			expectedHealth:    "Poor",
// 		},
// 		{
// 			name:              "poor_finances_overspending",
// 			monthlyIncome:     3000.00,
// 			monthlyExpenses:   3500.00,
// 			monthlyLoanPayments: 800.00, // Negative disposable income
// 			expectedHealth:    "Poor",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Arrange
// 			ctx := context.Background()
// 			userID := "user-123"

// 			incomes := []domain.Income{
// 				{Amount: tt.monthlyIncome, Frequency: "monthly", IsActive: true},
// 			}
// 			expenses := []domain.Expense{
// 				{Amount: tt.monthlyExpenses, Frequency: "monthly"},
// 			}
// 			loans := []domain.Loan{
// 				{MonthlyPayment: tt.monthlyLoanPayments},
// 			}

// 			mockIncomeRepo := &MockIncomeRepository{}
// 			mockExpenseRepo := &MockExpenseRepository{}
// 			mockLoanRepo := &MockLoanRepository{}

// 			mockIncomeRepo.On("GetByUserID", ctx, userID).Return(incomes, nil)
// 			mockExpenseRepo.On("GetByUserID", ctx, userID).Return(expenses, nil)
// 			mockLoanRepo.On("GetByUserID", ctx, userID).Return(loans, nil)

// 			service := NewFinanceService(mockIncomeRepo, mockExpenseRepo, mockLoanRepo)

// 			// Act
// 			health, err := service.EvaluateFinancialHealth(ctx, userID)

// 			// Assert
// 			assert.NoError(t, err)
// 			assert.Equal(t, tt.expectedHealth, health)
// 			mockIncomeRepo.AssertExpectations(t)
// 			mockExpenseRepo.AssertExpectations(t)
// 			mockLoanRepo.AssertExpectations(t)
// 		})
// 	}
// }

// func TestFinanceService_GetMaxAffordableAmount_ReturnsCorrectAmount(t *testing.T) {
// 	tests := []struct {
// 		name                  string
// 		monthlyIncome         float64
// 		monthlyExpenses       float64
// 		monthlyLoanPayments   float64
// 		expectedAffordability float64
// 	}{
// 		{
// 			name:                  "healthy_finances",
// 			monthlyIncome:         6000.00,
// 			monthlyExpenses:       3000.00,
// 			monthlyLoanPayments:   1500.00, // DTI: 25%, Disposable: 1500
// 			expectedAffordability: 4500.00, // 3x disposable income for healthy DTI
// 		},
// 		{
// 			name:                  "moderate_debt",
// 			monthlyIncome:         5000.00,
// 			monthlyExpenses:       2500.00,
// 			monthlyLoanPayments:   2000.00, // DTI: 40%, Disposable: 500
// 			expectedAffordability: 1000.00, // 2x disposable income for moderate DTI
// 		},
// 		{
// 			name:                  "high_debt",
// 			monthlyIncome:         4000.00,
// 			monthlyExpenses:       1500.00,
// 			monthlyLoanPayments:   2500.00, // DTI: 62.5%, Disposable: 0
// 			expectedAffordability: 0.00, // 0.5x of 0 disposable income
// 		},
// 		{
// 			name:                  "overspending",
// 			monthlyIncome:         3000.00,
// 			monthlyExpenses:       3500.00,
// 			monthlyLoanPayments:   1000.00, // Negative disposable income
// 			expectedAffordability: 0.00, // No affordability when overspending
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Arrange
// 			ctx := context.Background()
// 			userID := "user-123"

// 			incomes := []domain.Income{
// 				{Amount: tt.monthlyIncome, Frequency: "monthly", IsActive: true},
// 			}
// 			expenses := []domain.Expense{
// 				{Amount: tt.monthlyExpenses, Frequency: "monthly"},
// 			}
// 			loans := []domain.Loan{
// 				{MonthlyPayment: tt.monthlyLoanPayments},
// 			}

// 			mockIncomeRepo := &MockIncomeRepository{}
// 			mockExpenseRepo := &MockExpenseRepository{}
// 			mockLoanRepo := &MockLoanRepository{}

// 			mockIncomeRepo.On("GetByUserID", ctx, userID).Return(incomes, nil)
// 			mockExpenseRepo.On("GetByUserID", ctx, userID).Return(expenses, nil)
// 			mockLoanRepo.On("GetByUserID", ctx, userID).Return(loans, nil)

// 			service := NewFinanceService(mockIncomeRepo, mockExpenseRepo, mockLoanRepo)

// 			// Act
// 			maxAffordable, err := service.GetMaxAffordableAmount(ctx, userID)

// 			// Assert
// 			assert.NoError(t, err)
// 			assert.Equal(t, tt.expectedAffordability, maxAffordable)
// 			mockIncomeRepo.AssertExpectations(t)
// 			mockExpenseRepo.AssertExpectations(t)
// 			mockLoanRepo.AssertExpectations(t)
// 		})
// 	}
// }

// func TestFinanceService_NormalizeToMonthly_FrequencyConversions_ReturnsCorrectAmounts(t *testing.T) {
// 	tests := []struct {
// 		name            string
// 		amount          float64
// 		frequency       string
// 		expectedMonthly float64
// 	}{
// 		{"monthly_same", 1000.00, "monthly", 1000.00},
// 		{"weekly_conversion", 250.00, "weekly", 1083.33}, // 250 * 52 / 12
// 		{"daily_conversion", 50.00, "daily", 1521.88},    // 50 * 365.25 / 12
// 		{"one_time_zero", 1000.00, "one-time", 0.00},     // One-time doesn't count
// 		{"invalid_frequency", 500.00, "invalid", 0.00},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Arrange
// 			service := &financeService{} // Assuming this is the concrete type

// 			// Act
// 			result := service.NormalizeToMonthly(tt.amount, tt.frequency)

// 			// Assert
// 			assert.InDelta(t, tt.expectedMonthly, result, 0.01)
// 		})
// 	}
// }

// func TestFinanceService_CalculateFinanceSummary_CompleteScenario_ReturnsCorrectSummary(t *testing.T) {
// 	// Arrange
// 	ctx := context.Background()
// 	userID := "user-123"

// 	incomes := []domain.Income{
// 		{Amount: 5000.00, Frequency: "monthly", IsActive: true},
// 		{Amount: 500.00, Frequency: "weekly", IsActive: true}, // 2166.67 monthly
// 	}
// 	expenses := []domain.Expense{
// 		{Amount: 1200.00, Frequency: "monthly", Category: "housing"},
// 		{Amount: 100.00, Frequency: "weekly", Category: "food"}, // 433.33 monthly
// 		{Amount: 25.00, Frequency: "daily", Category: "transport"}, // 760.94 monthly
// 	}
// 	loans := []domain.Loan{
// 		{MonthlyPayment: 800.00, Type: "mortgage"},
// 		{MonthlyPayment: 350.00, Type: "auto"},
// 	}

// 	mockIncomeRepo := &MockIncomeRepository{}
// 	mockExpenseRepo := &MockExpenseRepository{}
// 	mockLoanRepo := &MockLoanRepository{}

// 	mockIncomeRepo.On("GetByUserID", ctx, userID).Return(incomes, nil)
// 	mockExpenseRepo.On("GetByUserID", ctx, userID).Return(expenses, nil)
// 	mockLoanRepo.On("GetByUserID", ctx, userID).Return(loans, nil)

// 	service := NewFinanceService(mockIncomeRepo, mockExpenseRepo, mockLoanRepo)

// 	// Act
// 	summary, err := service.CalculateFinanceSummary(ctx, userID)

// 	// Assert
// 	assert.NoError(t, err)
// 	assert.NotNil(t, summary)
// 	assert.Equal(t, userID, summary.UserID)

// 	// Total income: 5000 + 2166.67 = 7166.67
// 	assert.InDelta(t, 7166.67, summary.MonthlyIncome, 0.01)

// 	// Total expenses: 1200 + 433.33 + 760.94 = 2394.27
// 	assert.InDelta(t, 2394.27, summary.MonthlyExpenses, 0.01)

// 	// Total loan payments: 800 + 350 = 1150
// 	assert.Equal(t, 1150.00, summary.MonthlyLoanPayments)

// 	// Disposable income: 7166.67 - 2394.27 - 1150 = 3622.40
// 	assert.InDelta(t, 3622.40, summary.DisposableIncome, 0.01)

// 	// DTI ratio: 1150 / 7166.67 = 0.1604
// 	assert.InDelta(t, 0.16, summary.DebtToIncomeRatio, 0.01)

// 	// Savings rate: 3622.40 / 7166.67 = 0.5054
// 	assert.InDelta(t, 0.51, summary.SavingsRate, 0.01)

// 	// Should be excellent health (low DTI, high savings rate)
// 	assert.Equal(t, "Excellent", summary.FinancialHealth)

// 	// Budget remaining should equal disposable income
// 	assert.InDelta(t, 3622.40, summary.BudgetRemaining, 0.01)

// 	mockIncomeRepo.AssertExpectations(t)
// 	mockExpenseRepo.AssertExpectations(t)
// 	mockLoanRepo.AssertExpectations(t)
// }

// func TestFinanceService_CategorizeExpenses_ReturnsCorrectCategorization(t *testing.T) {
// 	// Arrange
// 	ctx := context.Background()
// 	userID := "user-123"

// 	expenses := []domain.Expense{
// 		{Amount: 1200.00, Frequency: "monthly", Category: "housing"},
// 		{Amount: 800.00, Frequency: "monthly", Category: "housing"}, // Total housing: 2000
// 		{Amount: 150.00, Frequency: "weekly", Category: "food"}, // 650 monthly
// 		{Amount: 100.00, Frequency: "weekly", Category: "food"}, // 433.33 monthly, total food: 1083.33
// 		{Amount: 50.00, Frequency: "daily", Category: "transport"}, // 1520.83 monthly
// 		{Amount: 200.00, Frequency: "monthly", Category: "entertainment"},
// 		{Amount: 100.00, Frequency: "monthly", Category: "utilities"},
// 	}

// 	mockExpenseRepo := &MockExpenseRepository{}
// 	mockExpenseRepo.On("GetByUserID", ctx, userID).Return(expenses, nil)

// 	service := NewFinanceService(nil, mockExpenseRepo, nil)

// 	// Act
// 	categoryMap, err := service.CategorizeExpenses(ctx, userID)

// 	// Assert
// 	assert.NoError(t, err)
// 	assert.NotNil(t, categoryMap)

// 	assert.Equal(t, 2000.00, categoryMap["housing"])
// 	assert.InDelta(t, 1083.33, categoryMap["food"], 0.01)
// 	assert.InDelta(t, 1520.83, categoryMap["transport"], 0.01)
// 	assert.Equal(t, 200.00, categoryMap["entertainment"])
// 	assert.Equal(t, 100.00, categoryMap["utilities"])
// 	assert.Equal(t, 0.0, categoryMap["other"]) // Should exist with zero value

// 	mockExpenseRepo.AssertExpectations(t)
// }
