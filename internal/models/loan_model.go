package models

import (
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LoanModel represents the loan table structure in the database
type LoanModel struct {
	ID               string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID           string         `gorm:"not null;index;type:varchar(36)" json:"user_id"`
	Lender           string         `gorm:"not null;type:varchar(255)" json:"lender"`
	Type             string         `gorm:"not null;type:varchar(50)" json:"type"`
	PrincipalAmount  float64        `gorm:"not null;type:decimal(12,2)" json:"principal_amount"`
	RemainingBalance float64        `gorm:"not null;type:decimal(12,2)" json:"remaining_balance"`
	MonthlyPayment   float64        `gorm:"not null;type:decimal(10,2)" json:"monthly_payment"`
	InterestRate     float64        `gorm:"not null;type:decimal(5,3)" json:"interest_rate"`
	EndDate          time.Time      `gorm:"not null" json:"end_date"`
	CreatedAt        time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Foreign key relationship
	User UserModel `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// TableName returns the table name for GORM
func (LoanModel) TableName() string {
	return "loans"
}

// BeforeCreate sets the ID if not provided
func (l *LoanModel) BeforeCreate(tx *gorm.DB) error {
	if l.ID == "" {
		l.ID = "loan-" + uuid.New().String()
	}
	if l.CreatedAt.IsZero() {
		l.CreatedAt = time.Now()
	}
	if l.UpdatedAt.IsZero() {
		l.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate updates the UpdatedAt timestamp
func (l *LoanModel) BeforeUpdate(tx *gorm.DB) error {
	l.UpdatedAt = time.Now()
	return nil
}

// ToDomain converts LoanModel to domain.Loan
func (l LoanModel) ToDomain() domain.Loan {
	return domain.Loan{
		ID:               l.ID,
		UserID:           l.UserID,
		Lender:           l.Lender,
		Type:             l.Type,
		PrincipalAmount:  l.PrincipalAmount,
		RemainingBalance: l.RemainingBalance,
		MonthlyPayment:   l.MonthlyPayment,
		InterestRate:     l.InterestRate,
		EndDate:          l.EndDate,
		CreatedAt:        l.CreatedAt,
		UpdatedAt:        l.UpdatedAt,
	}
}

// FromDomain creates LoanModel from domain.Loan
func (l *LoanModel) FromDomain(loan domain.Loan) {
	l.ID = loan.ID
	l.UserID = loan.UserID
	l.Lender = loan.Lender
	l.Type = loan.Type
	l.PrincipalAmount = loan.PrincipalAmount
	l.RemainingBalance = loan.RemainingBalance
	l.MonthlyPayment = loan.MonthlyPayment
	l.InterestRate = loan.InterestRate
	l.EndDate = loan.EndDate
	l.CreatedAt = loan.CreatedAt
	l.UpdatedAt = loan.UpdatedAt
}

// NewLoanModelFromDomain creates a new LoanModel from domain.Loan
func NewLoanModelFromDomain(loan domain.Loan) *LoanModel {
	model := &LoanModel{}
	model.FromDomain(loan)
	return model
}