package router

import (
	"github.com/gin-gonic/gin"
	
	"github.com/DuckDHD/BuyOrBye/internal/handlers"
	"github.com/DuckDHD/BuyOrBye/internal/middleware"
	"github.com/DuckDHD/BuyOrBye/internal/services"
)

// FinanceRouter handles all finance-related routes
type FinanceRouter struct {
	financeHandler *handlers.FinanceHandler
	jwtService     services.JWTService
}

// NewFinanceRouter creates a new finance router
func NewFinanceRouter(financeHandler *handlers.FinanceHandler, jwtService services.JWTService) *FinanceRouter {
	return &FinanceRouter{
		financeHandler: financeHandler,
		jwtService:     jwtService,
	}
}

// RegisterRoutes registers all finance routes to the given router group
// All finance routes require authentication
func (fr *FinanceRouter) RegisterRoutes(rg *gin.RouterGroup) {
	// All finance routes require authentication
	financeGroup := rg.Group("/finance")
	financeGroup.Use(middleware.JWTAuth(fr.jwtService))
	{
		// ==================== INCOME ENDPOINTS ====================
		
		// POST /api/finance/income - Add new income source
		financeGroup.POST("/income", fr.financeHandler.AddIncome)
		
		// GET /api/finance/income - Get user's income sources
		financeGroup.GET("/income", fr.financeHandler.GetIncomes)
		
		// PUT /api/finance/income/:id - Update specific income (owner only)
		financeGroup.PUT("/income/:id", fr.financeHandler.UpdateIncome)
		
		// DELETE /api/finance/income/:id - Delete specific income (soft delete, owner only)
		financeGroup.DELETE("/income/:id", fr.financeHandler.DeleteIncome)

		// ==================== EXPENSE ENDPOINTS ====================
		
		// POST /api/finance/expense - Add new expense
		financeGroup.POST("/expense", fr.financeHandler.AddExpense)
		
		// GET /api/finance/expenses - Get user's expenses (with optional category filter)
		financeGroup.GET("/expenses", fr.financeHandler.GetExpenses)
		
		// PUT /api/finance/expense/:id - Update specific expense (owner only)
		financeGroup.PUT("/expense/:id", fr.financeHandler.UpdateExpense)
		
		// DELETE /api/finance/expense/:id - Delete specific expense (owner only)
		financeGroup.DELETE("/expense/:id", fr.financeHandler.DeleteExpense)

		// ==================== LOAN ENDPOINTS ====================
		
		// POST /api/finance/loan - Add new loan
		financeGroup.POST("/loan", fr.financeHandler.AddLoan)
		
		// GET /api/finance/loans - Get user's loans
		financeGroup.GET("/loans", fr.financeHandler.GetLoans)
		
		// PUT /api/finance/loan/:id - Update specific loan (owner only)
		financeGroup.PUT("/loan/:id", fr.financeHandler.UpdateLoan)
		
		// ==================== FINANCIAL ANALYSIS ENDPOINTS ====================
		
		// GET /api/finance/summary - Get complete financial summary
		financeGroup.GET("/summary", fr.financeHandler.GetFinanceSummary)
		
		// GET /api/finance/affordability - Get max affordable purchase amount
		financeGroup.GET("/affordability", fr.financeHandler.GetAffordability)
	}
}