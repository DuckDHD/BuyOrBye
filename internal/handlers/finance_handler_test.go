package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/services"
	"github.com/DuckDHD/BuyOrBye/internal/types"
)

// MockFinanceService is a mock implementation of FinanceService for testing
type MockFinanceService struct {
	mock.Mock
}

// Income operations
func (m *MockFinanceService) AddIncome(ctx context.Context, income domain.Income) error {
	args := m.Called(ctx, income)
	return args.Error(0)
}

func (m *MockFinanceService) UpdateIncome(ctx context.Context, income domain.Income) error {
	args := m.Called(ctx, income)
	return args.Error(0)
}

func (m *MockFinanceService) DeleteIncome(ctx context.Context, userID, incomeID string) error {
	args := m.Called(ctx, userID, incomeID)
	return args.Error(0)
}

func (m *MockFinanceService) GetUserIncomes(ctx context.Context, userID string) ([]domain.Income, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Income), args.Error(1)
}

func (m *MockFinanceService) GetActiveUserIncomes(ctx context.Context, userID string) ([]domain.Income, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Income), args.Error(1)
}

// Expense operations
func (m *MockFinanceService) AddExpense(ctx context.Context, expense domain.Expense) error {
	args := m.Called(ctx, expense)
	return args.Error(0)
}

func (m *MockFinanceService) UpdateExpense(ctx context.Context, expense domain.Expense) error {
	args := m.Called(ctx, expense)
	return args.Error(0)
}

func (m *MockFinanceService) DeleteExpense(ctx context.Context, userID, expenseID string) error {
	args := m.Called(ctx, userID, expenseID)
	return args.Error(0)
}

func (m *MockFinanceService) GetUserExpenses(ctx context.Context, userID string) ([]domain.Expense, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Expense), args.Error(1)
}

func (m *MockFinanceService) GetUserExpensesByCategory(ctx context.Context, userID, category string) ([]domain.Expense, error) {
	args := m.Called(ctx, userID, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Expense), args.Error(1)
}

// Loan operations
func (m *MockFinanceService) AddLoan(ctx context.Context, loan domain.Loan) error {
	args := m.Called(ctx, loan)
	return args.Error(0)
}

func (m *MockFinanceService) UpdateLoan(ctx context.Context, loan domain.Loan) error {
	args := m.Called(ctx, loan)
	return args.Error(0)
}

func (m *MockFinanceService) GetUserLoans(ctx context.Context, userID string) ([]domain.Loan, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Loan), args.Error(1)
}

func (m *MockFinanceService) UpdateLoanBalance(ctx context.Context, userID, loanID string, newBalance float64) error {
	args := m.Called(ctx, userID, loanID, newBalance)
	return args.Error(0)
}

// Financial analysis
func (m *MockFinanceService) CalculateFinanceSummary(ctx context.Context, userID string) (domain.FinanceSummary, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return domain.FinanceSummary{}, args.Error(1)
	}
	return args.Get(0).(domain.FinanceSummary), args.Error(1)
}

func (m *MockFinanceService) CalculateDisposableIncome(ctx context.Context, userID string) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockFinanceService) CalculateDebtToIncomeRatio(ctx context.Context, userID string) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockFinanceService) EvaluateFinancialHealth(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockFinanceService) GetMaxAffordableAmount(ctx context.Context, userID string) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}

// Helper functions
func (m *MockFinanceService) NormalizeToMonthly(amount float64, frequency string) (float64, error) {
	args := m.Called(amount, frequency)
	return args.Get(0).(float64), args.Error(1)
}

func setupFinanceTestRouter(financeService services.FinanceService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Middleware to set authenticated user
	r.Use(func(c *gin.Context) {
		c.Set("userID", "test-user-123")
		c.Set("userEmail", "test@example.com")
		c.Next()
	})

	handler := NewFinanceHandler(financeService)

	// Set up finance routes
	finance := r.Group("/api/finance")
	{
		// Income routes
		finance.POST("/income", handler.AddIncome)
		finance.GET("/income", handler.GetIncomes)
		finance.PUT("/income/:id", handler.UpdateIncome)
		finance.DELETE("/income/:id", handler.DeleteIncome)

		// Expense routes
		finance.POST("/expense", handler.AddExpense)
		finance.GET("/expenses", handler.GetExpenses)
		finance.PUT("/expense/:id", handler.UpdateExpense)
		finance.DELETE("/expense/:id", handler.DeleteExpense)

		// Loan routes
		finance.POST("/loan", handler.AddLoan)
		finance.GET("/loans", handler.GetLoans)
		finance.PUT("/loan/:id", handler.UpdateLoan)

		// Financial analysis routes
		finance.GET("/summary", handler.GetFinanceSummary)
		finance.GET("/affordability", handler.GetAffordability)
	}

	return r
}

func setupUnauthenticatedTestRouter(financeService services.FinanceService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewFinanceHandler(financeService)

	// Set up finance routes without authentication middleware
	finance := r.Group("/api/finance")
	{
		finance.POST("/expense", handler.AddExpense)
	}

	return r
}

func createTestIncome() domain.Income {
	return domain.Income{
		ID:        "income-123",
		UserID:    "test-user-123",
		Source:    "Software Engineer Salary",
		Amount:    5000.00,
		Frequency: "monthly",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestExpense() domain.Expense {
	return domain.Expense{
		ID:        "expense-123",
		UserID:    "test-user-123",
		Category:  "housing",
		Name:      "Monthly Rent",
		Amount:    1200.00,
		Frequency: "monthly",
		IsFixed:   true,
		Priority:  1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestLoan() domain.Loan {
	return domain.Loan{
		ID:               "loan-123",
		UserID:           "test-user-123",
		Lender:           "Chase Bank",
		Type:             "mortgage",
		PrincipalAmount:  250000.00,
		RemainingBalance: 245000.00,
		MonthlyPayment:   1266.71,
		InterestRate:     4.5,
		EndDate:          time.Now().AddDate(30, 0, 0),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func createTestFinanceSummary() domain.FinanceSummary {
	return domain.FinanceSummary{
		UserID:              "test-user-123",
		MonthlyIncome:       5000.00,
		MonthlyExpenses:     3200.00,
		MonthlyLoanPayments: 1266.71,
		DisposableIncome:    533.29,
		DebtToIncomeRatio:   0.253,
		SavingsRate:         0.107,
		FinancialHealth:     "Good",
		BudgetRemaining:     533.29,
		UpdatedAt:          time.Now(),
	}
}

// ==================== INCOME TESTS ====================

func TestFinanceHandler_AddIncome_Success(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	addIncomeRequest := types.AddIncomeDTO{
		Source:    "Software Engineer Salary",
		Amount:    5000.00,
		Frequency: "monthly",
	}

	mockFinanceService.On("AddIncome", mock.Anything, mock.MatchedBy(func(income domain.Income) bool {
		return income.UserID == "test-user-123" &&
			income.Source == addIncomeRequest.Source &&
			income.Amount == addIncomeRequest.Amount &&
			income.Frequency == addIncomeRequest.Frequency &&
			income.IsActive == true
	})).Return(nil)

	requestBody, _ := json.Marshal(addIncomeRequest)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/finance/income", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Income added successfully", response["message"])

	mockFinanceService.AssertExpectations(t)
}

func TestFinanceHandler_AddIncome_ValidationError(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	addIncomeRequest := types.AddIncomeDTO{
		Source:    "", // Invalid - required field
		Amount:    -100.00, // Invalid - must be positive
		Frequency: "invalid", // Invalid - not in allowed values
	}

	requestBody, _ := json.Marshal(addIncomeRequest)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/finance/income", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response types.ValidationErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "validation_error", response.Error)
	assert.Contains(t, response.Fields, "Source")
	assert.Contains(t, response.Fields, "Amount")
	assert.Contains(t, response.Fields, "Frequency")

	// Service should not be called
	mockFinanceService.AssertNotCalled(t, "AddIncome")
}

func TestFinanceHandler_GetIncomes_Success(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	expectedIncomes := []domain.Income{createTestIncome()}

	mockFinanceService.On("GetUserIncomes", mock.Anything, "test-user-123").Return(expectedIncomes, nil)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/finance/income", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response []types.IncomeResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response, 1)
	assert.Equal(t, expectedIncomes[0].ID, response[0].ID)
	assert.Equal(t, expectedIncomes[0].Source, response[0].Source)
	assert.Equal(t, expectedIncomes[0].Amount, response[0].Amount)

	mockFinanceService.AssertExpectations(t)
}

func TestFinanceHandler_UpdateIncome_OnlyOwner(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	updateIncomeRequest := types.UpdateIncomeDTO{
		Source: stringPtr("Senior Software Engineer"),
		Amount: floatPtr(5500.00),
	}

	// Mock GetUserIncomes to return empty slice so income won't be found
	mockFinanceService.On("GetUserIncomes", mock.Anything, "test-user-123").
		Return([]domain.Income{}, nil)

	requestBody, _ := json.Marshal(updateIncomeRequest)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/finance/income/income-456", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response types.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Equal(t, "not_found", response.Error)

	mockFinanceService.AssertExpectations(t)
}

func TestFinanceHandler_UpdateIncome_ForbiddenAccess(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	updateIncomeRequest := types.UpdateIncomeDTO{
		Source: stringPtr("Senior Software Engineer"),
		Amount: floatPtr(5500.00),
	}

	// Mock GetUserIncomes to return income for different user
	existingIncome := createTestIncome()
	existingIncome.ID = "income-456" 
	existingIncome.UserID = "different-user-123" // Different user
	
	mockFinanceService.On("GetUserIncomes", mock.Anything, "test-user-123").
		Return([]domain.Income{existingIncome}, nil)

	requestBody, _ := json.Marshal(updateIncomeRequest)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/finance/income/income-456", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code) // Should return not found for security

	var response types.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Equal(t, "not_found", response.Error)

	mockFinanceService.AssertExpectations(t)
}

func TestFinanceHandler_DeleteIncome_SoftDelete(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	mockFinanceService.On("DeleteIncome", mock.Anything, "test-user-123", "income-123").Return(nil)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/finance/income/income-123", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Income deleted successfully", response["message"])

	mockFinanceService.AssertExpectations(t)
}

// ==================== EXPENSE TESTS ====================

func TestFinanceHandler_AddExpense_RequiresAuth(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupUnauthenticatedTestRouter(mockFinanceService) // No auth middleware

	addExpenseRequest := types.AddExpenseDTO{
		Category:  "housing",
		Name:      "Monthly Rent",
		Amount:    1200.00,
		Frequency: "monthly",
		IsFixed:   true,
		Priority:  1,
	}

	requestBody, _ := json.Marshal(addExpenseRequest)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/finance/expense", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response types.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Equal(t, "unauthorized", response.Error)

	// Service should not be called
	mockFinanceService.AssertNotCalled(t, "AddExpense")
}

func TestFinanceHandler_AddExpense_Success(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	addExpenseRequest := types.AddExpenseDTO{
		Category:  "housing",
		Name:      "Monthly Rent",
		Amount:    1200.00,
		Frequency: "monthly",
		IsFixed:   true,
		Priority:  1,
	}

	mockFinanceService.On("AddExpense", mock.Anything, mock.MatchedBy(func(expense domain.Expense) bool {
		return expense.UserID == "test-user-123" &&
			expense.Category == addExpenseRequest.Category &&
			expense.Name == addExpenseRequest.Name &&
			expense.Amount == addExpenseRequest.Amount &&
			expense.Priority == addExpenseRequest.Priority
	})).Return(nil)

	requestBody, _ := json.Marshal(addExpenseRequest)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/finance/expense", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Expense added successfully", response["message"])

	mockFinanceService.AssertExpectations(t)
}

func TestFinanceHandler_GetExpenses_Success(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	expectedExpenses := []domain.Expense{createTestExpense()}

	mockFinanceService.On("GetUserExpenses", mock.Anything, "test-user-123").Return(expectedExpenses, nil)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/finance/expenses", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response []types.ExpenseResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response, 1)
	assert.Equal(t, expectedExpenses[0].ID, response[0].ID)
	assert.Equal(t, expectedExpenses[0].Category, response[0].Category)
	assert.Equal(t, expectedExpenses[0].Name, response[0].Name)

	mockFinanceService.AssertExpectations(t)
}

func TestFinanceHandler_GetExpenses_WithCategoryFilter(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	expectedExpenses := []domain.Expense{createTestExpense()}

	mockFinanceService.On("GetUserExpensesByCategory", mock.Anything, "test-user-123", "housing").
		Return(expectedExpenses, nil)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/finance/expenses?category=housing", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response []types.ExpenseResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response, 1)
	assert.Equal(t, "housing", response[0].Category)

	mockFinanceService.AssertExpectations(t)
}

// ==================== LOAN TESTS ====================

func TestFinanceHandler_AddLoan_Success(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	addLoanRequest := types.AddLoanDTO{
		Lender:           "Chase Bank",
		Type:             "mortgage",
		PrincipalAmount:  250000.00,
		RemainingBalance: 245000.00,
		MonthlyPayment:   1266.71,
		InterestRate:     4.5,
		EndDate:          time.Now().AddDate(30, 0, 0),
	}

	mockFinanceService.On("AddLoan", mock.Anything, mock.MatchedBy(func(loan domain.Loan) bool {
		return loan.UserID == "test-user-123" &&
			loan.Lender == addLoanRequest.Lender &&
			loan.Type == addLoanRequest.Type &&
			loan.PrincipalAmount == addLoanRequest.PrincipalAmount
	})).Return(nil)

	requestBody, _ := json.Marshal(addLoanRequest)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/finance/loan", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Loan added successfully", response["message"])

	mockFinanceService.AssertExpectations(t)
}

func TestFinanceHandler_AddLoan_ValidationError(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	addLoanRequest := types.AddLoanDTO{
		Lender:           "", // Required field
		Type:             "invalid-type", // Must be from allowed values
		PrincipalAmount:  -1000.00, // Must be positive
		RemainingBalance: -500.00, // Must be non-negative
		MonthlyPayment:   0.00, // Must be positive
		InterestRate:     150.00, // Must be <= 100
		EndDate:          time.Time{}, // Required field
	}

	requestBody, _ := json.Marshal(addLoanRequest)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/finance/loan", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response types.ValidationErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "validation_error", response.Error)

	// Service should not be called
	mockFinanceService.AssertNotCalled(t, "AddLoan")
}

// ==================== SUMMARY & AFFORDABILITY TESTS ====================

func TestFinanceHandler_GetFinanceSummary_Success(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	expectedSummary := createTestFinanceSummary()

	mockFinanceService.On("CalculateFinanceSummary", mock.Anything, "test-user-123").
		Return(expectedSummary, nil)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/finance/summary", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response types.FinanceSummaryResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedSummary.UserID, response.UserID)
	assert.Equal(t, expectedSummary.MonthlyIncome, response.MonthlyIncome)
	assert.Equal(t, expectedSummary.MonthlyExpenses, response.MonthlyExpenses)
	assert.Equal(t, expectedSummary.FinancialHealth, response.FinancialHealth)

	mockFinanceService.AssertExpectations(t)
}

func TestFinanceHandler_GetAffordability_Success(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	expectedAffordability := 1500.00

	mockFinanceService.On("GetMaxAffordableAmount", mock.Anything, "test-user-123").
		Return(expectedAffordability, nil)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/finance/affordability", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test-user-123", response["user_id"])
	assert.Equal(t, expectedAffordability, response["max_affordable_amount"])

	mockFinanceService.AssertExpectations(t)
}

func TestFinanceHandler_GetAffordability_ServiceError(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	mockFinanceService.On("GetMaxAffordableAmount", mock.Anything, "test-user-123").
		Return(0.0, fmt.Errorf("failed to calculate affordability"))

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/finance/affordability", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response types.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "internal_error", response.Error)

	mockFinanceService.AssertExpectations(t)
}

// ==================== MALFORMED JSON TESTS ====================

func TestFinanceHandler_AddIncome_MalformedJSON_Returns400(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/finance/income", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response types.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "bad_request", response.Error)

	mockFinanceService.AssertNotCalled(t, "AddIncome")
}

func TestFinanceHandler_AddExpense_MalformedJSON_Returns400(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/finance/expense", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response types.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "bad_request", response.Error)

	mockFinanceService.AssertNotCalled(t, "AddExpense")
}

// ==================== NOT FOUND TESTS ====================

func TestFinanceHandler_UpdateIncome_NotFound(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	updateIncomeRequest := types.UpdateIncomeDTO{
		Source: stringPtr("Updated Source"),
		Amount: floatPtr(6000.00),
	}

	// Mock GetUserIncomes to return empty slice so income won't be found
	mockFinanceService.On("GetUserIncomes", mock.Anything, "test-user-123").
		Return([]domain.Income{}, nil)

	requestBody, _ := json.Marshal(updateIncomeRequest)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/api/finance/income/nonexistent", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response types.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Equal(t, "not_found", response.Error)

	mockFinanceService.AssertExpectations(t)
}

func TestFinanceHandler_DeleteIncome_NotFound(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	mockFinanceService.On("DeleteIncome", mock.Anything, "test-user-123", "nonexistent").
		Return(fmt.Errorf("income not found"))

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/finance/income/nonexistent", nil)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)

	var response types.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Equal(t, "not_found", response.Error)

	mockFinanceService.AssertExpectations(t)
}

// ==================== SERVICE ERROR TESTS ====================

func TestFinanceHandler_AddIncome_ServiceError_Returns500(t *testing.T) {
	// Arrange
	mockFinanceService := new(MockFinanceService)
	router := setupFinanceTestRouter(mockFinanceService)

	addIncomeRequest := types.AddIncomeDTO{
		Source:    "Test Income",
		Amount:    1000.00,
		Frequency: "monthly",
	}

	mockFinanceService.On("AddIncome", mock.Anything, mock.AnythingOfType("domain.Income")).
		Return(fmt.Errorf("database connection failed"))

	requestBody, _ := json.Marshal(addIncomeRequest)

	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/finance/income", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response types.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "internal_error", response.Error)

	mockFinanceService.AssertExpectations(t)
}

// ==================== HELPER FUNCTIONS ====================

func stringPtr(s string) *string {
	return &s
}

func floatPtr(f float64) *float64 {
	return &f
}

