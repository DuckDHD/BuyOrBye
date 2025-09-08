package types

import (
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// AddIncomeDTO represents a request to add a new income source
type AddIncomeDTO struct {
	Source    string  `json:"source" validate:"required,min=2" example:"Software Engineer Salary"`
	Amount    float64 `json:"amount" validate:"required,gt=0" example:"5000.00"`
	Frequency string  `json:"frequency" validate:"required,oneof=monthly weekly daily one-time" example:"monthly"`
} 

// UpdateIncomeDTO represents a request to update an existing income source
type UpdateIncomeDTO struct {
	Source    *string  `json:"source,omitempty" validate:"omitempty,min=2" example:"Senior Software Engineer"`
	Amount    *float64 `json:"amount,omitempty" validate:"omitempty,gt=0" example:"5500.00"`
	Frequency *string  `json:"frequency,omitempty" validate:"omitempty,oneof=monthly weekly daily one-time" example:"monthly"`
}

// AddExpenseDTO represents a request to add a new expense
type AddExpenseDTO struct {
	Category  string  `json:"category" validate:"required,oneof=housing food transport entertainment utilities other" example:"housing"`
	Name      string  `json:"name" validate:"required,min=2" example:"Monthly Rent"`
	Amount    float64 `json:"amount" validate:"required,gt=0" example:"1200.00"`
	Frequency string  `json:"frequency" validate:"required,oneof=monthly weekly daily" example:"monthly"`
	IsFixed   bool    `json:"is_fixed" example:"true"`
	Priority  int     `json:"priority" validate:"required,min=1,max=3" example:"1"`
}

// UpdateExpenseDTO represents a request to update an existing expense
type UpdateExpenseDTO struct {
	Category  *string  `json:"category,omitempty" validate:"omitempty,oneof=housing food transport entertainment utilities other" example:"utilities"`
	Name      *string  `json:"name,omitempty" validate:"omitempty,min=2" example:"Electricity Bill"`
	Amount    *float64 `json:"amount,omitempty" validate:"omitempty,gt=0" example:"150.00"`
	Frequency *string  `json:"frequency,omitempty" validate:"omitempty,oneof=monthly weekly daily" example:"monthly"`
	IsFixed   *bool    `json:"is_fixed,omitempty" example:"false"`
	Priority  *int     `json:"priority,omitempty" validate:"omitempty,min=1,max=3" example:"2"`
}

// AddLoanDTO represents a request to add a new loan
type AddLoanDTO struct {
	Lender           string    `json:"lender" validate:"required,min=2" example:"Chase Bank"`
	Type             string    `json:"type" validate:"required,oneof=mortgage auto personal student" example:"mortgage"`
	PrincipalAmount  float64   `json:"principal_amount" validate:"required,gt=0" example:"250000.00"`
	RemainingBalance float64   `json:"remaining_balance" validate:"required,gte=0" example:"245000.00"`
	MonthlyPayment   float64   `json:"monthly_payment" validate:"required,gt=0" example:"1266.71"`
	InterestRate     float64   `json:"interest_rate" validate:"required,gte=0,lte=100" example:"4.5"`
	EndDate          time.Time `json:"end_date" validate:"required" example:"2054-01-15T00:00:00Z"`
}

// UpdateLoanDTO represents a request to update an existing loan
type UpdateLoanDTO struct {
	Lender           *string    `json:"lender,omitempty" validate:"omitempty,min=2" example:"Wells Fargo"`
	Type             *string    `json:"type,omitempty" validate:"omitempty,oneof=mortgage auto personal student" example:"auto"`
	PrincipalAmount  *float64   `json:"principal_amount,omitempty" validate:"omitempty,gt=0" example:"240000.00"`
	RemainingBalance *float64   `json:"remaining_balance,omitempty" validate:"omitempty,gte=0" example:"235000.00"`
	MonthlyPayment   *float64   `json:"monthly_payment,omitempty" validate:"omitempty,gt=0" example:"1200.00"`
	InterestRate     *float64   `json:"interest_rate,omitempty" validate:"omitempty,gte=0,lte=100" example:"3.5"`
	EndDate          *time.Time `json:"end_date,omitempty" validate:"omitempty" example:"2050-01-15T00:00:00Z"`
}

// IncomeResponseDTO represents an income in API responses
type IncomeResponseDTO struct {
	ID        string    `json:"id" example:"income-123"`
	UserID    string    `json:"user_id" example:"user-456"`
	Source    string    `json:"source" example:"Software Engineer Salary"`
	Amount    float64   `json:"amount" example:"5000.00"`
	Frequency string    `json:"frequency" example:"monthly"`
	IsActive  bool      `json:"is_active" example:"true"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

// ExpenseResponseDTO represents an expense in API responses
type ExpenseResponseDTO struct {
	ID        string    `json:"id" example:"expense-123"`
	UserID    string    `json:"user_id" example:"user-456"`
	Category  string    `json:"category" example:"housing"`
	Name      string    `json:"name" example:"Monthly Rent"`
	Amount    float64   `json:"amount" example:"1200.00"`
	Frequency string    `json:"frequency" example:"monthly"`
	IsFixed   bool      `json:"is_fixed" example:"true"`
	Priority  int       `json:"priority" example:"1"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

// LoanResponseDTO represents a loan in API responses
type LoanResponseDTO struct {
	ID               string    `json:"id" example:"loan-123"`
	UserID           string    `json:"user_id" example:"user-456"`
	Lender           string    `json:"lender" example:"Chase Bank"`
	Type             string    `json:"type" example:"mortgage"`
	PrincipalAmount  float64   `json:"principal_amount" example:"250000.00"`
	RemainingBalance float64   `json:"remaining_balance" example:"245000.00"`
	MonthlyPayment   float64   `json:"monthly_payment" example:"1266.71"`
	InterestRate     float64   `json:"interest_rate" example:"4.5"`
	EndDate          time.Time `json:"end_date" example:"2054-01-15T00:00:00Z"`
	CreatedAt        time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt        time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

// FinanceSummaryResponseDTO represents a finance summary in API responses
type FinanceSummaryResponseDTO struct {
	UserID              string    `json:"user_id" example:"user-456"`
	MonthlyIncome       float64   `json:"monthly_income" example:"5000.00"`
	MonthlyExpenses     float64   `json:"monthly_expenses" example:"3200.00"`
	MonthlyLoanPayments float64   `json:"monthly_loan_payments" example:"1266.71"`
	DisposableIncome    float64   `json:"disposable_income" example:"533.29"`
	DebtToIncomeRatio   float64   `json:"debt_to_income_ratio" example:"0.253"`
	SavingsRate         float64   `json:"savings_rate" example:"0.107"`
	FinancialHealth     string    `json:"financial_health" example:"Good"`
	BudgetRemaining     float64   `json:"budget_remaining" example:"533.29"`
	UpdatedAt           time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
}

// ToDomain converts AddIncomeDTO to domain.Income
func (dto AddIncomeDTO) ToDomain(userID string) domain.Income {
	return domain.Income{
		UserID:    userID,
		Source:    dto.Source,
		Amount:    dto.Amount,
		Frequency: dto.Frequency,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ToDomain converts AddExpenseDTO to domain.Expense  
func (dto AddExpenseDTO) ToDomain(userID string) domain.Expense {
	return domain.Expense{
		UserID:    userID,
		Category:  dto.Category,
		Name:      dto.Name,
		Amount:    dto.Amount,
		Frequency: dto.Frequency,
		IsFixed:   dto.IsFixed,
		Priority:  dto.Priority,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ToDomain converts AddLoanDTO to domain.Loan
func (dto AddLoanDTO) ToDomain(userID string) domain.Loan {
	return domain.Loan{
		UserID:           userID,
		Lender:           dto.Lender,
		Type:             dto.Type,
		PrincipalAmount:  dto.PrincipalAmount,
		RemainingBalance: dto.RemainingBalance,
		MonthlyPayment:   dto.MonthlyPayment,
		InterestRate:     dto.InterestRate,
		EndDate:          dto.EndDate,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// FromDomain converts domain.Income to IncomeResponseDTO
func (dto *IncomeResponseDTO) FromDomain(income domain.Income) {
	dto.ID = income.ID
	dto.UserID = income.UserID
	dto.Source = income.Source
	dto.Amount = income.Amount
	dto.Frequency = income.Frequency
	dto.IsActive = income.IsActive
	dto.CreatedAt = income.CreatedAt
	dto.UpdatedAt = income.UpdatedAt
}

// FromDomain converts domain.Expense to ExpenseResponseDTO
func (dto *ExpenseResponseDTO) FromDomain(expense domain.Expense) {
	dto.ID = expense.ID
	dto.UserID = expense.UserID
	dto.Category = expense.Category
	dto.Name = expense.Name
	dto.Amount = expense.Amount
	dto.Frequency = expense.Frequency
	dto.IsFixed = expense.IsFixed
	dto.Priority = expense.Priority
	dto.CreatedAt = expense.CreatedAt
	dto.UpdatedAt = expense.UpdatedAt
}

// FromDomain converts domain.Loan to LoanResponseDTO
func (dto *LoanResponseDTO) FromDomain(loan domain.Loan) {
	dto.ID = loan.ID
	dto.UserID = loan.UserID
	dto.Lender = loan.Lender
	dto.Type = loan.Type
	dto.PrincipalAmount = loan.PrincipalAmount
	dto.RemainingBalance = loan.RemainingBalance
	dto.MonthlyPayment = loan.MonthlyPayment
	dto.InterestRate = loan.InterestRate
	dto.EndDate = loan.EndDate
	dto.CreatedAt = loan.CreatedAt
	dto.UpdatedAt = loan.UpdatedAt
}

// FromDomain converts domain.FinanceSummary to FinanceSummaryResponseDTO
func (dto *FinanceSummaryResponseDTO) FromDomain(summary domain.FinanceSummary) {
	dto.UserID = summary.UserID
	dto.MonthlyIncome = summary.MonthlyIncome
	dto.MonthlyExpenses = summary.MonthlyExpenses
	dto.MonthlyLoanPayments = summary.MonthlyLoanPayments
	dto.DisposableIncome = summary.DisposableIncome
	dto.DebtToIncomeRatio = summary.DebtToIncomeRatio
	dto.SavingsRate = summary.SavingsRate
	dto.FinancialHealth = summary.FinancialHealth
	dto.BudgetRemaining = summary.BudgetRemaining
	dto.UpdatedAt = summary.UpdatedAt
}

// ApplyUpdates applies UpdateIncomeDTO fields to domain.Income
func (dto UpdateIncomeDTO) ApplyUpdates(income *domain.Income) {
	if dto.Source != nil {
		income.Source = *dto.Source
	}
	if dto.Amount != nil {
		income.Amount = *dto.Amount
	}
	if dto.Frequency != nil {
		income.Frequency = *dto.Frequency
	}
	income.UpdatedAt = time.Now()
}

// ApplyUpdates applies UpdateExpenseDTO fields to domain.Expense
func (dto UpdateExpenseDTO) ApplyUpdates(expense *domain.Expense) {
	if dto.Category != nil {
		expense.Category = *dto.Category
	}
	if dto.Name != nil {
		expense.Name = *dto.Name
	}
	if dto.Amount != nil {
		expense.Amount = *dto.Amount
	}
	if dto.Frequency != nil {
		expense.Frequency = *dto.Frequency
	}
	if dto.IsFixed != nil {
		expense.IsFixed = *dto.IsFixed
	}
	if dto.Priority != nil {
		expense.Priority = *dto.Priority
	}
	expense.UpdatedAt = time.Now()
}

// ApplyUpdates applies UpdateLoanDTO fields to domain.Loan
func (dto UpdateLoanDTO) ApplyUpdates(loan *domain.Loan) {
	if dto.Lender != nil {
		loan.Lender = *dto.Lender
	}
	if dto.Type != nil {
		loan.Type = *dto.Type
	}
	if dto.PrincipalAmount != nil {
		loan.PrincipalAmount = *dto.PrincipalAmount
	}
	if dto.RemainingBalance != nil {
		loan.RemainingBalance = *dto.RemainingBalance
	}
	if dto.MonthlyPayment != nil {
		loan.MonthlyPayment = *dto.MonthlyPayment
	}
	if dto.InterestRate != nil {
		loan.InterestRate = *dto.InterestRate
	}
	if dto.EndDate != nil {
		loan.EndDate = *dto.EndDate
	}
	loan.UpdatedAt = time.Now()
}