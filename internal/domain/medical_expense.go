package domain

import (
	"fmt"
	"time"
)

// MedicalExpense represents a medical expense with insurance tracking
type MedicalExpense struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	ProfileID        string    `json:"profile_id"`
	Amount           float64   `json:"amount"`                  // total expense amount
	Category         string    `json:"category"`                // "doctor_visit", "medication", "hospital", "lab_test", "therapy", "equipment"
	Description      string    `json:"description"`
	IsRecurring      bool      `json:"is_recurring"`
	Frequency        string    `json:"frequency"`               // "monthly", "quarterly", "annually", "one_time"
	IsCovered        bool      `json:"is_covered"`              // covered by insurance
	InsurancePayment float64   `json:"insurance_payment"`       // amount paid by insurance
	OutOfPocket      float64   `json:"out_of_pocket"`           // actual user payment
	Date             time.Time `json:"date"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Validate validates the medical expense data
func (m *MedicalExpense) Validate() error {
	if m.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if m.ProfileID == "" {
		return fmt.Errorf("profile ID is required")
	}

	if m.Amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	validCategories := []string{"doctor_visit", "medication", "hospital", "lab_test", "therapy", "equipment"}
	isValidCategory := false
	for _, category := range validCategories {
		if m.Category == category {
			isValidCategory = true
			break
		}
	}
	if !isValidCategory {
		return fmt.Errorf("category must be one of: doctor_visit, medication, hospital, lab_test, therapy, equipment")
	}

	// If recurring, frequency must be specified and valid
	if m.IsRecurring {
		if m.Frequency == "" {
			return fmt.Errorf("frequency is required for recurring expenses")
		}

		validFrequencies := []string{"monthly", "quarterly", "annually"}
		isValidFrequency := false
		for _, frequency := range validFrequencies {
			if m.Frequency == frequency {
				isValidFrequency = true
				break
			}
		}
		if !isValidFrequency {
			return fmt.Errorf("frequency must be one of: monthly, quarterly, annually")
		}
	}

	if m.InsurancePayment < 0 {
		return fmt.Errorf("insurance payment must be non-negative")
	}

	if m.InsurancePayment > m.Amount {
		return fmt.Errorf("insurance payment cannot exceed total amount")
	}

	if m.OutOfPocket < 0 {
		return fmt.Errorf("out of pocket amount must be non-negative")
	}

	if m.OutOfPocket > m.Amount {
		return fmt.Errorf("out of pocket amount cannot exceed total amount")
	}

	// Check if expense date is not in the future
	if m.Date.After(time.Now()) {
		return fmt.Errorf("expense date cannot be in the future")
	}

	return nil
}

// CalculateOutOfPocket calculates the out-of-pocket amount from total and insurance payment
func (m *MedicalExpense) CalculateOutOfPocket() float64 {
	return m.Amount - m.InsurancePayment
}

// GetAnnualizedCost returns the annualized cost based on frequency
func (m *MedicalExpense) GetAnnualizedCost() float64 {
	if !m.IsRecurring {
		return m.Amount
	}

	multiplier := m.GetFrequencyMultiplier()
	return m.Amount * multiplier
}

// GetCoveragePercentage returns the insurance coverage percentage
func (m *MedicalExpense) GetCoveragePercentage() float64 {
	if m.Amount == 0 {
		return 0.0
	}
	return (m.InsurancePayment / m.Amount) * 100
}

// GetFrequencyMultiplier returns the annual multiplier for the frequency
func (m *MedicalExpense) GetFrequencyMultiplier() float64 {
	switch m.Frequency {
	case "monthly":
		return 12.0
	case "quarterly":
		return 4.0
	case "annually":
		return 1.0
	default:
		return 1.0 // For "one_time" or unknown
	}
}

// IsHighCostExpense determines if this is a high-cost expense (>= $500)
func (m *MedicalExpense) IsHighCostExpense() bool {
	return m.Amount >= 500.0
}