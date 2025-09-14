package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/go-playground/validator/v10"
)

// PurchaseIntentDTO represents the request DTO for purchase decisions
type PurchaseIntentDTO struct {
	ID          string  `json:"id,omitempty"`
	ItemName    string  `json:"item_name" validate:"required,min=2,max=200"`
	ItemCost    float64 `json:"item_cost" validate:"gt=0,max=1000000"`
	Category    string  `json:"category" validate:"required,oneof=electronics clothing food transport health entertainment other"`
	Urgency     string  `json:"urgency" validate:"required,oneof=low medium high critical"`
	Frequency   string  `json:"frequency" validate:"required,oneof=one_time monthly yearly"`
	Purpose     string  `json:"purpose" validate:"max=500"`
	Alternative string  `json:"alternative" validate:"max=200"`
}

// DecisionResponseDTO represents the response DTO for purchase decisions
type DecisionResponseDTO struct {
	ID              string                `json:"id"`
	UserID          string                `json:"user_id"`
	IntentID        string                `json:"intent_id"`
	Decision        string                `json:"decision"`
	Confidence      float64               `json:"confidence"`
	PrimaryReason   string                `json:"primary_reason"`
	Factors         []DecisionFactorDTO   `json:"factors"`
	Recommendations []string              `json:"recommendations"`
	WaitPeriod      int                   `json:"wait_period"`
	MaxBudget       float64               `json:"max_budget"`
	CreatedAt       time.Time             `json:"created_at"`
	ProcessingTime  int64                 `json:"processing_time_ms"`
}

// DecisionFactorDTO represents a decision factor
type DecisionFactorDTO struct {
	Category    string  `json:"category"`
	Impact      string  `json:"impact"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
}

// DecisionHistoryDTO represents historical decision data with statistics
type DecisionHistoryDTO struct {
	UserID           string                    `json:"user_id"`
	TotalDecisions   int                       `json:"total_decisions"`
	RecentDecisions  []DecisionSummaryDTO      `json:"recent_decisions"`
	CategoryStats    []CategoryStatsDTO        `json:"category_stats"`
	DecisionPattern  map[string]int            `json:"decision_pattern"`
	TotalSpending    float64                   `json:"total_spending"`
	AverageConfidence float64                  `json:"average_confidence"`
	Period           string                    `json:"period"`
}

// DecisionSummaryDTO represents a simplified decision for history
type DecisionSummaryDTO struct {
	ID            string    `json:"id"`
	ItemName      string    `json:"item_name"`
	ItemCost      float64   `json:"item_cost"`
	Category      string    `json:"category"`
	Decision      string    `json:"decision"`
	Confidence    float64   `json:"confidence"`
	CreatedAt     time.Time `json:"created_at"`
	DaysAgo       int       `json:"days_ago"`
}

// CategoryStatsDTO represents statistics for a specific category
type CategoryStatsDTO struct {
	Category      string  `json:"category"`
	TotalItems    int     `json:"total_items"`
	BuyCount      int     `json:"buy_count"`
	WaitCount     int     `json:"wait_count"`
	ByeCount      int     `json:"bye_count"`
	TotalSpent    float64 `json:"total_spent"`
	AverageCost   float64 `json:"average_cost"`
	BuyRate       float64 `json:"buy_rate"`
}

// ToDomain converts PurchaseIntentDTO to domain.PurchaseIntent
func (dto *PurchaseIntentDTO) ToDomain() (*domain.PurchaseIntent, error) {
	// Validate the DTO
	validate := validator.New()
	if err := validate.Struct(dto); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Generate ID if not provided
	if dto.ID == "" {
		dto.ID = generateID()
	}

	return &domain.PurchaseIntent{
		ID:          dto.ID,
		ItemName:    strings.TrimSpace(dto.ItemName),
		ItemCost:    dto.ItemCost,
		Category:    strings.ToLower(dto.Category),
		Urgency:     strings.ToLower(dto.Urgency),
		Frequency:   strings.ToLower(dto.Frequency),
		Purpose:     strings.TrimSpace(dto.Purpose),
		Alternative: strings.TrimSpace(dto.Alternative),
		CreatedAt:   time.Now(),
	}, nil
}

// FromDomain creates PurchaseIntentDTO from domain.PurchaseIntent
func (dto *PurchaseIntentDTO) FromDomain(intent domain.PurchaseIntent) {
	dto.ID = intent.ID
	dto.ItemName = intent.ItemName
	dto.ItemCost = intent.ItemCost
	dto.Category = intent.Category
	dto.Urgency = intent.Urgency
	dto.Frequency = intent.Frequency
	dto.Purpose = intent.Purpose
	dto.Alternative = intent.Alternative
}

// FromDomain creates DecisionResponseDTO from domain.DecisionOutcome
func (dto *DecisionResponseDTO) FromDomain(outcome domain.DecisionOutcome) {
	dto.ID = outcome.ID
	dto.UserID = outcome.UserID
	dto.IntentID = outcome.IntentID
	dto.Decision = outcome.Decision
	dto.Confidence = outcome.Confidence
	dto.PrimaryReason = outcome.PrimaryReason
	dto.WaitPeriod = outcome.WaitPeriod
	dto.MaxBudget = outcome.MaxBudget
	dto.CreatedAt = outcome.CreatedAt
	dto.ProcessingTime = outcome.ProcessingTime

	// Convert factors
	dto.Factors = make([]DecisionFactorDTO, len(outcome.Factors))
	for i, factor := range outcome.Factors {
		dto.Factors[i] = DecisionFactorDTO{
			Category:    factor.Category,
			Impact:      factor.Impact,
			Weight:      factor.Weight,
			Description: factor.Description,
		}
	}

	// Copy recommendations
	dto.Recommendations = make([]string, len(outcome.Recommendations))
	copy(dto.Recommendations, outcome.Recommendations)
}

// ToDomain converts DecisionResponseDTO to domain.DecisionOutcome
func (dto *DecisionResponseDTO) ToDomain() (*domain.DecisionOutcome, error) {
	// Convert factors
	factors := make([]domain.DecisionFactor, len(dto.Factors))
	for i, factorDTO := range dto.Factors {
		factors[i] = domain.DecisionFactor{
			Category:    factorDTO.Category,
			Impact:      factorDTO.Impact,
			Weight:      factorDTO.Weight,
			Description: factorDTO.Description,
		}
	}

	// Copy recommendations
	recommendations := make([]string, len(dto.Recommendations))
	copy(recommendations, dto.Recommendations)

	outcome := &domain.DecisionOutcome{
		ID:              dto.ID,
		UserID:          dto.UserID,
		IntentID:        dto.IntentID,
		Decision:        dto.Decision,
		Confidence:      dto.Confidence,
		PrimaryReason:   dto.PrimaryReason,
		Factors:         factors,
		Recommendations: recommendations,
		WaitPeriod:      dto.WaitPeriod,
		MaxBudget:       dto.MaxBudget,
		CreatedAt:       dto.CreatedAt,
		ProcessingTime:  dto.ProcessingTime,
	}

	// Validate the domain object
	if err := outcome.Validate(); err != nil {
		return nil, fmt.Errorf("invalid decision outcome: %w", err)
	}

	return outcome, nil
}

// FromDomainHistory creates DecisionHistoryDTO from domain decision history
func (dto *DecisionHistoryDTO) FromDomainHistory(userID string, decisions []domain.PastDecision, period string) {
	dto.UserID = userID
	dto.Period = period
	dto.TotalDecisions = len(decisions)

	dto.RecentDecisions = make([]DecisionSummaryDTO, 0, len(decisions))
	
	totalSpending := 0.0
	totalConfidence := 0.0
	decisionPattern := map[string]int{"BUY": 0, "WAIT": 0, "BYE": 0}
	categoryMap := make(map[string]*CategoryStatsDTO)

	for _, decision := range decisions {
		// Add to recent decisions
		dto.RecentDecisions = append(dto.RecentDecisions, DecisionSummaryDTO{
			ItemName:   decision.ItemName,
			ItemCost:   decision.ItemCost,
			Category:   decision.Category,
			Decision:   decision.Decision,
			DaysAgo:    decision.DaysAgo,
		})

		// Update statistics
		if decision.Decision == "BUY" {
			totalSpending += decision.ItemCost
		}
		// Note: PastDecision doesn't have confidence, so we'll set a default
		totalConfidence += 0.7 // Default medium confidence for history
		decisionPattern[decision.Decision]++

		// Update category statistics
		if _, exists := categoryMap[decision.Category]; !exists {
			categoryMap[decision.Category] = &CategoryStatsDTO{
				Category: decision.Category,
			}
		}
		
		stats := categoryMap[decision.Category]
		stats.TotalItems++
		switch decision.Decision {
		case "BUY":
			stats.BuyCount++
			stats.TotalSpent += decision.ItemCost
		case "WAIT":
			stats.WaitCount++
		case "BYE":
			stats.ByeCount++
		}
	}

	// Finalize statistics
	dto.TotalSpending = totalSpending
	if dto.TotalDecisions > 0 {
		dto.AverageConfidence = totalConfidence / float64(dto.TotalDecisions)
	}
	dto.DecisionPattern = decisionPattern

	// Convert category map to slice and calculate rates
	dto.CategoryStats = make([]CategoryStatsDTO, 0, len(categoryMap))
	for _, stats := range categoryMap {
		if stats.TotalItems > 0 {
			stats.BuyRate = float64(stats.BuyCount) / float64(stats.TotalItems)
			if stats.BuyCount > 0 {
				stats.AverageCost = stats.TotalSpent / float64(stats.BuyCount)
			}
		}
		dto.CategoryStats = append(dto.CategoryStats, *stats)
	}
}

// Validate performs validation on PurchaseIntentDTO
func (dto *PurchaseIntentDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}

// Validate performs validation on DecisionResponseDTO
func (dto *DecisionResponseDTO) Validate() error {
	if dto.Decision == "" {
		return fmt.Errorf("decision is required")
	}
	
	validDecisions := map[string]bool{"BUY": true, "WAIT": true, "BYE": true}
	if !validDecisions[dto.Decision] {
		return fmt.Errorf("invalid decision: %s", dto.Decision)
	}
	
	if dto.Confidence < 0.0 || dto.Confidence > 1.0 {
		return fmt.Errorf("confidence must be between 0.0 and 1.0")
	}
	
	if dto.WaitPeriod < 0 || dto.WaitPeriod > 365 {
		return fmt.Errorf("wait period must be between 0 and 365 days")
	}
	
	return nil
}

// GetDecisionColor returns a color code for UI display
func (dto *DecisionResponseDTO) GetDecisionColor() string {
	switch dto.Decision {
	case "BUY":
		return "#22c55e" // Green
	case "WAIT":
		return "#f59e0b" // Yellow/Orange
	case "BYE":
		return "#ef4444" // Red
	default:
		return "#6b7280" // Gray
	}
}

// GetConfidenceLevel returns a human-readable confidence level
func (dto *DecisionResponseDTO) GetConfidenceLevel() string {
	if dto.Confidence >= 0.8 {
		return "High"
	} else if dto.Confidence >= 0.6 {
		return "Medium"
	} else {
		return "Low"
	}
}

// IsHighConfidence returns true if confidence is high
func (dto *DecisionResponseDTO) IsHighConfidence() bool {
	return dto.Confidence >= 0.8
}

// ShouldWait returns true if this is a WAIT decision
func (dto *DecisionResponseDTO) ShouldWait() bool {
	return dto.Decision == "WAIT"
}

// generateID generates a simple ID for DTOs (in production, use UUID)
func generateID() string {
	return fmt.Sprintf("dto_%d", time.Now().UnixNano())
}