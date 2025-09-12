package services

import (
	"fmt"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// medicalCostAnalyzer implements the MedicalCostAnalyzer interface
type medicalCostAnalyzer struct{}

// NewMedicalCostAnalyzer creates a new medical cost analyzer instance
func NewMedicalCostAnalyzer() MedicalCostAnalyzer {
	return &medicalCostAnalyzer{}
}

// CalculateMonthlyAverage calculates average monthly medical expenses
// Normalize all frequencies to monthly equivalent
func (m *medicalCostAnalyzer) CalculateMonthlyAverage(expenses []domain.MedicalExpense) float64 {
	if len(expenses) == 0 {
		return 0.0
	}

	total := 0.0
	for _, expense := range expenses {
		monthlyAmount := m.normalizeToMonthly(expense)
		total += monthlyAmount
	}

	return total
}

// ProjectAnnualCosts projects total annual medical costs including recurring medications
func (m *medicalCostAnalyzer) ProjectAnnualCosts(expenses []domain.MedicalExpense, conditions []domain.MedicalCondition) float64 {
	total := 0.0

	// Sum actual recurring expenses (monthly * 12)
	for _, expense := range expenses {
		if expense.IsRecurring {
			total += m.normalizeToMonthly(expense) * 12
		} else {
			// For one-time expenses, include in projection if recent
			total += expense.Amount
		}
	}

	// Add estimated costs for chronic conditions requiring medication
	for _, condition := range conditions {
		if condition.IsActive && condition.RequiresMedication {
			// Use monthly medication cost if available
			if condition.MonthlyMedCost > 0 {
				total += condition.MonthlyMedCost * 12
			} else {
				// Fallback to severity-based estimates
				switch condition.Severity {
				case "mild":
					total += 1200.0 // $100/month
				case "moderate":
					total += 2400.0 // $200/month
				case "severe":
					total += 4800.0 // $400/month
				case "critical":
					total += 7200.0 // $600/month
				}
			}
		}
	}

	return total
}

// IdentifyCostReductionOpportunities identifies generic alternatives, preventive care opportunities
func (m *medicalCostAnalyzer) IdentifyCostReductionOpportunities(expenses []domain.MedicalExpense) []CostReductionOpportunity {
	opportunities := make([]CostReductionOpportunity, 0)

	medicationExpenses := 0.0
	doctorVisitExpenses := 0.0
	totalExpenses := 0.0
	hasPreventiveCare := false

	for _, expense := range expenses {
		totalExpenses += expense.Amount

		// Track by category
		switch expense.Category {
		case "medication":
			medicationExpenses += expense.Amount
			// Suggest generic alternatives for high medication costs
			if expense.Amount > 200.0 {
				opportunities = append(opportunities, CostReductionOpportunity{
					Type:             "generic_alternative",
					Description:      "High-cost medication may have generic alternative",
					PotentialSavings: expense.Amount * 0.4, // 40% potential savings with generics
					Recommendation:   "Ask doctor about generic alternatives or therapeutic substitutes",
				})
			}
		case "doctor_visit":
			doctorVisitExpenses += expense.Amount
			// Check if this is preventive care
			if expense.Description != "" && (expense.Description == "annual checkup" || expense.Description == "physical" || expense.Description == "preventive") {
				hasPreventiveCare = true
			}
		case "lab_test":
			// Suggest bundling lab tests
			if expense.Amount > 300.0 {
				opportunities = append(opportunities, CostReductionOpportunity{
					Type:             "lab_optimization",
					Description:      "Multiple lab tests could be bundled for cost savings",
					PotentialSavings: expense.Amount * 0.15,
					Recommendation:   "Schedule multiple lab tests together to reduce facility fees",
				})
			}
		}

		// Check for high out-of-pocket costs suggesting insurance gaps
		if expense.OutOfPocket > expense.Amount*0.8 { // >80% out-of-pocket
			opportunities = append(opportunities, CostReductionOpportunity{
				Type:             "insurance_gap",
				Description:      "High out-of-pocket expense suggests insurance coverage gap",
				PotentialSavings: expense.OutOfPocket * 0.3,
				Recommendation:   "Review insurance benefits or consider supplemental coverage",
			})
		}
	}

	// Suggest preventive care if missing
	if !hasPreventiveCare && totalExpenses > 1000.0 {
		opportunities = append(opportunities, CostReductionOpportunity{
			Type:             "preventive_care",
			Description:      "Lack of preventive care may lead to higher future costs",
			PotentialSavings: totalExpenses * 0.2, // Preventive care can reduce future costs
			Recommendation:   "Schedule annual physical and preventive screenings to catch issues early",
		})
	}

	// Medication cost optimization
	if medicationExpenses > totalExpenses*0.3 { // >30% on medications
		opportunities = append(opportunities, CostReductionOpportunity{
			Type:             "medication_review",
			Description:      "High medication costs warrant comprehensive review",
			PotentialSavings: medicationExpenses * 0.25,
			Recommendation:   "Request medication review with pharmacist and explore patient assistance programs",
		})
	}

	// Healthcare shopping for expensive procedures
	if totalExpenses > 3000.0 {
		opportunities = append(opportunities, CostReductionOpportunity{
			Type:             "healthcare_shopping",
			Description:      "High medical expenses could benefit from price comparison",
			PotentialSavings: totalExpenses * 0.15,
			Recommendation:   "Compare prices across providers for non-emergency procedures and services",
		})
	}

	return opportunities
}

// AnalyzeTrends identifies cost increases over time and spending patterns
func (m *medicalCostAnalyzer) AnalyzeTrends(expenses []domain.MedicalExpense) []string {
	trends := make([]string, 0)

	if len(expenses) < 2 {
		return trends
	}

	// Group expenses by category and track totals
	categoryTotals := make(map[string]float64)
	for _, expense := range expenses {
		categoryTotals[expense.Category] += expense.Amount
	}

	// Identify dominant categories
	totalExpenses := 0.0
	for _, amount := range categoryTotals {
		totalExpenses += amount
	}

	for category, amount := range categoryTotals {
		percentage := (amount / totalExpenses) * 100
		if percentage > 40 {
			trends = append(trends, fmt.Sprintf("High concentration in %s category (%.1f%% of total costs)", category, percentage))
		}
	}

	// Check for high out-of-pocket ratios
	totalOutOfPocket := 0.0
	for _, expense := range expenses {
		totalOutOfPocket += expense.OutOfPocket
	}

	outOfPocketRatio := (totalOutOfPocket / totalExpenses) * 100
	if outOfPocketRatio > 60 {
		trends = append(trends, fmt.Sprintf("High out-of-pocket burden (%.1f%% of total costs)", outOfPocketRatio))
	}

	// Check for recurring vs one-time expense patterns
	recurringTotal := 0.0
	oneTimeTotal := 0.0
	for _, expense := range expenses {
		if expense.IsRecurring {
			recurringTotal += m.normalizeToMonthly(expense) * 12
		} else {
			oneTimeTotal += expense.Amount
		}
	}

	if recurringTotal > oneTimeTotal*2 {
		trends = append(trends, "Recurring expenses dominate - consider long-term cost management strategies")
	} else if oneTimeTotal > recurringTotal*2 {
		trends = append(trends, "High one-time expenses - may indicate emergency care or delayed treatment")
	}

	// Medication dependency analysis
	medicationTotal := categoryTotals["medication"]
	if medicationTotal > totalExpenses*0.4 {
		trends = append(trends, "High medication dependency - explore cost reduction strategies")
	}

	return trends
}

// normalizeToMonthly converts expense amount to monthly equivalent
func (m *medicalCostAnalyzer) normalizeToMonthly(expense domain.MedicalExpense) float64 {
	if !expense.IsRecurring {
		return 0 // One-time expenses don't contribute to monthly recurring
	}

	switch expense.Frequency {
	case "daily":
		return expense.Amount * 30 // Approximate month
	case "weekly":
		return expense.Amount * 4.33 // Approximate weeks per month
	case "bi-weekly":
		return expense.Amount * 2.17 // Approximate bi-weeks per month
	case "monthly":
		return expense.Amount
	case "quarterly":
		return expense.Amount / 3
	case "semi-annually":
		return expense.Amount / 6
	case "annually":
		return expense.Amount / 12
	default: // "one_time" and others
		return 0
	}
}

// calculateRecurringAnnualCost calculates annual cost for a medical expense based on frequency
func (m *medicalCostAnalyzer) calculateRecurringAnnualCost(expense domain.MedicalExpense) float64 {
	if !expense.IsRecurring {
		return expense.Amount // One-time cost
	}

	return m.normalizeToMonthly(expense) * 12
}

// categorizeExpenseRisk categorizes an expense based on its cost and frequency
func (m *medicalCostAnalyzer) categorizeExpenseRisk(expense domain.MedicalExpense) string {
	annualCost := m.calculateRecurringAnnualCost(expense)

	switch {
	case annualCost >= 10000:
		return "high"
	case annualCost >= 3000:
		return "moderate"
	case annualCost >= 1000:
		return "low"
	default:
		return "minimal"
	}
}
