package models

import (
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"gorm.io/gorm"
)

// InsurancePolicyModel represents the insurance policy database model
type InsurancePolicyModel struct {
	gorm.Model
	
	// Foreign Keys
	UserID    string `gorm:"not null;size:36;index:idx_user_policies" json:"user_id"`
	ProfileID uint   `gorm:"not null;index:idx_profile_policies" json:"profile_id"`
	
	// Policy Details
	Provider           string    `gorm:"not null;size:100" json:"provider"`
	PolicyNumber       string    `gorm:"uniqueIndex;not null;size:50" json:"policy_number"` // Unique across all policies
	Type               string    `gorm:"not null;size:10;check:type IN ('health','dental','vision')" json:"type"`
	MonthlyPremium     float64   `gorm:"not null;check:monthly_premium > 0" json:"monthly_premium"`
	AnnualDeductible   float64   `gorm:"not null;check:annual_deductible >= 0" json:"annual_deductible"`
	OutOfPocketMax     float64   `gorm:"not null;check:out_of_pocket_max > 0" json:"out_of_pocket_max"`
	CoveragePercentage int       `gorm:"not null;check:coverage_percentage >= 0 AND coverage_percentage <= 100" json:"coverage_percentage"`
	StartDate          time.Time `gorm:"not null;index:idx_policy_dates" json:"start_date"`
	EndDate            time.Time `gorm:"not null;index:idx_policy_dates" json:"end_date"`
	IsActive           bool      `gorm:"not null;default:true;index:idx_active_policies" json:"is_active"`
	
	// Deductible Tracking
	DeductibleMet      float64 `gorm:"not null;default:0;check:deductible_met >= 0" json:"deductible_met"`
	OutOfPocketCurrent float64 `gorm:"not null;default:0;check:out_of_pocket_current >= 0" json:"out_of_pocket_current"`
	
	// Relationship
	Profile HealthProfileModel `gorm:"foreignKey:ProfileID;constraint:OnDelete:CASCADE" json:"-"`
}

// TableName overrides the table name used by InsurancePolicyModel to `insurance_policies`
func (InsurancePolicyModel) TableName() string {
	return "insurance_policies"
}

// BeforeCreate sets default values and validates before creating
func (i *InsurancePolicyModel) BeforeCreate(tx *gorm.DB) error {
	i.validateDates()
	i.capDeductibleTracking()
	return nil
}

// BeforeUpdate validates before updating
func (i *InsurancePolicyModel) BeforeUpdate(tx *gorm.DB) error {
	i.capDeductibleTracking()
	return nil
}

// validateDates ensures end date is after start date
func (i *InsurancePolicyModel) validateDates() {
	if i.EndDate.Before(i.StartDate) {
		// Swap dates if end is before start
		i.StartDate, i.EndDate = i.EndDate, i.StartDate
	}
}

// capDeductibleTracking ensures tracking values don't exceed limits
func (i *InsurancePolicyModel) capDeductibleTracking() {
	if i.DeductibleMet > i.AnnualDeductible {
		i.DeductibleMet = i.AnnualDeductible
	}
	if i.OutOfPocketCurrent > i.OutOfPocketMax {
		i.OutOfPocketCurrent = i.OutOfPocketMax
	}
}

// IsCurrentlyActive checks if policy is active based on current date
func (i *InsurancePolicyModel) IsCurrentlyActive() bool {
	now := time.Now()
	return i.IsActive && 
		   now.After(i.StartDate) && 
		   now.Before(i.EndDate.AddDate(0, 0, 1)) // Include end date
}

// GetRemainingDeductible calculates remaining deductible amount
func (i *InsurancePolicyModel) GetRemainingDeductible() float64 {
	remaining := i.AnnualDeductible - i.DeductibleMet
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetRemainingOutOfPocket calculates remaining out-of-pocket maximum
func (i *InsurancePolicyModel) GetRemainingOutOfPocket() float64 {
	remaining := i.OutOfPocketMax - i.OutOfPocketCurrent
	if remaining < 0 {
		return 0
	}
	return remaining
}

// CalculateCoverage calculates coverage for a given expense amount
func (i *InsurancePolicyModel) CalculateCoverage(expenseAmount float64) (covered, outOfPocket float64) {
	if !i.IsCurrentlyActive() || expenseAmount <= 0 {
		return 0, expenseAmount
	}
	
	remainingDeductible := i.GetRemainingDeductible()
	remainingOutOfPocket := i.GetRemainingOutOfPocket()
	
	// If deductible not met, patient pays deductible first
	if remainingDeductible > 0 {
		deductibleAmount := expenseAmount
		if deductibleAmount > remainingDeductible {
			deductibleAmount = remainingDeductible
		}
		
		// Apply coverage percentage to remaining amount after deductible
		remainingAmount := expenseAmount - deductibleAmount
		if remainingAmount > 0 {
			coveragePercent := float64(i.CoveragePercentage) / 100.0
			insurancePortion := remainingAmount * coveragePercent
			patientPortion := remainingAmount * (1 - coveragePercent)
			
			// Check out-of-pocket maximum
			totalPatientCost := deductibleAmount + patientPortion
			if totalPatientCost > remainingOutOfPocket {
				totalPatientCost = remainingOutOfPocket
				insurancePortion = expenseAmount - totalPatientCost
			}
			
			return insurancePortion, totalPatientCost
		}
		
		return 0, deductibleAmount
	}
	
	// Deductible is met, apply coverage percentage
	coveragePercent := float64(i.CoveragePercentage) / 100.0
	insurancePortion := expenseAmount * coveragePercent
	patientPortion := expenseAmount * (1 - coveragePercent)
	
	// Check out-of-pocket maximum
	if patientPortion > remainingOutOfPocket {
		patientPortion = remainingOutOfPocket
		insurancePortion = expenseAmount - patientPortion
	}
	
	return insurancePortion, patientPortion
}

// ToDomain converts InsurancePolicyModel to domain.InsurancePolicy
func (i *InsurancePolicyModel) ToDomain() *domain.InsurancePolicy {
	return &domain.InsurancePolicy{
		ID:                 fmt.Sprintf("%d", i.ID), // Convert uint to string
		UserID:             i.UserID,
		ProfileID:          fmt.Sprintf("%d", i.ProfileID),
		Provider:           i.Provider,
		PolicyNumber:       i.PolicyNumber,
		Type:               i.Type,
		MonthlyPremium:     i.MonthlyPremium,
		Deductible:         i.AnnualDeductible,
		DeductibleMet:      i.DeductibleMet,
		OutOfPocketMax:     i.OutOfPocketMax,
		OutOfPocketCurrent: i.OutOfPocketCurrent,
		CoveragePercentage: float64(i.CoveragePercentage),
		StartDate:          i.StartDate,
		EndDate:            i.EndDate,
		IsActive:           i.IsActive,
		CreatedAt:          i.CreatedAt,
		UpdatedAt:          i.UpdatedAt,
	}
}

// FromDomain creates InsurancePolicyModel from domain.InsurancePolicy
func (i *InsurancePolicyModel) FromDomain(policy *domain.InsurancePolicy, profileID uint) {
	// Note: We don't set ID since it's auto-generated
	i.UserID = policy.UserID
	i.ProfileID = profileID
	i.Provider = policy.Provider
	i.PolicyNumber = policy.PolicyNumber
	i.Type = policy.Type
	i.MonthlyPremium = policy.MonthlyPremium
	i.AnnualDeductible = policy.Deductible
	i.DeductibleMet = policy.DeductibleMet
	i.OutOfPocketMax = policy.OutOfPocketMax
	i.OutOfPocketCurrent = policy.OutOfPocketCurrent
	i.CoveragePercentage = int(policy.CoveragePercentage)
	i.StartDate = policy.StartDate
	i.EndDate = policy.EndDate
	i.IsActive = policy.IsActive
	i.CreatedAt = policy.CreatedAt
	i.UpdatedAt = policy.UpdatedAt
}