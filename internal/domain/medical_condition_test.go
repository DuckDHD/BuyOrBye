package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMedicalCondition_Validate(t *testing.T) {
	tests := []struct {
		name        string
		condition   MedicalCondition
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_chronic_condition",
			condition: MedicalCondition{
				ID:                 "condition-1",
				UserID:             "user-1",
				ProfileID:          "profile-1",
				Name:               "Type 2 Diabetes",
				Category:           "chronic",
				Severity:           "moderate",
				DiagnosedDate:      time.Now().AddDate(-2, 0, 0),
				IsActive:           true,
				RequiresMedication: true,
				MonthlyMedCost:     150.0,
				RiskFactor:         0.6,
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			},
			expectError: false,
		},
		{
			name: "valid_acute_condition",
			condition: MedicalCondition{
				ID:                 "condition-2",
				UserID:             "user-2",
				ProfileID:          "profile-2",
				Name:               "Bronchitis",
				Category:           "acute",
				Severity:           "mild",
				DiagnosedDate:      time.Now().AddDate(0, -1, 0),
				IsActive:           true,
				RequiresMedication: false,
				MonthlyMedCost:     0.0,
				RiskFactor:         0.1,
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			},
			expectError: false,
		},
		{
			name: "valid_mental_health_condition",
			condition: MedicalCondition{
				ID:                 "condition-3",
				UserID:             "user-3",
				ProfileID:          "profile-3",
				Name:               "Anxiety Disorder",
				Category:           "mental_health",
				Severity:           "severe",
				DiagnosedDate:      time.Now().AddDate(-1, 0, 0),
				IsActive:           true,
				RequiresMedication: true,
				MonthlyMedCost:     80.0,
				RiskFactor:         0.4,
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			},
			expectError: false,
		},
		{
			name: "valid_preventive_condition",
			condition: MedicalCondition{
				ID:                 "condition-4",
				UserID:             "user-4",
				ProfileID:          "profile-4",
				Name:               "Annual Physical",
				Category:           "preventive",
				Severity:           "mild",
				DiagnosedDate:      time.Now(),
				IsActive:           true,
				RequiresMedication: false,
				MonthlyMedCost:     0.0,
				RiskFactor:         0.0,
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			},
			expectError: false,
		},
		{
			name: "empty_name_invalid",
			condition: MedicalCondition{
				ID:            "condition-5",
				UserID:        "user-5",
				ProfileID:     "profile-5",
				Name:          "",
				Category:      "chronic",
				Severity:      "moderate",
				DiagnosedDate: time.Now(),
				IsActive:      true,
			},
			expectError: true,
			errorMsg:    "condition name is required",
		},
		{
			name: "invalid_category",
			condition: MedicalCondition{
				ID:            "condition-6",
				UserID:        "user-6",
				ProfileID:     "profile-6",
				Name:          "Some Condition",
				Category:      "invalid_category",
				Severity:      "moderate",
				DiagnosedDate: time.Now(),
				IsActive:      true,
			},
			expectError: true,
			errorMsg:    "category must be one of: chronic, acute, mental_health, preventive",
		},
		{
			name: "invalid_severity_mild",
			condition: MedicalCondition{
				ID:            "condition-7",
				UserID:        "user-7",
				ProfileID:     "profile-7",
				Name:          "Some Condition",
				Category:      "chronic",
				Severity:      "invalid_severity",
				DiagnosedDate: time.Now(),
				IsActive:      true,
			},
			expectError: true,
			errorMsg:    "severity must be one of: mild, moderate, severe, critical",
		},
		{
			name: "empty_severity_invalid",
			condition: MedicalCondition{
				ID:            "condition-8",
				UserID:        "user-8",
				ProfileID:     "profile-8",
				Name:          "Some Condition",
				Category:      "chronic",
				Severity:      "",
				DiagnosedDate: time.Now(),
				IsActive:      true,
			},
			expectError: true,
			errorMsg:    "severity must be one of: mild, moderate, severe, critical",
		},
		{
			name: "negative_monthly_med_cost_invalid",
			condition: MedicalCondition{
				ID:                 "condition-9",
				UserID:             "user-9",
				ProfileID:          "profile-9",
				Name:               "Some Condition",
				Category:           "chronic",
				Severity:           "moderate",
				DiagnosedDate:      time.Now(),
				IsActive:           true,
				RequiresMedication: true,
				MonthlyMedCost:     -50.0,
			},
			expectError: true,
			errorMsg:    "monthly medication cost must be non-negative",
		},
		{
			name: "invalid_risk_factor_negative",
			condition: MedicalCondition{
				ID:            "condition-10",
				UserID:        "user-10",
				ProfileID:     "profile-10",
				Name:          "Some Condition",
				Category:      "chronic",
				Severity:      "moderate",
				DiagnosedDate: time.Now(),
				IsActive:      true,
				RiskFactor:    -0.1,
			},
			expectError: true,
			errorMsg:    "risk factor must be between 0.0 and 1.0",
		},
		{
			name: "invalid_risk_factor_too_high",
			condition: MedicalCondition{
				ID:            "condition-11",
				UserID:        "user-11",
				ProfileID:     "profile-11",
				Name:          "Some Condition",
				Category:      "chronic",
				Severity:      "moderate",
				DiagnosedDate: time.Now(),
				IsActive:      true,
				RiskFactor:    1.5,
			},
			expectError: true,
			errorMsg:    "risk factor must be between 0.0 and 1.0",
		},
		{
			name: "valid_risk_factor_boundary_0",
			condition: MedicalCondition{
				ID:            "condition-12",
				UserID:        "user-12",
				ProfileID:     "profile-12",
				Name:          "Some Condition",
				Category:      "preventive",
				Severity:      "mild",
				DiagnosedDate: time.Now(),
				IsActive:      true,
				RiskFactor:    0.0,
			},
			expectError: false,
		},
		{
			name: "valid_risk_factor_boundary_1",
			condition: MedicalCondition{
				ID:            "condition-13",
				UserID:        "user-13",
				ProfileID:     "profile-13",
				Name:          "Some Condition",
				Category:      "chronic",
				Severity:      "critical",
				DiagnosedDate: time.Now(),
				IsActive:      true,
				RiskFactor:    1.0,
			},
			expectError: false,
		},
		{
			name: "future_diagnosed_date_invalid",
			condition: MedicalCondition{
				ID:            "condition-14",
				UserID:        "user-14",
				ProfileID:     "profile-14",
				Name:          "Some Condition",
				Category:      "chronic",
				Severity:      "moderate",
				DiagnosedDate: time.Now().AddDate(0, 1, 0), // Future date
				IsActive:      true,
			},
			expectError: true,
			errorMsg:    "diagnosed date cannot be in the future",
		},
		{
			name: "empty_user_id_invalid",
			condition: MedicalCondition{
				ID:            "condition-15",
				UserID:        "",
				ProfileID:     "profile-15",
				Name:          "Some Condition",
				Category:      "chronic",
				Severity:      "moderate",
				DiagnosedDate: time.Now(),
				IsActive:      true,
			},
			expectError: true,
			errorMsg:    "user ID is required",
		},
		{
			name: "empty_profile_id_invalid",
			condition: MedicalCondition{
				ID:            "condition-16",
				UserID:        "user-16",
				ProfileID:     "",
				Name:          "Some Condition",
				Category:      "chronic",
				Severity:      "moderate",
				DiagnosedDate: time.Now(),
				IsActive:      true,
			},
			expectError: true,
			errorMsg:    "profile ID is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.condition.Validate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMedicalCondition_CalculateRiskContribution(t *testing.T) {
	tests := []struct {
		name                   string
		condition              MedicalCondition
		expectedRiskContribution float64
	}{
		{
			name: "mild_chronic_condition",
			condition: MedicalCondition{
				Name:     "Mild Hypertension",
				Category: "chronic",
				Severity: "mild",
				IsActive: true,
			},
			expectedRiskContribution: 2.0,
		},
		{
			name: "moderate_chronic_condition",
			condition: MedicalCondition{
				Name:     "Type 2 Diabetes",
				Category: "chronic",
				Severity: "moderate",
				IsActive: true,
			},
			expectedRiskContribution: 5.0,
		},
		{
			name: "severe_chronic_condition",
			condition: MedicalCondition{
				Name:     "Heart Disease",
				Category: "chronic",
				Severity: "severe",
				IsActive: true,
			},
			expectedRiskContribution: 10.0,
		},
		{
			name: "critical_chronic_condition",
			condition: MedicalCondition{
				Name:     "Advanced Cancer",
				Category: "chronic",
				Severity: "critical",
				IsActive: true,
			},
			expectedRiskContribution: 15.0,
		},
		{
			name: "mild_acute_condition",
			condition: MedicalCondition{
				Name:     "Common Cold",
				Category: "acute",
				Severity: "mild",
				IsActive: true,
			},
			expectedRiskContribution: 1.0,
		},
		{
			name: "severe_acute_condition",
			condition: MedicalCondition{
				Name:     "Pneumonia",
				Category: "acute",
				Severity: "severe",
				IsActive: true,
			},
			expectedRiskContribution: 5.0,
		},
		{
			name: "critical_acute_condition",
			condition: MedicalCondition{
				Name:     "Acute MI",
				Category: "acute",
				Severity: "critical",
				IsActive: true,
			},
			expectedRiskContribution: 8.0,
		},
		{
			name: "moderate_mental_health_condition",
			condition: MedicalCondition{
				Name:     "Depression",
				Category: "mental_health",
				Severity: "moderate",
				IsActive: true,
			},
			expectedRiskContribution: 3.0,
		},
		{
			name: "severe_mental_health_condition",
			condition: MedicalCondition{
				Name:     "Bipolar Disorder",
				Category: "mental_health",
				Severity: "severe",
				IsActive: true,
			},
			expectedRiskContribution: 6.0,
		},
		{
			name: "critical_mental_health_condition",
			condition: MedicalCondition{
				Name:     "Severe Psychosis",
				Category: "mental_health",
				Severity: "critical",
				IsActive: true,
			},
			expectedRiskContribution: 10.0,
		},
		{
			name: "mild_preventive_condition",
			condition: MedicalCondition{
				Name:     "Annual Checkup",
				Category: "preventive",
				Severity: "mild",
				IsActive: true,
			},
			expectedRiskContribution: 0.0,
		},
		{
			name: "moderate_preventive_condition",
			condition: MedicalCondition{
				Name:     "Vaccination",
				Category: "preventive",
				Severity: "moderate",
				IsActive: true,
			},
			expectedRiskContribution: 0.0,
		},
		{
			name: "inactive_severe_chronic_condition",
			condition: MedicalCondition{
				Name:     "Resolved Heart Disease",
				Category: "chronic",
				Severity: "severe",
				IsActive: false,
			},
			expectedRiskContribution: 0.0,
		},
		{
			name: "inactive_critical_condition",
			condition: MedicalCondition{
				Name:     "Resolved Cancer",
				Category: "chronic",
				Severity: "critical",
				IsActive: false,
			},
			expectedRiskContribution: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			riskContribution := tt.condition.CalculateRiskContribution()
			assert.Equal(t, tt.expectedRiskContribution, riskContribution, "Risk contribution calculation should match expected value")
		})
	}
}

func TestMedicalCondition_GetSeverityScore(t *testing.T) {
	tests := []struct {
		name          string
		severity      string
		expectedScore int
	}{
		{
			name:          "mild_severity",
			severity:      "mild",
			expectedScore: 1,
		},
		{
			name:          "moderate_severity",
			severity:      "moderate",
			expectedScore: 2,
		},
		{
			name:          "severe_severity",
			severity:      "severe",
			expectedScore: 3,
		},
		{
			name:          "critical_severity",
			severity:      "critical",
			expectedScore: 4,
		},
		{
			name:          "unknown_severity",
			severity:      "unknown",
			expectedScore: 0,
		},
		{
			name:          "empty_severity",
			severity:      "",
			expectedScore: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := MedicalCondition{
				Severity: tt.severity,
			}

			severityScore := condition.GetSeverityScore()
			assert.Equal(t, tt.expectedScore, severityScore, "Severity score should match expected value")
		})
	}
}

func TestMedicalCondition_IsChronic(t *testing.T) {
	tests := []struct {
		name       string
		category   string
		isChronicExpected bool
	}{
		{
			name:              "chronic_category",
			category:          "chronic",
			isChronicExpected: true,
		},
		{
			name:              "acute_category",
			category:          "acute",
			isChronicExpected: false,
		},
		{
			name:              "mental_health_category",
			category:          "mental_health",
			isChronicExpected: true,
		},
		{
			name:              "preventive_category",
			category:          "preventive",
			isChronicExpected: false,
		},
		{
			name:              "unknown_category",
			category:          "unknown",
			isChronicExpected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := MedicalCondition{
				Category: tt.category,
			}

			isChronic := condition.IsChronic()
			assert.Equal(t, tt.isChronicExpected, isChronic, "IsChronic result should match expected value")
		})
	}
}

func TestMedicalCondition_GetAnnualMedCost(t *testing.T) {
	tests := []struct {
		name                string
		monthlyMedCost      float64
		expectedAnnualCost  float64
	}{
		{
			name:               "monthly_cost_100",
			monthlyMedCost:     100.0,
			expectedAnnualCost: 1200.0,
		},
		{
			name:               "monthly_cost_75_50",
			monthlyMedCost:     75.50,
			expectedAnnualCost: 906.0,
		},
		{
			name:               "monthly_cost_zero",
			monthlyMedCost:     0.0,
			expectedAnnualCost: 0.0,
		},
		{
			name:               "monthly_cost_decimal",
			monthlyMedCost:     33.33,
			expectedAnnualCost: 399.96,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := MedicalCondition{
				MonthlyMedCost: tt.monthlyMedCost,
			}

			annualCost := condition.GetAnnualMedCost()
			assert.InDelta(t, tt.expectedAnnualCost, annualCost, 0.01, "Annual medication cost should be accurate")
		})
	}
}

func TestMedicalCondition_RequiresHighRiskManagement(t *testing.T) {
	tests := []struct {
		name                      string
		condition                 MedicalCondition
		expectedHighRiskMgmt      bool
	}{
		{
			name: "severe_chronic_active_condition",
			condition: MedicalCondition{
				Category: "chronic",
				Severity: "severe",
				IsActive: true,
			},
			expectedHighRiskMgmt: true,
		},
		{
			name: "critical_chronic_active_condition",
			condition: MedicalCondition{
				Category: "chronic",
				Severity: "critical",
				IsActive: true,
			},
			expectedHighRiskMgmt: true,
		},
		{
			name: "critical_acute_active_condition",
			condition: MedicalCondition{
				Category: "acute",
				Severity: "critical",
				IsActive: true,
			},
			expectedHighRiskMgmt: true,
		},
		{
			name: "severe_mental_health_active_condition",
			condition: MedicalCondition{
				Category: "mental_health",
				Severity: "severe",
				IsActive: true,
			},
			expectedHighRiskMgmt: true,
		},
		{
			name: "moderate_chronic_active_condition",
			condition: MedicalCondition{
				Category: "chronic",
				Severity: "moderate",
				IsActive: true,
			},
			expectedHighRiskMgmt: false,
		},
		{
			name: "mild_chronic_active_condition",
			condition: MedicalCondition{
				Category: "chronic",
				Severity: "mild",
				IsActive: true,
			},
			expectedHighRiskMgmt: false,
		},
		{
			name: "severe_chronic_inactive_condition",
			condition: MedicalCondition{
				Category: "chronic",
				Severity: "severe",
				IsActive: false,
			},
			expectedHighRiskMgmt: false,
		},
		{
			name: "preventive_condition",
			condition: MedicalCondition{
				Category: "preventive",
				Severity: "critical",
				IsActive: true,
			},
			expectedHighRiskMgmt: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requiresHighRisk := tt.condition.RequiresHighRiskManagement()
			assert.Equal(t, tt.expectedHighRiskMgmt, requiresHighRisk, "High risk management requirement should match expected value")
		})
	}
}