package domain

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Loan represents a user's loan in the domain layer
type Loan struct {
	ID               string
	UserID           string
	Lender           string
	Type             string
	PrincipalAmount  float64
	RemainingBalance float64
	MonthlyPayment   float64
	InterestRate     float64
	EndDate          time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Loan type constants
const (
	LoanTypeMortgage = "mortgage"
	LoanTypeAuto     = "auto"
	LoanTypePersonal = "personal"
	LoanTypeStudent  = "student"
)

// Interest rate thresholds for each loan type (considered "high" if above these)
const (
	HighInterestThresholdMortgage = 6.0  // 6%
	HighInterestThresholdAuto     = 8.0  // 8%
	HighInterestThresholdPersonal = 15.0 // 15%
	HighInterestThresholdStudent  = 7.0  // 7%
)

// ValidLoanTypes contains all valid loan type values
var ValidLoanTypes = []string{
	LoanTypeMortgage,
	LoanTypeAuto,
	LoanTypePersonal,
	LoanTypeStudent,
}

// Validate validates the Loan struct and returns an error if validation fails
func (l *Loan) Validate() error {
	var errors []string

	// Validate required fields
	if l.UserID == "" {
		errors = append(errors, "user ID is required")
	}

	if l.Lender == "" {
		errors = append(errors, "lender is required")
	}

	if l.Type == "" {
		errors = append(errors, "loan type is required")
	} else if !isValidLoanType(l.Type) {
		errors = append(errors, "loan type must be one of: mortgage, auto, personal, student")
	}

	if l.PrincipalAmount <= 0 {
		errors = append(errors, "principal amount must be greater than 0")
	}

	if l.RemainingBalance <= 0 {
		errors = append(errors, "remaining balance must be greater than 0")
	}

	if l.MonthlyPayment <= 0 {
		errors = append(errors, "monthly payment must be greater than 0")
	}

	if l.InterestRate < 0 || l.InterestRate > 100 {
		errors = append(errors, "interest rate must be between 0 and 100")
	}

	// Remaining balance cannot exceed principal amount
	if l.RemainingBalance > l.PrincipalAmount {
		errors = append(errors, "remaining balance cannot exceed principal amount")
	}

	// End date must be in the future
	if !l.EndDate.IsZero() && l.EndDate.Before(time.Now()) {
		errors = append(errors, "end date must be in the future")
	}

	if l.CreatedAt.IsZero() {
		errors = append(errors, "created at is required")
	}

	if l.UpdatedAt.IsZero() {
		errors = append(errors, "updated at is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// CalculateProgress returns the percentage of the loan that has been paid off
func (l *Loan) CalculateProgress() float64 {
	if l.PrincipalAmount <= 0 {
		return 0.0
	}

	paidAmount := l.PrincipalAmount - l.RemainingBalance
	return (paidAmount / l.PrincipalAmount) * 100.0
}

// GetTypeDisplayName returns a user-friendly display name for the loan type
func (l *Loan) GetTypeDisplayName() string {
	switch l.Type {
	case LoanTypeMortgage:
		return "Mortgage"
	case LoanTypeAuto:
		return "Auto Loan"
	case LoanTypePersonal:
		return "Personal Loan"
	case LoanTypeStudent:
		return "Student Loan"
	default:
		return "Other"
	}
}

// CalculateMonthsRemaining calculates the approximate number of months remaining to pay off the loan
// This is a simplified calculation assuming fixed monthly payments and compound interest
func (l *Loan) CalculateMonthsRemaining() (int, error) {
	if l.MonthlyPayment <= 0 {
		return 0, fmt.Errorf("monthly payment must be greater than 0")
	}

	if l.RemainingBalance <= 0 {
		return 0, nil // Loan is already paid off
	}

	if l.InterestRate == 0 {
		// Simple calculation without interest
		return int(math.Ceil(l.RemainingBalance / l.MonthlyPayment)), nil
	}

	// Monthly interest rate
	monthlyRate := l.InterestRate / 100.0 / 12.0

	// Check if monthly payment is sufficient to cover interest
	monthlyInterest := l.RemainingBalance * monthlyRate
	if l.MonthlyPayment <= monthlyInterest {
		return 0, fmt.Errorf("monthly payment is insufficient to cover interest charges")
	}

	// Calculate months using loan payment formula
	// n = -log(1 - (P * r) / M) / log(1 + r)
	// where P = principal, r = monthly rate, M = monthly payment
	numerator := math.Log(1 - (l.RemainingBalance*monthlyRate)/l.MonthlyPayment)
	denominator := math.Log(1 + monthlyRate)

	months := -numerator / denominator
	return int(math.Ceil(months)), nil
}

// IsHighInterestRate returns true if the loan has a high interest rate for its type
func (l *Loan) IsHighInterestRate() bool {
	switch l.Type {
	case LoanTypeMortgage:
		return l.InterestRate > HighInterestThresholdMortgage
	case LoanTypeAuto:
		return l.InterestRate > HighInterestThresholdAuto
	case LoanTypePersonal:
		return l.InterestRate > HighInterestThresholdPersonal
	case LoanTypeStudent:
		return l.InterestRate > HighInterestThresholdStudent
	default:
		// Unknown loan type, use personal loan threshold as default
		return l.InterestRate > HighInterestThresholdPersonal
	}
}

// CalculateTotalInterest calculates the total interest that will be paid over the life of the loan
func (l *Loan) CalculateTotalInterest() (float64, error) {
	monthsRemaining, err := l.CalculateMonthsRemaining()
	if err != nil {
		return 0, err
	}

	if monthsRemaining == 0 {
		return 0, nil // No interest if loan is paid off
	}

	totalPayments := float64(monthsRemaining) * l.MonthlyPayment
	totalInterest := totalPayments - l.RemainingBalance

	// Ensure interest is not negative
	if totalInterest < 0 {
		return 0, nil
	}

	return totalInterest, nil
}

// GetPayoffDate calculates the estimated payoff date based on current monthly payments
func (l *Loan) GetPayoffDate() (time.Time, error) {
	monthsRemaining, err := l.CalculateMonthsRemaining()
	if err != nil {
		return time.Time{}, err
	}

	if monthsRemaining == 0 {
		return time.Now(), nil // Already paid off
	}

	return time.Now().AddDate(0, monthsRemaining, 0), nil
}

// IsNearPayoff returns true if the loan has less than 12 months remaining
func (l *Loan) IsNearPayoff() bool {
	monthsRemaining, err := l.CalculateMonthsRemaining()
	if err != nil {
		return false
	}
	return monthsRemaining <= 12 && monthsRemaining > 0
}

// isValidLoanType checks if the provided loan type is valid
func isValidLoanType(loanType string) bool {
	for _, validType := range ValidLoanTypes {
		if loanType == validType {
			return true
		}
	}
	return false
}