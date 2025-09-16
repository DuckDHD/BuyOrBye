package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/middleware"
	"github.com/DuckDHD/BuyOrBye/internal/services"
	"github.com/DuckDHD/BuyOrBye/internal/types"
)

// FinanceService interface from services package - handlers act as DTO-domain adapters

// FinanceHandler handles HTTP requests for finance endpoints
type FinanceHandler struct {
	financeService services.FinanceService
	validator      *validator.Validate
}

// NewFinanceHandler creates a new finance handler with dependency injection
func NewFinanceHandler(financeService services.FinanceService) *FinanceHandler {
	return &FinanceHandler{
		financeService: financeService,
		validator:      validator.New(),
	}
}

// ==================== INCOME ENDPOINTS ====================

// AddIncome handles POST /api/finance/income requests
// Adds a new income source for the authenticated user
func (h *FinanceHandler) AddIncome(c *gin.Context) {
	var request types.AddIncomeDTO

	// Parse and bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid JSON format",
		))
		return
	}

	// Validate request fields
	if err := h.validator.Struct(&request); err != nil {
		validationErrors := h.buildValidationErrors(err)
		c.JSON(http.StatusBadRequest, types.NewValidationErrorResponse(
			"Validation failed",
			validationErrors,
		))
		return
	}

	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Transform DTO to domain object
	income := request.ToDomain(userID)

	// Call service layer with domain object
	if err := h.financeService.AddIncome(c.Request.Context(), income); err != nil {
		h.handleFinanceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Income added successfully",
	})
}

// GetIncomes handles GET /api/finance/income requests
// Retrieves all income records for the authenticated user
func (h *FinanceHandler) GetIncomes(c *gin.Context) {
	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Call service layer
	incomes, err := h.financeService.GetUserIncomes(c.Request.Context(), userID)
	if err != nil {
		h.handleFinanceError(c, err)
		return
	}

	// Convert domain objects to DTOs
	response := &types.IncomeListResponseDTO{
		Incomes: types.FromDomainIncomeList(incomes),
		Total:   len(incomes),
	}

	c.JSON(http.StatusOK, response)
}

// UpdateIncome handles PUT /api/finance/income/:id requests
// Updates an existing income record for the authenticated user
func (h *FinanceHandler) UpdateIncome(c *gin.Context) {
	var request types.UpdateIncomeDTO
	incomeID := c.Param("id")

	// Parse and bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid JSON format",
		))
		return
	}

	// Validate request fields
	if err := h.validator.Struct(&request); err != nil {
		validationErrors := h.buildValidationErrors(err)
		c.JSON(http.StatusBadRequest, types.NewValidationErrorResponse(
			"Validation failed",
			validationErrors,
		))
		return
	}

	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// First get the existing income to update
	existingIncome, err := h.financeService.GetIncomeByID(c.Request.Context(), incomeID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, types.NewErrorResponse(
				http.StatusNotFound,
				"not_found",
				"Income record not found",
			))
			return
		}
		h.handleFinanceError(c, err)
		return
	}

	// Verify ownership
	if existingIncome.UserID != userID {
		c.JSON(http.StatusForbidden, types.NewErrorResponse(
			http.StatusForbidden,
			"forbidden",
			"Access denied: You can only update your own income records",
		))
		return
	}

	// Apply updates to the existing income
	request.ApplyUpdates(&existingIncome)

	// Call service layer with updated domain object
	if err := h.financeService.UpdateIncome(c.Request.Context(), existingIncome); err != nil {
		if strings.Contains(err.Error(), "does not belong to user") {
			c.JSON(http.StatusForbidden, types.NewErrorResponse(
				http.StatusForbidden,
				"forbidden",
				"Access denied: You can only update your own income records",
			))
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, types.NewErrorResponse(
				http.StatusNotFound,
				"not_found",
				"Income record not found",
			))
			return
		}
		h.handleFinanceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Income updated successfully",
	})
}

// DeleteIncome handles DELETE /api/finance/income/:id requests
// Soft deletes an income record for the authenticated user
func (h *FinanceHandler) DeleteIncome(c *gin.Context) {
	incomeID := c.Param("id")

	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Call service layer
	if err := h.financeService.DeleteIncome(c.Request.Context(), userID, incomeID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, types.NewErrorResponse(
				http.StatusNotFound,
				"not_found",
				"Income record not found",
			))
			return
		}
		h.handleFinanceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Income deleted successfully",
	})
}

// ==================== EXPENSE ENDPOINTS ====================

// AddExpense handles POST /api/finance/expense requests
// Adds a new expense for the authenticated user
func (h *FinanceHandler) AddExpense(c *gin.Context) {
	var request types.AddExpenseDTO

	// Extract user ID from authentication context first for auth check
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Parse and bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid JSON format",
		))
		return
	}

	// Validate request fields
	if err := h.validator.Struct(&request); err != nil {
		validationErrors := h.buildValidationErrors(err)
		c.JSON(http.StatusBadRequest, types.NewValidationErrorResponse(
			"Validation failed",
			validationErrors,
		))
		return
	}

	// Transform DTO to domain object
	expense := request.ToDomain(userID)

	// Call service layer with domain object
	if err := h.financeService.AddExpense(c.Request.Context(), expense); err != nil {
		h.handleFinanceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Expense added successfully",
	})
}

// GetExpenses handles GET /api/finance/expenses requests
// Retrieves expenses for the authenticated user, optionally filtered by category
func (h *FinanceHandler) GetExpenses(c *gin.Context) {
	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Check for category filter
	category := c.Query("category")

	var expenses []domain.Expense
	var err error

	// Call appropriate service method based on filter
	if category != "" {
		expenses, err = h.financeService.GetUserExpensesByCategory(c.Request.Context(), userID, category)
	} else {
		expenses, err = h.financeService.GetUserExpenses(c.Request.Context(), userID)
	}

	if err != nil {
		h.handleFinanceError(c, err)
		return
	}

	// Convert domain objects to DTOs
	response := &types.ExpenseListResponseDTO{
		Expenses: types.FromDomainExpenseList(expenses),
		Total:    len(expenses),
	}

	c.JSON(http.StatusOK, response)
}

// UpdateExpense handles PUT /api/finance/expense/:id requests
// Updates an existing expense record for the authenticated user
func (h *FinanceHandler) UpdateExpense(c *gin.Context) {
	var request types.UpdateExpenseDTO
	expenseID := c.Param("id")

	// Parse and bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid JSON format",
		))
		return
	}

	// Validate request fields
	if err := h.validator.Struct(&request); err != nil {
		validationErrors := h.buildValidationErrors(err)
		c.JSON(http.StatusBadRequest, types.NewValidationErrorResponse(
			"Validation failed",
			validationErrors,
		))
		return
	}

	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// First get the existing expense to update
	existingExpense, err := h.financeService.GetExpenseByID(c.Request.Context(), expenseID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, types.NewErrorResponse(
				http.StatusNotFound,
				"not_found",
				"Expense record not found",
			))
			return
		}
		h.handleFinanceError(c, err)
		return
	}

	// Verify ownership
	if existingExpense.UserID != userID {
		c.JSON(http.StatusForbidden, types.NewErrorResponse(
			http.StatusForbidden,
			"forbidden",
			"Access denied: You can only update your own expense records",
		))
		return
	}

	// Apply updates to the existing expense
	request.ApplyUpdates(&existingExpense)

	// Call service layer with updated domain object
	if err := h.financeService.UpdateExpense(c.Request.Context(), existingExpense); err != nil {
		if strings.Contains(err.Error(), "does not belong to user") {
			c.JSON(http.StatusForbidden, types.NewErrorResponse(
				http.StatusForbidden,
				"forbidden",
				"Access denied: You can only update your own expense records",
			))
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, types.NewErrorResponse(
				http.StatusNotFound,
				"not_found",
				"Expense not found or access denied",
			))
			return
		}
		h.handleFinanceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Expense updated successfully",
	})
}

// DeleteExpense handles DELETE /api/finance/expense/:id requests
// Soft deletes an expense record for the authenticated user
func (h *FinanceHandler) DeleteExpense(c *gin.Context) {
	expenseID := c.Param("id")

	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Call service layer
	if err := h.financeService.DeleteExpense(c.Request.Context(), userID, expenseID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, types.NewErrorResponse(
				http.StatusNotFound,
				"not_found",
				"Expense record not found",
			))
			return
		}
		h.handleFinanceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Expense deleted successfully",
	})
}

// ==================== LOAN ENDPOINTS ====================

// AddLoan handles POST /api/finance/loan requests
// Adds a new loan for the authenticated user
func (h *FinanceHandler) AddLoan(c *gin.Context) {
	var request types.AddLoanDTO

	// Parse and bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid JSON format",
		))
		return
	}

	// Validate request fields
	if err := h.validator.Struct(&request); err != nil {
		validationErrors := h.buildValidationErrors(err)
		c.JSON(http.StatusBadRequest, types.NewValidationErrorResponse(
			"Validation failed",
			validationErrors,
		))
		return
	}

	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Transform DTO to domain object
	loan := request.ToDomain(userID)

	// Call service layer with domain object
	if err := h.financeService.AddLoan(c.Request.Context(), loan); err != nil {
		h.handleFinanceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Loan added successfully",
	})
}

// GetLoans handles GET /api/finance/loans requests
// Retrieves all loan records for the authenticated user
func (h *FinanceHandler) GetLoans(c *gin.Context) {
	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Call service layer
	loans, err := h.financeService.GetUserLoans(c.Request.Context(), userID)
	if err != nil {
		h.handleFinanceError(c, err)
		return
	}

	// Convert domain objects to DTOs
	response := &types.LoanListResponseDTO{
		Loans: types.FromDomainLoanList(loans),
		Total: len(loans),
	}

	c.JSON(http.StatusOK, response)
}

// UpdateLoan handles PUT /api/finance/loan/:id requests
// Updates an existing loan record for the authenticated user
func (h *FinanceHandler) UpdateLoan(c *gin.Context) {
	var request types.UpdateLoanDTO
	loanID := c.Param("id")

	// Parse and bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid JSON format",
		))
		return
	}

	// Validate request fields
	if err := h.validator.Struct(&request); err != nil {
		validationErrors := h.buildValidationErrors(err)
		c.JSON(http.StatusBadRequest, types.NewValidationErrorResponse(
			"Validation failed",
			validationErrors,
		))
		return
	}

	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// First get the existing loan to update
	existingLoan, err := h.financeService.GetLoanByID(c.Request.Context(), loanID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, types.NewErrorResponse(
				http.StatusNotFound,
				"not_found",
				"Loan record not found",
			))
			return
		}
		h.handleFinanceError(c, err)
		return
	}

	// Verify ownership
	if existingLoan.UserID != userID {
		c.JSON(http.StatusForbidden, types.NewErrorResponse(
			http.StatusForbidden,
			"forbidden",
			"Access denied: You can only update your own loan records",
		))
		return
	}

	// Apply updates to the existing loan
	request.ApplyUpdates(&existingLoan)

	// Call service layer with updated domain object
	if err := h.financeService.UpdateLoan(c.Request.Context(), existingLoan); err != nil {
		if strings.Contains(err.Error(), "does not belong to user") {
			c.JSON(http.StatusForbidden, types.NewErrorResponse(
				http.StatusForbidden,
				"forbidden",
				"Access denied: You can only update your own loan records",
			))
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, types.NewErrorResponse(
				http.StatusNotFound,
				"not_found",
				"Loan not found or access denied",
			))
			return
		}
		h.handleFinanceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Loan updated successfully",
	})
}

// ==================== FINANCIAL ANALYSIS ENDPOINTS ====================

// GetFinanceSummary handles GET /api/finance/summary requests
// Returns comprehensive financial overview for the authenticated user
func (h *FinanceHandler) GetFinanceSummary(c *gin.Context) {
	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Call service layer
	summary, err := h.financeService.CalculateFinanceSummary(c.Request.Context(), userID)
	if err != nil {
		h.handleFinanceError(c, err)
		return
	}

	// Convert domain object to DTO
	response := types.FromDomainFinanceSummary(summary)

	c.JSON(http.StatusOK, response)
}

// GetAffordability handles GET /api/finance/affordability requests
// Returns maximum affordable amount for purchases based on user's financial situation
func (h *FinanceHandler) GetAffordability(c *gin.Context) {
	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Call service layer
	maxAffordable, err := h.financeService.GetMaxAffordableAmount(c.Request.Context(), userID)
	if err != nil {
		h.handleFinanceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":               userID,
		"max_affordable_amount": maxAffordable,
		"currency":              "USD",
		"calculation_date":      "now", // Could be actual timestamp
	})
}

// ==================== HELPER METHODS ====================

// buildValidationErrors constructs a map of validation errors from validator.ValidationErrors
func (h *FinanceHandler) buildValidationErrors(err error) map[string]interface{} {
	validationErrors := make(map[string]interface{})
	for _, err := range err.(validator.ValidationErrors) {
		field := err.Field()
		switch err.Tag() {
		case "required":
			validationErrors[field] = field + " is required"
		case "email":
			validationErrors[field] = field + " must be a valid email address"
		case "min":
			validationErrors[field] = field + " must be at least " + err.Param() + " characters"
		case "gt":
			validationErrors[field] = field + " must be greater than " + err.Param()
		case "gte":
			validationErrors[field] = field + " must be greater than or equal to " + err.Param()
		case "lte":
			validationErrors[field] = field + " must be less than or equal to " + err.Param()
		case "oneof":
			validationErrors[field] = field + " must be one of: " + err.Param()
		default:
			validationErrors[field] = field + " is invalid"
		}
	}
	return validationErrors
}

// handleFinanceError handles finance-specific errors and maps them to appropriate HTTP responses
func (h *FinanceHandler) handleFinanceError(c *gin.Context, err error) {
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "income not found"):
		c.JSON(http.StatusNotFound, types.NewErrorResponse(
			http.StatusNotFound,
			"not_found",
			"Income record not found",
		))
	case strings.Contains(errMsg, "expense not found"):
		c.JSON(http.StatusNotFound, types.NewErrorResponse(
			http.StatusNotFound,
			"not_found",
			"Expense record not found",
		))
	case strings.Contains(errMsg, "loan not found"):
		c.JSON(http.StatusNotFound, types.NewErrorResponse(
			http.StatusNotFound,
			"not_found",
			"Loan record not found",
		))
	case strings.Contains(errMsg, "financial summary not found") || strings.Contains(errMsg, "finance summary not found"):
		c.JSON(http.StatusNotFound, types.NewErrorResponse(
			http.StatusNotFound,
			"not_found",
			"Financial summary not found",
		))
	case strings.Contains(errMsg, "unauthorized access") || strings.Contains(errMsg, "access denied"):
		c.JSON(http.StatusForbidden, types.NewErrorResponse(
			http.StatusForbidden,
			"forbidden",
			"Access denied: You can only access your own financial records",
		))
	case strings.Contains(errMsg, "invalid finance data") || strings.Contains(errMsg, "invalid financial data"):
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid financial data provided",
		))
	default:
		// Internal server error for unexpected errors
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			http.StatusInternalServerError,
			"internal_error",
			"An internal error occurred. Please try again later",
		))
	}
}
