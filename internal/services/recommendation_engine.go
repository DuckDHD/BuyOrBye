package services

import (
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// RecommendationEngine provides business rules fallback when AI fails
type RecommendationEngine struct{}

// NewRecommendationEngine creates a new recommendation engine
func NewRecommendationEngine() *RecommendationEngine {
	return &RecommendationEngine{}
}

// MakeFallbackDecision applies business rules when AI is unavailable
// Never rely solely on AI - always have business rules fallback
func (re *RecommendationEngine) MakeFallbackDecision(intent domain.PurchaseIntent, context domain.DecisionContext) *domain.DecisionOutcome {
	startTime := time.Now()
	
	// Apply business rules in priority order
	
	// Rule 1: Emergency fund < 3 months → BYE (except critical health)
	if context.FinancialContext.EmergencyFundMonths < 3.0 {
		if !(intent.Category == "health" && intent.Urgency == "critical") {
			return re.createDecision(intent, "BYE", 0.95,
				"Insufficient emergency fund (less than 3 months of expenses). Build emergency fund first.",
				[]string{
					"Build emergency fund to 3-6 months of expenses",
					"Focus on essential purchases only",
					"Consider if this purchase is truly necessary",
					"Look for lower-cost alternatives",
				}, 60, 0, startTime)
		}
	}
	
	// Rule 2: DTI > 50% → WAIT (90 days, except health)
	if context.FinancialContext.DebtToIncomeRatio > 0.5 {
		if intent.Category != "health" {
			return re.createDecision(intent, "WAIT", 0.90,
				"High debt-to-income ratio indicates financial stress. Wait and focus on debt reduction.",
				[]string{
					"Focus on debt reduction first",
					"Create a debt repayment plan", 
					"Consider debt consolidation options",
					"Revisit this purchase in 90 days",
				}, 90, intent.ItemCost*0.8, startTime)
		}
	}
	
	// Rule 3: Health + critical → BUY (health emergencies take priority)
	if intent.Category == "health" && intent.Urgency == "critical" {
		return re.createDecision(intent, "BUY", 0.98,
			"Critical health expenses take priority over financial constraints.",
			[]string{
				"Proceed with health purchase immediately",
				"Consider payment plans if available",
				"Review health insurance coverage",
				"Look into health savings account options",
			}, 0, intent.ItemCost*1.1, startTime)
	}
	
	// Rule 4: Cost > 30% disposable income → BYE
	monthlyImpact := intent.GetMonthlyImpact()
	disposableThreshold := context.FinancialContext.DisposableIncome * 0.3
	if monthlyImpact > disposableThreshold {
		return re.createDecision(intent, "BYE", 0.85,
			"Purchase exceeds 30% of disposable income, risking financial stability.",
			[]string{
				"Look for a less expensive alternative",
				"Save up for this purchase over time",
				"Consider if this purchase is truly necessary",
				fmt.Sprintf("Current monthly impact: $%.2f, recommended limit: $%.2f", monthlyImpact, disposableThreshold),
			}, 30, disposableThreshold*12, startTime)
	}
	
	// Rule 5: High debt + non-essential → WAIT
	if context.FinancialContext.DebtToIncomeRatio > 0.4 && !intent.IsEssential() {
		return re.createDecision(intent, "WAIT", 0.80,
			"High debt levels suggest waiting for non-essential purchases.",
			[]string{
				"Focus on debt reduction",
				"Wait until debt-to-income ratio improves",
				"Consider if this purchase can be delayed",
				"Look for ways to increase income",
			}, 60, intent.ItemCost*0.9, startTime)
	}
	
	// Rule 6: Low savings rate + expensive purchase → WAIT
	if context.FinancialContext.SavingsRate < 0.1 && intent.ItemCost > context.FinancialContext.MonthlyIncome*0.5 {
		return re.createDecision(intent, "WAIT", 0.75,
			"Low savings rate suggests building financial cushion before large purchases.",
			[]string{
				"Increase savings rate to at least 10%",
				"Build emergency fund first",
				"Consider smaller, more affordable options",
				"Wait until financial position improves",
			}, 45, intent.ItemCost*0.7, startTime)
	}
	
	// Rule 7: Recent excessive spending → WAIT
	recentSpending := context.GetTotalRecentSpending(30)
	if recentSpending > context.FinancialContext.MonthlyIncome*0.5 {
		return re.createDecision(intent, "WAIT", 0.70,
			"Recent high spending suggests taking a break to reassess budget.",
			[]string{
				"Review recent spending patterns",
				"Take a spending break for 30 days", 
				"Focus on essential purchases only",
				"Create a more structured budget",
			}, 30, intent.ItemCost*0.6, startTime)
	}
	
	// Rule 8: Otherwise → BUY (if passes all checks)
	if context.CanAfford(monthlyImpact, 0.25) && !context.IsFinanciallyStressed() {
		confidence := 0.65
		if intent.IsEssential() {
			confidence = 0.80
		}
		
		return re.createDecision(intent, "BUY", confidence,
			"Purchase appears affordable and fits within budget constraints.",
			[]string{
				"Proceed with purchase",
				"Shop around for the best price",
				"Consider timing for sales or discounts",
				"Monitor your budget after purchase",
			}, 0, intent.ItemCost*1.05, startTime)
	}
	
	// Default fallback: BYE (conservative approach)
	return re.createDecision(intent, "BYE", 0.60,
		"Unable to determine if purchase is advisable based on current financial situation.",
		[]string{
			"Review your financial situation",
			"Consider if this purchase is essential",
			"Look for lower-cost alternatives",
			"Wait until financial position is clearer",
		}, 14, intent.ItemCost*0.5, startTime)
}

// ValidateBusinessRules checks if business rules would block a decision
func (re *RecommendationEngine) ValidateBusinessRules(intent domain.PurchaseIntent, context domain.DecisionContext) []string {
	var warnings []string
	
	// Check emergency fund
	if context.FinancialContext.EmergencyFundMonths < 3.0 {
		if !(intent.Category == "health" && intent.Urgency == "critical") {
			warnings = append(warnings, "Emergency fund below 3 months")
		}
	}
	
	// Check debt-to-income ratio  
	if context.FinancialContext.DebtToIncomeRatio > 0.5 {
		warnings = append(warnings, "High debt-to-income ratio (>50%)")
	}
	
	// Check affordability
	monthlyImpact := intent.GetMonthlyImpact()
	if monthlyImpact > context.FinancialContext.DisposableIncome*0.3 {
		warnings = append(warnings, "Purchase exceeds 30% of disposable income")
	}
	
	// Check recent spending
	recentSpending := context.GetTotalRecentSpending(30)
	if recentSpending > context.FinancialContext.MonthlyIncome*0.5 {
		warnings = append(warnings, "High recent spending in last 30 days")
	}
	
	// Check savings rate
	if context.FinancialContext.SavingsRate < 0.1 {
		warnings = append(warnings, "Low savings rate (<10%)")
	}
	
	return warnings
}

// GetRecommendationStrength returns how strong the recommendation is (0-100)
func (re *RecommendationEngine) GetRecommendationStrength(intent domain.PurchaseIntent, context domain.DecisionContext, decision string) int {
	strength := 50 // Base strength
	
	// Increase strength for clear financial indicators
	if context.FinancialContext.EmergencyFundMonths >= 6.0 {
		strength += 20
	}
	if context.FinancialContext.DebtToIncomeRatio < 0.3 {
		strength += 15
	}
	if context.FinancialContext.SavingsRate > 0.15 {
		strength += 10
	}
	
	// Adjust based on decision type
	switch decision {
	case "BUY":
		if intent.IsEssential() {
			strength += 15
		}
		if intent.Urgency == "critical" {
			strength += 20
		}
	case "BYE":
		if context.IsFinanciallyStressed() {
			strength += 25
		}
		if intent.ItemCost > context.FinancialContext.MonthlyIncome {
			strength += 20
		}
	case "WAIT":
		if context.FinancialContext.DebtToIncomeRatio > 0.4 {
			strength += 15
		}
	}
	
	// Cap at 100
	if strength > 100 {
		strength = 100
	}
	
	return strength
}

// createDecision creates a standardized decision outcome
func (re *RecommendationEngine) createDecision(
	intent domain.PurchaseIntent,
	decision string,
	confidence float64,
	reason string,
	recommendations []string,
	waitPeriod int,
	maxBudget float64,
	startTime time.Time,
) *domain.DecisionOutcome {
	// Create decision factors based on the decision
	factors := re.createDecisionFactors(decision, confidence)
	
	return &domain.DecisionOutcome{
		UserID:          intent.UserID,
		IntentID:        intent.ID,
		Decision:        decision,
		Confidence:      confidence,
		PrimaryReason:   reason,
		Factors:         factors,
		Recommendations: recommendations,
		WaitPeriod:      waitPeriod,
		MaxBudget:       maxBudget,
		CreatedAt:       time.Now(),
		ProcessingTime:  time.Since(startTime).Milliseconds(),
	}
}

// createDecisionFactors creates appropriate factors for the decision
func (re *RecommendationEngine) createDecisionFactors(decision string, confidence float64) []domain.DecisionFactor {
	var factors []domain.DecisionFactor
	
	switch decision {
	case "BUY":
		factors = append(factors, domain.DecisionFactor{
			Category:    "financial",
			Impact:      "positive",
			Weight:      confidence,
			Description: "Purchase aligns with financial capacity and priorities",
		})
		factors = append(factors, domain.DecisionFactor{
			Category:    "practical",
			Impact:      "positive",
			Weight:      0.7,
			Description: "Business rule analysis supports purchase decision",
		})
		
	case "WAIT":
		factors = append(factors, domain.DecisionFactor{
			Category:    "timing",
			Impact:      "negative",
			Weight:      confidence,
			Description: "Current timing is not optimal for this purchase",
		})
		factors = append(factors, domain.DecisionFactor{
			Category:    "financial",
			Impact:      "negative",
			Weight:      0.8,
			Description: "Financial indicators suggest waiting for better conditions",
		})
		
	case "BYE":
		factors = append(factors, domain.DecisionFactor{
			Category:    "financial",
			Impact:      "negative",
			Weight:      confidence,
			Description: "Purchase exceeds financial constraints or risk tolerance",
		})
		factors = append(factors, domain.DecisionFactor{
			Category:    "practical",
			Impact:      "negative",
			Weight:      0.9,
			Description: "Business rule analysis recommends avoiding this purchase",
		})
	}
	
	return factors
}

// GetRulesSummary provides a summary of which rules apply to the decision
func (re *RecommendationEngine) GetRulesSummary(intent domain.PurchaseIntent, context domain.DecisionContext) string {
	rules := []string{}
	
	if context.FinancialContext.EmergencyFundMonths < 3.0 {
		rules = append(rules, "Emergency fund rule")
	}
	if context.FinancialContext.DebtToIncomeRatio > 0.5 {
		rules = append(rules, "High DTI rule") 
	}
	if intent.Category == "health" && intent.Urgency == "critical" {
		rules = append(rules, "Critical health rule")
	}
	
	monthlyImpact := intent.GetMonthlyImpact()
	if monthlyImpact > context.FinancialContext.DisposableIncome*0.3 {
		rules = append(rules, "Affordability rule")
	}
	
	if len(rules) == 0 {
		return "No major rules triggered"
	}
	
	return fmt.Sprintf("Applied rules: %v", rules)
}