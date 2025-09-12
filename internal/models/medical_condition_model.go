package models

import (
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"gorm.io/gorm"
)

// MedicalConditionModel represents the medical condition database model
type MedicalConditionModel struct {
	gorm.Model
	
	// Foreign Keys
	UserID    string `gorm:"not null;size:36;index:idx_user_conditions" json:"user_id"`
	ProfileID uint   `gorm:"not null;index:idx_profile_conditions" json:"profile_id"`
	
	// Condition Details
	Name               string    `gorm:"not null;size:100;index:idx_condition_name" json:"name"`
	Category           string    `gorm:"not null;size:20;index:idx_condition_category;check:category IN ('chronic','acute','mental_health','preventive')" json:"category"`
	Severity           string    `gorm:"not null;size:10;check:severity IN ('mild','moderate','severe','critical')" json:"severity"`
	DiagnosedDate      time.Time `gorm:"not null" json:"diagnosed_date"`
	IsActive           bool      `gorm:"not null;default:true;index:idx_active_conditions" json:"is_active"`
	RequiresMedication bool      `gorm:"not null;default:false" json:"requires_medication"`
	MonthlyMedCost     float64   `gorm:"not null;default:0;check:monthly_med_cost >= 0" json:"monthly_med_cost"`
	RiskFactor         float64   `gorm:"not null;default:0.1;check:risk_factor >= 0 AND risk_factor <= 1" json:"risk_factor"`
	
	// Relationship
	Profile HealthProfileModel `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName overrides the table name used by MedicalConditionModel to `medical_conditions`
func (MedicalConditionModel) TableName() string {
	return "medical_conditions"
}

// BeforeCreate sets default values before creating
func (m *MedicalConditionModel) BeforeCreate(tx *gorm.DB) error {
	if m.RiskFactor == 0 {
		m.RiskFactor = 0.1 // Default risk factor
	}
	return nil
}

// ToDomain converts MedicalConditionModel to domain.MedicalCondition
func (m *MedicalConditionModel) ToDomain() *domain.MedicalCondition {
	return &domain.MedicalCondition{
		ID:                 fmt.Sprintf("%d", m.ID), // Convert uint to string
		UserID:             m.UserID,
		ProfileID:          fmt.Sprintf("%d", m.ProfileID),
		Name:               m.Name,
		Category:           m.Category,
		Severity:           m.Severity,
		DiagnosedDate:      m.DiagnosedDate,
		IsActive:           m.IsActive,
		RequiresMedication: m.RequiresMedication,
		MonthlyMedCost:     m.MonthlyMedCost,
		RiskFactor:         m.RiskFactor,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}

// FromDomain creates MedicalConditionModel from domain.MedicalCondition
func (m *MedicalConditionModel) FromDomain(condition *domain.MedicalCondition, profileID uint) {
	// Note: We don't set ID since it's auto-generated
	m.UserID = condition.UserID
	m.ProfileID = profileID
	m.Name = condition.Name
	m.Category = condition.Category
	m.Severity = condition.Severity
	m.DiagnosedDate = condition.DiagnosedDate
	m.IsActive = condition.IsActive
	m.RequiresMedication = condition.RequiresMedication
	m.MonthlyMedCost = condition.MonthlyMedCost
	m.RiskFactor = condition.RiskFactor
	m.CreatedAt = condition.CreatedAt
	m.UpdatedAt = condition.UpdatedAt
}