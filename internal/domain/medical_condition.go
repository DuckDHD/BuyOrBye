package domain

import (
	"fmt"
	"time"
)

// MedicalCondition represents a medical condition with severity and risk assessment
type MedicalCondition struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	ProfileID          string    `json:"profile_id"`
	Name               string    `json:"name"`                  // standardized condition name
	Category           string    `json:"category"`              // "chronic", "acute", "mental_health", "preventive"
	Severity           string    `json:"severity"`              // "mild", "moderate", "severe", "critical"
	DiagnosedDate      time.Time `json:"diagnosed_date"`
	IsActive           bool      `json:"is_active"`
	RequiresMedication bool      `json:"requires_medication"`
	MonthlyMedCost     float64   `json:"monthly_med_cost"`      // estimated monthly medication cost
	RiskFactor         float64   `json:"risk_factor"`           // 0.0 to 1.0 risk multiplier
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// Validate validates the medical condition data
func (m *MedicalCondition) Validate() error {
	if m.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if m.ProfileID == "" {
		return fmt.Errorf("profile ID is required")
	}

	if m.Name == "" {
		return fmt.Errorf("condition name is required")
	}

	validCategories := []string{"chronic", "acute", "mental_health", "preventive"}
	isValidCategory := false
	for _, category := range validCategories {
		if m.Category == category {
			isValidCategory = true
			break
		}
	}
	if !isValidCategory {
		return fmt.Errorf("category must be one of: chronic, acute, mental_health, preventive")
	}

	validSeverities := []string{"mild", "moderate", "severe", "critical"}
	isValidSeverity := false
	for _, severity := range validSeverities {
		if m.Severity == severity {
			isValidSeverity = true
			break
		}
	}
	if !isValidSeverity {
		return fmt.Errorf("severity must be one of: mild, moderate, severe, critical")
	}

	if m.MonthlyMedCost < 0 {
		return fmt.Errorf("monthly medication cost must be non-negative")
	}

	if m.RiskFactor < 0.0 || m.RiskFactor > 1.0 {
		return fmt.Errorf("risk factor must be between 0.0 and 1.0")
	}

	// Check if diagnosed date is not in the future
	if m.DiagnosedDate.After(time.Now()) {
		return fmt.Errorf("diagnosed date cannot be in the future")
	}

	return nil
}

// CalculateRiskContribution calculates the risk contribution of this condition
// Returns risk points based on category and severity, only if condition is active
func (m *MedicalCondition) CalculateRiskContribution() float64 {
	if !m.IsActive {
		return 0.0
	}

	// Preventive conditions don't contribute to risk
	if m.Category == "preventive" {
		return 0.0
	}

	// Base risk points by severity
	var severityPoints float64
	switch m.Severity {
	case "mild":
		severityPoints = 2.0
	case "moderate":
		severityPoints = 5.0
	case "severe":
		severityPoints = 10.0
	case "critical":
		severityPoints = 15.0
	default:
		severityPoints = 0.0
	}

	// Adjust by category
	var categoryMultiplier float64
	switch m.Category {
	case "chronic":
		categoryMultiplier = 1.0 // Full impact
	case "acute":
		// Acute conditions have lower long-term risk impact
		if m.Severity == "critical" {
			categoryMultiplier = 0.53 // 8/15
		} else if m.Severity == "severe" {
			categoryMultiplier = 0.5 // 5/10
		} else {
			categoryMultiplier = 0.5
		}
	case "mental_health":
		// Mental health conditions have moderate impact
		if m.Severity == "critical" {
			categoryMultiplier = 0.67 // 10/15
		} else if m.Severity == "severe" {
			categoryMultiplier = 0.6 // 6/10
		} else {
			categoryMultiplier = 0.6
		}
	default:
		categoryMultiplier = 1.0
	}

	return severityPoints * categoryMultiplier
}

// GetSeverityScore returns numeric severity score (1-4)
func (m *MedicalCondition) GetSeverityScore() int {
	switch m.Severity {
	case "mild":
		return 1
	case "moderate":
		return 2
	case "severe":
		return 3
	case "critical":
		return 4
	default:
		return 0
	}
}

// IsChronic returns true if the condition is chronic or mental health (long-term)
func (m *MedicalCondition) IsChronic() bool {
	return m.Category == "chronic" || m.Category == "mental_health"
}

// GetAnnualMedCost returns the annual medication cost
func (m *MedicalCondition) GetAnnualMedCost() float64 {
	return m.MonthlyMedCost * 12
}

// RequiresHighRiskManagement determines if condition requires high-risk management
// True for severe/critical conditions that are active, except preventive
func (m *MedicalCondition) RequiresHighRiskManagement() bool {
	if !m.IsActive || m.Category == "preventive" {
		return false
	}

	return m.Severity == "severe" || m.Severity == "critical"
}