package handlers

import (
	"context"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// FinanceService interface is defined in handlers package following consumer-defined principle
// This interface is consumed by FinanceHandler in this package
type FinanceServiceInterface interface {
	// Income operations
	AddIncome(ctx context.Context, income domain.Income) error
	UpdateIncome(ctx context.Context, income domain.Income) error
	DeleteIncome(ctx context.Context, userID, incomeID string) error
	GetUserIncomes(ctx context.Context, userID string) ([]domain.Income, error)
	GetActiveUserIncomes(ctx context.Context, userID string) ([]domain.Income, error)
	GetIncomeByID(ctx context.Context, incomeID string) (domain.Income, error)

	// Expense operations
	AddExpense(ctx context.Context, expense domain.Expense) error
	UpdateExpense(ctx context.Context, expense domain.Expense) error
	DeleteExpense(ctx context.Context, userID, expenseID string) error
	GetUserExpenses(ctx context.Context, userID string) ([]domain.Expense, error)
	GetUserExpensesByCategory(ctx context.Context, userID, category string) ([]domain.Expense, error)
	GetExpenseByID(ctx context.Context, expenseID string) (domain.Expense, error)

	// Loan operations
	AddLoan(ctx context.Context, loan domain.Loan) error
	UpdateLoan(ctx context.Context, loan domain.Loan) error
	GetUserLoans(ctx context.Context, userID string) ([]domain.Loan, error)
	UpdateLoanBalance(ctx context.Context, userID, loanID string, newBalance float64) error
	GetLoanByID(ctx context.Context, loanID string) (domain.Loan, error)

	// Financial analysis
	CalculateFinanceSummary(ctx context.Context, userID string) (domain.FinanceSummary, error)
	CalculateDisposableIncome(ctx context.Context, userID string) (float64, error)
	CalculateDebtToIncomeRatio(ctx context.Context, userID string) (float64, error)
	EvaluateFinancialHealth(ctx context.Context, userID string) (string, error)
	GetMaxAffordableAmount(ctx context.Context, userID string) (float64, error)

	// Helper functions
	NormalizeToMonthly(amount float64, frequency string) (float64, error)

	// Additional UI helper methods
	GetFinanceSummary(userID string) (domain.FinanceSummary, error)
	GetIncomes(userID string) ([]domain.Income, error)
	GetExpenseByIDUI(userID string, expenseID uint) (domain.Expense, error)
}
