package models

import (
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"gorm.io/gorm"
)

// FinanceSummaryModel represents the finance_summaries table structure in the database
type FinanceSummaryModel struct {
	UserID              string         `gorm:"primaryKey;type:varchar(36)" json:"user_id"`
	MonthlyIncome       float64        `gorm:"not null;type:decimal(10,2)" json:"monthly_income"`
	MonthlyExpenses     float64        `gorm:"not null;type:decimal(10,2)" json:"monthly_expenses"`
	MonthlyLoanPayments float64        `gorm:"not null;type:decimal(10,2)" json:"monthly_loan_payments"`
	DisposableIncome    float64        `gorm:"not null;type:decimal(10,2)" json:"disposable_income"`
	DebtToIncomeRatio   float64        `gorm:"not null;type:decimal(5,4)" json:"debt_to_income_ratio"`
	SavingsRate         float64        `gorm:"not null;type:decimal(5,4)" json:"savings_rate"`
	FinancialHealth     string         `gorm:"not null;type:varchar(20)" json:"financial_health"`
	BudgetRemaining     float64        `gorm:"not null;type:decimal(10,2)" json:"budget_remaining"`
	UpdatedAt           time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Foreign key relationship
	User UserModel `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// TableName returns the table name for GORM
func (FinanceSummaryModel) TableName() string {
	return "finance_summaries"
}

// BeforeCreate sets timestamps
func (f *FinanceSummaryModel) BeforeCreate(tx *gorm.DB) error {
	if f.UpdatedAt.IsZero() {
		f.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate updates the UpdatedAt timestamp
func (f *FinanceSummaryModel) BeforeUpdate(tx *gorm.DB) error {
	f.UpdatedAt = time.Now()
	return nil
}

// ToDomain converts FinanceSummaryModel to domain.FinanceSummary
func (f FinanceSummaryModel) ToDomain() domain.FinanceSummary {
	return domain.FinanceSummary{
		UserID:              f.UserID,
		MonthlyIncome:       f.MonthlyIncome,
		MonthlyExpenses:     f.MonthlyExpenses,
		MonthlyLoanPayments: f.MonthlyLoanPayments,
		DisposableIncome:    f.DisposableIncome,
		DebtToIncomeRatio:   f.DebtToIncomeRatio,
		SavingsRate:         f.SavingsRate,
		FinancialHealth:     f.FinancialHealth,
		BudgetRemaining:     f.BudgetRemaining,
		UpdatedAt:           f.UpdatedAt,
	}
}

// FromDomain creates FinanceSummaryModel from domain.FinanceSummary
func (f *FinanceSummaryModel) FromDomain(summary domain.FinanceSummary) {
	f.UserID = summary.UserID
	f.MonthlyIncome = summary.MonthlyIncome
	f.MonthlyExpenses = summary.MonthlyExpenses
	f.MonthlyLoanPayments = summary.MonthlyLoanPayments
	f.DisposableIncome = summary.DisposableIncome
	f.DebtToIncomeRatio = summary.DebtToIncomeRatio
	f.SavingsRate = summary.SavingsRate
	f.FinancialHealth = summary.FinancialHealth
	f.BudgetRemaining = summary.BudgetRemaining
	f.UpdatedAt = summary.UpdatedAt
}

// NewFinanceSummaryModelFromDomain creates a new FinanceSummaryModel from domain.FinanceSummary
func NewFinanceSummaryModelFromDomain(summary domain.FinanceSummary) *FinanceSummaryModel {
	model := &FinanceSummaryModel{}
	model.FromDomain(summary)
	return model
}