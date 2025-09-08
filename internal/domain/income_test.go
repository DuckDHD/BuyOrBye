package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIncome_Validate_AllFieldsValid_ReturnsNil(t *testing.T) {
	// Arrange
	income := Income{
		ID:        "income-123",
		UserID:    "user-123",
		Source:    "Salary",
		Amount:    5000.50,
		Frequency: "monthly",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Act
	err := income.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestIncome_Validate_ZeroAmount_ReturnsError(t *testing.T) {
	// Arrange
	income := Income{
		ID:        "income-123",
		UserID:    "user-123",
		Source:    "Salary",
		Amount:    0.0, // Invalid amount
		Frequency: "monthly",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Act
	err := income.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be greater than 0")
}

func TestIncome_Validate_NegativeAmount_ReturnsError(t *testing.T) {
	// Arrange
	income := Income{
		ID:        "income-123",
		UserID:    "user-123",
		Source:    "Salary",
		Amount:    -100.0, // Invalid negative amount
		Frequency: "monthly",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Act
	err := income.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be greater than 0")
}

func TestIncome_Validate_EmptySource_ReturnsError(t *testing.T) {
	// Arrange
	income := Income{
		ID:        "income-123",
		UserID:    "user-123",
		Source:    "", // Empty source
		Amount:    5000.50,
		Frequency: "monthly",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Act
	err := income.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source is required")
}

func TestIncome_Validate_InvalidFrequency_ReturnsError(t *testing.T) {
	// Arrange
	invalidFrequencies := []string{
		"yearly",
		"biweekly",
		"quarterly",
		"",
		"MONTHLY", // Case sensitive
		"invalid",
	}

	for _, frequency := range invalidFrequencies {
		t.Run("frequency_"+frequency, func(t *testing.T) {
			income := Income{
				ID:        "income-123",
				UserID:    "user-123",
				Source:    "Salary",
				Amount:    5000.50,
				Frequency: frequency,
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Act
			err := income.Validate()

			// Assert
			assert.Error(t, err)
			if frequency == "" {
				assert.Contains(t, err.Error(), "frequency is required")
			} else {
				assert.Contains(t, err.Error(), "frequency must be one of: monthly, weekly, daily, one-time")
			}
		})
	}
}

func TestIncome_Validate_ValidFrequencies_ReturnsNil(t *testing.T) {
	// Arrange
	validFrequencies := []string{
		"monthly",
		"weekly", 
		"daily",
		"one-time",
	}

	for _, frequency := range validFrequencies {
		t.Run("valid_frequency_"+frequency, func(t *testing.T) {
			income := Income{
				ID:        "income-123",
				UserID:    "user-123",
				Source:    "Salary",
				Amount:    5000.50,
				Frequency: frequency,
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Act
			err := income.Validate()

			// Assert
			assert.NoError(t, err)
		})
	}
}

func TestIncome_Validate_EmptyUserID_ReturnsError(t *testing.T) {
	// Arrange
	income := Income{
		ID:        "income-123",
		UserID:    "", // Empty user ID
		Source:    "Salary",
		Amount:    5000.50,
		Frequency: "monthly",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Act
	err := income.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID is required")
}

func TestIncome_Validate_ZeroCreatedAt_ReturnsError(t *testing.T) {
	// Arrange
	income := Income{
		ID:        "income-123",
		UserID:    "user-123",
		Source:    "Salary",
		Amount:    5000.50,
		Frequency: "monthly",
		IsActive:  true,
		CreatedAt: time.Time{}, // Zero time
		UpdatedAt: time.Now(),
	}

	// Act
	err := income.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "created at is required")
}

func TestIncome_Validate_MultipleErrors_ReturnsAllErrors(t *testing.T) {
	// Arrange
	income := Income{
		ID:        "income-123",
		UserID:    "",        // Missing user ID
		Source:    "",        // Missing source
		Amount:    -100.0,    // Invalid amount
		Frequency: "yearly",  // Invalid frequency
		IsActive:  true,
		CreatedAt: time.Time{}, // Zero time
		UpdatedAt: time.Now(),
	}

	// Act
	err := income.Validate()

	// Assert
	assert.Error(t, err)
	errorMsg := err.Error()
	assert.Contains(t, errorMsg, "user ID is required")
	assert.Contains(t, errorMsg, "source is required")
	assert.Contains(t, errorMsg, "amount must be greater than 0")
	assert.Contains(t, errorMsg, "frequency must be one of: monthly, weekly, daily, one-time")
	assert.Contains(t, errorMsg, "created at is required")
}

func TestIncome_NormalizeToMonthly_MonthlyFrequency_ReturnsSameAmount(t *testing.T) {
	// Arrange
	income := Income{
		Amount:    5000.50,
		Frequency: "monthly",
	}

	// Act
	monthlyAmount := income.NormalizeToMonthly()

	// Assert
	assert.Equal(t, 5000.50, monthlyAmount)
}

func TestIncome_NormalizeToMonthly_WeeklyFrequency_ReturnsMonthlyEquivalent(t *testing.T) {
	// Arrange
	tests := []struct {
		name           string
		weeklyAmount   float64
		expectedMonthly float64
	}{
		{"round_number", 1000.0, 4333.33},
		{"with_decimals", 1250.75, 5419.58},
		{"small_amount", 100.0, 433.33},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			income := Income{
				Amount:    tt.weeklyAmount,
				Frequency: "weekly",
			}

			// Act
			monthlyAmount := income.NormalizeToMonthly()

			// Assert
			// Weekly to monthly: weekly * 52 weeks / 12 months = weekly * 4.3333
			assert.InDelta(t, tt.expectedMonthly, monthlyAmount, 0.01)
		})
	}
}

func TestIncome_NormalizeToMonthly_DailyFrequency_ReturnsMonthlyEquivalent(t *testing.T) {
	// Arrange
	tests := []struct {
		name           string
		dailyAmount    float64
		expectedMonthly float64
	}{
		{"round_number", 100.0, 3041.67}, // 100 * 365.25 / 12
		{"with_decimals", 150.75, 4587.44},
		{"small_amount", 50.0, 1520.83},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			income := Income{
				Amount:    tt.dailyAmount,
				Frequency: "daily",
			}

			// Act
			monthlyAmount := income.NormalizeToMonthly()

			// Assert
			// Daily to monthly: daily * 365.25 days / 12 months = daily * 30.4375
			assert.InDelta(t, tt.expectedMonthly, monthlyAmount, 0.01)
		})
	}
}

func TestIncome_NormalizeToMonthly_OneTimeFrequency_ReturnsZero(t *testing.T) {
	// Arrange
	income := Income{
		Amount:    1000.0,
		Frequency: "one-time",
	}

	// Act
	monthlyAmount := income.NormalizeToMonthly()

	// Assert
	// One-time income doesn't contribute to regular monthly calculations
	assert.Equal(t, 0.0, monthlyAmount)
}

func TestIncome_NormalizeToMonthly_InvalidFrequency_ReturnsZero(t *testing.T) {
	// Arrange
	income := Income{
		Amount:    1000.0,
		Frequency: "invalid",
	}

	// Act
	monthlyAmount := income.NormalizeToMonthly()

	// Assert
	assert.Equal(t, 0.0, monthlyAmount)
}

func TestIncome_NormalizeToMonthly_EdgeCases_HandlesCorrectly(t *testing.T) {
	tests := []struct {
		name           string
		amount         float64
		frequency      string
		expectedResult float64
	}{
		{"zero_amount_monthly", 0.0, "monthly", 0.0},
		{"zero_amount_weekly", 0.0, "weekly", 0.0},
		{"very_small_amount", 0.01, "weekly", 0.04},
		{"very_large_amount", 1000000.0, "daily", 30437500.0},
		{"empty_frequency", 1000.0, "", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			income := Income{
				Amount:    tt.amount,
				Frequency: tt.frequency,
			}

			// Act
			result := income.NormalizeToMonthly()

			// Assert
			assert.InDelta(t, tt.expectedResult, result, 0.01)
		})
	}
}