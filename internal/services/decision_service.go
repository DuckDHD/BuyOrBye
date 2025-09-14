package services

import (
	"context"
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

type LegacyDecisionRepository interface {
	Save(ctx context.Context, decision domain.DecisionOutcome) error
	GetByIntentID(ctx context.Context, intentID string) (*domain.DecisionOutcome, error)
	GetRecentDecisions(ctx context.Context, userID string, days int) ([]domain.PastDecision, error)
}

type FinanceServiceInterface interface {
	GetFinancialSnapshot(ctx context.Context, userID string) (*domain.FinancialSnapshot, error)
}

type HealthServiceInterface interface {
	GetHealthSnapshot(ctx context.Context, userID string) (*domain.HealthSnapshot, error)
}

type AIClient interface {
	GenerateDecision(ctx context.Context, prompt domain.AIPrompt) (*domain.AIResponse, error)
}

type DecisionService struct {
	repo           LegacyDecisionRepository
	financeService FinanceServiceInterface
	healthService  HealthServiceInterface
	aiClient       AIClient
	promptBuilder  *PromptBuilder
	interpreter    *DecisionInterpreter
}

func NewDecisionService(
	repo LegacyDecisionRepository,
	financeService FinanceServiceInterface,
	healthService HealthServiceInterface,
	aiClient AIClient,
	promptBuilder *PromptBuilder,
	interpreter *DecisionInterpreter,
) *DecisionService {
	return &DecisionService{
		repo:           repo,
		financeService: financeService,
		healthService:  healthService,
		aiClient:       aiClient,
		promptBuilder:  promptBuilder,
		interpreter:    interpreter,
	}
}

func (s *DecisionService) MakeDecision(ctx context.Context, intent domain.PurchaseIntent) (*domain.DecisionOutcome, error) {
	startTime := time.Now()

	// Validate intent
	if err := intent.Validate(); err != nil {
		return nil, fmt.Errorf("invalid purchase intent: %w", err)
	}

	// Build decision context
	context, err := s.buildDecisionContext(ctx, intent.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to build decision context: %w", err)
	}

	// Apply business rules first
	if outcome := s.applyBusinessRules(intent, *context); outcome != nil {
		outcome.ProcessingTime = time.Since(startTime).Milliseconds()
		
		// Save decision
		if err := s.repo.Save(ctx, *outcome); err != nil {
			return nil, fmt.Errorf("failed to save decision: %w", err)
		}
		
		return outcome, nil
	}

	// Fallback to AI for complex decisions
	outcome, err := s.makeAIDecision(ctx, intent, *context)
	if err != nil {
		return nil, fmt.Errorf("failed to make AI decision: %w", err)
	}

	outcome.ProcessingTime = time.Since(startTime).Milliseconds()

	// Save decision
	if err := s.repo.Save(ctx, *outcome); err != nil {
		return nil, fmt.Errorf("failed to save decision: %w", err)
	}

	return outcome, nil
}

func (s *DecisionService) buildDecisionContext(ctx context.Context, userID string) (*domain.DecisionContext, error) {
	// Get financial snapshot
	financial, err := s.financeService.GetFinancialSnapshot(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get financial snapshot: %w", err)
	}

	// Get health snapshot
	health, err := s.healthService.GetHealthSnapshot(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health snapshot: %w", err)
	}

	// Get recent decisions
	recentDecisions, err := s.repo.GetRecentDecisions(ctx, userID, 30)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent decisions: %w", err)
	}

	return &domain.DecisionContext{
		UserID:           userID,
		FinancialContext: *financial,
		HealthContext:    *health,
		TransportContext: domain.TransportSnapshot{
			HasVehicle:           true,  // Default values - would come from transport service
			MonthlyTransportCost: 200.0,
			PublicTransitAccess:  true,
			CommuteDistance:      15.0,
		},
		PurchaseHistory: recentDecisions,
		CurrentDate:     time.Now(),
	}, nil
}

func (s *DecisionService) applyBusinessRules(intent domain.PurchaseIntent, context domain.DecisionContext) *domain.DecisionOutcome {
	// Rule 1: Emergency Fund Check (< 3 months = BYE, except critical health)
	if context.FinancialContext.EmergencyFundMonths < 3.0 {
		if !(intent.Category == "health" && intent.Urgency == "critical") {
			return &domain.DecisionOutcome{
				UserID:        intent.UserID,
				IntentID:      intent.ID,
				Decision:      "BYE",
				Confidence:    0.9,
				PrimaryReason: "Insufficient emergency fund (less than 3 months of expenses). Focus on building emergency fund first.",
				Factors: []domain.DecisionFactor{
					{
						Category:    "financial",
						Impact:      "negative",
						Weight:      0.9,
						Description: fmt.Sprintf("Emergency fund covers only %.1f months", context.FinancialContext.EmergencyFundMonths),
					},
				},
				Recommendations: []string{
					"Build emergency fund to 3-6 months of expenses",
					"Consider if this purchase is truly essential",
					"Look for lower-cost alternatives",
				},
				WaitPeriod: 60,
				MaxBudget:  0,
				CreatedAt:  time.Now(),
			}
		}
	}

	// Rule 2: High Debt-to-Income Ratio (> 50% = WAIT, except health)
	if context.FinancialContext.DebtToIncomeRatio > 0.5 {
		if intent.Category != "health" {
			return &domain.DecisionOutcome{
				UserID:        intent.UserID,
				IntentID:      intent.ID,
				Decision:      "WAIT",
				Confidence:    0.85,
				PrimaryReason: "High debt-to-income ratio indicates financial stress. Wait and focus on debt reduction.",
				Factors: []domain.DecisionFactor{
					{
						Category:    "financial",
						Impact:      "negative",
						Weight:      0.8,
						Description: fmt.Sprintf("Debt-to-income ratio is %.1f%%", context.FinancialContext.DebtToIncomeRatio*100),
					},
				},
				Recommendations: []string{
					"Focus on debt reduction",
					"Create a debt repayment plan",
					"Consider debt consolidation options",
				},
				WaitPeriod: 90,
				MaxBudget:  intent.ItemCost * 0.7,
				CreatedAt:  time.Now(),
			}
		}
	}

	// Rule 3: Health Critical Override (health + critical = BUY)
	if intent.Category == "health" && intent.Urgency == "critical" {
		return &domain.DecisionOutcome{
			UserID:        intent.UserID,
			IntentID:      intent.ID,
			Decision:      "BUY",
			Confidence:    0.95,
			PrimaryReason: "Critical health expenses take priority over financial constraints.",
			Factors: []domain.DecisionFactor{
				{
					Category:    "health",
					Impact:      "positive",
					Weight:      1.0,
					Description: "Critical health purchase overrides financial concerns",
				},
			},
			Recommendations: []string{
				"Proceed with health purchase immediately",
				"Consider payment plans if available",
				"Review health insurance coverage",
			},
			WaitPeriod: 0,
			MaxBudget:  intent.ItemCost * 1.1,
			CreatedAt:  time.Now(),
		}
	}

	// Rule 4: Exceeds Disposable Income (> 30% = BYE)
	disposableThreshold := context.FinancialContext.DisposableIncome * 0.3
	if intent.GetMonthlyImpact() > disposableThreshold {
		return &domain.DecisionOutcome{
			UserID:        intent.UserID,
			IntentID:      intent.ID,
			Decision:      "BYE",
			Confidence:    0.8,
			PrimaryReason: "Purchase exceeds 30% of disposable income, risking financial stability.",
			Factors: []domain.DecisionFactor{
				{
					Category:    "financial",
					Impact:      "negative",
					Weight:      0.8,
					Description: fmt.Sprintf("Monthly impact $%.2f exceeds 30%% of disposable income $%.2f", intent.GetMonthlyImpact(), disposableThreshold),
				},
			},
			Recommendations: []string{
				"Look for a less expensive alternative",
				"Save up for this purchase over time",
				"Consider if this purchase is truly necessary",
			},
			WaitPeriod: 30,
			MaxBudget:  disposableThreshold * 12, // Annual budget limit
			CreatedAt:  time.Now(),
		}
	}

	// Rule 5: Affordable Purchase (within all limits = BUY)
	if context.CanAfford(intent.GetMonthlyImpact(), 0.3) && !context.IsFinanciallyStressed() {
		confidence := 0.7
		if intent.IsEssential() {
			confidence = 0.85
		}

		return &domain.DecisionOutcome{
			UserID:        intent.UserID,
			IntentID:      intent.ID,
			Decision:      "BUY",
			Confidence:    confidence,
			PrimaryReason: "Purchase is affordable and fits within budget constraints.",
			Factors: []domain.DecisionFactor{
				{
					Category:    "financial",
					Impact:      "positive",
					Weight:      0.7,
					Description: "Purchase fits comfortably within budget",
				},
			},
			Recommendations: []string{
				"Proceed with purchase",
				"Shop around for best price",
				"Consider timing for best deals",
			},
			WaitPeriod: 0,
			MaxBudget:  intent.ItemCost * 1.05,
			CreatedAt:  time.Now(),
		}
	}

	// No business rule matched - needs AI decision
	return nil
}

func (s *DecisionService) makeAIDecision(ctx context.Context, intent domain.PurchaseIntent, context domain.DecisionContext) (*domain.DecisionOutcome, error) {
	// Build AI prompt
	prompt, err := s.promptBuilder.BuildPrompt(intent, context)
	if err != nil {
		return nil, fmt.Errorf("failed to build AI prompt: %w", err)
	}

	// Get AI response
	aiResponse, err := s.aiClient.GenerateDecision(ctx, *prompt)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI response: %w", err)
	}

	// Parse AI response
	outcome, err := s.interpreter.ParseResponse(*aiResponse, intent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return outcome, nil
}

func (s *DecisionService) GetDecisionHistory(ctx context.Context, userID string, days int) ([]domain.PastDecision, error) {
	decisions, err := s.repo.GetRecentDecisions(ctx, userID, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get decision history: %w", err)
	}

	return decisions, nil
}

func (s *DecisionService) GetDecisionByIntent(ctx context.Context, intentID string) (*domain.DecisionOutcome, error) {
	decision, err := s.repo.GetByIntentID(ctx, intentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get decision by intent ID: %w", err)
	}

	return decision, nil
}