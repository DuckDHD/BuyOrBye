package types

import (
	"testing"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

func TestAddIncomeDTO_Validation_ValidData_PassesValidation(t *testing.T) {
	// Arrange
	validator := validator.New()
	dto := AddIncomeDTO{
		Source:    "Software Engineer Salary",
		Amount:    5000.00,
		Frequency: "monthly",
	}

	// Act
	err := validator.Struct(dto)

	// Assert
	assert.NoError(t, err)
}

func TestAddIncomeDTO_Validation_InvalidData_FailsValidation(t *testing.T) {
	validator := validator.New()
	
	tests := []struct {
		name     string
		dto      AddIncomeDTO
		expected string
	}{
		{
			"empty_source",
			AddIncomeDTO{Source: "", Amount: 5000.00, Frequency: "monthly"},
			"Source",
		},
		{
			"short_source",
			AddIncomeDTO{Source: "a", Amount: 5000.00, Frequency: "monthly"},
			"Source",
		},
		{
			"zero_amount",
			AddIncomeDTO{Source: "Test Source", Amount: 0.00, Frequency: "monthly"},
			"Amount",
		},
		{
			"negative_amount",
			AddIncomeDTO{Source: "Test Source", Amount: -100.00, Frequency: "monthly"},
			"Amount",
		},
		{
			"invalid_frequency",
			AddIncomeDTO{Source: "Test Source", Amount: 5000.00, Frequency: "yearly"},
			"Frequency",
		},
		{
			"empty_frequency",
			AddIncomeDTO{Source: "Test Source", Amount: 5000.00, Frequency: ""},
			"Frequency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := validator.Struct(tt.dto)

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expected)
		})
	}
}

func TestAddIncomeDTO_ToDomain_CreatesCorrectDomainObject(t *testing.T) {
	// Arrange
	dto := AddIncomeDTO{
		Source:    "Software Engineer Salary",
		Amount:    5000.00,
		Frequency: "monthly",
	}
	userID := "user-123"

	// Act
	domain := dto.ToDomain(userID)

	// Assert
	assert.Equal(t, userID, domain.UserID)
	assert.Equal(t, dto.Source, domain.Source)
	assert.Equal(t, dto.Amount, domain.Amount)
	assert.Equal(t, dto.Frequency, domain.Frequency)
	assert.True(t, domain.IsActive)
	assert.False(t, domain.CreatedAt.IsZero())
	assert.False(t, domain.UpdatedAt.IsZero())
}

func TestAddExpenseDTO_Validation_ValidData_PassesValidation(t *testing.T) {
	// Arrange
	validator := validator.New()
	dto := AddExpenseDTO{
		Category:  "housing",
		Name:      "Monthly Rent",
		Amount:    1200.00,
		Frequency: "monthly",
		IsFixed:   true,
		Priority:  1,
	}

	// Act
	err := validator.Struct(dto)

	// Assert
	assert.NoError(t, err)
}

func TestAddExpenseDTO_Validation_InvalidData_FailsValidation(t *testing.T) {
	validator := validator.New()
	
	tests := []struct {
		name     string
		dto      AddExpenseDTO
		expected string
	}{
		{
			"invalid_category",
			AddExpenseDTO{Category: "invalid", Name: "Test", Amount: 100.00, Frequency: "monthly", Priority: 1},
			"Category",
		},
		{
			"empty_name",
			AddExpenseDTO{Category: "housing", Name: "", Amount: 100.00, Frequency: "monthly", Priority: 1},
			"Name",
		},
		{
			"zero_amount",
			AddExpenseDTO{Category: "housing", Name: "Test", Amount: 0.00, Frequency: "monthly", Priority: 1},
			"Amount",
		},
		{
			"invalid_frequency",
			AddExpenseDTO{Category: "housing", Name: "Test", Amount: 100.00, Frequency: "yearly", Priority: 1},
			"Frequency",
		},
		{
			"invalid_priority_low",
			AddExpenseDTO{Category: "housing", Name: "Test", Amount: 100.00, Frequency: "monthly", Priority: 0},
			"Priority",
		},
		{
			"invalid_priority_high",
			AddExpenseDTO{Category: "housing", Name: "Test", Amount: 100.00, Frequency: "monthly", Priority: 4},
			"Priority",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := validator.Struct(tt.dto)

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expected)
		})
	}
}

func TestAddLoanDTO_Validation_ValidData_PassesValidation(t *testing.T) {
	// Arrange
	validator := validator.New()
	dto := AddLoanDTO{
		Lender:           "Chase Bank",
		Type:             "mortgage",
		PrincipalAmount:  250000.00,
		RemainingBalance: 245000.00,
		MonthlyPayment:   1266.71,
		InterestRate:     4.5,
		EndDate:          time.Date(2054, 1, 15, 0, 0, 0, 0, time.UTC),
	}

	// Act
	err := validator.Struct(dto)

	// Assert
	assert.NoError(t, err)
}

func TestAddLoanDTO_Validation_InvalidData_FailsValidation(t *testing.T) {
	validator := validator.New()
	futureDate := time.Date(2054, 1, 15, 0, 0, 0, 0, time.UTC)
	
	tests := []struct {
		name     string
		dto      AddLoanDTO
		expected string
	}{
		{
			"empty_lender",
			AddLoanDTO{Lender: "", Type: "mortgage", PrincipalAmount: 250000, RemainingBalance: 245000, MonthlyPayment: 1266, InterestRate: 4.5, EndDate: futureDate},
			"Lender",
		},
		{
			"invalid_type",
			AddLoanDTO{Lender: "Chase", Type: "invalid", PrincipalAmount: 250000, RemainingBalance: 245000, MonthlyPayment: 1266, InterestRate: 4.5, EndDate: futureDate},
			"Type",
		},
		{
			"zero_principal",
			AddLoanDTO{Lender: "Chase", Type: "mortgage", PrincipalAmount: 0, RemainingBalance: 245000, MonthlyPayment: 1266, InterestRate: 4.5, EndDate: futureDate},
			"PrincipalAmount",
		},
		{
			"negative_remaining",
			AddLoanDTO{Lender: "Chase", Type: "mortgage", PrincipalAmount: 250000, RemainingBalance: -1000, MonthlyPayment: 1266, InterestRate: 4.5, EndDate: futureDate},
			"RemainingBalance",
		},
		{
			"zero_monthly_payment",
			AddLoanDTO{Lender: "Chase", Type: "mortgage", PrincipalAmount: 250000, RemainingBalance: 245000, MonthlyPayment: 0, InterestRate: 4.5, EndDate: futureDate},
			"MonthlyPayment",
		},
		{
			"negative_interest_rate",
			AddLoanDTO{Lender: "Chase", Type: "mortgage", PrincipalAmount: 250000, RemainingBalance: 245000, MonthlyPayment: 1266, InterestRate: -1, EndDate: futureDate},
			"InterestRate",
		},
		{
			"excessive_interest_rate",
			AddLoanDTO{Lender: "Chase", Type: "mortgage", PrincipalAmount: 250000, RemainingBalance: 245000, MonthlyPayment: 1266, InterestRate: 101, EndDate: futureDate},
			"InterestRate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := validator.Struct(tt.dto)

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expected)
		})
	}
}

func TestIncomeResponseDTO_FromDomain_ConvertsCorrectly(t *testing.T) {
	// Arrange
	income := domain.Income{
		ID:        "income-123",
		UserID:    "user-456",
		Source:    "Software Engineer Salary",
		Amount:    5000.00,
		Frequency: "monthly",
		IsActive:  true,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	dto := &IncomeResponseDTO{}

	// Act
	dto.FromDomain(income)

	// Assert
	assert.Equal(t, income.ID, dto.ID)
	assert.Equal(t, income.UserID, dto.UserID)
	assert.Equal(t, income.Source, dto.Source)
	assert.Equal(t, income.Amount, dto.Amount)
	assert.Equal(t, income.Frequency, dto.Frequency)
	assert.Equal(t, income.IsActive, dto.IsActive)
	assert.Equal(t, income.CreatedAt, dto.CreatedAt)
	assert.Equal(t, income.UpdatedAt, dto.UpdatedAt)
}

func TestExpenseResponseDTO_FromDomain_ConvertsCorrectly(t *testing.T) {
	// Arrange
	expense := domain.Expense{
		ID:        "expense-123",
		UserID:    "user-456",
		Category:  "housing",
		Name:      "Monthly Rent",
		Amount:    1200.00,
		Frequency: "monthly",
		IsFixed:   true,
		Priority:  1,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	dto := &ExpenseResponseDTO{}

	// Act
	dto.FromDomain(expense)

	// Assert
	assert.Equal(t, expense.ID, dto.ID)
	assert.Equal(t, expense.UserID, dto.UserID)
	assert.Equal(t, expense.Category, dto.Category)
	assert.Equal(t, expense.Name, dto.Name)
	assert.Equal(t, expense.Amount, dto.Amount)
	assert.Equal(t, expense.Frequency, dto.Frequency)
	assert.Equal(t, expense.IsFixed, dto.IsFixed)
	assert.Equal(t, expense.Priority, dto.Priority)
	assert.Equal(t, expense.CreatedAt, dto.CreatedAt)
	assert.Equal(t, expense.UpdatedAt, dto.UpdatedAt)
}

func TestLoanResponseDTO_FromDomain_ConvertsCorrectly(t *testing.T) {
	// Arrange
	loan := domain.Loan{
		ID:               "loan-123",
		UserID:           "user-456",
		Lender:           "Chase Bank",
		Type:             "mortgage",
		PrincipalAmount:  250000.00,
		RemainingBalance: 245000.00,
		MonthlyPayment:   1266.71,
		InterestRate:     4.5,
		EndDate:          time.Date(2054, 1, 15, 0, 0, 0, 0, time.UTC),
		CreatedAt:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	dto := &LoanResponseDTO{}

	// Act
	dto.FromDomain(loan)

	// Assert
	assert.Equal(t, loan.ID, dto.ID)
	assert.Equal(t, loan.UserID, dto.UserID)
	assert.Equal(t, loan.Lender, dto.Lender)
	assert.Equal(t, loan.Type, dto.Type)
	assert.Equal(t, loan.PrincipalAmount, dto.PrincipalAmount)
	assert.Equal(t, loan.RemainingBalance, dto.RemainingBalance)
	assert.Equal(t, loan.MonthlyPayment, dto.MonthlyPayment)
	assert.Equal(t, loan.InterestRate, dto.InterestRate)
	assert.Equal(t, loan.EndDate, dto.EndDate)
	assert.Equal(t, loan.CreatedAt, dto.CreatedAt)
	assert.Equal(t, loan.UpdatedAt, dto.UpdatedAt)
}

func TestUpdateIncomeDTO_ApplyUpdates_UpdatesFields(t *testing.T) {
	// Arrange
	income := &domain.Income{
		Source:    "Old Source",
		Amount:    4000.00,
		Frequency: "monthly",
		UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	
	newSource := "New Source"
	newAmount := 5000.00
	newFrequency := "weekly"
	
	dto := UpdateIncomeDTO{
		Source:    &newSource,
		Amount:    &newAmount,
		Frequency: &newFrequency,
	}

	// Act
	dto.ApplyUpdates(income)

	// Assert
	assert.Equal(t, newSource, income.Source)
	assert.Equal(t, newAmount, income.Amount)
	assert.Equal(t, newFrequency, income.Frequency)
	assert.True(t, income.UpdatedAt.After(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
}

func TestUpdateIncomeDTO_ApplyUpdates_PartialUpdate_OnlyUpdatesProvidedFields(t *testing.T) {
	// Arrange
	originalSource := "Original Source"
	originalFrequency := "monthly"
	income := &domain.Income{
		Source:    originalSource,
		Amount:    4000.00,
		Frequency: originalFrequency,
		UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	
	newAmount := 5000.00
	dto := UpdateIncomeDTO{
		Amount: &newAmount,
		// Source and Frequency are nil, should not be updated
	}

	// Act
	dto.ApplyUpdates(income)

	// Assert
	assert.Equal(t, originalSource, income.Source)    // Should remain unchanged
	assert.Equal(t, newAmount, income.Amount)         // Should be updated
	assert.Equal(t, originalFrequency, income.Frequency) // Should remain unchanged
	assert.True(t, income.UpdatedAt.After(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
}

func TestUpdateExpenseDTO_ApplyUpdates_UpdatesAllFields(t *testing.T) {
	// Arrange
	expense := &domain.Expense{
		Category:  "housing",
		Name:      "Old Rent",
		Amount:    1000.00,
		Frequency: "monthly",
		IsFixed:   true,
		Priority:  1,
		UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	
	newCategory := "utilities"
	newName := "New Expense"
	newAmount := 150.00
	newFrequency := "weekly"
	newIsFixed := false
	newPriority := 2
	
	dto := UpdateExpenseDTO{
		Category:  &newCategory,
		Name:      &newName,
		Amount:    &newAmount,
		Frequency: &newFrequency,
		IsFixed:   &newIsFixed,
		Priority:  &newPriority,
	}

	// Act
	dto.ApplyUpdates(expense)

	// Assert
	assert.Equal(t, newCategory, expense.Category)
	assert.Equal(t, newName, expense.Name)
	assert.Equal(t, newAmount, expense.Amount)
	assert.Equal(t, newFrequency, expense.Frequency)
	assert.Equal(t, newIsFixed, expense.IsFixed)
	assert.Equal(t, newPriority, expense.Priority)
	assert.True(t, expense.UpdatedAt.After(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
}

func TestUpdateLoanDTO_ApplyUpdates_UpdatesAllFields(t *testing.T) {
	// Arrange
	loan := &domain.Loan{
		Lender:           "Old Bank",
		Type:             "mortgage",
		PrincipalAmount:  200000.00,
		RemainingBalance: 180000.00,
		MonthlyPayment:   1200.00,
		InterestRate:     5.0,
		EndDate:          time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	
	newLender := "New Bank"
	newType := "auto"
	newPrincipal := 250000.00
	newBalance := 230000.00
	newPayment := 1500.00
	newRate := 4.0
	newEndDate := time.Date(2055, 1, 1, 0, 0, 0, 0, time.UTC)
	
	dto := UpdateLoanDTO{
		Lender:           &newLender,
		Type:             &newType,
		PrincipalAmount:  &newPrincipal,
		RemainingBalance: &newBalance,
		MonthlyPayment:   &newPayment,
		InterestRate:     &newRate,
		EndDate:          &newEndDate,
	}

	// Act
	dto.ApplyUpdates(loan)

	// Assert
	assert.Equal(t, newLender, loan.Lender)
	assert.Equal(t, newType, loan.Type)
	assert.Equal(t, newPrincipal, loan.PrincipalAmount)
	assert.Equal(t, newBalance, loan.RemainingBalance)
	assert.Equal(t, newPayment, loan.MonthlyPayment)
	assert.Equal(t, newRate, loan.InterestRate)
	assert.Equal(t, newEndDate, loan.EndDate)
	assert.True(t, loan.UpdatedAt.After(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)))
}