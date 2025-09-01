package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/config"
	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/DuckDHD/BuyOrBye/pkg/ai"
	"github.com/DuckDHD/BuyOrBye/pkg/fx"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DecisionService struct {
	db     *pgxpool.Pool
	config *config.Config
	ai     *ai.Client
	fx     *fx.Client
	logger *slog.Logger
}

func NewDecisionService(db *pgxpool.Pool, cfg *config.Config) *DecisionService {
	return &DecisionService{
		db:     db,
		config: cfg,
		ai:     ai.NewClient(cfg.AIAPIKey),
		fx:     fx.NewClient(),
		logger: slog.Default(),
	}
}

// Optional helper if you want to inject a custom logger later.
func (s *DecisionService) WithLogger(l *slog.Logger) *DecisionService {
	if l != nil {
		s.logger = l
	}
	return s
}

func (s *DecisionService) MakeDecision(ctx context.Context, input models.DecisionInput, userID *uuid.UUID, sessionID uuid.UUID) (*models.DecisionResponse, error) {
	l := s.reqLogger(ctx).With(
		"user_id", userID,
		"session_id", sessionID,
	)
	l.Info("make_decision.start",
		"item", safeTrunc(input.ItemName, 80),
		"currency", input.Currency,
		"is_necessity", input.IsNecessity,
		"price", input.Price,
		"desc_len", len(input.Description),
	)

	// Get user profile if authenticated
	var profile *models.Profile
	if userID != nil {
		start := time.Now()
		p, err := s.getUserProfile(ctx, *userID)
		dur := time.Since(start)
		if err != nil {
			l.Error("make_decision.get_user_profile.error", "err", err, "duration_ms", dur.Milliseconds())
			return nil, fmt.Errorf("failed to get user profile: %w", err)
		}
		l.Info("make_decision.get_user_profile.ok", "duration_ms", dur.Milliseconds())
		profile = p
	}

	// Convert currency if needed
	startFX := time.Now()
	costInUSD, err := s.fx.ConvertToUSD(input.Price, input.Currency)
	durFX := time.Since(startFX)
	if err != nil {
		l.Error("make_decision.fx_convert.error", "err", err, "from_amount", input.Price, "from_currency", input.Currency, "duration_ms", durFX.Milliseconds())
		return nil, fmt.Errorf("failed to convert currency: %w", err)
	}
	l.Info("make_decision.fx_convert.ok",
		"from_amount", input.Price,
		"from_currency", input.Currency,
		"usd_amount", costInUSD,
		"duration_ms", durFX.Milliseconds(),
	)

	// Calculate decision using rules engine
	startCalc := time.Now()
	verdict, score, reasons := s.calculateDecision(input, profile, costInUSD)
	l.Info("make_decision.calculated",
		"verdict", verdict,
		"score", score,
		"reasons_count", len(reasons),
		"duration_ms", time.Since(startCalc).Milliseconds(),
	)

	// Generate AI explanation (best-effort)
	startAI := time.Now()
	aiReason, err := s.ai.ExplainDecision(ctx, input, verdict, reasons)
	durAI := time.Since(startAI)
	if err != nil {
		l.Warn("make_decision.ai_explain.error", "err", err, "duration_ms", durAI.Milliseconds())
	} else if aiReason != "" {
		reasons = append(reasons, aiReason)
		l.Info("make_decision.ai_explain.ok", "added_reason_len", len(aiReason), "duration_ms", durAI.Milliseconds())
	} else {
		l.Info("make_decision.ai_explain.empty", "duration_ms", durAI.Milliseconds())
	}

	// Save decision to database
	decision := &models.Decision{
		ID:        uuid.New(),
		UserID:    userID,
		SessionID: sessionID,
		Input:     input,
		Verdict:   verdict,
		Score:     score,
		Reasons:   models.ReasonsJSON(reasons),
		Cost:      input.Price,
		Currency:  input.Currency,
		CreatedAt: time.Now(),
	}

	startSave := time.Now()
	if err := s.saveDecision(ctx, decision); err != nil {
		l.Error("make_decision.save.error", "err", err, "duration_ms", time.Since(startSave).Milliseconds())
		return nil, fmt.Errorf("failed to save decision: %w", err)
	}
	l.Info("make_decision.save.ok",
		"decision_id", decision.ID,
		"duration_ms", time.Since(startSave).Milliseconds(),
	)

	// Generate flip-to-yes suggestion (if applicable)
	flipToYes := s.generateFlipToYes(input, verdict, reasons)
	if flipToYes != nil {
		l.Info("make_decision.flip_to_yes", "suggestion_len", len(*flipToYes))
	}

	// Check remaining credits (best-effort, don’t block decision)
	creditsLeft := 0
	if userID != nil {
		startCredits := time.Now()
		if c, err := s.getUserCreditsRemaining(ctx, *userID); err != nil {
			l.Warn("make_decision.credits_remaining.error", "err", err, "duration_ms", time.Since(startCredits).Milliseconds())
		} else {
			creditsLeft = c
			l.Info("make_decision.credits_remaining.ok", "credits_left", creditsLeft, "duration_ms", time.Since(startCredits).Milliseconds())
		}
	}

	l.Info("make_decision.done",
		"decision_id", decision.ID,
		"verdict", decision.Verdict,
		"score", decision.Score,
		"reasons_count", len(decision.Reasons),
		"credits_left", creditsLeft,
	)

	return &models.DecisionResponse{
		Decision:    decision,
		FlipToYes:   flipToYes,
		CreditsLeft: creditsLeft,
	}, nil
}

func (s *DecisionService) calculateDecision(input models.DecisionInput, profile *models.Profile, costInUSD float64) (models.DecisionVerdict, float64, []string) {
	reasons := []string{}
	score := 0.5 // Default neutral score

	// If no profile, use basic rules
	if profile == nil {
		if costInUSD > 100 {
			return models.VerdictNo, 0.2, []string{"Large purchase without financial profile - consider setting up your profile for better guidance"}
		}
		if input.IsNecessity {
			return models.VerdictYes, 0.8, []string{"Marked as necessity and reasonably priced"}
		}
		return models.VerdictYes, 0.6, []string{"Small purchase amount - likely safe to proceed"}
	}

	// Calculate financial health metrics
	monthlyIncome := profile.Income
	monthlyExpenses := profile.FixedExpenses + profile.DebtsMin
	monthlySurplus := monthlyIncome - monthlyExpenses
	emergencyBuffer := profile.SavingsLiquid

	// Rule 1: Check if purchase would cause negative balance
	if costInUSD > emergencyBuffer {
		reasons = append(reasons, "This would exceed your available savings")
		score -= 0.4
	}

	// Rule 2: Check surplus percentage
	if monthlySurplus > 0 {
		surplusPercentage := costInUSD / monthlySurplus
		if surplusPercentage > 0.4 {
			reasons = append(reasons, fmt.Sprintf("This represents %.1f%% of your monthly surplus", surplusPercentage*100))
			score -= 0.3
		}
	}

	// Rule 3: Emergency fund check
	monthlyFixed := profile.FixedExpenses
	if emergencyBuffer < monthlyFixed {
		reasons = append(reasons, "Your emergency fund is below one month of expenses")
		score -= 0.2
	}

	// Rule 4: Necessity bonus
	if input.IsNecessity {
		reasons = append(reasons, "Marked as necessity")
		score += 0.3
	}

	// Rule 5: Small purchase allowance
	if costInUSD < 20 {
		score += 0.2
	}

	// Determine verdict based on score
	if score >= 0.6 {
		return models.VerdictYes, score, append(reasons, "Financial situation supports this purchase")
	}

	return models.VerdictNo, score, append(reasons, "Consider waiting or finding alternatives")
}

func (s *DecisionService) generateFlipToYes(input models.DecisionInput, verdict models.DecisionVerdict, reasons []string) *string {
	if verdict == models.VerdictYes {
		return nil
	}

	// Generate suggestions to make it a "yes"
	suggestions := []string{
		fmt.Sprintf("Wait 10 days and reassess"),
		fmt.Sprintf("Look for a used option under %.0f %s", input.Price*0.7, input.Currency),
		fmt.Sprintf("Save up for 2-3 weeks first"),
	}

	suggestion := suggestions[0] // Simple logic for now
	return &suggestion
}

func (s *DecisionService) getUserProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	l := s.reqLogger(ctx).With("user_id", userID)
	l.Info("db.get_user_profile.start")

	query := `
		SELECT id, user_id, income, fixed_expenses, debts_min, 
		       savings_liquid, guardrails_json, updated_at
		FROM profiles 
		WHERE user_id = $1
	`

	var profile models.Profile
	var guardrailsJSON []byte

	start := time.Now()
	err := s.db.QueryRow(ctx, query, userID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.Income,
		&profile.FixedExpenses,
		&profile.DebtsMin,
		&profile.SavingsLiquid,
		&guardrailsJSON,
		&profile.UpdatedAt,
	)
	dur := time.Since(start)

	if err != nil {
		l.Error("db.get_user_profile.error", "err", err, "duration_ms", dur.Milliseconds())
		return nil, err
	}

	if guardrailsJSON != nil {
		if err := json.Unmarshal(guardrailsJSON, &profile.Guardrails); err != nil {
			l.Error("db.get_user_profile.unmarshal_guardrails.error", "err", err)
			return nil, fmt.Errorf("failed to unmarshal guardrails: %w", err)
		}
	}

	l.Info("db.get_user_profile.ok", "duration_ms", dur.Milliseconds())
	return &profile, nil
}

func (s *DecisionService) saveDecision(ctx context.Context, decision *models.Decision) error {
	l := s.reqLogger(ctx).With("decision_id", decision.ID, "user_id", decision.UserID)
	l.Info("db.save_decision.start")

	query := `
		INSERT INTO decisions (id, user_id, session_id, input_json, verdict, score, 
		                     reasons_json, cost, currency, created_at, shared)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	inputJSON, errInput := json.Marshal(decision.Input)
	if errInput != nil {
		l.Warn("db.save_decision.marshal_input.warn", "err", errInput)
	}
	reasonsJSON, errReasons := json.Marshal(decision.Reasons)
	if errReasons != nil {
		l.Warn("db.save_decision.marshal_reasons.warn", "err", errReasons, "reasons_count", len(decision.Reasons))
	}

	start := time.Now()
	_, err := s.db.Exec(ctx, query,
		decision.ID,
		decision.UserID,
		decision.SessionID,
		inputJSON,
		decision.Verdict,
		decision.Score,
		reasonsJSON,
		decision.Cost,
		decision.Currency,
		decision.CreatedAt,
		decision.Shared,
	)
	dur := time.Since(start)

	if err != nil {
		l.Error("db.save_decision.error", "err", err, "duration_ms", dur.Milliseconds())
		return err
	}

	l.Info("db.save_decision.ok", "duration_ms", dur.Milliseconds())
	return nil
}

func (s *DecisionService) getUserCreditsRemaining(ctx context.Context, userID uuid.UUID) (int, error) {
	l := s.reqLogger(ctx).With("user_id", userID)
	l.Info("db.credits_remaining.start")

	// Check if user is pro
	var isPro bool
	startPro := time.Now()
	err := s.db.QueryRow(ctx, "SELECT is_pro FROM users WHERE id = $1", userID).Scan(&isPro)
	durPro := time.Since(startPro)
	if err != nil {
		l.Error("db.credits_remaining.is_pro.error", "err", err, "duration_ms", durPro.Milliseconds())
		return 0, err
	}
	l.Info("db.credits_remaining.is_pro.ok", "is_pro", isPro, "duration_ms", durPro.Milliseconds())

	if isPro {
		l.Info("db.credits_remaining.unlimited", "value", 1000)
		return 1000, nil // Unlimited for pro users (effectively very high)
	}

	// Count decisions today for free users
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	var count int
	startCount := time.Now()
	err = s.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM decisions 
		WHERE user_id = $1 AND created_at >= $2 AND created_at < $3
	`, userID, today, tomorrow).Scan(&count)
	durCount := time.Since(startCount)

	if err != nil {
		l.Error("db.credits_remaining.count.error", "err", err, "duration_ms", durCount.Milliseconds())
		return 0, err
	}

	freeLimit := 3
	remaining := freeLimit - count
	if remaining < 0 {
		remaining = 0
	}

	l.Info("db.credits_remaining.ok",
		"count_today", count,
		"free_limit", freeLimit,
		"remaining", remaining,
		"duration_ms", durCount.Milliseconds(),
	)

	return remaining, nil
}

/* ===========================
 * Logging helpers
 * ===========================
 */

func (s *DecisionService) reqLogger(ctx context.Context) *slog.Logger {
	reqID := middleware.GetReqID(ctx)
	return s.logger.With("req_id", reqID)
}

func safeTrunc(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
