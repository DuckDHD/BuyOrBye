package services

import (
	"context"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// Analysis service interfaces defined in services package
// These are consumed by other services in this package

// BudgetAnalyzer defines budget analysis operations
type BudgetAnalyzer interface {
	// AnalyzeBudget identifies overspending categories and budget issues
	AnalyzeBudget(ctx context.Context, userID string) (BudgetAnalysis, error)

	// GetSpendingInsights provides category-wise spending analysis
	GetSpendingInsights(ctx context.Context, userID string) (SpendingInsights, error)

	// RecommendSavings applies the 50/30/20 rule and provides savings recommendations
	RecommendSavings(ctx context.Context, userID string) (SavingsRecommendation, error)

	// IdentifyUnnecessaryExpenses finds expenses that can be optimized
	IdentifyUnnecessaryExpenses(ctx context.Context, userID string) ([]ExpenseOptimization, error)
}

// DebtCalculator defines debt analysis and calculation operations
type DebtCalculator interface {
	// CalculateTotalDebt sums all loan balances for a user
	CalculateTotalDebt(ctx context.Context, userID string) (float64, error)

	// ProjectDebtFreeDate estimates when the user will be debt-free based on current payments
	ProjectDebtFreeDate(ctx context.Context, userID string) (time.Time, error)

	// SuggestPaymentStrategy recommends avalanche vs snowball approach
	SuggestPaymentStrategy(ctx context.Context, userID string, extraPayment float64) (PaymentStrategy, error)

	// CalculateInterestSavings calculates savings from making extra payments
	CalculateInterestSavings(ctx context.Context, userID string, extraPayment float64) (InterestSavings, error)

	// GetDebtAnalysis provides comprehensive debt analysis
	GetDebtAnalysis(ctx context.Context, userID string) (DebtAnalysis, error)

	// CalculateMinimumPayment calculates the minimum required payment for a loan
	CalculateMinimumPayment(principalAmount, interestRate float64, termMonths int) float64
}

// FinanceService interface is consumed by analyzer services in this package
// Following the consumer-defined principle - this belongs here because analyzers consume it
type FinanceService interface {
	// Income operations
	GetUserIncomes(ctx context.Context, userID string) ([]domain.Income, error)
	GetActiveUserIncomes(ctx context.Context, userID string) ([]domain.Income, error)

	// Expense operations  
	GetUserExpenses(ctx context.Context, userID string) ([]domain.Expense, error)
	GetUserExpensesByCategory(ctx context.Context, userID, category string) ([]domain.Expense, error)

	// Loan operations
	GetUserLoans(ctx context.Context, userID string) ([]domain.Loan, error)

	// Financial analysis
	CalculateFinanceSummary(ctx context.Context, userID string) (domain.FinanceSummary, error)
	CalculateDisposableIncome(ctx context.Context, userID string) (float64, error)
	CalculateDebtToIncomeRatio(ctx context.Context, userID string) (float64, error)
	EvaluateFinancialHealth(ctx context.Context, userID string) (string, error)
	GetMaxAffordableAmount(ctx context.Context, userID string) (float64, error)

	// Helper functions
	NormalizeToMonthly(amount float64, frequency string) (float64, error)
}