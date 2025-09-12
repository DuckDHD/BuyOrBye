package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsurancePolicy_Validate(t *testing.T) {
	tests := []struct {
		name        string
		policy      InsurancePolicy
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_health_insurance_policy",
			policy: InsurancePolicy{
				ID:                  "policy-1",
				UserID:              "user-1",
				ProfileID:           "profile-1",
				Provider:            "Blue Cross Blue Shield",
				PolicyNumber:        "BCBS123456789",
				Type:                "health",
				MonthlyPremium:      300.0,
				Deductible:          2000.0,
				DeductibleMet:       500.0,
				OutOfPocketMax:      8000.0,
				OutOfPocketCurrent:  1200.0,
				CoveragePercentage:  80.0,
				StartDate:           time.Now().AddDate(-1, 0, 0),
				EndDate:             time.Now().AddDate(0, 6, 0),
				IsActive:            true,
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
			},
			expectError: false,
		},
		{
			name: "valid_dental_insurance_policy",
			policy: InsurancePolicy{
				ID:                  "policy-2",
				UserID:              "user-2",
				ProfileID:           "profile-2",
				Provider:            "Delta Dental",
				PolicyNumber:        "DD987654321",
				Type:                "dental",
				MonthlyPremium:      50.0,
				Deductible:          100.0,
				DeductibleMet:       25.0,
				OutOfPocketMax:      1500.0,
				OutOfPocketCurrent:  150.0,
				CoveragePercentage:  70.0,
				StartDate:           time.Now().AddDate(-2, 0, 0),
				EndDate:             time.Now().AddDate(0, 10, 0),
				IsActive:            true,
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
			},
			expectError: false,
		},
		{
			name: "valid_vision_insurance_policy",
			policy: InsurancePolicy{
				ID:                  "policy-3",
				UserID:              "user-3",
				ProfileID:           "profile-3",
				Provider:            "VSP Vision Care",
				PolicyNumber:        "VSP555666777",
				Type:                "vision",
				MonthlyPremium:      25.0,
				Deductible:          0.0,
				DeductibleMet:       0.0,
				OutOfPocketMax:      500.0,
				OutOfPocketCurrent:  100.0,
				CoveragePercentage:  90.0,
				StartDate:           time.Now().AddDate(0, -3, 0),
				EndDate:             time.Now().AddDate(1, 0, 0),
				IsActive:            true,
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
			},
			expectError: false,
		},
		{
			name: "valid_comprehensive_insurance_policy",
			policy: InsurancePolicy{
				ID:                  "policy-4",
				UserID:              "user-4",
				ProfileID:           "profile-4",
				Provider:            "Aetna Comprehensive",
				PolicyNumber:        "AETNA111222333",
				Type:                "comprehensive",
				MonthlyPremium:      450.0,
				Deductible:          1500.0,
				DeductibleMet:       800.0,
				OutOfPocketMax:      6000.0,
				OutOfPocketCurrent:  2000.0,
				CoveragePercentage:  85.0,
				StartDate:           time.Now().AddDate(-1, -6, 0),
				EndDate:             time.Now().AddDate(0, 6, 0),
				IsActive:            true,
				CreatedAt:           time.Now(),
				UpdatedAt:           time.Now(),
			},
			expectError: false,
		},
		{
			name: "empty_provider_invalid",
			policy: InsurancePolicy{
				ID:                 "policy-5",
				UserID:             "user-5",
				ProfileID:          "profile-5",
				Provider:           "",
				PolicyNumber:       "POL123456",
				Type:               "health",
				MonthlyPremium:     300.0,
				Deductible:         2000.0,
				OutOfPocketMax:     8000.0,
				CoveragePercentage: 80.0,
				StartDate:          time.Now().AddDate(-1, 0, 0),
				EndDate:            time.Now().AddDate(0, 6, 0),
			},
			expectError: true,
			errorMsg:    "provider is required",
		},
		{
			name: "empty_policy_number_invalid",
			policy: InsurancePolicy{
				ID:                 "policy-6",
				UserID:             "user-6",
				ProfileID:          "profile-6",
				Provider:           "Some Insurance",
				PolicyNumber:       "",
				Type:               "health",
				MonthlyPremium:     300.0,
				Deductible:         2000.0,
				OutOfPocketMax:     8000.0,
				CoveragePercentage: 80.0,
				StartDate:          time.Now().AddDate(-1, 0, 0),
				EndDate:            time.Now().AddDate(0, 6, 0),
			},
			expectError: true,
			errorMsg:    "policy number is required",
		},
		{
			name: "invalid_type",
			policy: InsurancePolicy{
				ID:                 "policy-7",
				UserID:             "user-7",
				ProfileID:          "profile-7",
				Provider:           "Some Insurance",
				PolicyNumber:       "POL123456",
				Type:               "invalid_type",
				MonthlyPremium:     300.0,
				Deductible:         2000.0,
				OutOfPocketMax:     8000.0,
				CoveragePercentage: 80.0,
				StartDate:          time.Now().AddDate(-1, 0, 0),
				EndDate:            time.Now().AddDate(0, 6, 0),
			},
			expectError: true,
			errorMsg:    "type must be one of: health, dental, vision, comprehensive",
		},
		{
			name: "zero_monthly_premium_invalid",
			policy: InsurancePolicy{
				ID:                 "policy-8",
				UserID:             "user-8",
				ProfileID:          "profile-8",
				Provider:           "Some Insurance",
				PolicyNumber:       "POL123456",
				Type:               "health",
				MonthlyPremium:     0.0,
				Deductible:         2000.0,
				OutOfPocketMax:     8000.0,
				CoveragePercentage: 80.0,
				StartDate:          time.Now().AddDate(-1, 0, 0),
				EndDate:            time.Now().AddDate(0, 6, 0),
			},
			expectError: true,
			errorMsg:    "monthly premium must be positive",
		},
		{
			name: "negative_monthly_premium_invalid",
			policy: InsurancePolicy{
				ID:                 "policy-9",
				UserID:             "user-9",
				ProfileID:          "profile-9",
				Provider:           "Some Insurance",
				PolicyNumber:       "POL123456",
				Type:               "health",
				MonthlyPremium:     -100.0,
				Deductible:         2000.0,
				OutOfPocketMax:     8000.0,
				CoveragePercentage: 80.0,
				StartDate:          time.Now().AddDate(-1, 0, 0),
				EndDate:            time.Now().AddDate(0, 6, 0),
			},
			expectError: true,
			errorMsg:    "monthly premium must be positive",
		},
		{
			name: "negative_deductible_invalid",
			policy: InsurancePolicy{
				ID:                 "policy-10",
				UserID:             "user-10",
				ProfileID:          "profile-10",
				Provider:           "Some Insurance",
				PolicyNumber:       "POL123456",
				Type:               "health",
				MonthlyPremium:     300.0,
				Deductible:         -100.0,
				OutOfPocketMax:     8000.0,
				CoveragePercentage: 80.0,
				StartDate:          time.Now().AddDate(-1, 0, 0),
				EndDate:            time.Now().AddDate(0, 6, 0),
			},
			expectError: true,
			errorMsg:    "deductible must be non-negative",
		},
		{
			name: "zero_out_of_pocket_max_invalid",
			policy: InsurancePolicy{
				ID:                 "policy-11",
				UserID:             "user-11",
				ProfileID:          "profile-11",
				Provider:           "Some Insurance",
				PolicyNumber:       "POL123456",
				Type:               "health",
				MonthlyPremium:     300.0,
				Deductible:         2000.0,
				OutOfPocketMax:     0.0,
				CoveragePercentage: 80.0,
				StartDate:          time.Now().AddDate(-1, 0, 0),
				EndDate:            time.Now().AddDate(0, 6, 0),
			},
			expectError: true,
			errorMsg:    "out of pocket maximum must be positive",
		},
		{
			name: "negative_coverage_percentage_invalid",
			policy: InsurancePolicy{
				ID:                 "policy-12",
				UserID:             "user-12",
				ProfileID:          "profile-12",
				Provider:           "Some Insurance",
				PolicyNumber:       "POL123456",
				Type:               "health",
				MonthlyPremium:     300.0,
				Deductible:         2000.0,
				OutOfPocketMax:     8000.0,
				CoveragePercentage: -10.0,
				StartDate:          time.Now().AddDate(-1, 0, 0),
				EndDate:            time.Now().AddDate(0, 6, 0),
			},
			expectError: true,
			errorMsg:    "coverage percentage must be between 0 and 100",
		},
		{
			name: "coverage_percentage_over_100_invalid",
			policy: InsurancePolicy{
				ID:                 "policy-13",
				UserID:             "user-13",
				ProfileID:          "profile-13",
				Provider:           "Some Insurance",
				PolicyNumber:       "POL123456",
				Type:               "health",
				MonthlyPremium:     300.0,
				Deductible:         2000.0,
				OutOfPocketMax:     8000.0,
				CoveragePercentage: 110.0,
				StartDate:          time.Now().AddDate(-1, 0, 0),
				EndDate:            time.Now().AddDate(0, 6, 0),
			},
			expectError: true,
			errorMsg:    "coverage percentage must be between 0 and 100",
		},
		{
			name: "boundary_coverage_percentage_0",
			policy: InsurancePolicy{
				ID:                 "policy-14",
				UserID:             "user-14",
				ProfileID:          "profile-14",
				Provider:           "Some Insurance",
				PolicyNumber:       "POL123456",
				Type:               "health",
				MonthlyPremium:     300.0,
				Deductible:         2000.0,
				OutOfPocketMax:     8000.0,
				CoveragePercentage: 0.0,
				StartDate:          time.Now().AddDate(-1, 0, 0),
				EndDate:            time.Now().AddDate(0, 6, 0),
			},
			expectError: false,
		},
		{
			name: "boundary_coverage_percentage_100",
			policy: InsurancePolicy{
				ID:                 "policy-15",
				UserID:             "user-15",
				ProfileID:          "profile-15",
				Provider:           "Some Insurance",
				PolicyNumber:       "POL123456",
				Type:               "health",
				MonthlyPremium:     300.0,
				Deductible:         2000.0,
				OutOfPocketMax:     8000.0,
				CoveragePercentage: 100.0,
				StartDate:          time.Now().AddDate(-1, 0, 0),
				EndDate:            time.Now().AddDate(0, 6, 0),
			},
			expectError: false,
		},
		{
			name: "end_date_before_start_date_invalid",
			policy: InsurancePolicy{
				ID:                 "policy-16",
				UserID:             "user-16",
				ProfileID:          "profile-16",
				Provider:           "Some Insurance",
				PolicyNumber:       "POL123456",
				Type:               "health",
				MonthlyPremium:     300.0,
				Deductible:         2000.0,
				OutOfPocketMax:     8000.0,
				CoveragePercentage: 80.0,
				StartDate:          time.Now(),
				EndDate:            time.Now().AddDate(-1, 0, 0),
			},
			expectError: true,
			errorMsg:    "end date must be after start date",
		},
		{
			name: "deductible_met_exceeds_deductible_invalid",
			policy: InsurancePolicy{
				ID:                 "policy-17",
				UserID:             "user-17",
				ProfileID:          "profile-17",
				Provider:           "Some Insurance",
				PolicyNumber:       "POL123456",
				Type:               "health",
				MonthlyPremium:     300.0,
				Deductible:         2000.0,
				DeductibleMet:      2500.0,
				OutOfPocketMax:     8000.0,
				CoveragePercentage: 80.0,
				StartDate:          time.Now().AddDate(-1, 0, 0),
				EndDate:            time.Now().AddDate(0, 6, 0),
			},
			expectError: true,
			errorMsg:    "deductible met cannot exceed total deductible",
		},
		{
			name: "out_of_pocket_current_exceeds_max_invalid",
			policy: InsurancePolicy{
				ID:                 "policy-18",
				UserID:             "user-18",
				ProfileID:          "profile-18",
				Provider:           "Some Insurance",
				PolicyNumber:       "POL123456",
				Type:               "health",
				MonthlyPremium:     300.0,
				Deductible:         2000.0,
				OutOfPocketMax:     8000.0,
				OutOfPocketCurrent: 8500.0,
				CoveragePercentage: 80.0,
				StartDate:          time.Now().AddDate(-1, 0, 0),
				EndDate:            time.Now().AddDate(0, 6, 0),
			},
			expectError: true,
			errorMsg:    "current out of pocket cannot exceed maximum",
		},
		{
			name: "negative_deductible_met_invalid",
			policy: InsurancePolicy{
				ID:                 "policy-19",
				UserID:             "user-19",
				ProfileID:          "profile-19",
				Provider:           "Some Insurance",
				PolicyNumber:       "POL123456",
				Type:               "health",
				MonthlyPremium:     300.0,
				Deductible:         2000.0,
				DeductibleMet:      -100.0,
				OutOfPocketMax:     8000.0,
				CoveragePercentage: 80.0,
				StartDate:          time.Now().AddDate(-1, 0, 0),
				EndDate:            time.Now().AddDate(0, 6, 0),
			},
			expectError: true,
			errorMsg:    "deductible met must be non-negative",
		},
		{
			name: "empty_user_id_invalid",
			policy: InsurancePolicy{
				ID:                 "policy-20",
				UserID:             "",
				ProfileID:          "profile-20",
				Provider:           "Some Insurance",
				PolicyNumber:       "POL123456",
				Type:               "health",
				MonthlyPremium:     300.0,
				Deductible:         2000.0,
				OutOfPocketMax:     8000.0,
				CoveragePercentage: 80.0,
				StartDate:          time.Now().AddDate(-1, 0, 0),
				EndDate:            time.Now().AddDate(0, 6, 0),
			},
			expectError: true,
			errorMsg:    "user ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInsurancePolicy_CalculateCoverage(t *testing.T) {
	tests := []struct {
		name                  string
		policy                InsurancePolicy
		expenseAmount         float64
		expectedCoverage      float64
		expectedOutOfPocket   float64
		expectedDeductibleMet float64
	}{
		{
			name: "expense_below_deductible",
			policy: InsurancePolicy{
				Deductible:         2000.0,
				DeductibleMet:      500.0,
				CoveragePercentage: 80.0,
				OutOfPocketMax:     8000.0,
				OutOfPocketCurrent: 1000.0,
			},
			expenseAmount:         300.0,
			expectedCoverage:      0.0,
			expectedOutOfPocket:   300.0,
			expectedDeductibleMet: 800.0,
		},
		{
			name: "expense_meets_remaining_deductible",
			policy: InsurancePolicy{
				Deductible:         2000.0,
				DeductibleMet:      1800.0,
				CoveragePercentage: 80.0,
				OutOfPocketMax:     8000.0,
				OutOfPocketCurrent: 2500.0,
			},
			expenseAmount:         500.0,
			expectedCoverage:      240.0, // 300 * 0.80
			expectedOutOfPocket:   260.0, // 200 (remaining deductible) + 60 (20% of 300)
			expectedDeductibleMet: 2000.0,
		},
		{
			name: "expense_after_deductible_met",
			policy: InsurancePolicy{
				Deductible:         2000.0,
				DeductibleMet:      2000.0,
				CoveragePercentage: 80.0,
				OutOfPocketMax:     8000.0,
				OutOfPocketCurrent: 3000.0,
			},
			expenseAmount:         1000.0,
			expectedCoverage:      800.0,
			expectedOutOfPocket:   200.0,
			expectedDeductibleMet: 2000.0,
		},
		{
			name: "expense_hits_out_of_pocket_maximum",
			policy: InsurancePolicy{
				Deductible:         2000.0,
				DeductibleMet:      2000.0,
				CoveragePercentage: 80.0,
				OutOfPocketMax:     8000.0,
				OutOfPocketCurrent: 7800.0,
			},
			expenseAmount:         1000.0,
			expectedCoverage:      800.0, // Full coverage after OOP max hit
			expectedOutOfPocket:   200.0,
			expectedDeductibleMet: 2000.0,
		},
		{
			name: "expense_exceeds_out_of_pocket_maximum",
			policy: InsurancePolicy{
				Deductible:         1000.0,
				DeductibleMet:      1000.0,
				CoveragePercentage: 70.0,
				OutOfPocketMax:     5000.0,
				OutOfPocketCurrent: 4900.0,
			},
			expenseAmount:         2000.0,
			expectedCoverage:      1900.0, // Only pay remaining $100 OOP, rest covered
			expectedOutOfPocket:   100.0,
			expectedDeductibleMet: 1000.0,
		},
		{
			name: "zero_deductible_policy",
			policy: InsurancePolicy{
				Deductible:         0.0,
				DeductibleMet:      0.0,
				CoveragePercentage: 90.0,
				OutOfPocketMax:     3000.0,
				OutOfPocketCurrent: 500.0,
			},
			expenseAmount:         500.0,
			expectedCoverage:      450.0,
			expectedOutOfPocket:   50.0,
			expectedDeductibleMet: 0.0,
		},
		{
			name: "100_percent_coverage_policy",
			policy: InsurancePolicy{
				Deductible:         1000.0,
				DeductibleMet:      1000.0,
				CoveragePercentage: 100.0,
				OutOfPocketMax:     5000.0,
				OutOfPocketCurrent: 2000.0,
			},
			expenseAmount:         800.0,
			expectedCoverage:      800.0,
			expectedOutOfPocket:   0.0,
			expectedDeductibleMet: 1000.0,
		},
		{
			name: "zero_coverage_percentage_policy",
			policy: InsurancePolicy{
				Deductible:         1000.0,
				DeductibleMet:      1000.0,
				CoveragePercentage: 0.0,
				OutOfPocketMax:     10000.0,
				OutOfPocketCurrent: 2000.0,
			},
			expenseAmount:         500.0,
			expectedCoverage:      0.0,
			expectedOutOfPocket:   500.0,
			expectedDeductibleMet: 1000.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coverage, outOfPocket, newDeductibleMet := tt.policy.CalculateCoverage(tt.expenseAmount)

			assert.InDelta(t, tt.expectedCoverage, coverage, 0.01, "Coverage calculation should be accurate")
			assert.InDelta(t, tt.expectedOutOfPocket, outOfPocket, 0.01, "Out of pocket calculation should be accurate")
			assert.InDelta(t, tt.expectedDeductibleMet, newDeductibleMet, 0.01, "Deductible met calculation should be accurate")
		})
	}
}

func TestInsurancePolicy_GetRemainingDeductible(t *testing.T) {
	tests := []struct {
		name                      string
		deductible                float64
		deductibleMet             float64
		expectedRemainingDeductible float64
	}{
		{
			name:                        "no_deductible_met",
			deductible:                  2000.0,
			deductibleMet:               0.0,
			expectedRemainingDeductible: 2000.0,
		},
		{
			name:                        "partial_deductible_met",
			deductible:                  2000.0,
			deductibleMet:               800.0,
			expectedRemainingDeductible: 1200.0,
		},
		{
			name:                        "full_deductible_met",
			deductible:                  2000.0,
			deductibleMet:               2000.0,
			expectedRemainingDeductible: 0.0,
		},
		{
			name:                        "zero_deductible",
			deductible:                  0.0,
			deductibleMet:               0.0,
			expectedRemainingDeductible: 0.0,
		},
		{
			name:                        "small_remaining_deductible",
			deductible:                  500.0,
			deductibleMet:               450.0,
			expectedRemainingDeductible: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := InsurancePolicy{
				Deductible:    tt.deductible,
				DeductibleMet: tt.deductibleMet,
			}

			remainingDeductible := policy.GetRemainingDeductible()
			assert.Equal(t, tt.expectedRemainingDeductible, remainingDeductible, "Remaining deductible calculation should be accurate")
		})
	}
}

func TestInsurancePolicy_GetRemainingOutOfPocket(t *testing.T) {
	tests := []struct {
		name                        string
		outOfPocketMax              float64
		outOfPocketCurrent          float64
		expectedRemainingOutOfPocket float64
	}{
		{
			name:                         "no_out_of_pocket_spent",
			outOfPocketMax:               8000.0,
			outOfPocketCurrent:           0.0,
			expectedRemainingOutOfPocket: 8000.0,
		},
		{
			name:                         "partial_out_of_pocket_spent",
			outOfPocketMax:               8000.0,
			outOfPocketCurrent:           3000.0,
			expectedRemainingOutOfPocket: 5000.0,
		},
		{
			name:                         "out_of_pocket_max_reached",
			outOfPocketMax:               8000.0,
			outOfPocketCurrent:           8000.0,
			expectedRemainingOutOfPocket: 0.0,
		},
		{
			name:                         "small_remaining_out_of_pocket",
			outOfPocketMax:               5000.0,
			outOfPocketCurrent:           4900.0,
			expectedRemainingOutOfPocket: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := InsurancePolicy{
				OutOfPocketMax:     tt.outOfPocketMax,
				OutOfPocketCurrent: tt.outOfPocketCurrent,
			}

			remainingOutOfPocket := policy.GetRemainingOutOfPocket()
			assert.Equal(t, tt.expectedRemainingOutOfPocket, remainingOutOfPocket, "Remaining out of pocket calculation should be accurate")
		})
	}
}

func TestInsurancePolicy_IsDeductibleMet(t *testing.T) {
	tests := []struct {
		name             string
		deductible       float64
		deductibleMet    float64
		expectedResult   bool
	}{
		{
			name:           "deductible_not_met",
			deductible:     2000.0,
			deductibleMet:  1500.0,
			expectedResult: false,
		},
		{
			name:           "deductible_exactly_met",
			deductible:     2000.0,
			deductibleMet:  2000.0,
			expectedResult: true,
		},
		{
			name:           "deductible_over_met",
			deductible:     2000.0,
			deductibleMet:  2100.0,
			expectedResult: true,
		},
		{
			name:           "zero_deductible_policy",
			deductible:     0.0,
			deductibleMet:  0.0,
			expectedResult: true,
		},
		{
			name:           "no_amount_towards_deductible",
			deductible:     1000.0,
			deductibleMet:  0.0,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := InsurancePolicy{
				Deductible:    tt.deductible,
				DeductibleMet: tt.deductibleMet,
			}

			result := policy.IsDeductibleMet()
			assert.Equal(t, tt.expectedResult, result, "Deductible met check should be accurate")
		})
	}
}

func TestInsurancePolicy_IsOutOfPocketMaxReached(t *testing.T) {
	tests := []struct {
		name                  string
		outOfPocketMax        float64
		outOfPocketCurrent    float64
		expectedResult        bool
	}{
		{
			name:               "out_of_pocket_max_not_reached",
			outOfPocketMax:     8000.0,
			outOfPocketCurrent: 5000.0,
			expectedResult:     false,
		},
		{
			name:               "out_of_pocket_max_exactly_reached",
			outOfPocketMax:     8000.0,
			outOfPocketCurrent: 8000.0,
			expectedResult:     true,
		},
		{
			name:               "out_of_pocket_max_exceeded",
			outOfPocketMax:     8000.0,
			outOfPocketCurrent: 8500.0,
			expectedResult:     true,
		},
		{
			name:               "no_out_of_pocket_spent",
			outOfPocketMax:     5000.0,
			outOfPocketCurrent: 0.0,
			expectedResult:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := InsurancePolicy{
				OutOfPocketMax:     tt.outOfPocketMax,
				OutOfPocketCurrent: tt.outOfPocketCurrent,
			}

			result := policy.IsOutOfPocketMaxReached()
			assert.Equal(t, tt.expectedResult, result, "Out of pocket max reached check should be accurate")
		})
	}
}

func TestInsurancePolicy_GetAnnualPremium(t *testing.T) {
	tests := []struct {
		name                   string
		monthlyPremium         float64
		expectedAnnualPremium  float64
	}{
		{
			name:                  "standard_monthly_premium",
			monthlyPremium:        300.0,
			expectedAnnualPremium: 3600.0,
		},
		{
			name:                  "low_monthly_premium",
			monthlyPremium:        25.0,
			expectedAnnualPremium: 300.0,
		},
		{
			name:                  "high_monthly_premium",
			monthlyPremium:        500.0,
			expectedAnnualPremium: 6000.0,
		},
		{
			name:                  "decimal_monthly_premium",
			monthlyPremium:        125.50,
			expectedAnnualPremium: 1506.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := InsurancePolicy{
				MonthlyPremium: tt.monthlyPremium,
			}

			annualPremium := policy.GetAnnualPremium()
			assert.InDelta(t, tt.expectedAnnualPremium, annualPremium, 0.01, "Annual premium calculation should be accurate")
		})
	}
}