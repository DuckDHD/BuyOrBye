package services

import (
	"context"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// Repository interfaces are defined in the services package following the consumer-defined principle
// These interfaces are consumed by services in this package, so they belong here

// UserRepository defines the interface for user persistence operations
// This interface is consumed by AuthService
type UserRepository interface {
	// Create saves a new user to the database
	Create(ctx context.Context, user *domain.User) error

	// GetByEmail retrieves a user by their email address
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// GetByID retrieves a user by their ID
	GetByID(ctx context.Context, userID string) (*domain.User, error)

	// Update modifies an existing user's data
	Update(ctx context.Context, user *domain.User) error

	// UpdateLastLogin updates the last login timestamp for a user
	UpdateLastLogin(ctx context.Context, userID string, loginTime time.Time) error
}

// TokenRepository defines the interface for refresh token persistence operations
// This interface is consumed by AuthService
type TokenRepository interface {
	// SaveRefreshToken stores a refresh token for a user
	SaveRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error

	// GetRefreshToken retrieves a refresh token by the token string
	GetRefreshToken(ctx context.Context, token string) (userID string, err error)

	// RevokeToken marks a refresh token as revoked
	RevokeToken(ctx context.Context, token string) error

	// RevokeAllUserTokens marks all refresh tokens for a user as revoked
	RevokeAllUserTokens(ctx context.Context, userID string) error

	// CleanupExpiredTokens removes expired tokens from the database
	CleanupExpiredTokens(ctx context.Context) error
}

// IncomeRepository defines the interface for income data persistence
// This interface is consumed by FinanceService
type IncomeRepository interface {
	// Basic CRUD operations
	SaveIncome(ctx context.Context, income domain.Income) error
	GetIncomeByID(ctx context.Context, id string) (domain.Income, error)
	UpdateIncome(ctx context.Context, income domain.Income) error
	DeleteIncome(ctx context.Context, id string) error

	// User-scoped queries
	GetUserIncomes(ctx context.Context, userID string) ([]domain.Income, error)
	GetActiveIncomes(ctx context.Context, userID string) ([]domain.Income, error)
	GetUserIncomesByFrequency(ctx context.Context, userID string, frequency string) ([]domain.Income, error)

	// Aggregation queries
	CalculateUserTotalIncome(ctx context.Context, userID string, activeOnly bool) (float64, error)
}

// ExpenseRepository defines the interface for expense data persistence
// This interface is consumed by FinanceService
type ExpenseRepository interface {
	// Basic CRUD operations
	SaveExpense(ctx context.Context, expense domain.Expense) error
	GetExpenseByID(ctx context.Context, id string) (domain.Expense, error)
	UpdateExpense(ctx context.Context, expense domain.Expense) error
	DeleteExpense(ctx context.Context, id string) error

	// User-scoped queries
	GetUserExpenses(ctx context.Context, userID string) ([]domain.Expense, error)
	GetExpensesByCategory(ctx context.Context, userID string, category string) ([]domain.Expense, error)
	GetExpensesByFrequency(ctx context.Context, userID string, frequency string) ([]domain.Expense, error)
	GetExpensesByPriority(ctx context.Context, userID string, priority int) ([]domain.Expense, error)

	// Filtered queries
	GetFixedExpenses(ctx context.Context, userID string) ([]domain.Expense, error)
	GetVariableExpenses(ctx context.Context, userID string) ([]domain.Expense, error)

	// Aggregation queries
	CalculateUserTotalExpenses(ctx context.Context, userID string) (float64, error)
	CalculateTotalByCategory(ctx context.Context, userID string, category string) (float64, error)
}

// LoanRepository defines the interface for loan data persistence
// This interface is consumed by FinanceService
type LoanRepository interface {
	// Basic CRUD operations
	SaveLoan(ctx context.Context, loan domain.Loan) error
	GetLoanByID(ctx context.Context, id string) (domain.Loan, error)
	UpdateLoan(ctx context.Context, loan domain.Loan) error
	DeleteLoan(ctx context.Context, id string) error

	// User-scoped queries
	GetUserLoans(ctx context.Context, userID string) ([]domain.Loan, error)
	GetLoansByType(ctx context.Context, userID string, loanType string) ([]domain.Loan, error)
	GetLoansByInterestRateRange(ctx context.Context, userID string, minRate, maxRate float64) ([]domain.Loan, error)

	// Balance management
	UpdateLoanBalance(ctx context.Context, loanID string, newBalance float64) error
	GetNearPayoffLoans(ctx context.Context, userID string, threshold float64) ([]domain.Loan, error)

	// Aggregation queries
	CalculateUserTotalDebt(ctx context.Context, userID string) (float64, error)
	CalculateUserMonthlyPayments(ctx context.Context, userID string) (float64, error)
}

// FinanceSummaryRepository defines the interface for finance summary data persistence
// This interface is consumed by FinanceService
type FinanceSummaryRepository interface {
	// Basic CRUD operations
	SaveFinanceSummary(ctx context.Context, summary domain.FinanceSummary) error
	GetFinanceSummaryByUserID(ctx context.Context, userID string) (domain.FinanceSummary, error)
	UpdateFinanceSummary(ctx context.Context, summary domain.FinanceSummary) error
	DeleteFinanceSummary(ctx context.Context, userID string) error

	// Aggregation and analysis
	GetFinanceSummariesByHealthStatus(ctx context.Context, healthStatus string) ([]domain.FinanceSummary, error)
	GetUsersWithHighDebtRatio(ctx context.Context, threshold float64) ([]domain.FinanceSummary, error)
	GetUsersWithLowSavingsRate(ctx context.Context, threshold float64) ([]domain.FinanceSummary, error)
}

// FinanceRepositories aggregates all finance-related repositories
// Used by FinanceService for dependency injection
type FinanceRepositories struct {
	Income         IncomeRepository
	Expense        ExpenseRepository
	Loan           LoanRepository
	FinanceSummary FinanceSummaryRepository
}

// NewFinanceRepositories creates a new FinanceRepositories instance
func NewFinanceRepositories(
	income IncomeRepository,
	expense ExpenseRepository,
	loan LoanRepository,
	financeSummary FinanceSummaryRepository,
) *FinanceRepositories {
	return &FinanceRepositories{
		Income:         income,
		Expense:        expense,
		Loan:           loan,
		FinanceSummary: financeSummary,
	}
}