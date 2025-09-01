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

func (s *DecisionService) WithLogger(l *slog.Logger) *DecisionService {
	if l != nil {
		s.logger = l
	}
	return s
}

// StartValidation begins the multi-step validation process
func (s *DecisionService) StartValidation(ctx context.Context, input models.DecisionInput, userID *uuid.UUID, sessionID uuid.UUID) (*ValidationResponse, error) {
	l := s.reqLogger(ctx).With("user_id", userID, "session_id", sessionID)
	l.Info("validation.start", "item", safeTrunc(input.ItemName, 80))

	// Create new validation state
	state := &ValidationState{
		ID:            uuid.New(),
		SessionID:     sessionID,
		UserID:        userID,
		CurrentStep:   StepFinancialProfile,
		OriginalInput: &input,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Save validation state
	if err := s.saveValidationState(ctx, state); err != nil {
		l.Error("validation.save_state.error", "err", err)
		return nil, fmt.Errorf("failed to save validation state: %w", err)
	}

	return &ValidationResponse{
		State:    state,
		NextStep: StepFinancialProfile,
		Questions: []string{
			"What is your monthly income after taxes?",
			"What are your fixed monthly charges (rent, utilities, insurance)?",
			"What is your total monthly loan payments?",
		},
		Complete: false,
	}, nil
}

// UpdateValidation processes a validation step
func (s *DecisionService) UpdateValidation(ctx context.Context, validationID uuid.UUID, stepData map[string]interface{}) (*ValidationResponse, error) {
	l := s.reqLogger(ctx).With("validation_id", validationID)
	l.Info("validation.update.start")

	// Get current validation state
	state, err := s.getValidationState(ctx, validationID)
	if err != nil {
		l.Error("validation.get_state.error", "err", err)
		return nil, fmt.Errorf("failed to get validation state: %w", err)
	}

	// Process current step
	nextStep, questions, err := s.processValidationStep(ctx, state, stepData)
	if err != nil {
		l.Error("validation.process_step.error", "err", err, "current_step", state.CurrentStep)
		return nil, fmt.Errorf("failed to process validation step: %w", err)
	}

	// Update state
	state.CurrentStep = nextStep
	state.UpdatedAt = time.Now()

	// Save updated state
	if err := s.saveValidationState(ctx, state); err != nil {
		l.Error("validation.save_updated_state.error", "err", err)
		return nil, fmt.Errorf("failed to save updated validation state: %w", err)
	}

	response := &ValidationResponse{
		State:    state,
		NextStep: nextStep,
		Questions: questions,
		Complete: nextStep == StepComplete,
	}

	// If validation is complete, generate final decision
	if nextStep == StepComplete {
		decision, err := s.makeEnhancedDecision(ctx, state)
		if err != nil {
			l.Error("validation.make_decision.error", "err", err)
			return nil, fmt.Errorf("failed to make enhanced decision: %w", err)
		}
		response.Decision = decision
	}

	return response, nil
}

// processValidationStep handles the logic for each validation step
func (s *DecisionService) processValidationStep(ctx context.Context, state *ValidationState, stepData map[string]interface{}) (ValidationStep, []string, error) {
	switch state.CurrentStep {
	case StepFinancialProfile:
		return s.processFinancialProfile(state, stepData)
	case StepDebtAssessment:
		return s.processDebtAssessment(state, stepData)
	case StepHealthContext:
		return s.processHealthContext(state, stepData)
	case StepTransportation:
		return s.processTransportation(state, stepData)
	case StepPurchaseReason:
		return s.processPurchaseReason(state, stepData)
	default:
		return StepComplete, nil, nil
	}
}

func (s *DecisionService) processFinancialProfile(state *ValidationState, data map[string]interface{}) (ValidationStep, []string, error) {
	financial := &FinancialValidation{}
	
	if income, ok := data["income"].(float64); ok {
		financial.Income = income
	}
	if charges, ok := data["fixed_charges"].(float64); ok {
		financial.FixedCharges = charges
	}
	if loans, ok := data["monthly_loans"].(float64); ok {
		financial.MonthlyLoans = loans
	}
	if confirmed, ok := data["confirmed"].(bool); ok {
		financial.Confirmed = confirmed
	}

	state.FinancialValidation = financial

	if !financial.Confirmed {
		return StepFinancialProfile, []string{
			"Please confirm your financial information is accurate.",
		}, nil
	}

	return StepDebtAssessment, []string{
		"Please provide details about each of your loans:",
		"- Type of loan (mortgage, car, personal, etc.)",
		"- Monthly payment amount",
		"- Remaining balance",
		"- Interest rate (if known)",
	}, nil
}

func (s *DecisionService) processDebtAssessment(state *ValidationState, data map[string]interface{}) (ValidationStep, []string, error) {
	assessment := &DebtAssessment{}
	
	if loansData, ok := data["loans"].([]interface{}); ok {
		for _, loanData := range loansData {
			if loanMap, ok := loanData.(map[string]interface{}); ok {
				loan := LoanInfo{}
				if lType, ok := loanMap["type"].(string); ok {
					loan.Type = lType
				}
				if payment, ok := loanMap["monthly_payment"].(float64); ok {
					loan.MonthlyPayment = payment
				}
				if remaining, ok := loanMap["remaining_amount"].(float64); ok {
					loan.RemainingAmount = remaining
				}
				if rate, ok := loanMap["interest_rate"].(float64); ok {
					loan.InterestRate = rate
				}
				assessment.Loans = append(assessment.Loans, loan)
			}
		}
	}
	
	if confirmed, ok := data["confirmed"].(bool); ok {
		assessment.Confirmed = confirmed
	}

	state.DebtAssessment = assessment

	if !assessment.Confirmed {
		return StepDebtAssessment, []string{
			"Please confirm your debt information is complete and accurate.",
		}, nil
	}

	return StepHealthContext, []string{
		"Let's understand your household's health context:",
		"- How many people in your household?",
		"- Any special medical needs or ongoing conditions?",
		"- Approximate monthly healthcare costs?",
	}, nil
}

func (s *DecisionService) processHealthContext(state *ValidationState, data map[string]interface{}) (ValidationStep, []string, error) {
	health := &HealthContext{}
	
	if size, ok := data["household_size"].(float64); ok {
		health.HouseholdSize = int(size)
	}
	if needs, ok := data["special_needs"].(bool); ok {
		health.SpecialNeeds = needs
	}
	if conditions, ok := data["medical_conditions"].([]interface{}); ok {
		for _, condition := range conditions {
			if condStr, ok := condition.(string); ok {
				health.MedicalConditions = append(health.MedicalConditions, condStr)
			}
		}
	}
	if costs, ok := data["monthly_health_costs"].(float64); ok {
		health.MonthlyHealthCosts = costs
	}
	if confirmed, ok := data["confirmed"].(bool); ok {
		health.Confirmed = confirmed
	}

	state.HealthContext = health

	if !health.Confirmed {
		return StepHealthContext, []string{
			"Please confirm your household health information.",
		}, nil
	}

	return StepTransportation, []string{
		"Tell us about your transportation:",
		"- What's your main mode of transport?",
		"- If you have a car: brand, model, year, kilometers, engine type?",
		"- What's the general condition of your vehicle?",
		"- Any maintenance needed and estimated cost?",
		"- Monthly transportation costs?",
	}, nil
}

func (s *DecisionService) processTransportation(state *ValidationState, data map[string]interface{}) (ValidationStep, []string, error) {
	transport := &TransportationInfo{}
	
	if mainTransport, ok := data["main_transport"].(string); ok {
		transport.MainTransport = mainTransport
	}
	
	if vehicleData, ok := data["vehicle_info"].(map[string]interface{}); ok {
		vehicle := &VehicleInfo{}
		if brand, ok := vehicleData["brand"].(string); ok {
			vehicle.Brand = brand
		}
		if model, ok := vehicleData["model"].(string); ok {
			vehicle.Model = model
		}
		if year, ok := vehicleData["year"].(float64); ok {
			vehicle.Year = int(year)
		}
		if km, ok := vehicleData["kilometers"].(float64); ok {
			vehicle.Kilometers = int(km)
		}
		if engine, ok := vehicleData["engine"].(string); ok {
			vehicle.Engine = engine
		}
		if state, ok := vehicleData["general_state"].(string); ok {
			vehicle.GeneralState = state
		}
		if needed, ok := vehicleData["maintenance_needed"].(bool); ok {
			vehicle.MaintenanceNeeded = needed
		}
		if cost, ok := vehicleData["estimated_maintenance_cost"].(float64); ok {
			vehicle.EstimatedMaintenanceCost = cost
		}
		transport.VehicleInfo = vehicle
	}
	
	if costs, ok := data["monthly_transport_costs"].(float64); ok {
		transport.MonthlyTransportCosts = costs
	}
	if confirmed, ok := data["confirmed"].(bool); ok {
		transport.Confirmed = confirmed
	}

	state.TransportationInfo = transport

	if !transport.Confirmed {
		return StepTransportation, []string{
			"Please confirm your transportation information.",
		}, nil
	}

	return StepPurchaseReason, []string{
		"Finally, let's understand why you want to make this purchase:",
		"- What practical need does this fulfill?",
		"- Is there an emotional reason for wanting this?",
		"- How urgent is this purchase?",
		"- What alternatives have you considered?",
	}, nil
}

func (s *DecisionService) processPurchaseReason(state *ValidationState, data map[string]interface{}) (ValidationStep, []string, error) {
	reason := &PurchaseReason{}
	
	if practical, ok := data["practical_need"].(string); ok {
		reason.PracticalNeed = practical
	}
	if emotional, ok := data["emotional_want"].(string); ok {
		reason.EmotionalWant = emotional
	}
	if urgency, ok := data["urgency"].(string); ok {
		reason.Urgency = urgency
	}
	if alternatives, ok := data["alternatives"].(string); ok {
		reason.Alternatives = alternatives
	}
	if confirmed, ok := data["confirmed"].(bool); ok {
		reason.Confirmed = confirmed
	}

	state.PurchaseReason = reason

	if !reason.Confirmed {
		return StepPurchaseReason, []string{
			"Please confirm your purchase reasoning information.",
		}, nil
	}

	return StepComplete, nil, nil
}

// makeEnhancedDecision creates a comprehensive decision using all validation data
func (s *DecisionService) makeEnhancedDecision(ctx context.Context, state *ValidationState) (*models.DecisionResponse, error) {
	l := s.reqLogger(ctx).With("validation_id", state.ID)
	l.Info("enhanced_decision.start")

	if state.OriginalInput == nil {
		return nil, fmt.Errorf("original input missing from validation state")
	}

	// Convert currency if needed
	costInUSD, err := s.fx.ConvertToUSD(state.OriginalInput.Price, state.OriginalInput.Currency)
	if err != nil {
		return nil, fmt.Errorf("failed to convert currency: %w", err)
	}

	// Generate comprehensive prompt for AI
	prompt := s.buildComprehensivePrompt(state, costInUSD)
	
	// Get AI analysis
	startAI := time.Now()
	aiAnalysis, err := s.ai.AnalyzeComprehensiveDecision(ctx, prompt)
	durAI := time.Since(startAI)
	
	if err != nil {
		l.Warn("enhanced_decision.ai_analysis.error", "err", err, "duration_ms", durAI.Milliseconds())
		// Fallback to basic decision if AI fails
		return s.makeBasicDecision(ctx, *state.OriginalInput, state.UserID, state.SessionID)
	}

	l.Info("enhanced_decision.ai_analysis.ok", "duration_ms", durAI.Milliseconds())

	// Parse AI response and create decision
	verdict, score, reasons := s.parseAIAnalysis(aiAnalysis)

	// Create and save decision
	decision := &models.Decision{
		ID:        uuid.New(),
		UserID:    state.UserID,
		SessionID: state.SessionID,
		Input:     *state.OriginalInput,
		Verdict:   verdict,
		Score:     score,
		Reasons:   models.ReasonsJSON(reasons),
		Cost:      state.OriginalInput.Price,
		Currency:  state.OriginalInput.Currency,
		CreatedAt: time.Now(),
	}

	if err := s.saveDecision(ctx, decision); err != nil {
		return nil, fmt.Errorf("failed to save enhanced decision: %w", err)
	}

	// Generate flip-to-yes suggestion
	flipToYes := s.generateEnhancedFlipToYes(state, verdict, reasons)

	// Get remaining credits
	creditsLeft := 0
	if state.UserID != nil {
		if c, err := s.getUserCreditsRemaining(ctx, *state.UserID); err == nil {
			creditsLeft = c
		}
	}

	return &models.DecisionResponse{
		Decision:    decision,
		FlipToYes:   flipToYes,
		CreditsLeft: creditsLeft,
	}, nil
}

func (s *DecisionService) buildComprehensivePrompt(state *ValidationState, costInUSD float64) string {
	prompt := fmt.Sprintf(`
Analyze this purchase decision comprehensively:

PURCHASE DETAILS:
- Item: %s
- Price: %.2f %s (%.2f USD)
- Description: %s
- Marked as necessity: %t

FINANCIAL PROFILE:`, 
		state.OriginalInput.ItemName,
		state.OriginalInput.Price,
		state.OriginalInput.Currency,
		costInUSD,
		state.OriginalInput.Description,
		state.OriginalInput.IsNecessity)

	if state.FinancialValidation != nil {
		prompt += fmt.Sprintf(`
- Monthly Income: %.2f
- Fixed Charges: %.2f
- Monthly Loans: %.2f
- Net Available: %.2f`, 
			state.FinancialValidation.Income,
			state.FinancialValidation.FixedCharges,
			state.FinancialValidation.MonthlyLoans,
			state.FinancialValidation.Income - state.FinancialValidation.FixedCharges - state.FinancialValidation.MonthlyLoans)
	}

	if state.DebtAssessment != nil && len(state.DebtAssessment.Loans) > 0 {
		prompt += "\n\nDEBT BREAKDOWN:"
		for _, loan := range state.DebtAssessment.Loans {
			prompt += fmt.Sprintf("\n- %s: %.2f/month, %.2f remaining", loan.Type, loan.MonthlyPayment, loan.RemainingAmount)
		}
	}

	if state.HealthContext != nil {
		prompt += fmt.Sprintf(`

HOUSEHOLD CONTEXT:
- Household Size: %d
- Special Health Needs: %t
- Monthly Health Costs: %.2f`, 
			state.HealthContext.HouseholdSize,
			state.HealthContext.SpecialNeeds,
			state.HealthContext.MonthlyHealthCosts)
		
		if len(state.HealthContext.MedicalConditions) > 0 {
			prompt += "\n- Medical Conditions: " + fmt.Sprintf("%v", state.HealthContext.MedicalConditions)
		}
	}

	if state.TransportationInfo != nil {
		prompt += fmt.Sprintf("\n\nTRANSPORTATION:\n- Main Transport: %s\n- Monthly Transport Costs: %.2f", 
			state.TransportationInfo.MainTransport,
			state.TransportationInfo.MonthlyTransportCosts)
		
		if state.TransportationInfo.VehicleInfo != nil {
			v := state.TransportationInfo.VehicleInfo
			prompt += fmt.Sprintf(`
- Vehicle: %s %s (%d)
- Kilometers: %d
- Engine: %s
- Condition: %s
- Maintenance Needed: %t`, 
				v.Brand, v.Model, v.Year, v.Kilometers, v.Engine, v.GeneralState, v.MaintenanceNeeded)
			
			if v.MaintenanceNeeded {
				prompt += fmt.Sprintf("\n- Estimated Maintenance Cost: %.2f", v.EstimatedMaintenanceCost)
			}
		}
	}

	if state.PurchaseReason != nil {
		prompt += fmt.Sprintf(`

PURCHASE REASONING:
- Practical Need: %s
- Emotional Want: %s
- Urgency: %s
- Alternatives Considered: %s`, 
			state.PurchaseReason.PracticalNeed,
			state.PurchaseReason.EmotionalWant,
			state.PurchaseReason.Urgency,
			state.PurchaseReason.Alternatives)
	}

	prompt += `

Please analyze this purchase and provide:
1. A verdict (YES/NO)
2. A confidence score (0.0-1.0)
3. Detailed reasoning considering financial health, household needs, transportation requirements, and purchase motivation
4. Specific recommendations or alternatives if applicable

Format your response as:
VERDICT: [YES/NO]
SCORE: [0.0-1.0]
REASONING: [Detailed analysis]`

	return prompt
}

func (s *DecisionService) parseAIAnalysis(analysis string) (models.DecisionVerdict, float64, []string) {
	// This would parse the AI response
	// For now, implementing basic parsing logic
	// You should enhance this based on your AI response format
	
	if analysis == "" {
		return models.VerdictNo, 0.3, []string{"AI analysis unavailable, defaulting to conservative recommendation"}
	}

	// Basic parsing - enhance this based on your AI's response format
	verdict := models.VerdictNo
	score := 0.5
	reasons := []string{analysis} // Simplified - you'd want to parse this properly

	return verdict, score, reasons
}

func (s *DecisionService) generateEnhancedFlipToYes(state *ValidationState, verdict models.DecisionVerdict, reasons []string) *string {
	if verdict == models.VerdictYes {
		return nil
	}

	// Generate contextual suggestions based on validation data
	suggestions := []string{}
	
	if state.DebtAssessment != nil && len(state.DebtAssessment.Loans) > 0 {
		suggestions = append(suggestions, "Consider paying down high-interest debt first")
	}
	
	if state.TransportationInfo != nil && state.TransportationInfo.VehicleInfo != nil && state.TransportationInfo.VehicleInfo.MaintenanceNeeded {
		suggestions = append(suggestions, fmt.Sprintf("Address vehicle maintenance (%.0f) before new purchases", state.TransportationInfo.VehicleInfo.EstimatedMaintenanceCost))
	}
	
	if state.HealthContext != nil && state.HealthContext.SpecialNeeds {
		suggestions = append(suggestions, "Build health emergency fund before discretionary purchases")
	}
	
	if state.PurchaseReason != nil && state.PurchaseReason.Urgency == "can_wait" {
		suggestions = append(suggestions, "Wait 30 days and save specifically for this item")
	}

	if len(suggestions) == 0 {
		suggestion := fmt.Sprintf("Save for 2-3 months to build a buffer before purchasing")
		return &suggestion
	}

	suggestion := suggestions[0]
	return &suggestion
}

// Database operations for validation state
func (s *DecisionService) saveValidationState(ctx context.Context, state *ValidationState) error {
	l := s.reqLogger(ctx).With("validation_id", state.ID)
	
	stateJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal validation state: %w", err)
	}

	query := `
		INSERT INTO validation_states (id, session_id, user_id, current_step, state_json, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			current_step = EXCLUDED.current_step,
			state_json = EXCLUDED.state_json,
			updated_at = EXCLUDED.updated_at
	`

	_, err = s.db.Exec(ctx, query,
		state.ID,
		state.SessionID,
		state.UserID,
		state.CurrentStep,
		stateJSON,
		state.CreatedAt,
		state.UpdatedAt,
	)

	if err != nil {
		l.Error("db.save_validation_state.error", "err", err)
		return err
	}

	l.Info("db.save_validation_state.ok")
	return nil
}

func (s *DecisionService) getValidationState(ctx context.Context, validationID uuid.UUID) (*ValidationState, error) {
	l := s.reqLogger(ctx).With("validation_id", validationID)

	query := `SELECT state_json FROM validation_states WHERE id = $1`
	
	var stateJSON []byte
	err := s.db.QueryRow(ctx, query, validationID).Scan(&stateJSON)
	if err != nil {
		l.Error("db.get_validation_state.error", "err", err)
		return nil, err
	}

	var state ValidationState
	if err := json.Unmarshal(stateJSON, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal validation state: %w", err)
	}

	l.Info("db.get_validation_state.ok")
	return &state, nil
}

// Fallback to existing basic decision logic
func (s *DecisionService) makeBasicDecision(ctx context.Context, input models.DecisionInput, userID *uuid.UUID, sessionID uuid.UUID) (*models.DecisionResponse, error) {
	return s.MakeDecision(ctx, input, userID, sessionID)
}

// Keep all existing methods for backward compatibility
func (s *DecisionService) MakeDecision(ctx context.Context, input models.DecisionInput, userID *uuid.UUID, sessionID uuid.UUID) (*models.DecisionResponse, error) {
	l := s.reqLogger(ctx).With("user_id", userID, "session_id", sessionID)
	l.Info("make_decision.start",
		"item", safeTrunc(input.ItemName, 80),
		"currency", input.Currency,
		"is_necessity", input.IsNecessity,
		"price", input.Price,
		"desc_len", len(input.Description),
	)

	// NEW: ensure the referenced session exists in DB
	if err := s.ensureSession(ctx, sessionID, userID); err != nil {
		l.Error("make_decision.ensure_session.error", "err", err)
		return nil, fmt.Errorf("failed to ensure session: %w", err)
	}

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

	// Check remaining credits (best-effort, don't block decision)
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

func (s *DecisionService) ensureSession(ctx context.Context, sessionID uuid.UUID, userID *uuid.UUID) error {
	l := s.reqLogger(ctx).With("session_id", sessionID, "user_id", userID)
	l.Info("db.ensure_session.start")

	anon := true
	var uid any = nil
	if userID != nil {
		anon = false
		uid = *userID
	}

	start := time.Now()
	_, err := s.db.Exec(ctx, `
		INSERT INTO sessions (id, user_id, anon)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
	`, sessionID, uid, anon)
	dur := time.Since(start)

	if err != nil {
		l.Error("db.ensure_session.error", "err", err, "duration_ms", dur.Milliseconds())
		return err
	}
	l.Info("db.ensure_session.ok", "duration_ms", dur.Milliseconds())
	return nil
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