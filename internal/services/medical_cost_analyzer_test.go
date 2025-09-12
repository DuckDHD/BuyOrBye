package services

import (
	"testing"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/stretchr/testify/assert"
)

// Test CalculateMonthlyAverage with different frequencies
func TestMedicalCostAnalyzer_CalculateMonthlyAverage(t *testing.T) {
	tests := []struct {
		name                     string
		expenses                 []domain.MedicalExpense
		expectedMonthlyAverage   float64
	}{
		{
			name:                   "no_expenses",
			expenses:               []domain.MedicalExpense{},
			expectedMonthlyAverage: 0.0,
		},
		{
			name: "only_monthly_recurring_expenses",
			expenses: []domain.MedicalExpense{
				{
					Amount:      80.0,
					OutOfPocket: 20.0,
					IsRecurring: true,
					Frequency:   "monthly",
				},
				{
					Amount:      120.0,
					OutOfPocket: 40.0,
					IsRecurring: true,
					Frequency:   "monthly",
				},
			},
			expectedMonthlyAverage: 60.0, // (20 + 40) per month
		},
		{
			name: "only_quarterly_recurring_expenses",
			expenses: []domain.MedicalExpense{
				{
					Amount:      600.0,
					OutOfPocket: 150.0,
					IsRecurring: true,
					Frequency:   "quarterly",
				},
				{
					Amount:      400.0,
					OutOfPocket: 100.0,
					IsRecurring: true,
					Frequency:   "quarterly",
				},
			},
			expectedMonthlyAverage: 83.33, // (150 + 100) / 3 months
		},
		{
			name: "only_annual_recurring_expenses",
			expenses: []domain.MedicalExpense{
				{
					Amount:      1200.0,
					OutOfPocket: 300.0,
					IsRecurring: true,
					Frequency:   "annually",
				},
				{
					Amount:      2400.0,
					OutOfPocket: 600.0,
					IsRecurring: true,
					Frequency:   "annually",
				},
			},
			expectedMonthlyAverage: 75.0, // (300 + 600) / 12 months
		},
		{
			name: "mixed_recurring_frequencies",
			expenses: []domain.MedicalExpense{
				{
					Amount:      80.0,
					OutOfPocket: 20.0,
					IsRecurring: true,
					Frequency:   "monthly",
				},
				{
					Amount:      600.0,
					OutOfPocket: 150.0,
					IsRecurring: true,
					Frequency:   "quarterly",
				},
				{
					Amount:      1200.0,
					OutOfPocket: 300.0,
					IsRecurring: true,
					Frequency:   "annually",
				},
			},
			expectedMonthlyAverage: 95.0, // 20 + (150/3) + (300/12) = 20 + 50 + 25
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMedicalCostAnalyzer()
			
			monthlyAverage := analyzer.CalculateMonthlyAverage(tt.expenses)
			
			assert.InDelta(t, tt.expectedMonthlyAverage, monthlyAverage, 0.01, "Monthly average calculation should be accurate")
		})
	}
}

// Test ProjectAnnualCosts including recurring medications
func TestMedicalCostAnalyzer_ProjectAnnualCosts(t *testing.T) {
	tests := []struct {
		name                 string
		expenses             []domain.MedicalExpense
		conditions           []domain.MedicalCondition
		expectedAnnualCosts  float64
	}{
		{
			name:                "no_expenses_or_conditions",
			expenses:            []domain.MedicalExpense{},
			conditions:          []domain.MedicalCondition{},
			expectedAnnualCosts: 0.0,
		},
		{
			name: "only_recurring_expenses",
			expenses: []domain.MedicalExpense{
				{
					Amount:      80.0,
					OutOfPocket: 20.0,
					IsRecurring: true,
					Frequency:   "monthly",
				},
				{
					Amount:      600.0,
					OutOfPocket: 150.0,
					IsRecurring: true,
					Frequency:   "quarterly",
				},
				{
					Amount:      1200.0,
					OutOfPocket: 300.0,
					IsRecurring: true,
					Frequency:   "annually",
				},
			},
			conditions:          []domain.MedicalCondition{},
			expectedAnnualCosts: 1140.0, // (20*12) + (150*4) + 300 = 240 + 600 + 300
		},
		{
			name: "conditions_with_estimated_costs",
			expenses: []domain.MedicalExpense{},
			conditions: []domain.MedicalCondition{
				{
					Name:     "Type 2 Diabetes",
					Category: "chronic",
					Severity: "moderate",
					IsActive: true,
				},
				{
					Name:     "Hypertension",
					Category: "chronic", 
					Severity: "mild",
					IsActive: true,
				},
			},
			expectedAnnualCosts: 5000.0, // 3500 (moderate) + 1500 (mild)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMedicalCostAnalyzer()
			
			annualCosts := analyzer.ProjectAnnualCosts(tt.expenses, tt.conditions)
			
			assert.InDelta(t, tt.expectedAnnualCosts, annualCosts, 0.01, "Annual cost projection should be accurate")
		})
	}
}

// Test IdentifyCostReductionOpportunities
func TestMedicalCostAnalyzer_IdentifyCostReductionOpportunities(t *testing.T) {
	tests := []struct {
		name                        string
		expenses                    []domain.MedicalExpense
		expectedOpportunityCount    int
	}{
		{
			name:                     "no_expenses_no_opportunities",
			expenses:                 []domain.MedicalExpense{},
			expectedOpportunityCount: 0,
		},
		{
			name: "high_cost_services_create_opportunities",
			expenses: []domain.MedicalExpense{
				{
					Amount:      600.0, // High cost service
					OutOfPocket: 150.0,
					Category:    "doctor_visit",
				},
			},
			expectedOpportunityCount: 1, // Should identify high cost service
		},
		{
			name: "high_medication_costs_create_opportunities",
			expenses: []domain.MedicalExpense{
				{
					Amount:      200.0,
					OutOfPocket: 50.0,
					Category:    "medication",
				},
				{
					Amount:      150.0,
					OutOfPocket: 40.0,
					Category:    "medication",
				},
				{
					Amount:      100.0,
					OutOfPocket: 25.0,
					Category:    "doctor_visit",
				},
			},
			expectedOpportunityCount: 1, // Should identify medication optimization
		},
		{
			name: "recurring_high_expenses_create_opportunities",
			expenses: []domain.MedicalExpense{
				{
					Amount:      150.0,
					OutOfPocket: 150.0,
					Category:    "therapy",
					IsRecurring: true,
					Frequency:   "monthly",
				},
			},
			expectedOpportunityCount: 1, // Should identify recurring expense opportunity
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzer := NewMedicalCostAnalyzer()
			
			opportunities := analyzer.IdentifyCostReductionOpportunities(tt.expenses)
			
			assert.Equal(t, tt.expectedOpportunityCount, len(opportunities), "Number of opportunities should match expected")
			
			// Verify each opportunity has required fields
			for _, opp := range opportunities {
				assert.NotEmpty(t, opp.Type, "Opportunity should have a type")
				assert.NotEmpty(t, opp.Description, "Opportunity should have a description")
				assert.NotEmpty(t, opp.Recommendation, "Opportunity should have a recommendation")
				assert.GreaterOrEqual(t, opp.PotentialSavings, 0.0, "Potential savings should be non-negative")
			}
		})
	}
}