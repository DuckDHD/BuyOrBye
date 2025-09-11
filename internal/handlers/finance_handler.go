package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/dtos"
	"github.com/DuckDHD/BuyOrBye/internal/middleware"
)

// FinanceService interface is consumed by this handler and defined in this package
// Following the consumer-defined interface principle from CLAUDE.md

// FinanceHandler handles HTTP requests for finance endpoints
type FinanceHandler struct {
	financeService FinanceService
	validator      *validator.Validate
}

// NewFinanceHandler creates a new finance handler with dependency injection
func NewFinanceHandler(financeService FinanceService) *FinanceHandler {
	return &FinanceHandler{
		financeService: financeService,
		validator:      validator.New(),
	}
}

// ==================== INCOME ENDPOINTS ====================

// AddIncome handles POST /api/finance/income requests
// Adds a new income source for the authenticated user
func (h *FinanceHandler) AddIncome(c *gin.Context) {
	var request dtos.AddIncomeDTO

	// Parse and bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dtos.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid JSON format",
		))
		return
	}

	// Validate request fields
	if err := h.validator.Struct(&request); err != nil {
		validationErrors := h.buildValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.NewValidationErrorResponse(
			"Validation failed",
			validationErrors,
		))
		return
	}

	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dtos.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Convert DTO to domain struct
	income := request.ToDomain(userID)

	// Call service layer
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
		c.JSON(http.StatusUnauthorized, dtos.NewErrorResponse(
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

	// Convert domain structs to DTOs
	var response []dtos.IncomeResponseDTO
	for _, income := range incomes {
		var dto dtos.IncomeResponseDTO
		dto.FromDomain(income)
		response = append(response, dto)
	}

	c.JSON(http.StatusOK, response)
}

// UpdateIncome handles PUT /api/finance/income/:id requests
// Updates an existing income record for the authenticated user
func (h *FinanceHandler) UpdateIncome(c *gin.Context) {
	var request dtos.UpdateIncomeDTO
	incomeID := c.Param("id")

	// Parse and bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dtos.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid JSON format",
		))
		return
	}

	// Validate request fields
	if err := h.validator.Struct(&request); err != nil {
		validationErrors := h.buildValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.NewValidationErrorResponse(
			"Validation failed",
			validationErrors,
		))
		return
	}

	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dtos.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Get existing income first to apply partial updates
	existingIncomes, err := h.financeService.GetUserIncomes(c.Request.Context(), userID)
	if err != nil {
		h.handleFinanceError(c, err)
		return
	}

	// Find the income to update
	var income *domain.Income
	for _, inc := range existingIncomes {
		if inc.ID == incomeID && inc.UserID == userID {
			income = &inc
			break
		}
	}

	if income == nil {
		c.JSON(http.StatusNotFound, dtos.NewErrorResponse(
			http.StatusNotFound,
			"not_found",
			"Income not found or access denied",
		))
		return
	}

	// Apply updates
	request.ApplyUpdates(income)

	// Call service layer
	if err := h.financeService.UpdateIncome(c.Request.Context(), *income); err != nil {
		if strings.Contains(err.Error(), "does not belong to user") {
			c.JSON(http.StatusForbidden, dtos.NewErrorResponse(
				http.StatusForbidden,
				"forbidden",
				"Access denied: You can only update your own income records",
			))
			return
		}
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dtos.NewErrorResponse(
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
		c.JSON(http.StatusUnauthorized, dtos.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Call service layer
	if err := h.financeService.DeleteIncome(c.Request.Context(), userID, incomeID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dtos.NewErrorResponse(
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
	var request dtos.AddExpenseDTO

	// Extract user ID from authentication context first for auth check
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dtos.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Parse and bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dtos.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid JSON format",
		))
		return
	}

	// Validate request fields
	if err := h.validator.Struct(&request); err != nil {
		validationErrors := h.buildValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.NewValidationErrorResponse(
			"Validation failed",
			validationErrors,
		))
		return
	}

	// Convert DTO to domain struct
	expense := request.ToDomain(userID)

	// Call service layer
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
		c.JSON(http.StatusUnauthorized, dtos.NewErrorResponse(
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

	// Convert domain structs to DTOs
	var response []dtos.ExpenseResponseDTO
	for _, expense := range expenses {
		var dto dtos.ExpenseResponseDTO
		dto.FromDomain(expense)
		response = append(response, dto)
	}

	c.JSON(http.StatusOK, response)
}

// UpdateExpense handles PUT /api/finance/expense/:id requests
// Updates an existing expense record for the authenticated user
func (h *FinanceHandler) UpdateExpense(c *gin.Context) {
	var request dtos.UpdateExpenseDTO
	expenseID := c.Param("id")

	// Parse and bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dtos.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid JSON format",
		))
		return
	}

	// Validate request fields
	if err := h.validator.Struct(&request); err != nil {
		validationErrors := h.buildValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.NewValidationErrorResponse(
			"Validation failed",
			validationErrors,
		))
		return
	}

	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dtos.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Get existing expenses to find the one to update
	expenses, err := h.financeService.GetUserExpenses(c.Request.Context(), userID)
	if err != nil {
		h.handleFinanceError(c, err)
		return
	}

	// Find the expense to update
	var expense *domain.Expense
	for _, exp := range expenses {
		if exp.ID == expenseID && exp.UserID == userID {
			expense = &exp
			break
		}
	}

	if expense == nil {
		c.JSON(http.StatusNotFound, dtos.NewErrorResponse(
			http.StatusNotFound,
			"not_found",
			"Expense not found or access denied",
		))
		return
	}

	// Apply updates
	request.ApplyUpdates(expense)

	// Call service layer
	if err := h.financeService.UpdateExpense(c.Request.Context(), *expense); err != nil {
		if strings.Contains(err.Error(), "does not belong to user") {
			c.JSON(http.StatusForbidden, dtos.NewErrorResponse(
				http.StatusForbidden,
				"forbidden",
				"Access denied: You can only update your own expense records",
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
		c.JSON(http.StatusUnauthorized, dtos.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Call service layer
	if err := h.financeService.DeleteExpense(c.Request.Context(), userID, expenseID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, dtos.NewErrorResponse(
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
	var request dtos.AddLoanDTO

	// Parse and bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dtos.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid JSON format",
		))
		return
	}

	// Validate request fields
	if err := h.validator.Struct(&request); err != nil {
		validationErrors := h.buildValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.NewValidationErrorResponse(
			"Validation failed",
			validationErrors,
		))
		return
	}

	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dtos.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Convert DTO to domain struct
	loan := request.ToDomain(userID)

	// Call service layer
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
		c.JSON(http.StatusUnauthorized, dtos.NewErrorResponse(
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

	// Convert domain structs to DTOs
	var response []dtos.LoanResponseDTO
	for _, loan := range loans {
		var dto dtos.LoanResponseDTO
		dto.FromDomain(loan)
		response = append(response, dto)
	}

	c.JSON(http.StatusOK, response)
}

// UpdateLoan handles PUT /api/finance/loan/:id requests
// Updates an existing loan record for the authenticated user
func (h *FinanceHandler) UpdateLoan(c *gin.Context) {
	var request dtos.UpdateLoanDTO
	loanID := c.Param("id")

	// Parse and bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, dtos.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid JSON format",
		))
		return
	}

	// Validate request fields
	if err := h.validator.Struct(&request); err != nil {
		validationErrors := h.buildValidationErrors(err)
		c.JSON(http.StatusBadRequest, dtos.NewValidationErrorResponse(
			"Validation failed",
			validationErrors,
		))
		return
	}

	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dtos.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Authentication required",
		))
		return
	}

	// Get existing loans to find the one to update
	loans, err := h.financeService.GetUserLoans(c.Request.Context(), userID)
	if err != nil {
		h.handleFinanceError(c, err)
		return
	}

	// Find the loan to update
	var loan *domain.Loan
	for _, l := range loans {
		if l.ID == loanID && l.UserID == userID {
			loan = &l
			break
		}
	}

	if loan == nil {
		c.JSON(http.StatusNotFound, dtos.NewErrorResponse(
			http.StatusNotFound,
			"not_found",
			"Loan not found or access denied",
		))
		return
	}

	// Apply updates
	request.ApplyUpdates(loan)

	// Call service layer
	if err := h.financeService.UpdateLoan(c.Request.Context(), *loan); err != nil {
		if strings.Contains(err.Error(), "does not belong to user") {
			c.JSON(http.StatusForbidden, dtos.NewErrorResponse(
				http.StatusForbidden,
				"forbidden",
				"Access denied: You can only update your own loan records",
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
		c.JSON(http.StatusUnauthorized, dtos.NewErrorResponse(
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

	// Convert domain struct to DTO
	var response dtos.FinanceSummaryResponseDTO
	response.FromDomain(summary)

	c.JSON(http.StatusOK, response)
}

// GetAffordability handles GET /api/finance/affordability requests
// Returns maximum affordable amount for purchases based on user's financial situation
func (h *FinanceHandler) GetAffordability(c *gin.Context) {
	// Extract user ID from authentication context
	userID := middleware.GetUserID(c)
	if userID == "" {
		c.JSON(http.StatusUnauthorized, dtos.NewErrorResponse(
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
	switch {
	case errors.Is(err, domain.ErrIncomeNotFound):
		c.JSON(http.StatusNotFound, dtos.NewErrorResponse(
			http.StatusNotFound,
			"not_found",
			"Income record not found",
		))
	case errors.Is(err, domain.ErrExpenseNotFound):
		c.JSON(http.StatusNotFound, dtos.NewErrorResponse(
			http.StatusNotFound,
			"not_found",
			"Expense record not found",
		))
	case errors.Is(err, domain.ErrLoanNotFound):
		c.JSON(http.StatusNotFound, dtos.NewErrorResponse(
			http.StatusNotFound,
			"not_found",
			"Loan record not found",
		))
	case errors.Is(err, domain.ErrFinanceSummaryNotFound):
		c.JSON(http.StatusNotFound, dtos.NewErrorResponse(
			http.StatusNotFound,
			"not_found",
			"Financial summary not found",
		))
	case errors.Is(err, domain.ErrUnauthorizedAccess):
		c.JSON(http.StatusForbidden, dtos.NewErrorResponse(
			http.StatusForbidden,
			"forbidden",
			"Access denied: You can only access your own financial records",
		))
	case errors.Is(err, domain.ErrInvalidFinanceData):
		c.JSON(http.StatusBadRequest, dtos.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid financial data provided",
		))
	default:
		// Internal server error for unexpected errors
		c.JSON(http.StatusInternalServerError, dtos.NewErrorResponse(
			http.StatusInternalServerError,
			"internal_error",
			"An internal error occurred. Please try again later",
		))
	}
}