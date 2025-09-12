package services

import (
	"fmt"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// insuranceEvaluator implements the InsuranceEvaluator interface
type insuranceEvaluator struct{}

// NewInsuranceEvaluator creates a new insurance evaluator instance
func NewInsuranceEvaluator() InsuranceEvaluator {
	return &insuranceEvaluator{}
}

// CalculateCoverage applies deductible, then percentage, respects max out-of-pocket
func (i *insuranceEvaluator) CalculateCoverage(policy *domain.InsurancePolicy, expenseAmount float64) (*CoverageResult, error) {
	if expenseAmount <= 0 {
		return nil, fmt.Errorf("expense amount must be positive")
	}

	if !policy.IsActive {
		return &CoverageResult{
			TotalCovered:       0,
			PatientPays:        expenseAmount,
			DeductibleApplied:  0,
			CopayAmount:        expenseAmount,
			NewDeductibleMet:   policy.DeductibleMet,
			NewOutOfPocketUsed: policy.OutOfPocketCurrent,
		}, nil
	}

	// Use domain logic for coverage calculation
	insuranceCoverage, outOfPocketAmount, newDeductibleMet := policy.CalculateCoverage(expenseAmount)

	// Calculate how much was applied to deductible
	deductibleApplied := 0.0
	remainingDeductible := policy.GetRemainingDeductible()
	if remainingDeductible > 0 {
		deductibleApplied = expenseAmount
		if deductibleApplied > remainingDeductible {
			deductibleApplied = remainingDeductible
		}
	}

	return &CoverageResult{
		TotalCovered:       insuranceCoverage,
		PatientPays:        outOfPocketAmount,
		DeductibleApplied:  deductibleApplied,
		CopayAmount:        outOfPocketAmount - deductibleApplied,
		NewDeductibleMet:   newDeductibleMet,
		NewOutOfPocketUsed: policy.OutOfPocketCurrent + outOfPocketAmount,
	}, nil
}

// EvaluateCoverageGaps identifies uncovered conditions and services
func (i *insuranceEvaluator) EvaluateCoverageGaps(policies []domain.InsurancePolicy, conditions []domain.MedicalCondition, expenses []domain.MedicalExpense) []CoverageGap {
	gaps := make([]CoverageGap, 0)

	// Check for missing policy types
	hasPolicyTypes := make(map[string]bool)
	for _, policy := range policies {
		if policy.IsActive {
			hasPolicyTypes[policy.Type] = true
		}
	}

	// Recommend missing essential policy types
	if !hasPolicyTypes["health"] {
		gaps = append(gaps, CoverageGap{
			Type:        "missing_health_insurance",
			Description: "No active health insurance policy found",
			RiskLevel:   "critical",
			Recommendation: "Obtain comprehensive health insurance coverage immediately",
			EstimatedExposure: 50000.0, // High exposure without health insurance
		})
	}

	if !hasPolicyTypes["dental"] {
		gaps = append(gaps, CoverageGap{
			Type:        "missing_dental_coverage",
			Description: "No dental insurance coverage",
			RiskLevel:   "moderate", 
			Recommendation: "Consider dental insurance for routine and emergency dental care",
			EstimatedExposure: 3000.0,
		})
	}

	if !hasPolicyTypes["vision"] {
		gaps = append(gaps, CoverageGap{
			Type:        "missing_vision_coverage",
			Description: "No vision insurance coverage",
			RiskLevel:   "low",
			Recommendation: "Consider vision insurance if you wear glasses or contacts",
			EstimatedExposure: 1000.0,
		})
	}

	// Analyze conditions for coverage gaps
	for _, condition := range conditions {
		if condition.IsActive {
			coveredByInsurance := false
			
			// Check if any active policy might cover this condition
			for _, policy := range policies {
				if policy.IsActive && policy.Type == "health" {
					coveredByInsurance = true
					break
				}
			}

			if !coveredByInsurance {
				riskLevel := "moderate"
				estimatedCost := 2000.0
				
				switch condition.Severity {
				case "critical":
					riskLevel = "critical"
					estimatedCost = 20000.0
				case "severe":
					riskLevel = "high"
					estimatedCost = 10000.0
				case "moderate":
					riskLevel = "moderate" 
					estimatedCost = 5000.0
				}

				gaps = append(gaps, CoverageGap{
					Type:        "uncovered_condition",
					Description: fmt.Sprintf("Condition '%s' may not be adequately covered", condition.Name),
					RiskLevel:   riskLevel,
					Recommendation: "Review insurance benefits for condition-specific coverage",
					EstimatedExposure: estimatedCost,
				})
			}
		}
	}

	// Analyze expenses for high out-of-pocket patterns
	totalExpenses := 0.0
	totalOutOfPocket := 0.0
	for _, expense := range expenses {
		totalExpenses += expense.Amount
		totalOutOfPocket += expense.OutOfPocket
	}

	if totalExpenses > 0 {
		outOfPocketRatio := totalOutOfPocket / totalExpenses
		if outOfPocketRatio > 0.6 { // >60% out-of-pocket
			gaps = append(gaps, CoverageGap{
				Type:        "high_out_of_pocket",
				Description: fmt.Sprintf("High out-of-pocket expenses (%.1f%% of total)", outOfPocketRatio*100),
				RiskLevel:   "high",
				Recommendation: "Review deductible levels and consider supplemental insurance",
				EstimatedExposure: totalOutOfPocket * 1.5, // Potential future exposure
			})
		}
	}

	return gaps
}

// RecommendPolicyAdjustments suggests improvements based on usage patterns
func (i *insuranceEvaluator) RecommendPolicyAdjustments(policies []domain.InsurancePolicy, expenses []domain.MedicalExpense) []PolicyRecommendation {
	recommendations := make([]PolicyRecommendation, 0)

	for _, policy := range policies {
		if !policy.IsActive {
			continue
		}

		// Analyze deductible utilization
		deductibleUtilization := policy.DeductibleMet / policy.Deductible
		
		if deductibleUtilization < 0.3 && policy.Deductible > 2000 {
			// Low deductible usage with high deductible - might save on premiums
			recommendations = append(recommendations, PolicyRecommendation{
				PolicyID:    policy.ID,
				Type:        "deductible_adjustment",
				Description: "Low deductible utilization suggests you could increase deductible to lower premiums",
				Impact:      "Lower monthly premiums, higher potential out-of-pocket costs",
				Savings:     policy.MonthlyPremium * 0.15 * 12, // Estimated 15% premium savings
			})
		} else if deductibleUtilization > 0.8 && policy.Deductible > 1000 {
			// High deductible usage - might benefit from lower deductible
			recommendations = append(recommendations, PolicyRecommendation{
				Type:        "deductible_reduction", 
				Description: "High deductible utilization suggests you might benefit from lower deductible",
				Impact:      "Higher monthly premiums, lower out-of-pocket costs",
				Savings:     (policy.Deductible - 500) * 0.7, // Potential out-of-pocket savings
			})
		}

		// Analyze out-of-pocket utilization
		outOfPocketUtilization := policy.OutOfPocketCurrent / policy.OutOfPocketMax
		
		if outOfPocketUtilization > 0.8 {
			recommendations = append(recommendations, PolicyRecommendation{
				PolicyID:    policy.ID,
				Type:        "supplemental_coverage",
				Description: "High out-of-pocket utilization suggests need for supplemental coverage",
				Impact:      "Reduced financial exposure for future medical expenses",
				Savings:     policy.OutOfPocketMax * 0.3, // Potential future savings
			})
		}

		// Analyze coverage percentage effectiveness
		if policy.CoveragePercentage < 70 {
			recommendations = append(recommendations, PolicyRecommendation{
				PolicyID:    policy.ID,
				Type:        "coverage_upgrade",
				Description: fmt.Sprintf("Low coverage percentage (%.0f%%) may result in high costs", policy.CoveragePercentage),
				Impact:      "Better coverage for major medical expenses",
				Savings:     0, // Would increase costs short-term but reduce long-term exposure
			})
		}
	}

	// Analyze expense patterns for policy recommendations
	medicationExpenses := 0.0
	for _, expense := range expenses {
		if expense.Category == "medication" {
			medicationExpenses += expense.Amount
		}
	}

	if medicationExpenses > 3000 { // High medication costs
		recommendations = append(recommendations, PolicyRecommendation{
			Type:        "prescription_coverage",
			Description: "High medication expenses suggest need for better prescription coverage",
			Impact:      "Reduced medication costs through better formulary coverage",
			Savings:     medicationExpenses * 0.3, // Potential 30% savings with better coverage
		})
	}

	return recommendations
}

// TrackDeductibleProgress updates deductible tracking as expenses are added
func (i *insuranceEvaluator) TrackDeductibleProgress(policy *domain.InsurancePolicy, newExpenseAmount float64) (*DeductibleUpdate, error) {
	if newExpenseAmount <= 0 {
		return nil, fmt.Errorf("expense amount must be positive")
	}

	if !policy.IsActive {
		return &DeductibleUpdate{
			ExpenseAmount:      newExpenseAmount,
			AmountToDeductible: 0,
			AmountToOutOfPocket: newExpenseAmount,
			NewDeductibleMet:   policy.DeductibleMet,
			NewOutOfPocketUsed: policy.OutOfPocketCurrent,
			DeductibleCompleted: policy.DeductibleMet >= policy.Deductible,
			OutOfPocketMaxReached: policy.OutOfPocketCurrent >= policy.OutOfPocketMax,
		}, nil
	}

	// Calculate coverage using domain logic
	insuranceCoverage, outOfPocketAmount, newDeductibleMet := policy.CalculateCoverage(newExpenseAmount)
	
	amountToDeductible := newDeductibleMet - policy.DeductibleMet
	newOutOfPocketUsed := policy.OutOfPocketCurrent + outOfPocketAmount

	return &DeductibleUpdate{
		ExpenseAmount:       newExpenseAmount,
		InsuranceCoverage:   insuranceCoverage,
		AmountToDeductible:  amountToDeductible,
		AmountToOutOfPocket: outOfPocketAmount,
		NewDeductibleMet:    newDeductibleMet,
		NewOutOfPocketUsed:  newOutOfPocketUsed,
		DeductibleCompleted: newDeductibleMet >= policy.Deductible,
		OutOfPocketMaxReached: newOutOfPocketUsed >= policy.OutOfPocketMax,
	}, nil
}

