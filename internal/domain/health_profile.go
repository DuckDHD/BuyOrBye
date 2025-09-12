package domain

import (
	"fmt"
	"math"
	"time"
)

// HealthProfile represents a user's health profile with BMI calculation and risk assessment
type HealthProfile struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"user_id"`
	Age                  int       `json:"age"`
	Gender               string    `json:"gender"`                 // "male", "female", "other"
	Height               float64   `json:"height"`                 // in cm
	Weight               float64   `json:"weight"`                 // in kg
	BMI                  float64   `json:"bmi"`                    // calculated
	FamilySize           int       `json:"family_size"`            // household members
	HasChronicConditions bool      `json:"has_chronic_conditions"` // quick flag
	EmergencyFundHealth  float64   `json:"emergency_fund_health"`  // health-specific emergency fund
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// CalculateBMI calculates BMI from height and weight
// BMI = weight (kg) / (height (m))^2
func (h *HealthProfile) CalculateBMI() (float64, error) {
	if h.Height <= 0 {
		return 0, fmt.Errorf("height must be positive")
	}
	if h.Weight <= 0 {
		return 0, fmt.Errorf("weight must be positive")
	}

	// Convert height from cm to meters
	heightInMeters := h.Height / 100
	bmi := h.Weight / (heightInMeters * heightInMeters)

	// Round to 2 decimal places
	return math.Round(bmi*100) / 100, nil
}

// UpdateBMI calculates and updates the BMI field
func (h *HealthProfile) UpdateBMI() error {
	bmi, err := h.CalculateBMI()
	if err != nil {
		return err
	}
	h.BMI = bmi
	return nil
}

// Validate validates the health profile data
func (h *HealthProfile) Validate() error {
	if h.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if h.Age < 1 || h.Age > 150 {
		return fmt.Errorf("age must be between 1 and 150")
	}

	validGenders := []string{"male", "female", "other"}
	isValidGender := false
	for _, gender := range validGenders {
		if h.Gender == gender {
			isValidGender = true
			break
		}
	}
	if !isValidGender {
		return fmt.Errorf("gender must be one of: male, female, other")
	}

	if h.Height <= 0 {
		return fmt.Errorf("height must be positive")
	}

	if h.Weight <= 0 {
		return fmt.Errorf("weight must be positive")
	}

	if h.FamilySize < 1 {
		return fmt.Errorf("family size must be at least 1")
	}

	return nil
}

// HasHighRisk determines if the person has high health risk
// Factors: Age >= 65, BMI < 18.5 or BMI >= 30, or has chronic conditions
func (h *HealthProfile) HasHighRisk() bool {
	// Age factor: 65 and older is high risk
	if h.Age >= 65 {
		return true
	}

	// BMI factor: underweight or obese is high risk
	if h.BMI < 18.5 || h.BMI >= 30.0 {
		return true
	}

	// Chronic conditions factor
	if h.HasChronicConditions {
		return true
	}

	return false
}

// GetAgeGroup returns the age group category
func (h *HealthProfile) GetAgeGroup() string {
	switch {
	case h.Age < 18:
		return "child"
	case h.Age < 30:
		return "young_adult"
	case h.Age < 65:
		return "middle_aged"
	default:
		return "senior"
	}
}

// GetBMICategory returns the BMI category
func (h *HealthProfile) GetBMICategory() string {
	switch {
	case h.BMI < 18.5:
		return "underweight"
	case h.BMI < 25:
		return "normal"
	case h.BMI < 30:
		return "overweight"
	default:
		return "obese"
	}
}
