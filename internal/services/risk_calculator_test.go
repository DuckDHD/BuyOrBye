package services

import (
	"testing"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/stretchr/testify/assert"
)

// Test CalculateHealthRiskScore with different scenarios
func TestRiskCalculator_CalculateHealthRiskScore(t *testing.T) {
	calculator := NewRiskCalculator()

	tests := []struct {
		name       string
		profile    *domain.HealthProfile
		conditions []domain.MedicalCondition
		expected   struct {
			min int
			max int
		}
	}{
		{
			name: "young_healthy_low_risk",
			profile: &domain.HealthProfile{
				Age:        25,
				Height:     175.0,
				Weight:     70.0,
				FamilySize: 2,
			},
			conditions: []domain.MedicalCondition{},
			expected: struct {
				min int
				max int
			}{min: 0, max: 10}, // Should be low risk
		},
		{
			name: "middle_aged_chronic_moderate_risk",
			profile: &domain.HealthProfile{
				Age:        45,
				Height:     170.0,
				Weight:     85.0, // Overweight
				FamilySize: 4,
			},
			conditions: []domain.MedicalCondition{
				{
					Severity:   "moderate",
					Category:   "chronic",
					RiskFactor: 0.6,
				},
			},
			expected: struct {
				min int
				max int
			}{min: 20, max: 40},
		},
		{
			name: "elderly_severe_high_risk",
			profile: &domain.HealthProfile{
				Age:        70,
				Height:     160.0,
				Weight:     95.0, // Obese
				FamilySize: 5,
			},
			conditions: []domain.MedicalCondition{
				{
					Severity:   "severe",
					Category:   "chronic",
					RiskFactor: 0.8,
				},
				{
					Severity:   "moderate",
					Category:   "chronic",
					RiskFactor: 0.5,
				},
			},
			expected: struct {
				min int
				max int
			}{min: 50, max: 80},
		},
		{
			name: "critical_multiple_conditions",
			profile: &domain.HealthProfile{
				Age:        80,
				Height:     155.0,
				Weight:     45.0, // Underweight
				FamilySize: 6,
			},
			conditions: []domain.MedicalCondition{
				{
					Severity:   "critical",
					Category:   "chronic",
					RiskFactor: 1.0,
				},
				{
					Severity:   "severe",
					Category:   "chronic",
					RiskFactor: 0.9,
				},
			},
			expected: struct {
				min int
				max int
			}{min: 75, max: 100},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculator.CalculateHealthRiskScore(tt.profile, tt.conditions)
			
			assert.GreaterOrEqual(t, score, tt.expected.min, "Risk score should be at least minimum expected")
			assert.LessOrEqual(t, score, tt.expected.max, "Risk score should not exceed maximum expected")
			assert.GreaterOrEqual(t, score, 0, "Risk score should not be negative")
			assert.LessOrEqual(t, score, 100, "Risk score should not exceed 100")
		})
	}
}

// Test AssessFinancialVulnerability
func TestRiskCalculator_AssessFinancialVulnerability(t *testing.T) {
	calculator := NewRiskCalculator()

	tests := []struct {
		name        string
		healthCosts float64
		income      float64
		expected    string
	}{
		{
			name:        "secure_low_costs",
			healthCosts: 2400.0, // $2400/year
			income:      60000.0, // 4% of income
			expected:    "secure",
		},
		{
			name:        "moderate_medium_costs",
			healthCosts: 6000.0,  // $6000/year
			income:      60000.0, // 10% of income
			expected:    "moderate",
		},
		{
			name:        "vulnerable_high_costs",
			healthCosts: 12000.0, // $12000/year
			income:      60000.0, // 20% of income
			expected:    "vulnerable",
		},
		{
			name:        "critical_very_high_costs",
			healthCosts: 20000.0, // $20000/year
			income:      60000.0, // 33% of income
			expected:    "critical",
		},
		{
			name:        "critical_zero_income",
			healthCosts: 5000.0,
			income:      0.0,
			expected:    "critical",
		},
		{
			name:        "critical_negative_income",
			healthCosts: 5000.0,
			income:      -1000.0,
			expected:    "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculator.AssessFinancialVulnerability(tt.healthCosts, tt.income)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test RecommendEmergencyFund
func TestRiskCalculator_RecommendEmergencyFund(t *testing.T) {
	calculator := NewRiskCalculator()

	tests := []struct {
		name             string
		riskScore        int
		monthlyExpenses  float64
		expectedMultiple float64
	}{
		{
			name:             "low_risk_3_months",
			riskScore:        15,
			monthlyExpenses:  3000.0,
			expectedMultiple: 3.0,
		},
		{
			name:             "moderate_risk_6_months",
			riskScore:        35,
			monthlyExpenses:  3000.0,
			expectedMultiple: 6.0,
		},
		{
			name:             "high_risk_8_months",
			riskScore:        60,
			monthlyExpenses:  3000.0,
			expectedMultiple: 8.0,
		},
		{
			name:             "critical_risk_12_months",
			riskScore:        85,
			monthlyExpenses:  3000.0,
			expectedMultiple: 12.0,
		},
		{
			name:             "boundary_low_moderate_25",
			riskScore:        25,
			monthlyExpenses:  2500.0,
			expectedMultiple: 3.0,
		},
		{
			name:             "boundary_moderate_high_26",
			riskScore:        26,
			monthlyExpenses:  2500.0,
			expectedMultiple: 6.0,
		},
		{
			name:             "boundary_high_critical_51",
			riskScore:        51,
			monthlyExpenses:  4000.0,
			expectedMultiple: 8.0,
		},
		{
			name:             "boundary_critical_76",
			riskScore:        76,
			monthlyExpenses:  4000.0,
			expectedMultiple: 12.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculator.RecommendEmergencyFund(tt.riskScore, tt.monthlyExpenses)
			expected := tt.monthlyExpenses * tt.expectedMultiple
			assert.Equal(t, expected, result)
		})
	}
}

// Test DetermineRiskLevel
func TestRiskCalculator_DetermineRiskLevel(t *testing.T) {
	calculator := NewRiskCalculator()

	tests := []struct {
		name     string
		score    int
		expected string
	}{
		{
			name:     "low_risk_0",
			score:    0,
			expected: "low",
		},
		{
			name:     "low_risk_25",
			score:    25,
			expected: "low",
		},
		{
			name:     "moderate_risk_26",
			score:    26,
			expected: "moderate",
		},
		{
			name:     "moderate_risk_50",
			score:    50,
			expected: "moderate",
		},
		{
			name:     "high_risk_51",
			score:    51,
			expected: "high",
		},
		{
			name:     "high_risk_75",
			score:    75,
			expected: "high",
		},
		{
			name:     "critical_risk_76",
			score:    76,
			expected: "critical",
		},
		{
			name:     "critical_risk_100",
			score:    100,
			expected: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculator.DetermineRiskLevel(tt.score)
			assert.Equal(t, tt.expected, result)
		})
	}
}