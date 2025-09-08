package services

import (
	"context"
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)


// financeService implements the FinanceService interface
type financeService struct {
	repos *FinanceRepositories
}

// NewFinanceService creates a new FinanceService instance
func NewFinanceService(repos *FinanceRepositories) FinanceService {
	return &financeService{
		repos: repos,
	}
}

// AddIncome validates and adds a new income record
func (s *financeService) AddIncome(ctx context.Context, income domain.Income) error {
	if err := income.Validate(); err != nil {
		return fmt.Errorf("invalid income data: %w", err)
	}

	return s.repos.Income.SaveIncome(ctx, income)
}

// UpdateIncome validates and updates an existing income record
func (s *financeService) UpdateIncome(ctx context.Context, income domain.Income) error {
	if err := income.Validate(); err != nil {
		return fmt.Errorf("invalid income data: %w", err)
	}

	// Verify the income belongs to the user by getting it first
	existing, err := s.repos.Income.GetIncomeByID(ctx, income.ID)
	if err != nil {
		return fmt.Errorf("failed to verify income ownership: %w", err)
	}

	if existing.UserID != income.UserID {
		return fmt.Errorf("income does not belong to user")
	}

	return s.repos.Income.UpdateIncome(ctx, income)
}

// DeleteIncome removes an income record after verifying ownership
func (s *financeService) DeleteIncome(ctx context.Context, userID, incomeID string) error {
	// Verify ownership
	existing, err := s.repos.Income.GetIncomeByID(ctx, incomeID)
	if err != nil {
		return fmt.Errorf("income not found: %w", err)
	}

	if existing.UserID != userID {
		return fmt.Errorf("income does not belong to user")
	}

	return s.repos.Income.DeleteIncome(ctx, incomeID)
}

// GetUserIncomes retrieves all income records for a user
func (s *financeService) GetUserIncomes(ctx context.Context, userID string) ([]domain.Income, error) {
	return s.repos.Income.GetUserIncomes(ctx, userID)
}

// GetActiveUserIncomes retrieves only active income records for a user
func (s *financeService) GetActiveUserIncomes(ctx context.Context, userID string) ([]domain.Income, error) {
	return s.repos.Income.GetActiveIncomes(ctx, userID)
}

// AddExpense validates and adds a new expense record
func (s *financeService) AddExpense(ctx context.Context, expense domain.Expense) error {
	if err := expense.Validate(); err != nil {
		return fmt.Errorf("invalid expense data: %w", err)
	}

	return s.repos.Expense.SaveExpense(ctx, expense)
}

// UpdateExpense validates and updates an existing expense record
func (s *financeService) UpdateExpense(ctx context.Context, expense domain.Expense) error {
	if err := expense.Validate(); err != nil {
		return fmt.Errorf("invalid expense data: %w", err)
	}

	// Verify ownership
	existing, err := s.repos.Expense.GetExpenseByID(ctx, expense.ID)
	if err != nil {
		return fmt.Errorf("failed to verify expense ownership: %w", err)
	}

	if existing.UserID != expense.UserID {
		return fmt.Errorf("expense does not belong to user")
	}

	return s.repos.Expense.UpdateExpense(ctx, expense)
}

// DeleteExpense removes an expense record after verifying ownership
func (s *financeService) DeleteExpense(ctx context.Context, userID, expenseID string) error {
	// Verify ownership
	existing, err := s.repos.Expense.GetExpenseByID(ctx, expenseID)
	if err != nil {
		return fmt.Errorf("expense not found: %w", err)
	}

	if existing.UserID != userID {
		return fmt.Errorf("expense does not belong to user")
	}

	return s.repos.Expense.DeleteExpense(ctx, expenseID)
}

// GetUserExpenses retrieves all expense records for a user
func (s *financeService) GetUserExpenses(ctx context.Context, userID string) ([]domain.Expense, error) {
	return s.repos.Expense.GetUserExpenses(ctx, userID)
}

// GetUserExpensesByCategory retrieves expense records by category for a user
func (s *financeService) GetUserExpensesByCategory(ctx context.Context, userID, category string) ([]domain.Expense, error) {
	return s.repos.Expense.GetExpensesByCategory(ctx, userID, category)
}

// AddLoan validates and adds a new loan record
func (s *financeService) AddLoan(ctx context.Context, loan domain.Loan) error {
	if err := loan.Validate(); err != nil {
		return fmt.Errorf("invalid loan data: %w", err)
	}

	return s.repos.Loan.SaveLoan(ctx, loan)
}

// UpdateLoan validates and updates an existing loan record
func (s *financeService) UpdateLoan(ctx context.Context, loan domain.Loan) error {
	if err := loan.Validate(); err != nil {
		return fmt.Errorf("invalid loan data: %w", err)
	}

	// Verify ownership
	existing, err := s.repos.Loan.GetLoanByID(ctx, loan.ID)
	if err != nil {
		return fmt.Errorf("failed to verify loan ownership: %w", err)
	}

	if existing.UserID != loan.UserID {
		return fmt.Errorf("loan does not belong to user")
	}

	return s.repos.Loan.UpdateLoan(ctx, loan)
}

// GetUserLoans retrieves all loan records for a user
func (s *financeService) GetUserLoans(ctx context.Context, userID string) ([]domain.Loan, error) {
	return s.repos.Loan.GetUserLoans(ctx, userID)
}

// UpdateLoanBalance updates the remaining balance for a loan after verifying ownership
func (s *financeService) UpdateLoanBalance(ctx context.Context, userID, loanID string, newBalance float64) error {
	// Verify ownership
	existing, err := s.repos.Loan.GetLoanByID(ctx, loanID)
	if err != nil {
		return fmt.Errorf("loan not found: %w", err)
	}

	if existing.UserID != userID {
		return fmt.Errorf("loan does not belong to user")
	}

	if newBalance < 0 {
		return fmt.Errorf("loan balance cannot be negative")
	}

	return s.repos.Loan.UpdateLoanBalance(ctx, loanID, newBalance)
}

// CalculateFinanceSummary aggregates all financial data for a user
func (s *financeService) CalculateFinanceSummary(ctx context.Context, userID string) (domain.FinanceSummary, error) {
	// Get all active incomes
	incomes, err := s.repos.Income.GetActiveIncomes(ctx, userID)
	if err != nil {
		return domain.FinanceSummary{}, fmt.Errorf("failed to get user incomes: %w", err)
	}

	// Get all expenses
	expenses, err := s.repos.Expense.GetUserExpenses(ctx, userID)
	if err != nil {
		return domain.FinanceSummary{}, fmt.Errorf("failed to get user expenses: %w", err)
	}

	// Get all loans
	loans, err := s.repos.Loan.GetUserLoans(ctx, userID)
	if err != nil {
		return domain.FinanceSummary{}, fmt.Errorf("failed to get user loans: %w", err)
	}

	// Calculate monthly totals
	monthlyIncome := 0.0
	for _, income := range incomes {
		normalized, err := s.NormalizeToMonthly(income.Amount, income.Frequency)
		if err != nil {
			continue // Skip invalid frequencies
		}
		monthlyIncome += normalized
	}

	monthlyExpenses := 0.0
	for _, expense := range expenses {
		normalized, err := s.NormalizeToMonthly(expense.Amount, expense.Frequency)
		if err != nil {
			continue // Skip invalid frequencies
		}
		monthlyExpenses += normalized
	}

	monthlyLoanPayments := 0.0
	for _, loan := range loans {
		monthlyLoanPayments += loan.MonthlyPayment
	}

	// Calculate derived metrics
	disposableIncome := monthlyIncome - monthlyExpenses - monthlyLoanPayments
	budgetRemaining := disposableIncome

	var debtToIncomeRatio float64
	if monthlyIncome > 0 {
		debtToIncomeRatio = monthlyLoanPayments / monthlyIncome
	}

	var savingsRate float64
	if monthlyIncome > 0 {
		savingsRate = disposableIncome / monthlyIncome
		if savingsRate < 0 {
			savingsRate = 0 // Can't have negative savings rate for calculation
		}
	}

	summary := domain.FinanceSummary{
		UserID:              userID,
		MonthlyIncome:       monthlyIncome,
		MonthlyExpenses:     monthlyExpenses,
		MonthlyLoanPayments: monthlyLoanPayments,
		DisposableIncome:    disposableIncome,
		DebtToIncomeRatio:   debtToIncomeRatio,
		SavingsRate:         savingsRate,
		BudgetRemaining:     budgetRemaining,
		UpdatedAt:          time.Now(),
	}

	// Calculate financial health
	summary.FinancialHealth = summary.CalculateHealth()

	return summary, nil
}

// CalculateDisposableIncome calculates disposable income by normalizing all frequencies
func (s *financeService) CalculateDisposableIncome(ctx context.Context, userID string) (float64, error) {
	summary, err := s.CalculateFinanceSummary(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate finance summary: %w", err)
	}

	return summary.DisposableIncome, nil
}

// CalculateDebtToIncomeRatio calculates the debt-to-income ratio as a percentage
func (s *financeService) CalculateDebtToIncomeRatio(ctx context.Context, userID string) (float64, error) {
	summary, err := s.CalculateFinanceSummary(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate finance summary: %w", err)
	}

	return summary.DebtToIncomeRatio * 100, nil // Return as percentage
}

// EvaluateFinancialHealth applies business rules for health scoring
func (s *financeService) EvaluateFinancialHealth(ctx context.Context, userID string) (string, error) {
	summary, err := s.CalculateFinanceSummary(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to calculate finance summary: %w", err)
	}

	return summary.FinancialHealth, nil
}

// GetMaxAffordableAmount calculates the maximum affordable purchase amount
func (s *financeService) GetMaxAffordableAmount(ctx context.Context, userID string) (float64, error) {
	summary, err := s.CalculateFinanceSummary(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate finance summary: %w", err)
	}

	return summary.CalculateAffordability(), nil
}

// NormalizeToMonthly converts different frequencies to monthly amounts
func (s *financeService) NormalizeToMonthly(amount float64, frequency string) (float64, error) {
	switch frequency {
	case "daily":
		return amount * 30, nil // 30 days per month
	case "weekly":
		return amount * 4.33, nil // Average weeks per month (52/12)
	case "biweekly":
		return amount * 2.17, nil // Every two weeks (26 pay periods / 12 months)
	case "monthly":
		return amount, nil
	case "quarterly":
		return amount / 3, nil // 4 quarters / 12 months
	case "semiannual":
		return amount / 6, nil // 2 periods / 12 months
	case "annual", "yearly":
		return amount / 12, nil
	case "one-time":
		return amount / 12, nil // Spread one-time over 12 months
	default:
		return 0, fmt.Errorf("unsupported frequency: %s", frequency)
	}
}