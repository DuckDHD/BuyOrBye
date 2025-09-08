package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExpense_Validate_AllFieldsValid_ReturnsNil(t *testing.T) {
	// Arrange
	expense := Expense{
		ID:        "expense-123",
		UserID:    "user-123",
		Category:  "housing",
		Name:      "Monthly Rent",
		Amount:    1200.00,
		Frequency: "monthly",
		IsFixed:   true,
		Priority:  1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Act
	err := expense.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestExpense_Validate_ZeroAmount_ReturnsError(t *testing.T) {
	// Arrange
	expense := Expense{
		ID:        "expense-123",
		UserID:    "user-123",
		Category:  "food",
		Name:      "Groceries",
		Amount:    0.0, // Invalid amount
		Frequency: "weekly",
		IsFixed:   false,
		Priority:  1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Act
	err := expense.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be greater than 0")
}

func TestExpense_Validate_NegativeAmount_ReturnsError(t *testing.T) {
	// Arrange
	expense := Expense{
		ID:        "expense-123",
		UserID:    "user-123",
		Category:  "food",
		Name:      "Groceries",
		Amount:    -50.0, // Invalid negative amount
		Frequency: "weekly",
		IsFixed:   false,
		Priority:  1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Act
	err := expense.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "amount must be greater than 0")
}

func TestExpense_Validate_EmptyName_ReturnsError(t *testing.T) {
	// Arrange
	expense := Expense{
		ID:        "expense-123",
		UserID:    "user-123",
		Category:  "transport",
		Name:      "", // Empty name
		Amount:    100.00,
		Frequency: "monthly",
		IsFixed:   false,
		Priority:  2,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Act
	err := expense.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestExpense_Validate_InvalidCategory_ReturnsError(t *testing.T) {
	// Arrange
	invalidCategories := []string{
		"invalid",
		"rent", // Should be "housing"
		"car",  // Should be "transport"
		"",
		"FOOD", // Case sensitive
		"misc", // Should be "other"
	}

	for _, category := range invalidCategories {
		t.Run("category_"+category, func(t *testing.T) {
			expense := Expense{
				ID:        "expense-123",
				UserID:    "user-123",
				Category:  category,
				Name:      "Test Expense",
				Amount:    100.00,
				Frequency: "monthly",
				IsFixed:   false,
				Priority:  2,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Act
			err := expense.Validate()

			// Assert
			assert.Error(t, err)
			if category == "" {
				assert.Contains(t, err.Error(), "category is required")
			} else {
				assert.Contains(t, err.Error(), "category must be one of: housing, food, transport, entertainment, utilities, other")
			}
		})
	}
}

func TestExpense_Validate_ValidCategories_ReturnsNil(t *testing.T) {
	// Arrange
	validCategories := []string{
		"housing",
		"food",
		"transport",
		"entertainment",
		"utilities",
		"other",
	}

	for _, category := range validCategories {
		t.Run("valid_category_"+category, func(t *testing.T) {
			expense := Expense{
				ID:        "expense-123",
				UserID:    "user-123",
				Category:  category,
				Name:      "Test Expense",
				Amount:    100.00,
				Frequency: "monthly",
				IsFixed:   false,
				Priority:  2,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Act
			err := expense.Validate()

			// Assert
			assert.NoError(t, err)
		})
	}
}

func TestExpense_Validate_InvalidFrequency_ReturnsError(t *testing.T) {
	// Arrange
	invalidFrequencies := []string{
		"yearly",
		"biweekly",
		"one-time", // Not valid for expenses
		"",
		"MONTHLY", // Case sensitive
	}

	for _, frequency := range invalidFrequencies {
		t.Run("frequency_"+frequency, func(t *testing.T) {
			expense := Expense{
				ID:        "expense-123",
				UserID:    "user-123",
				Category:  "food",
				Name:      "Test Expense",
				Amount:    100.00,
				Frequency: frequency,
				IsFixed:   false,
				Priority:  2,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Act
			err := expense.Validate()

			// Assert
			assert.Error(t, err)
			if frequency == "" {
				assert.Contains(t, err.Error(), "frequency is required")
			} else {
				assert.Contains(t, err.Error(), "frequency must be one of: monthly, weekly, daily")
			}
		})
	}
}

func TestExpense_Validate_ValidFrequencies_ReturnsNil(t *testing.T) {
	// Arrange
	validFrequencies := []string{
		"monthly",
		"weekly",
		"daily",
	}

	for _, frequency := range validFrequencies {
		t.Run("valid_frequency_"+frequency, func(t *testing.T) {
			expense := Expense{
				ID:        "expense-123",
				UserID:    "user-123",
				Category:  "food",
				Name:      "Test Expense",
				Amount:    100.00,
				Frequency: frequency,
				IsFixed:   false,
				Priority:  2,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Act
			err := expense.Validate()

			// Assert
			assert.NoError(t, err)
		})
	}
}

func TestExpense_Validate_InvalidPriority_ReturnsError(t *testing.T) {
	// Arrange
	invalidPriorities := []int{0, 4, 5, -1, 10}

	for _, priority := range invalidPriorities {
		t.Run("priority_"+string(rune(priority+'0')), func(t *testing.T) {
			expense := Expense{
				ID:        "expense-123",
				UserID:    "user-123",
				Category:  "food",
				Name:      "Test Expense",
				Amount:    100.00,
				Frequency: "monthly",
				IsFixed:   false,
				Priority:  priority,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Act
			err := expense.Validate()

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "priority must be between 1 and 3")
		})
	}
}

func TestExpense_Validate_ValidPriorities_ReturnsNil(t *testing.T) {
	// Arrange
	validPriorities := []int{1, 2, 3}

	for _, priority := range validPriorities {
		t.Run("valid_priority_"+string(rune(priority+'0')), func(t *testing.T) {
			expense := Expense{
				ID:        "expense-123",
				UserID:    "user-123",
				Category:  "food",
				Name:      "Test Expense",
				Amount:    100.00,
				Frequency: "monthly",
				IsFixed:   false,
				Priority:  priority,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// Act
			err := expense.Validate()

			// Assert
			assert.NoError(t, err)
		})
	}
}

func TestExpense_Validate_EmptyUserID_ReturnsError(t *testing.T) {
	// Arrange
	expense := Expense{
		ID:        "expense-123",
		UserID:    "", // Empty user ID
		Category:  "food",
		Name:      "Test Expense",
		Amount:    100.00,
		Frequency: "monthly",
		IsFixed:   false,
		Priority:  2,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Act
	err := expense.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID is required")
}

func TestExpense_Validate_MultipleErrors_ReturnsAllErrors(t *testing.T) {
	// Arrange
	expense := Expense{
		ID:        "expense-123",
		UserID:    "",         // Missing user ID
		Category:  "invalid",  // Invalid category
		Name:      "",         // Missing name
		Amount:    -100.0,     // Invalid amount
		Frequency: "yearly",   // Invalid frequency
		Priority:  5,          // Invalid priority
		IsFixed:   false,
		CreatedAt: time.Time{}, // Zero time
		UpdatedAt: time.Now(),
	}

	// Act
	err := expense.Validate()

	// Assert
	assert.Error(t, err)
	errorMsg := err.Error()
	assert.Contains(t, errorMsg, "user ID is required")
	assert.Contains(t, errorMsg, "category must be one of: housing, food, transport, entertainment, utilities, other")
	assert.Contains(t, errorMsg, "name is required")
	assert.Contains(t, errorMsg, "amount must be greater than 0")
	assert.Contains(t, errorMsg, "frequency must be one of: monthly, weekly, daily")
	assert.Contains(t, errorMsg, "priority must be between 1 and 3")
	assert.Contains(t, errorMsg, "created at is required")
}

func TestExpense_NormalizeToMonthly_MonthlyFrequency_ReturnsSameAmount(t *testing.T) {
	// Arrange
	expense := Expense{
		Amount:    800.00,
		Frequency: "monthly",
	}

	// Act
	monthlyAmount := expense.NormalizeToMonthly()

	// Assert
	assert.Equal(t, 800.00, monthlyAmount)
}

func TestExpense_NormalizeToMonthly_WeeklyFrequency_ReturnsMonthlyEquivalent(t *testing.T) {
	// Arrange
	tests := []struct {
		name           string
		weeklyAmount   float64
		expectedMonthly float64
	}{
		{"groceries", 150.0, 650.00}, // 150 * 52 / 12 = 650
		{"gas", 75.50, 326.83},
		{"dining_out", 100.0, 433.33},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expense := Expense{
				Amount:    tt.weeklyAmount,
				Frequency: "weekly",
			}

			// Act
			monthlyAmount := expense.NormalizeToMonthly()

			// Assert
			assert.InDelta(t, tt.expectedMonthly, monthlyAmount, 0.01)
		})
	}
}

func TestExpense_NormalizeToMonthly_DailyFrequency_ReturnsMonthlyEquivalent(t *testing.T) {
	// Arrange
	tests := []struct {
		name           string
		dailyAmount    float64
		expectedMonthly float64
	}{
		{"coffee", 5.0, 152.08},    // 5 * 365.25 / 12
		{"parking", 10.0, 304.17}, // 10 * 365.25 / 12
		{"lunch", 12.50, 380.21},  // 12.50 * 365.25 / 12
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expense := Expense{
				Amount:    tt.dailyAmount,
				Frequency: "daily",
			}

			// Act
			monthlyAmount := expense.NormalizeToMonthly()

			// Assert
			assert.InDelta(t, tt.expectedMonthly, monthlyAmount, 0.01)
		})
	}
}

func TestExpense_NormalizeToMonthly_InvalidFrequency_ReturnsZero(t *testing.T) {
	// Arrange
	expense := Expense{
		Amount:    100.0,
		Frequency: "invalid",
	}

	// Act
	monthlyAmount := expense.NormalizeToMonthly()

	// Assert
	assert.Equal(t, 0.0, monthlyAmount)
}

func TestExpense_NormalizeToMonthly_EdgeCases_HandlesCorrectly(t *testing.T) {
	tests := []struct {
		name           string
		amount         float64
		frequency      string
		expectedResult float64
	}{
		{"zero_amount_monthly", 0.0, "monthly", 0.0},
		{"zero_amount_weekly", 0.0, "weekly", 0.0},
		{"very_small_amount", 0.01, "daily", 0.30},
		{"very_large_amount", 10000.0, "weekly", 43333.33},
		{"empty_frequency", 1000.0, "", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expense := Expense{
				Amount:    tt.amount,
				Frequency: tt.frequency,
			}

			// Act
			result := expense.NormalizeToMonthly()

			// Assert
			assert.InDelta(t, tt.expectedResult, result, 0.01)
		})
	}
}

func TestExpense_GetCategoryDisplayName_ReturnsCorrectDisplayNames(t *testing.T) {
	tests := []struct {
		category    string
		expected    string
	}{
		{"housing", "Housing"},
		{"food", "Food & Groceries"},
		{"transport", "Transportation"},
		{"entertainment", "Entertainment"},
		{"utilities", "Utilities"},
		{"other", "Other"},
		{"invalid", "Other"}, // Default fallback
	}

	for _, tt := range tests {
		t.Run("category_"+tt.category, func(t *testing.T) {
			expense := Expense{
				Category: tt.category,
			}

			// Act
			displayName := expense.GetCategoryDisplayName()

			// Assert
			assert.Equal(t, tt.expected, displayName)
		})
	}
}

func TestExpense_GetPriorityName_ReturnsCorrectPriorityNames(t *testing.T) {
	tests := []struct {
		priority int
		expected string
	}{
		{1, "Essential"},
		{2, "Important"},
		{3, "Nice-to-have"},
		{0, "Unknown"}, // Invalid priority
		{4, "Unknown"}, // Invalid priority
	}

	for _, tt := range tests {
		t.Run("priority_"+string(rune(tt.priority+'0')), func(t *testing.T) {
			expense := Expense{
				Priority: tt.priority,
			}

			// Act
			priorityName := expense.GetPriorityName()

			// Assert
			assert.Equal(t, tt.expected, priorityName)
		})
	}
}