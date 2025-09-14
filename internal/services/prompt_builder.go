package services

import (
	"fmt"
	"strings"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

type PromptBuilder struct {
	maxTokens int
}

func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		maxTokens: 4000,
	}
}

func (pb *PromptBuilder) BuildPrompt(intent domain.PurchaseIntent, context domain.DecisionContext) (*domain.AIPrompt, error) {
	// Validate inputs
	if err := intent.Validate(); err != nil {
		return nil, fmt.Errorf("invalid purchase intent: %w", err)
	}

	// Build system context
	systemContext := pb.buildSystemContext()

	// Build user context
	userContext := pb.buildUserContext(context)

	// Build purchase details
	purchaseDetails := pb.buildPurchaseDetails(intent)

	// Build decision criteria
	decisionCriteria := pb.buildDecisionCriteria()

	// Build response format
	responseFormat := pb.buildResponseFormat()

	prompt := &domain.AIPrompt{
		SystemContext:    systemContext,
		UserContext:      userContext,
		PurchaseDetails:  purchaseDetails,
		DecisionCriteria: decisionCriteria,
		ResponseFormat:   responseFormat,
		MaxTokens:        500,
		Temperature:      0.7,
	}

	// Validate token limits
	if err := prompt.Validate(); err != nil {
		return nil, fmt.Errorf("prompt validation failed: %w", err)
	}

	// Check if we need to truncate due to token limits
	if prompt.EstimateTokens() > pb.maxTokens {
		prompt = pb.truncatePrompt(prompt)
	}

	// Sanitize PII
	prompt = pb.sanitizePrompt(prompt)

	return prompt, nil
}

func (pb *PromptBuilder) buildSystemContext() string {
	return `You are a financial advisor helping users make purchase decisions.
Analyze the context and respond with BUY, WAIT, or BYE.

Decision criteria:
- BUY: Purchase fits budget, won't harm stability
- WAIT: Delay for better timing (specify days)
- BYE: Not recommended given circumstances

Response format:
DECISION: [BUY/WAIT/BYE]
CONFIDENCE: [0.0-1.0]
REASON: [One sentence explanation]
RECOMMENDATIONS: 
• [Bullet point 1]
• [Bullet point 2]
• [Bullet point 3]

Key factors to consider:
- Emergency fund should be 3+ months of expenses
- Debt-to-income ratio should be under 40%
- Health emergencies take priority over financial rules
- Purchase should not exceed 30% of disposable income

Always be conservative with money decisions. When in doubt, recommend WAIT or BYE.`
}

func (pb *PromptBuilder) buildUserContext(context domain.DecisionContext) string {
	financial := context.FinancialContext
	health := context.HealthContext
	
	var contextParts []string

	// Financial context
	contextParts = append(contextParts, fmt.Sprintf(`Financial Situation:
- Monthly Income: $%.2f
- Monthly Expenses: $%.2f  
- Disposable Income: $%.2f
- Emergency Fund: %.1f months
- Debt-to-Income Ratio: %.1f%%
- Savings Rate: %.1f%%
- Financial Health: %s
- Budget Remaining: $%.2f`,
		financial.MonthlyIncome,
		financial.MonthlyExpenses,
		financial.DisposableIncome,
		financial.EmergencyFundMonths,
		financial.DebtToIncomeRatio*100,
		financial.SavingsRate*100,
		financial.FinancialHealth,
		financial.BudgetRemaining))

	// Health context
	if health.HealthRiskScore > 0 || health.HasCriticalConditions {
		contextParts = append(contextParts, fmt.Sprintf(`Health Context:
- Health Risk Score: %d/100
- Monthly Health Costs: $%.2f
- Insurance Coverage: %.0f%%
- Financial Vulnerability: %s
- Critical Conditions: %t
- Emergency Fund Needed: $%.2f`,
			health.HealthRiskScore,
			health.MonthlyHealthCosts,
			health.InsuranceCoverage*100,
			health.FinancialVulnerability,
			health.HasCriticalConditions,
			health.EmergencyFundNeeded))
	}

	// Recent spending patterns
	if len(context.PurchaseHistory) > 0 {
		recentSpending := context.GetTotalRecentSpending(30)
		contextParts = append(contextParts, fmt.Sprintf(`Recent Spending:
- Total spending last 30 days: $%.2f
- Number of purchases: %d`, recentSpending, len(context.PurchaseHistory)))
	}

	// Financial stress indicators
	if context.IsFinanciallyStressed() {
		contextParts = append(contextParts, "⚠️ Current financial stress indicators detected")
	}

	return strings.Join(contextParts, "\n\n")
}

func (pb *PromptBuilder) buildPurchaseDetails(intent domain.PurchaseIntent) string {
	monthlyImpact := intent.GetMonthlyImpact()
	
	return fmt.Sprintf(`Purchase Request:
- Item: %s
- Cost: $%.2f
- Category: %s
- Urgency: %s
- Frequency: %s
- Monthly Impact: $%.2f
- Purpose: %s
- Alternative Considered: %s
- Essential Purchase: %t`,
		intent.ItemName,
		intent.ItemCost,
		intent.Category,
		intent.Urgency,
		intent.Frequency,
		monthlyImpact,
		intent.Purpose,
		intent.Alternative,
		intent.IsEssential())
}

func (pb *PromptBuilder) buildDecisionCriteria() string {
	return `Decision Framework:
1. Emergency Fund Priority: Minimum 3 months expenses (6+ preferred)
2. Debt Management: Keep total DTI below 40% (50% max)
3. Essential Needs: Health, food, transportation take priority
4. Affordability: Purchase should not exceed 30% of disposable income
5. Timing: Consider market conditions and personal circumstances

Critical Override: Health emergencies may justify financial stretching
Wait Criteria: High debt, low emergency fund, or major upcoming expenses
Buy Criteria: Affordable, aligns with priorities, good financial health`
}

func (pb *PromptBuilder) buildResponseFormat() string {
	return `Respond ONLY with valid JSON in this exact format:
{
  "decision": "BUY|WAIT|BYE",
  "confidence": 0.0,
  "reasoning": "Clear explanation of the decision",
  "factors": [
    {
      "category": "financial|health|practical|timing",
      "impact": "positive|negative|neutral", 
      "weight": 0.0,
      "description": "Factor description"
    }
  ],
  "recommendations": [
    "Specific actionable advice"
  ],
  "wait_period": 0,
  "max_budget": 0.0
}

Ensure decision is exactly BUY, WAIT, or BYE.
Confidence must be between 0.0 and 1.0.
Include 2-5 relevant factors with appropriate weights.
Provide 2-4 specific, actionable recommendations.`
}

func (pb *PromptBuilder) truncatePrompt(prompt *domain.AIPrompt) *domain.AIPrompt {
	// If over token limit, truncate user context first
	currentTokens := prompt.EstimateTokens()
	if currentTokens <= pb.maxTokens {
		return prompt
	}

	// Calculate how much to truncate
	targetTokens := pb.maxTokens - 500 // Leave buffer
	reductionNeeded := currentTokens - targetTokens

	// Truncate user context proportionally
	userContextTokens := int(float64(len(strings.Fields(prompt.UserContext))) * 1.3) // Rough token estimation
	if userContextTokens > reductionNeeded {
		// Truncate to fit
		words := strings.Fields(prompt.UserContext)
		targetWords := int(float64(len(words)) * (1.0 - float64(reductionNeeded)/float64(userContextTokens)))
		if targetWords > 50 { // Keep minimum context
			truncatedWords := words[:targetWords]
			prompt.UserContext = strings.Join(truncatedWords, " ") + "... [context truncated for token limits]"
		}
	}

	return prompt
}

func (pb *PromptBuilder) sanitizePrompt(prompt *domain.AIPrompt) *domain.AIPrompt {
	// Create a copy to avoid modifying original
	sanitized := *prompt

	// Sanitize each field
	sanitized.SystemContext = pb.sanitizeString(sanitized.SystemContext)
	sanitized.UserContext = pb.sanitizeString(sanitized.UserContext)
	sanitized.PurchaseDetails = pb.sanitizeString(sanitized.PurchaseDetails)
	sanitized.DecisionCriteria = pb.sanitizeString(sanitized.DecisionCriteria)
	sanitized.ResponseFormat = pb.sanitizeString(sanitized.ResponseFormat)

	return &sanitized
}

func (pb *PromptBuilder) sanitizeString(input string) string {
	result := input

	// Remove SSN patterns (XXX-XX-XXXX)
	result = domain.SSNPattern.ReplaceAllString(result, "[SSN-REDACTED]")

	// Remove credit card patterns (16 digits with optional separators)
	result = domain.CreditCardPattern.ReplaceAllString(result, "[CARD-REDACTED]")

	// Remove phone number patterns
	result = domain.PhonePattern.ReplaceAllString(result, "[PHONE-REDACTED]")

	// Remove email patterns
	result = domain.EmailPattern.ReplaceAllString(result, "[EMAIL-REDACTED]")

	return result
}

func (pb *PromptBuilder) SetMaxTokens(tokens int) {
	if tokens >= 10 && tokens <= 4000 {
		pb.maxTokens = tokens
	}
}

// Public wrapper methods for testing
func (pb *PromptBuilder) BuildDecisionPrompt(intent domain.PurchaseIntent, context domain.DecisionContext) (*domain.AIPrompt, error) {
	return pb.BuildPrompt(intent, context)
}

func (pb *PromptBuilder) BuildSystemPrompt() string {
	return pb.buildSystemContext()
}

func (pb *PromptBuilder) BuildUserContext(context domain.DecisionContext) string {
	return pb.buildUserContext(context)
}

func (pb *PromptBuilder) BuildDecisionCriteria() string {
	return pb.buildDecisionCriteria()
}

func (pb *PromptBuilder) BuildResponseFormat() string {
	return pb.buildResponseFormat()
}