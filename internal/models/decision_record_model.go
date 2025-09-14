package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"gorm.io/gorm"
)

// DecisionRecordModel represents a purchase decision record in the database
type DecisionRecordModel struct {
	gorm.Model
	
	// Core identifiers
	UserID   string `gorm:"type:varchar(255);not null;index:idx_user_decisions" json:"user_id"`
	IntentID string `gorm:"type:varchar(255);not null;uniqueIndex" json:"intent_id"`
	
	// Purchase details
	ItemName string  `gorm:"type:varchar(500);not null" json:"item_name"`
	ItemCost float64 `gorm:"type:decimal(15,2);not null" json:"item_cost"`
	Category string  `gorm:"type:varchar(50);not null;index:idx_category_decisions" json:"category"`
	Urgency  string  `gorm:"type:varchar(20);not null" json:"urgency"`
	Frequency string `gorm:"type:varchar(20);not null" json:"frequency"`
	Purpose  string  `gorm:"type:text" json:"purpose"`
	Alternative string `gorm:"type:varchar(500)" json:"alternative"`
	
	// Decision outcome
	Decision        string  `gorm:"type:varchar(10);not null;index:idx_decision_type" json:"decision"`
	Confidence      float64 `gorm:"type:decimal(3,2);not null" json:"confidence"`
	PrimaryReason   string  `gorm:"type:text" json:"primary_reason"`
	WaitPeriod      int     `gorm:"type:int;default:0" json:"wait_period"`
	MaxBudget       float64 `gorm:"type:decimal(15,2);default:0" json:"max_budget"`
	
	// Decision factors (JSON field)
	FactorsJSON string `gorm:"type:text;column:factors_json" json:"-"`
	
	// Recommendations (JSON field)
	RecommendationsJSON string `gorm:"type:text;column:recommendations_json" json:"-"`
	
	// Performance tracking
	ProcessingTime int64 `gorm:"type:bigint;not null" json:"processing_time_ms"`
	
	// Metadata
	DecisionSource string `gorm:"type:varchar(50);default:'business_rules'" json:"decision_source"` // business_rules, ai_assisted, ai_fallback
	AIProvider     string `gorm:"type:varchar(50)" json:"ai_provider"`
	AIModel        string `gorm:"type:varchar(100)" json:"ai_model"`
	
	// Composite indexes will be created in migration
}

// TableName specifies the table name for GORM
func (DecisionRecordModel) TableName() string {
	return "decision_records"
}

// BeforeCreate hook to validate data before creation
func (m *DecisionRecordModel) BeforeCreate(tx *gorm.DB) error {
	return m.validate()
}

// BeforeUpdate hook to validate data before updates
func (m *DecisionRecordModel) BeforeUpdate(tx *gorm.DB) error {
	return m.validate()
}

// validate performs model validation
func (m *DecisionRecordModel) validate() error {
	if m.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	
	if m.IntentID == "" {
		return fmt.Errorf("intent_id is required")
	}
	
	if m.ItemName == "" {
		return fmt.Errorf("item_name is required")
	}
	
	if m.ItemCost <= 0 {
		return fmt.Errorf("item_cost must be greater than 0")
	}
	
	if m.Category == "" {
		return fmt.Errorf("category is required")
	}
	
	validCategories := map[string]bool{
		"electronics": true, "clothing": true, "food": true,
		"transport": true, "health": true, "entertainment": true, "other": true,
	}
	if !validCategories[m.Category] {
		return fmt.Errorf("invalid category: %s", m.Category)
	}
	
	validDecisions := map[string]bool{"BUY": true, "WAIT": true, "BYE": true}
	if !validDecisions[m.Decision] {
		return fmt.Errorf("invalid decision: %s", m.Decision)
	}
	
	if m.Confidence < 0.0 || m.Confidence > 1.0 {
		return fmt.Errorf("confidence must be between 0.0 and 1.0")
	}
	
	if m.WaitPeriod < 0 || m.WaitPeriod > 365 {
		return fmt.Errorf("wait_period must be between 0 and 365 days")
	}
	
	return nil
}

// FromDomain converts domain.DecisionOutcome to DecisionRecordModel
func (m *DecisionRecordModel) FromDomain(outcome domain.DecisionOutcome, intent domain.PurchaseIntent) error {
	// Set basic fields
	m.UserID = outcome.UserID
	m.IntentID = outcome.IntentID
	m.ItemName = intent.ItemName
	m.ItemCost = intent.ItemCost
	m.Category = intent.Category
	m.Urgency = intent.Urgency
	m.Frequency = intent.Frequency
	m.Purpose = intent.Purpose
	m.Alternative = intent.Alternative
	
	// Set decision fields
	m.Decision = outcome.Decision
	m.Confidence = outcome.Confidence
	m.PrimaryReason = outcome.PrimaryReason
	m.WaitPeriod = outcome.WaitPeriod
	m.MaxBudget = outcome.MaxBudget
	m.ProcessingTime = outcome.ProcessingTime
	
	// Convert factors to JSON
	if len(outcome.Factors) > 0 {
		factorsBytes, err := json.Marshal(outcome.Factors)
		if err != nil {
			return fmt.Errorf("failed to marshal factors: %w", err)
		}
		m.FactorsJSON = string(factorsBytes)
	}
	
	// Convert recommendations to JSON
	if len(outcome.Recommendations) > 0 {
		recBytes, err := json.Marshal(outcome.Recommendations)
		if err != nil {
			return fmt.Errorf("failed to marshal recommendations: %w", err)
		}
		m.RecommendationsJSON = string(recBytes)
	}
	
	// Set timestamps
	if !outcome.CreatedAt.IsZero() {
		m.CreatedAt = outcome.CreatedAt
	}
	m.UpdatedAt = time.Now()
	
	return nil
}

// ToDomain converts DecisionRecordModel to domain.DecisionOutcome
func (m *DecisionRecordModel) ToDomain() (*domain.DecisionOutcome, error) {
	// Parse factors from JSON
	var factors []domain.DecisionFactor
	if m.FactorsJSON != "" {
		if err := json.Unmarshal([]byte(m.FactorsJSON), &factors); err != nil {
			return nil, fmt.Errorf("failed to unmarshal factors: %w", err)
		}
	}
	
	// Parse recommendations from JSON
	var recommendations []string
	if m.RecommendationsJSON != "" {
		if err := json.Unmarshal([]byte(m.RecommendationsJSON), &recommendations); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recommendations: %w", err)
		}
	}
	
	// Create domain object
	outcome := &domain.DecisionOutcome{
		ID:              fmt.Sprintf("%d", m.ID),
		UserID:          m.UserID,
		IntentID:        m.IntentID,
		Decision:        m.Decision,
		Confidence:      m.Confidence,
		PrimaryReason:   m.PrimaryReason,
		Factors:         factors,
		Recommendations: recommendations,
		WaitPeriod:      m.WaitPeriod,
		MaxBudget:       m.MaxBudget,
		CreatedAt:       m.CreatedAt,
		ProcessingTime:  m.ProcessingTime,
	}
	
	// Validate domain object
	if err := outcome.Validate(); err != nil {
		return nil, fmt.Errorf("invalid domain object: %w", err)
	}
	
	return outcome, nil
}

// ToPastDecision converts DecisionRecordModel to domain.PastDecision for history
func (m *DecisionRecordModel) ToPastDecision() domain.PastDecision {
	daysAgo := int(time.Since(m.CreatedAt).Hours() / 24)
	
	return domain.PastDecision{
		ItemName: m.ItemName,
		ItemCost: m.ItemCost,
		Category: m.Category,
		Decision: m.Decision,
		DaysAgo:  daysAgo,
	}
}

// GetFactors returns parsed factors
func (m *DecisionRecordModel) GetFactors() ([]domain.DecisionFactor, error) {
	if m.FactorsJSON == "" {
		return nil, nil
	}
	
	var factors []domain.DecisionFactor
	if err := json.Unmarshal([]byte(m.FactorsJSON), &factors); err != nil {
		return nil, fmt.Errorf("failed to parse factors: %w", err)
	}
	
	return factors, nil
}

// GetRecommendations returns parsed recommendations
func (m *DecisionRecordModel) GetRecommendations() ([]string, error) {
	if m.RecommendationsJSON == "" {
		return nil, nil
	}
	
	var recommendations []string
	if err := json.Unmarshal([]byte(m.RecommendationsJSON), &recommendations); err != nil {
		return nil, fmt.Errorf("failed to parse recommendations: %w", err)
	}
	
	return recommendations, nil
}

// SetFactors sets factors as JSON
func (m *DecisionRecordModel) SetFactors(factors []domain.DecisionFactor) error {
	if factors == nil {
		m.FactorsJSON = ""
		return nil
	}
	
	factorsBytes, err := json.Marshal(factors)
	if err != nil {
		return fmt.Errorf("failed to marshal factors: %w", err)
	}
	
	m.FactorsJSON = string(factorsBytes)
	return nil
}

// SetRecommendations sets recommendations as JSON
func (m *DecisionRecordModel) SetRecommendations(recommendations []string) error {
	if recommendations == nil {
		m.RecommendationsJSON = ""
		return nil
	}
	
	recBytes, err := json.Marshal(recommendations)
	if err != nil {
		return fmt.Errorf("failed to marshal recommendations: %w", err)
	}
	
	m.RecommendationsJSON = string(recBytes)
	return nil
}

// IsSuccessful returns true if this was a BUY decision
func (m *DecisionRecordModel) IsSuccessful() bool {
	return m.Decision == "BUY"
}

// IsWaiting returns true if this is a WAIT decision
func (m *DecisionRecordModel) IsWaiting() bool {
	return m.Decision == "WAIT"
}

// IsRejected returns true if this is a BYE decision
func (m *DecisionRecordModel) IsRejected() bool {
	return m.Decision == "BYE"
}

// GetMonthlyImpact calculates monthly financial impact based on frequency
func (m *DecisionRecordModel) GetMonthlyImpact() float64 {
	if m.Decision != "BUY" {
		return 0.0 // No impact if not buying
	}
	
	switch m.Frequency {
	case "one_time":
		return m.ItemCost / 12 // Amortize over 12 months
	case "monthly":
		return m.ItemCost
	case "yearly":
		return m.ItemCost / 12
	default:
		return m.ItemCost / 12 // Default to one-time
	}
}

// IsExpensiveDecision returns true if this decision involved a high-cost item
func (m *DecisionRecordModel) IsExpensiveDecision(threshold float64) bool {
	return m.ItemCost > threshold
}

// GetAgeInDays returns how many days ago this decision was made
func (m *DecisionRecordModel) GetAgeInDays() int {
	return int(time.Since(m.CreatedAt).Hours() / 24)
}

// IsRecent returns true if decision was made within specified days
func (m *DecisionRecordModel) IsRecent(days int) bool {
	return m.GetAgeInDays() <= days
}

// GetDecisionQuality returns a quality score based on confidence and outcome
func (m *DecisionRecordModel) GetDecisionQuality() string {
	if m.Confidence >= 0.8 {
		return "High"
	} else if m.Confidence >= 0.6 {
		return "Medium"
	} else {
		return "Low"
	}
}