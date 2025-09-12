package domain

import (
	"fmt"
	"time"
)

// InsurancePolicy represents an insurance policy with deductible tracking
type InsurancePolicy struct {
	ID                  string    `json:"id"`
	UserID              string    `json:"user_id"`
	ProfileID           string    `json:"profile_id"`
	Provider            string    `json:"provider"`
	PolicyNumber        string    `json:"policy_number"`
	Type                string    `json:"type"`                      // "health", "dental", "vision", "comprehensive"
	MonthlyPremium      float64   `json:"monthly_premium"`
	Deductible          float64   `json:"deductible"`
	DeductibleMet       float64   `json:"deductible_met"`            // amount already paid toward deductible
	OutOfPocketMax      float64   `json:"out_of_pocket_max"`
	OutOfPocketCurrent  float64   `json:"out_of_pocket_current"`     // current OOP expenses
	CoveragePercentage  float64   `json:"coverage_percentage"`       // after deductible (e.g., 80%)
	StartDate           time.Time `json:"start_date"`
	EndDate             time.Time `json:"end_date"`
	IsActive            bool      `json:"is_active"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// Validate validates the insurance policy data
func (i *InsurancePolicy) Validate() error {
	if i.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if i.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	if i.PolicyNumber == "" {
		return fmt.Errorf("policy number is required")
	}

	validTypes := []string{"health", "dental", "vision", "comprehensive"}
	isValidType := false
	for _, policyType := range validTypes {
		if i.Type == policyType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return fmt.Errorf("type must be one of: health, dental, vision, comprehensive")
	}

	if i.MonthlyPremium <= 0 {
		return fmt.Errorf("monthly premium must be positive")
	}

	if i.Deductible < 0 {
		return fmt.Errorf("deductible must be non-negative")
	}

	if i.OutOfPocketMax <= 0 {
		return fmt.Errorf("out of pocket maximum must be positive")
	}

	if i.CoveragePercentage < 0 || i.CoveragePercentage > 100 {
		return fmt.Errorf("coverage percentage must be between 0 and 100")
	}

	if i.EndDate.Before(i.StartDate) || i.EndDate.Equal(i.StartDate) {
		return fmt.Errorf("end date must be after start date")
	}

	if i.DeductibleMet < 0 {
		return fmt.Errorf("deductible met must be non-negative")
	}

	if i.DeductibleMet > i.Deductible {
		return fmt.Errorf("deductible met cannot exceed total deductible")
	}

	if i.OutOfPocketCurrent > i.OutOfPocketMax {
		return fmt.Errorf("current out of pocket cannot exceed maximum")
	}

	return nil
}

// CalculateCoverage calculates coverage for an expense considering deductible
// Returns: insuranceCoverage, outOfPocketAmount, newDeductibleMet
func (i *InsurancePolicy) CalculateCoverage(expenseAmount float64) (float64, float64, float64) {
	remainingDeductible := i.GetRemainingDeductible()
	newDeductibleMet := i.DeductibleMet
	
	var amountSubjectToCoverage float64
	var deductiblePortion float64

	if remainingDeductible > 0 {
		// Still have deductible to meet
		deductiblePortion = expenseAmount
		if deductiblePortion > remainingDeductible {
			deductiblePortion = remainingDeductible
		}
		
		newDeductibleMet += deductiblePortion
		amountSubjectToCoverage = expenseAmount - deductiblePortion
	} else {
		// Deductible already met
		amountSubjectToCoverage = expenseAmount
	}

	// Calculate insurance coverage on the amount subject to coverage
	insuranceCoverage := amountSubjectToCoverage * (i.CoveragePercentage / 100)
	
	// Check out-of-pocket maximum
	remainingOutOfPocket := i.GetRemainingOutOfPocket()
	outOfPocketForThisExpense := expenseAmount - insuranceCoverage
	
	if outOfPocketForThisExpense > remainingOutOfPocket {
		// Hit out-of-pocket max, insurance covers the rest
		outOfPocketForThisExpense = remainingOutOfPocket
		insuranceCoverage = expenseAmount - outOfPocketForThisExpense
	}

	return insuranceCoverage, outOfPocketForThisExpense, newDeductibleMet
}

// GetRemainingDeductible returns the remaining deductible amount
func (i *InsurancePolicy) GetRemainingDeductible() float64 {
	remaining := i.Deductible - i.DeductibleMet
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GetRemainingOutOfPocket returns the remaining out-of-pocket maximum
func (i *InsurancePolicy) GetRemainingOutOfPocket() float64 {
	remaining := i.OutOfPocketMax - i.OutOfPocketCurrent
	if remaining < 0 {
		return 0
	}
	return remaining
}

// IsDeductibleMet returns true if the deductible has been met
func (i *InsurancePolicy) IsDeductibleMet() bool {
	return i.DeductibleMet >= i.Deductible
}

// IsOutOfPocketMaxReached returns true if the out-of-pocket maximum has been reached
func (i *InsurancePolicy) IsOutOfPocketMaxReached() bool {
	return i.OutOfPocketCurrent >= i.OutOfPocketMax
}

// GetAnnualPremium returns the annual premium amount
func (i *InsurancePolicy) GetAnnualPremium() float64 {
	return i.MonthlyPremium * 12
}