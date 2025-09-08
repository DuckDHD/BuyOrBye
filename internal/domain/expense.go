package domain

import (
	"fmt"
	"strings"
	"time"
)

// Expense represents a user's expense in the domain layer
type Expense struct {
	ID        string
	UserID    string
	Category  string
	Name      string
	Amount    float64
	Frequency string
	IsFixed   bool
	Priority  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Expense category constants
const (
	CategoryHousing       = "housing"
	CategoryFood         = "food"
	CategoryTransport    = "transport"
	CategoryEntertainment = "entertainment"
	CategoryUtilities    = "utilities"
	CategoryOther        = "other"
)

// Expense frequency constants (subset of income frequencies - no one-time)
const (
	ExpenseFrequencyMonthly = "monthly"
	ExpenseFrequencyWeekly  = "weekly"
	ExpenseFrequencyDaily   = "daily"
)

// Priority constants
const (
	PriorityEssential    = 1 // Essential expenses (rent, utilities, groceries)
	PriorityImportant    = 2 // Important expenses (car payment, insurance)
	PriorityNiceToHave   = 3 // Nice-to-have expenses (entertainment, dining out)
)

// ValidCategories contains all valid category values for expenses
var ValidCategories = []string{
	CategoryHousing,
	CategoryFood,
	CategoryTransport,
	CategoryEntertainment,
	CategoryUtilities,
	CategoryOther,
}

// ValidExpenseFrequencies contains all valid frequency values for expenses
var ValidExpenseFrequencies = []string{
	ExpenseFrequencyMonthly,
	ExpenseFrequencyWeekly,
	ExpenseFrequencyDaily,
}

// Validate validates the Expense struct and returns an error if validation fails
func (e *Expense) Validate() error {
	var errors []string

	// Validate required fields
	if e.UserID == "" {
		errors = append(errors, "user ID is required")
	}

	if e.Name == "" {
		errors = append(errors, "name is required")
	}

	if e.Category == "" {
		errors = append(errors, "category is required")
	} else if !isValidCategory(e.Category) {
		errors = append(errors, "category must be one of: housing, food, transport, entertainment, utilities, other")
	}

	if e.Amount <= 0 {
		errors = append(errors, "amount must be greater than 0")
	}

	if e.Frequency == "" {
		errors = append(errors, "frequency is required")
	} else if !isValidExpenseFrequency(e.Frequency) {
		errors = append(errors, "frequency must be one of: monthly, weekly, daily")
	}

	if e.Priority < 1 || e.Priority > 3 {
		errors = append(errors, "priority must be between 1 and 3")
	}

	if e.CreatedAt.IsZero() {
		errors = append(errors, "created at is required")
	}

	if e.UpdatedAt.IsZero() {
		errors = append(errors, "updated at is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// NormalizeToMonthly converts expense amount to monthly equivalent based on frequency
func (e *Expense) NormalizeToMonthly() float64 {
	if e.Amount <= 0 {
		return 0.0
	}

	switch e.Frequency {
	case ExpenseFrequencyMonthly:
		return e.Amount
	case ExpenseFrequencyWeekly:
		// 52 weeks per year / 12 months = 4.3333...
		return e.Amount * 52.0 / 12.0
	case ExpenseFrequencyDaily:
		// 365.25 days per year (accounting for leap years) / 12 months
		return e.Amount * 365.25 / 12.0
	default:
		// Invalid frequency
		return 0.0
	}
}

// GetCategoryDisplayName returns a user-friendly display name for the category
func (e *Expense) GetCategoryDisplayName() string {
	switch e.Category {
	case CategoryHousing:
		return "Housing"
	case CategoryFood:
		return "Food & Groceries"
	case CategoryTransport:
		return "Transportation"
	case CategoryEntertainment:
		return "Entertainment"
	case CategoryUtilities:
		return "Utilities"
	case CategoryOther:
		return "Other"
	default:
		return "Other"
	}
}

// GetPriorityName returns a user-friendly display name for the priority level
func (e *Expense) GetPriorityName() string {
	switch e.Priority {
	case PriorityEssential:
		return "Essential"
	case PriorityImportant:
		return "Important"
	case PriorityNiceToHave:
		return "Nice-to-have"
	default:
		return "Unknown"
	}
}

// GetFrequencyDisplayName returns a user-friendly display name for the frequency
func (e *Expense) GetFrequencyDisplayName() string {
	switch e.Frequency {
	case ExpenseFrequencyMonthly:
		return "Monthly"
	case ExpenseFrequencyWeekly:
		return "Weekly"
	case ExpenseFrequencyDaily:
		return "Daily"
	default:
		return "Unknown"
	}
}

// IsEssential returns true if the expense is marked as essential priority
func (e *Expense) IsEssential() bool {
	return e.Priority == PriorityEssential
}

// CalculateAnnualAmount calculates the annual equivalent of this expense
func (e *Expense) CalculateAnnualAmount() float64 {
	monthlyAmount := e.NormalizeToMonthly()
	return monthlyAmount * 12.0
}

// isValidCategory checks if the provided category is valid
func isValidCategory(category string) bool {
	for _, validCat := range ValidCategories {
		if category == validCat {
			return true
		}
	}
	return false
}

// isValidExpenseFrequency checks if the provided frequency is valid for expenses
func isValidExpenseFrequency(frequency string) bool {
	for _, validFreq := range ValidExpenseFrequencies {
		if frequency == validFreq {
			return true
		}
	}
	return false
}