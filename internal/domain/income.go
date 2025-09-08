package domain

import (
	"fmt"
	"strings"
	"time"
)

// Income represents a user's income source in the domain layer
type Income struct {
	ID        string
	UserID    string
	Source    string
	Amount    float64
	Frequency string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Frequency constants for income
const (
	FrequencyMonthly  = "monthly"
	FrequencyWeekly   = "weekly"
	FrequencyDaily    = "daily"
	FrequencyOneTime  = "one-time"
)

// ValidFrequencies contains all valid frequency values for income
var ValidFrequencies = []string{
	FrequencyMonthly,
	FrequencyWeekly,
	FrequencyDaily,
	FrequencyOneTime,
}

// Validate validates the Income struct and returns an error if validation fails
func (i *Income) Validate() error {
	var errors []string

	// Validate required fields
	if i.UserID == "" {
		errors = append(errors, "user ID is required")
	}

	if i.Source == "" {
		errors = append(errors, "source is required")
	}

	if i.Amount <= 0 {
		errors = append(errors, "amount must be greater than 0")
	}

	if i.Frequency == "" {
		errors = append(errors, "frequency is required")
	} else if !isValidFrequency(i.Frequency) {
		errors = append(errors, "frequency must be one of: monthly, weekly, daily, one-time")
	}

	if i.CreatedAt.IsZero() {
		errors = append(errors, "created at is required")
	}

	if i.UpdatedAt.IsZero() {
		errors = append(errors, "updated at is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// NormalizeToMonthly converts income amount to monthly equivalent based on frequency
func (i *Income) NormalizeToMonthly() float64 {
	if i.Amount <= 0 {
		return 0.0
	}

	switch i.Frequency {
	case FrequencyMonthly:
		return i.Amount
	case FrequencyWeekly:
		// 52 weeks per year / 12 months = 4.3333...
		return i.Amount * 52.0 / 12.0
	case FrequencyDaily:
		// 365.25 days per year (accounting for leap years) / 12 months
		return i.Amount * 365.25 / 12.0
	case FrequencyOneTime:
		// One-time income doesn't contribute to regular monthly calculations
		return 0.0
	default:
		// Invalid frequency
		return 0.0
	}
}

// isValidFrequency checks if the provided frequency is valid
func isValidFrequency(frequency string) bool {
	for _, validFreq := range ValidFrequencies {
		if frequency == validFreq {
			return true
		}
	}
	return false
}

// GetFrequencyDisplayName returns a user-friendly display name for the frequency
func (i *Income) GetFrequencyDisplayName() string {
	switch i.Frequency {
	case FrequencyMonthly:
		return "Monthly"
	case FrequencyWeekly:
		return "Weekly"
	case FrequencyDaily:
		return "Daily"
	case FrequencyOneTime:
		return "One-time"
	default:
		return "Unknown"
	}
}

// IsRecurring returns true if the income is recurring (not one-time)
func (i *Income) IsRecurring() bool {
	return i.Frequency != FrequencyOneTime && i.Frequency != ""
}

// CalculateAnnualAmount calculates the annual equivalent of this income
func (i *Income) CalculateAnnualAmount() float64 {
	monthlyAmount := i.NormalizeToMonthly()
	if i.Frequency == FrequencyOneTime {
		// One-time income is just the amount itself annually
		return i.Amount
	}
	return monthlyAmount * 12.0
}