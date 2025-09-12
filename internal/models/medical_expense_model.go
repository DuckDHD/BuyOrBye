package models

import (
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"gorm.io/gorm"
)

// MedicalExpenseModel represents the medical expense database model
type MedicalExpenseModel struct {
	gorm.Model
	
	// Foreign Keys
	UserID    string `gorm:"not null;size:36;index:idx_user_expenses" json:"user_id"`
	ProfileID uint   `gorm:"not null;index:idx_profile_expenses" json:"profile_id"`
	
	// Expense Details
	Amount           float64   `gorm:"not null;check:amount > 0" json:"amount"`
	Category         string    `gorm:"not null;size:20;index:idx_expense_category;check:category IN ('doctor_visit','medication','hospital','lab_test','therapy','equipment')" json:"category"`
	Description      string    `gorm:"not null;size:200" json:"description"`
	IsRecurring      bool      `gorm:"not null;default:false;index:idx_recurring_expenses" json:"is_recurring"`
	Frequency        string    `gorm:"size:20;check:frequency IN ('daily','weekly','bi-weekly','monthly','quarterly','semi-annually','annually','one_time')" json:"frequency"`
	IsCovered        bool      `gorm:"not null;default:false" json:"is_covered"`
	InsurancePayment float64   `gorm:"not null;default:0;check:insurance_payment >= 0" json:"insurance_payment"`
	OutOfPocket      float64   `gorm:"not null;check:out_of_pocket >= 0" json:"out_of_pocket"`
	Date             time.Time `gorm:"not null;index:idx_expense_date" json:"date"`
	
	// Relationship
	Profile HealthProfileModel `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName overrides the table name used by MedicalExpenseModel to `medical_expenses`
func (MedicalExpenseModel) TableName() string {
	return "medical_expenses"
}

// BeforeCreate calculates out-of-pocket and sets defaults before creating
func (m *MedicalExpenseModel) BeforeCreate(tx *gorm.DB) error {
	m.calculateOutOfPocket()
	m.setDefaultFrequency()
	return nil
}

// BeforeUpdate recalculates out-of-pocket before updating
func (m *MedicalExpenseModel) BeforeUpdate(tx *gorm.DB) error {
	m.calculateOutOfPocket()
	return nil
}

// calculateOutOfPocket calculates the out-of-pocket amount
func (m *MedicalExpenseModel) calculateOutOfPocket() {
	if m.InsurancePayment > m.Amount {
		m.InsurancePayment = m.Amount // Cap insurance payment at total amount
	}
	m.OutOfPocket = m.Amount - m.InsurancePayment
}

// setDefaultFrequency sets default frequency based on recurring flag
func (m *MedicalExpenseModel) setDefaultFrequency() {
	if m.Frequency == "" {
		if m.IsRecurring {
			m.Frequency = "monthly"
		} else {
			m.Frequency = "one_time"
		}
	}
}

// ToDomain converts MedicalExpenseModel to domain.MedicalExpense
func (m *MedicalExpenseModel) ToDomain() *domain.MedicalExpense {
	return &domain.MedicalExpense{
		ID:               fmt.Sprintf("%d", m.ID), // Convert uint to string
		UserID:           m.UserID,
		ProfileID:        fmt.Sprintf("%d", m.ProfileID),
		Amount:           m.Amount,
		Category:         m.Category,
		Description:      m.Description,
		IsRecurring:      m.IsRecurring,
		Frequency:        m.Frequency,
		IsCovered:        m.IsCovered,
		InsurancePayment: m.InsurancePayment,
		OutOfPocket:      m.OutOfPocket,
		Date:             m.Date,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

// FromDomain creates MedicalExpenseModel from domain.MedicalExpense
func (m *MedicalExpenseModel) FromDomain(expense *domain.MedicalExpense, profileID uint) {
	// Note: We don't set ID since it's auto-generated
	m.UserID = expense.UserID
	m.ProfileID = profileID
	m.Amount = expense.Amount
	m.Category = expense.Category
	m.Description = expense.Description
	m.IsRecurring = expense.IsRecurring
	m.Frequency = expense.Frequency
	m.IsCovered = expense.IsCovered
	m.InsurancePayment = expense.InsurancePayment
	m.OutOfPocket = expense.OutOfPocket
	m.Date = expense.Date
	m.CreatedAt = expense.CreatedAt
	m.UpdatedAt = expense.UpdatedAt
}