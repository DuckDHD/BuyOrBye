package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMedicalExpense_Validate(t *testing.T) {
	tests := []struct {
		name        string
		expense     MedicalExpense
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_doctor_visit_expense",
			expense: MedicalExpense{
				ID:               "expense-1",
				UserID:           "user-1",
				ProfileID:        "profile-1",
				Amount:           150.0,
				Category:         "doctor_visit",
				Description:      "Annual checkup with primary care physician",
				IsRecurring:      false,
				Frequency:        "one_time",
				IsCovered:        true,
				InsurancePayment: 120.0,
				OutOfPocket:      30.0,
				Date:             time.Now(),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expectError: false,
		},
		{
			name: "valid_medication_expense_recurring",
			expense: MedicalExpense{
				ID:               "expense-2",
				UserID:           "user-2",
				ProfileID:        "profile-2",
				Amount:           80.0,
				Category:         "medication",
				Description:      "Monthly diabetes medication",
				IsRecurring:      true,
				Frequency:        "monthly",
				IsCovered:        true,
				InsurancePayment: 60.0,
				OutOfPocket:      20.0,
				Date:             time.Now(),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expectError: false,
		},
		{
			name: "valid_hospital_expense",
			expense: MedicalExpense{
				ID:               "expense-3",
				UserID:           "user-3",
				ProfileID:        "profile-3",
				Amount:           5000.0,
				Category:         "hospital",
				Description:      "Emergency room visit",
				IsRecurring:      false,
				Frequency:        "one_time",
				IsCovered:        true,
				InsurancePayment: 4000.0,
				OutOfPocket:      1000.0,
				Date:             time.Now(),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expectError: false,
		},
		{
			name: "valid_lab_test_expense",
			expense: MedicalExpense{
				ID:               "expense-4",
				UserID:           "user-4",
				ProfileID:        "profile-4",
				Amount:           200.0,
				Category:         "lab_test",
				Description:      "Blood work and cholesterol screening",
				IsRecurring:      false,
				Frequency:        "one_time",
				IsCovered:        true,
				InsurancePayment: 180.0,
				OutOfPocket:      20.0,
				Date:             time.Now(),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expectError: false,
		},
		{
			name: "valid_therapy_expense",
			expense: MedicalExpense{
				ID:               "expense-5",
				UserID:           "user-5",
				ProfileID:        "profile-5",
				Amount:           120.0,
				Category:         "therapy",
				Description:      "Physical therapy session",
				IsRecurring:      true,
				Frequency:        "quarterly",
				IsCovered:        false,
				InsurancePayment: 0.0,
				OutOfPocket:      120.0,
				Date:             time.Now(),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expectError: false,
		},
		{
			name: "valid_equipment_expense",
			expense: MedicalExpense{
				ID:               "expense-6",
				UserID:           "user-6",
				ProfileID:        "profile-6",
				Amount:           300.0,
				Category:         "equipment",
				Description:      "Blood glucose monitor",
				IsRecurring:      false,
				Frequency:        "one_time",
				IsCovered:        false,
				InsurancePayment: 0.0,
				OutOfPocket:      300.0,
				Date:             time.Now(),
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			},
			expectError: false,
		},
		{
			name: "zero_amount_invalid",
			expense: MedicalExpense{
				ID:        "expense-7",
				UserID:    "user-7",
				ProfileID: "profile-7",
				Amount:    0.0,
				Category:  "doctor_visit",
				Date:      time.Now(),
			},
			expectError: true,
			errorMsg:    "amount must be positive",
		},
		{
			name: "negative_amount_invalid",
			expense: MedicalExpense{
				ID:        "expense-8",
				UserID:    "user-8",
				ProfileID: "profile-8",
				Amount:    -100.0,
				Category:  "doctor_visit",
				Date:      time.Now(),
			},
			expectError: true,
			errorMsg:    "amount must be positive",
		},
		{
			name: "invalid_category",
			expense: MedicalExpense{
				ID:        "expense-9",
				UserID:    "user-9",
				ProfileID: "profile-9",
				Amount:    100.0,
				Category:  "invalid_category",
				Date:      time.Now(),
			},
			expectError: true,
			errorMsg:    "category must be one of: doctor_visit, medication, hospital, lab_test, therapy, equipment",
		},
		{
			name: "empty_category_invalid",
			expense: MedicalExpense{
				ID:        "expense-10",
				UserID:    "user-10",
				ProfileID: "profile-10",
				Amount:    100.0,
				Category:  "",
				Date:      time.Now(),
			},
			expectError: true,
			errorMsg:    "category must be one of: doctor_visit, medication, hospital, lab_test, therapy, equipment",
		},
		{
			name: "recurring_without_frequency_invalid",
			expense: MedicalExpense{
				ID:          "expense-11",
				UserID:      "user-11",
				ProfileID:   "profile-11",
				Amount:      100.0,
				Category:    "medication",
				IsRecurring: true,
				Frequency:   "",
				Date:        time.Now(),
			},
			expectError: true,
			errorMsg:    "frequency is required for recurring expenses",
		},
		{
			name: "recurring_invalid_frequency",
			expense: MedicalExpense{
				ID:          "expense-12",
				UserID:      "user-12",
				ProfileID:   "profile-12",
				Amount:      100.0,
				Category:    "medication",
				IsRecurring: true,
				Frequency:   "invalid_frequency",
				Date:        time.Now(),
			},
			expectError: true,
			errorMsg:    "frequency must be one of: monthly, quarterly, annually",
		},
		{
			name: "non_recurring_with_frequency_valid",
			expense: MedicalExpense{
				ID:          "expense-13",
				UserID:      "user-13",
				ProfileID:   "profile-13",
				Amount:      100.0,
				Category:    "doctor_visit",
				IsRecurring: false,
				Frequency:   "one_time",
				Date:        time.Now(),
			},
			expectError: false,
		},
		{
			name: "negative_insurance_payment_invalid",
			expense: MedicalExpense{
				ID:               "expense-14",
				UserID:           "user-14",
				ProfileID:        "profile-14",
				Amount:           100.0,
				Category:         "doctor_visit",
				InsurancePayment: -50.0,
				Date:             time.Now(),
			},
			expectError: true,
			errorMsg:    "insurance payment must be non-negative",
		},
		{
			name: "insurance_payment_exceeds_amount_invalid",
			expense: MedicalExpense{
				ID:               "expense-15",
				UserID:           "user-15",
				ProfileID:        "profile-15",
				Amount:           100.0,
				Category:         "doctor_visit",
				InsurancePayment: 150.0,
				Date:             time.Now(),
			},
			expectError: true,
			errorMsg:    "insurance payment cannot exceed total amount",
		},
		{
			name: "negative_out_of_pocket_invalid",
			expense: MedicalExpense{
				ID:          "expense-16",
				UserID:      "user-16",
				ProfileID:   "profile-16",
				Amount:      100.0,
				Category:    "doctor_visit",
				OutOfPocket: -25.0,
				Date:        time.Now(),
			},
			expectError: true,
			errorMsg:    "out of pocket amount must be non-negative",
		},
		{
			name: "out_of_pocket_exceeds_amount_invalid",
			expense: MedicalExpense{
				ID:          "expense-17",
				UserID:      "user-17",
				ProfileID:   "profile-17",
				Amount:      100.0,
				Category:    "doctor_visit",
				OutOfPocket: 150.0,
				Date:        time.Now(),
			},
			expectError: true,
			errorMsg:    "out of pocket amount cannot exceed total amount",
		},
		{
			name: "future_date_invalid",
			expense: MedicalExpense{
				ID:        "expense-18",
				UserID:    "user-18",
				ProfileID: "profile-18",
				Amount:    100.0,
				Category:  "doctor_visit",
				Date:      time.Now().AddDate(0, 1, 0),
			},
			expectError: true,
			errorMsg:    "expense date cannot be in the future",
		},
		{
			name: "empty_user_id_invalid",
			expense: MedicalExpense{
				ID:        "expense-19",
				UserID:    "",
				ProfileID: "profile-19",
				Amount:    100.0,
				Category:  "doctor_visit",
				Date:      time.Now(),
			},
			expectError: true,
			errorMsg:    "user ID is required",
		},
		{
			name: "empty_profile_id_invalid",
			expense: MedicalExpense{
				ID:        "expense-20",
				UserID:    "user-20",
				ProfileID: "",
				Amount:    100.0,
				Category:  "doctor_visit",
				Date:      time.Now(),
			},
			expectError: true,
			errorMsg:    "profile ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.expense.Validate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMedicalExpense_CalculateOutOfPocket(t *testing.T) {
	tests := []struct {
		name                     string
		totalAmount              float64
		insurancePayment         float64
		expectedOutOfPocket      float64
	}{
		{
			name:                "full_insurance_coverage",
			totalAmount:         100.0,
			insurancePayment:    100.0,
			expectedOutOfPocket: 0.0,
		},
		{
			name:                "partial_insurance_coverage",
			totalAmount:         100.0,
			insurancePayment:    80.0,
			expectedOutOfPocket: 20.0,
		},
		{
			name:                "no_insurance_coverage",
			totalAmount:         100.0,
			insurancePayment:    0.0,
			expectedOutOfPocket: 100.0,
		},
		{
			name:                "high_cost_with_partial_coverage",
			totalAmount:         5000.0,
			insurancePayment:    4000.0,
			expectedOutOfPocket: 1000.0,
		},
		{
			name:                "exact_coverage",
			totalAmount:         250.75,
			insurancePayment:    200.60,
			expectedOutOfPocket: 50.15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expense := MedicalExpense{
				Amount:           tt.totalAmount,
				InsurancePayment: tt.insurancePayment,
			}

			outOfPocket := expense.CalculateOutOfPocket()
			assert.InDelta(t, tt.expectedOutOfPocket, outOfPocket, 0.01, "Out of pocket calculation should be accurate")
		})
	}
}

func TestMedicalExpense_GetAnnualizedCost(t *testing.T) {
	tests := []struct {
		name               string
		amount             float64
		isRecurring        bool
		frequency          string
		expectedAnnualCost float64
	}{
		{
			name:               "one_time_expense",
			amount:             100.0,
			isRecurring:        false,
			frequency:          "one_time",
			expectedAnnualCost: 100.0,
		},
		{
			name:               "monthly_recurring_expense",
			amount:             80.0,
			isRecurring:        true,
			frequency:          "monthly",
			expectedAnnualCost: 960.0,
		},
		{
			name:               "quarterly_recurring_expense",
			amount:             200.0,
			isRecurring:        true,
			frequency:          "quarterly",
			expectedAnnualCost: 800.0,
		},
		{
			name:               "annually_recurring_expense",
			amount:             1200.0,
			isRecurring:        true,
			frequency:          "annually",
			expectedAnnualCost: 1200.0,
		},
		{
			name:               "monthly_medication_cost",
			amount:             45.50,
			isRecurring:        true,
			frequency:          "monthly",
			expectedAnnualCost: 546.0,
		},
		{
			name:               "quarterly_therapy_cost",
			amount:             120.0,
			isRecurring:        true,
			frequency:          "quarterly",
			expectedAnnualCost: 480.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expense := MedicalExpense{
				Amount:      tt.amount,
				IsRecurring: tt.isRecurring,
				Frequency:   tt.frequency,
			}

			annualCost := expense.GetAnnualizedCost()
			assert.InDelta(t, tt.expectedAnnualCost, annualCost, 0.01, "Annualized cost calculation should be accurate")
		})
	}
}

func TestMedicalExpense_GetCoveragePercentage(t *testing.T) {
	tests := []struct {
		name                       string
		totalAmount                float64
		insurancePayment           float64
		expectedCoveragePercentage float64
	}{
		{
			name:                       "full_coverage_100_percent",
			totalAmount:                100.0,
			insurancePayment:           100.0,
			expectedCoveragePercentage: 100.0,
		},
		{
			name:                       "eighty_percent_coverage",
			totalAmount:                100.0,
			insurancePayment:           80.0,
			expectedCoveragePercentage: 80.0,
		},
		{
			name:                       "no_coverage_zero_percent",
			totalAmount:                100.0,
			insurancePayment:           0.0,
			expectedCoveragePercentage: 0.0,
		},
		{
			name:                       "fifty_percent_coverage",
			totalAmount:                200.0,
			insurancePayment:           100.0,
			expectedCoveragePercentage: 50.0,
		},
		{
			name:                       "ninety_percent_coverage",
			totalAmount:                1000.0,
			insurancePayment:           900.0,
			expectedCoveragePercentage: 90.0,
		},
		{
			name:                       "partial_coverage_75_25",
			totalAmount:                400.0,
			insurancePayment:           301.0,
			expectedCoveragePercentage: 75.25,
		},
		{
			name:                       "zero_amount_edge_case",
			totalAmount:                0.0,
			insurancePayment:           0.0,
			expectedCoveragePercentage: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expense := MedicalExpense{
				Amount:           tt.totalAmount,
				InsurancePayment: tt.insurancePayment,
			}

			coveragePercentage := expense.GetCoveragePercentage()
			
			if tt.totalAmount == 0 {
				assert.Equal(t, 0.0, coveragePercentage, "Coverage percentage should be 0 for zero amount")
			} else {
				assert.InDelta(t, tt.expectedCoveragePercentage, coveragePercentage, 0.01, "Coverage percentage calculation should be accurate")
			}
		})
	}
}

func TestMedicalExpense_IsCovered(t *testing.T) {
	tests := []struct {
		name            string
		isCovered       bool
		insurancePayment float64
		expectedResult  bool
	}{
		{
			name:            "covered_with_payment",
			isCovered:       true,
			insurancePayment: 80.0,
			expectedResult:  true,
		},
		{
			name:            "covered_no_payment",
			isCovered:       true,
			insurancePayment: 0.0,
			expectedResult:  true,
		},
		{
			name:            "not_covered_no_payment",
			isCovered:       false,
			insurancePayment: 0.0,
			expectedResult:  false,
		},
		{
			name:            "not_covered_with_payment_anomaly",
			isCovered:       false,
			insurancePayment: 50.0,
			expectedResult:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expense := MedicalExpense{
				IsCovered:        tt.isCovered,
				InsurancePayment: tt.insurancePayment,
			}

			result := expense.IsCovered
			assert.Equal(t, tt.expectedResult, result, "IsCovered flag should match expected value")
		})
	}
}

func TestMedicalExpense_GetFrequencyMultiplier(t *testing.T) {
	tests := []struct {
		name                      string
		frequency                 string
		expectedMultiplier        float64
	}{
		{
			name:               "monthly_frequency",
			frequency:          "monthly",
			expectedMultiplier: 12.0,
		},
		{
			name:               "quarterly_frequency",
			frequency:          "quarterly",
			expectedMultiplier: 4.0,
		},
		{
			name:               "annually_frequency",
			frequency:          "annually",
			expectedMultiplier: 1.0,
		},
		{
			name:               "one_time_frequency",
			frequency:          "one_time",
			expectedMultiplier: 1.0,
		},
		{
			name:               "unknown_frequency",
			frequency:          "unknown",
			expectedMultiplier: 1.0,
		},
		{
			name:               "empty_frequency",
			frequency:          "",
			expectedMultiplier: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expense := MedicalExpense{
				Frequency: tt.frequency,
			}

			multiplier := expense.GetFrequencyMultiplier()
			assert.Equal(t, tt.expectedMultiplier, multiplier, "Frequency multiplier should match expected value")
		})
	}
}

func TestMedicalExpense_IsHighCostExpense(t *testing.T) {
	tests := []struct {
		name                string
		amount              float64
		expectedHighCost    bool
	}{
		{
			name:             "low_cost_expense",
			amount:           50.0,
			expectedHighCost: false,
		},
		{
			name:             "moderate_cost_expense",
			amount:           250.0,
			expectedHighCost: false,
		},
		{
			name:             "boundary_high_cost_500",
			amount:           500.0,
			expectedHighCost: true,
		},
		{
			name:             "high_cost_expense",
			amount:           1000.0,
			expectedHighCost: true,
		},
		{
			name:             "very_high_cost_expense",
			amount:           5000.0,
			expectedHighCost: true,
		},
		{
			name:             "boundary_below_high_cost",
			amount:           499.99,
			expectedHighCost: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expense := MedicalExpense{
				Amount: tt.amount,
			}

			isHighCost := expense.IsHighCostExpense()
			assert.Equal(t, tt.expectedHighCost, isHighCost, "High cost expense determination should match expected value")
		})
	}
}