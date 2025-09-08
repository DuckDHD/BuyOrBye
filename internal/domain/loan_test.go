package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoan_Validate_AllFieldsValid_ReturnsNil(t *testing.T) {
	// Arrange
	futureDate := time.Now().AddDate(2, 0, 0) // 2 years from now
	loan := Loan{
		ID:               "loan-123",
		UserID:           "user-123",
		Lender:           "Chase Bank",
		Type:             "mortgage",
		PrincipalAmount:  250000.00,
		RemainingBalance: 200000.00,
		MonthlyPayment:   1500.00,
		InterestRate:     4.5,
		EndDate:          futureDate,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Act
	err := loan.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestLoan_Validate_ZeroPrincipalAmount_ReturnsError(t *testing.T) {
	// Arrange
	futureDate := time.Now().AddDate(1, 0, 0)
	loan := Loan{
		ID:               "loan-123",
		UserID:           "user-123",
		Lender:           "Chase Bank",
		Type:             "auto",
		PrincipalAmount:  0.0, // Invalid amount
		RemainingBalance: 15000.00,
		MonthlyPayment:   300.00,
		InterestRate:     5.5,
		EndDate:          futureDate,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Act
	err := loan.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "principal amount must be greater than 0")
}

func TestLoan_Validate_NegativePrincipalAmount_ReturnsError(t *testing.T) {
	// Arrange
	futureDate := time.Now().AddDate(1, 0, 0)
	loan := Loan{
		ID:               "loan-123",
		UserID:           "user-123",
		Lender:           "Chase Bank",
		Type:             "personal",
		PrincipalAmount:  -5000.0, // Invalid negative amount
		RemainingBalance: 3000.00,
		MonthlyPayment:   150.00,
		InterestRate:     8.0,
		EndDate:          futureDate,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Act
	err := loan.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "principal amount must be greater than 0")
}

func TestLoan_Validate_ZeroRemainingBalance_ReturnsError(t *testing.T) {
	// Arrange
	futureDate := time.Now().AddDate(1, 0, 0)
	loan := Loan{
		ID:               "loan-123",
		UserID:           "user-123",
		Lender:           "Student Loan Corp",
		Type:             "student",
		PrincipalAmount:  30000.00,
		RemainingBalance: 0.0, // Invalid - should be greater than 0 for active loans
		MonthlyPayment:   300.00,
		InterestRate:     6.0,
		EndDate:          futureDate,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Act
	err := loan.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "remaining balance must be greater than 0")
}

func TestLoan_Validate_ZeroMonthlyPayment_ReturnsError(t *testing.T) {
	// Arrange
	futureDate := time.Now().AddDate(1, 0, 0)
	loan := Loan{
		ID:               "loan-123",
		UserID:           "user-123",
		Lender:           "Auto Finance",
		Type:             "auto",
		PrincipalAmount:  25000.00,
		RemainingBalance: 20000.00,
		MonthlyPayment:   0.0, // Invalid payment
		InterestRate:     4.0,
		EndDate:          futureDate,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Act
	err := loan.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "monthly payment must be greater than 0")
}

func TestLoan_Validate_NegativeInterestRate_ReturnsError(t *testing.T) {
	// Arrange
	futureDate := time.Now().AddDate(1, 0, 0)
	loan := Loan{
		ID:               "loan-123",
		UserID:           "user-123",
		Lender:           "Credit Union",
		Type:             "personal",
		PrincipalAmount:  10000.00,
		RemainingBalance: 8000.00,
		MonthlyPayment:   200.00,
		InterestRate:     -2.5, // Invalid negative rate
		EndDate:          futureDate,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Act
	err := loan.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interest rate must be between 0 and 100")
}

func TestLoan_Validate_ExcessiveInterestRate_ReturnsError(t *testing.T) {
	// Arrange
	futureDate := time.Now().AddDate(1, 0, 0)
	loan := Loan{
		ID:               "loan-123",
		UserID:           "user-123",
		Lender:           "Payday Loans Inc",
		Type:             "personal",
		PrincipalAmount:  1000.00,
		RemainingBalance: 900.00,
		MonthlyPayment:   150.00,
		InterestRate:     150.0, // Invalid - over 100%
		EndDate:          futureDate,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Act
	err := loan.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "interest rate must be between 0 and 100")
}

func TestLoan_Validate_InvalidType_ReturnsError(t *testing.T) {
	// Arrange
	invalidTypes := []string{
		"credit-card",
		"business",
		"payday",
		"",
		"MORTGAGE", // Case sensitive
		"home",     // Should be "mortgage"
	}

	futureDate := time.Now().AddDate(1, 0, 0)

	for _, loanType := range invalidTypes {
		t.Run("type_"+loanType, func(t *testing.T) {
			loan := Loan{
				ID:               "loan-123",
				UserID:           "user-123",
				Lender:           "Test Bank",
				Type:             loanType,
				PrincipalAmount:  10000.00,
				RemainingBalance: 8000.00,
				MonthlyPayment:   200.00,
				InterestRate:     5.0,
				EndDate:          futureDate,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}

			// Act
			err := loan.Validate()

			// Assert
			assert.Error(t, err)
			if loanType == "" {
				assert.Contains(t, err.Error(), "loan type is required")
			} else {
				assert.Contains(t, err.Error(), "loan type must be one of: mortgage, auto, personal, student")
			}
		})
	}
}

func TestLoan_Validate_ValidTypes_ReturnsNil(t *testing.T) {
	// Arrange
	validTypes := []string{
		"mortgage",
		"auto",
		"personal", 
		"student",
	}

	futureDate := time.Now().AddDate(1, 0, 0)

	for _, loanType := range validTypes {
		t.Run("valid_type_"+loanType, func(t *testing.T) {
			loan := Loan{
				ID:               "loan-123",
				UserID:           "user-123",
				Lender:           "Test Bank",
				Type:             loanType,
				PrincipalAmount:  10000.00,
				RemainingBalance: 8000.00,
				MonthlyPayment:   200.00,
				InterestRate:     5.0,
				EndDate:          futureDate,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}

			// Act
			err := loan.Validate()

			// Assert
			assert.NoError(t, err)
		})
	}
}

func TestLoan_Validate_EmptyLender_ReturnsError(t *testing.T) {
	// Arrange
	futureDate := time.Now().AddDate(1, 0, 0)
	loan := Loan{
		ID:               "loan-123",
		UserID:           "user-123",
		Lender:           "", // Empty lender
		Type:             "auto",
		PrincipalAmount:  15000.00,
		RemainingBalance: 12000.00,
		MonthlyPayment:   300.00,
		InterestRate:     4.5,
		EndDate:          futureDate,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Act
	err := loan.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lender is required")
}

func TestLoan_Validate_EmptyUserID_ReturnsError(t *testing.T) {
	// Arrange
	futureDate := time.Now().AddDate(1, 0, 0)
	loan := Loan{
		ID:               "loan-123",
		UserID:           "", // Empty user ID
		Lender:           "Test Bank",
		Type:             "personal",
		PrincipalAmount:  5000.00,
		RemainingBalance: 4000.00,
		MonthlyPayment:   150.00,
		InterestRate:     7.0,
		EndDate:          futureDate,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Act
	err := loan.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user ID is required")
}

func TestLoan_Validate_EndDateInPast_ReturnsError(t *testing.T) {
	// Arrange
	pastDate := time.Now().AddDate(-1, 0, 0) // 1 year ago
	loan := Loan{
		ID:               "loan-123",
		UserID:           "user-123",
		Lender:           "Old Bank",
		Type:             "student",
		PrincipalAmount:  20000.00,
		RemainingBalance: 15000.00,
		MonthlyPayment:   250.00,
		InterestRate:     6.0,
		EndDate:          pastDate,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Act
	err := loan.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "end date must be in the future")
}

func TestLoan_Validate_RemainingBalanceExceedsPrincipal_ReturnsError(t *testing.T) {
	// Arrange
	futureDate := time.Now().AddDate(1, 0, 0)
	loan := Loan{
		ID:               "loan-123",
		UserID:           "user-123",
		Lender:           "Test Bank",
		Type:             "personal",
		PrincipalAmount:  10000.00,
		RemainingBalance: 15000.00, // Exceeds principal - invalid
		MonthlyPayment:   200.00,
		InterestRate:     5.0,
		EndDate:          futureDate,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Act
	err := loan.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "remaining balance cannot exceed principal amount")
}

func TestLoan_Validate_MultipleErrors_ReturnsAllErrors(t *testing.T) {
	// Arrange
	pastDate := time.Now().AddDate(-1, 0, 0)
	loan := Loan{
		ID:               "loan-123",
		UserID:           "",        // Missing user ID
		Lender:           "",        // Missing lender
		Type:             "invalid", // Invalid type
		PrincipalAmount:  -1000.0,  // Invalid amount
		RemainingBalance: 0.0,      // Invalid balance
		MonthlyPayment:   -50.0,    // Invalid payment
		InterestRate:     150.0,    // Invalid interest rate
		EndDate:          pastDate, // Past date
		CreatedAt:        time.Time{}, // Zero time
		UpdatedAt:        time.Now(),
	}

	// Act
	err := loan.Validate()

	// Assert
	assert.Error(t, err)
	errorMsg := err.Error()
	assert.Contains(t, errorMsg, "user ID is required")
	assert.Contains(t, errorMsg, "lender is required")
	assert.Contains(t, errorMsg, "loan type must be one of: mortgage, auto, personal, student")
	assert.Contains(t, errorMsg, "principal amount must be greater than 0")
	assert.Contains(t, errorMsg, "remaining balance must be greater than 0")
	assert.Contains(t, errorMsg, "monthly payment must be greater than 0")
	assert.Contains(t, errorMsg, "interest rate must be between 0 and 100")
	assert.Contains(t, errorMsg, "end date must be in the future")
	assert.Contains(t, errorMsg, "created at is required")
}

func TestLoan_CalculateProgress_ReturnsCorrectPercentage(t *testing.T) {
	// Arrange
	tests := []struct {
		name             string
		principalAmount  float64
		remainingBalance float64
		expectedProgress float64
	}{
		{"half_paid", 10000.0, 5000.0, 50.0},
		{"quarter_paid", 20000.0, 15000.0, 25.0},
		{"almost_done", 5000.0, 500.0, 90.0},
		{"just_started", 100000.0, 95000.0, 5.0},
		{"fully_paid", 15000.0, 0.0, 100.0}, // Edge case - should not occur in validation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loan := Loan{
				PrincipalAmount:  tt.principalAmount,
				RemainingBalance: tt.remainingBalance,
			}

			// Act
			progress := loan.CalculateProgress()

			// Assert
			assert.InDelta(t, tt.expectedProgress, progress, 0.01)
		})
	}
}

func TestLoan_CalculateProgress_ZeroPrincipal_ReturnsZero(t *testing.T) {
	// Arrange
	loan := Loan{
		PrincipalAmount:  0.0, // Edge case
		RemainingBalance: 5000.0,
	}

	// Act
	progress := loan.CalculateProgress()

	// Assert
	assert.Equal(t, 0.0, progress)
}

func TestLoan_GetTypeDisplayName_ReturnsCorrectDisplayNames(t *testing.T) {
	tests := []struct {
		loanType string
		expected string
	}{
		{"mortgage", "Mortgage"},
		{"auto", "Auto Loan"},
		{"personal", "Personal Loan"},
		{"student", "Student Loan"},
		{"invalid", "Other"}, // Default fallback
	}

	for _, tt := range tests {
		t.Run("type_"+tt.loanType, func(t *testing.T) {
			loan := Loan{
				Type: tt.loanType,
			}

			// Act
			displayName := loan.GetTypeDisplayName()

			// Assert
			assert.Equal(t, tt.expected, displayName)
		})
	}
}

func TestLoan_CalculateMonthsRemaining_ReturnsCorrectMonths(t *testing.T) {
	// Arrange
	tests := []struct {
		name               string
		remainingBalance   float64
		monthlyPayment     float64
		interestRate       float64
		expectedMonths     int
		expectError        bool
	}{
		{"simple_no_interest", 1200.0, 100.0, 0.0, 12, false},
		{"with_interest", 10000.0, 500.0, 5.0, 21, false}, // Approximate
		{"zero_payment", 1000.0, 0.0, 5.0, 0, true}, // Should error
		{"zero_balance", 0.0, 100.0, 5.0, 0, false}, // Paid off
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loan := Loan{
				RemainingBalance: tt.remainingBalance,
				MonthlyPayment:   tt.monthlyPayment,
				InterestRate:     tt.interestRate,
			}

			// Act
			months, err := loan.CalculateMonthsRemaining()

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expectedMonths, months, 2) // Allow some variance for interest calculations
			}
		})
	}
}

func TestLoan_IsHighInterestRate_ReturnsCorrectResult(t *testing.T) {
	tests := []struct {
		name           string
		interestRate   float64
		loanType       string
		expectedResult bool
	}{
		{"mortgage_low_rate", 3.5, "mortgage", false},
		{"mortgage_high_rate", 7.0, "mortgage", true}, // Above 6% threshold
		{"auto_low_rate", 4.0, "auto", false},
		{"auto_high_rate", 9.0, "auto", true}, // Above 8% threshold
		{"personal_moderate_rate", 10.0, "personal", false},
		{"personal_high_rate", 16.0, "personal", true}, // Above 15% threshold
		{"student_low_rate", 5.0, "student", false},
		{"student_high_rate", 8.0, "student", true}, // Above 7% threshold
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loan := Loan{
				Type:         tt.loanType,
				InterestRate: tt.interestRate,
			}

			// Act
			isHigh := loan.IsHighInterestRate()

			// Assert
			assert.Equal(t, tt.expectedResult, isHigh)
		})
	}
}