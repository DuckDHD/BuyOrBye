package domain

import (
	"fmt"
	"time"
)

// DecisionOutcome represents the result of a purchase decision analysis
type DecisionOutcome struct {
	ID              string           `json:"id"`
	UserID          string           `json:"user_id"`
	IntentID        string           `json:"intent_id"`
	Decision        string           `json:"decision"`        // "BUY", "WAIT", "BYE"
	Confidence      float64          `json:"confidence"`      // 0.0 to 1.0
	PrimaryReason   string           `json:"primary_reason"`  // Main reason for decision
	Factors         []DecisionFactor `json:"factors"`         // Detailed factors that influenced decision
	Recommendations []string         `json:"recommendations"` // Actionable advice
	WaitPeriod      int              `json:"wait_period"`     // Days to wait (if WAIT)
	MaxBudget       float64          `json:"max_budget"`      // Recommended max spend
	CreatedAt       time.Time        `json:"created_at"`
	ProcessingTime  int64            `json:"processing_time"` // Milliseconds
}

// DecisionFactor represents an individual factor that influenced the decision
type DecisionFactor struct {
	Category    string  `json:"category"`    // "financial", "health", "practical", "timing"
	Impact      string  `json:"impact"`      // "positive", "negative", "neutral"
	Weight      float64 `json:"weight"`      // 0.0 to 1.0 importance
	Description string  `json:"description"` // Human-readable explanation
}

// Valid decision values
var validDecisions = map[string]bool{
	"BUY":  true,
	"WAIT": true,
	"BYE":  true,
}

// Valid factor categories
var validFactorCategories = map[string]bool{
	"financial": true,
	"health":    true,
	"practical": true,
	"timing":    true,
}

// Valid impact types
var validImpacts = map[string]bool{
	"positive": true,
	"negative": true,
	"neutral":  true,
}

// Validate performs comprehensive validation of the decision outcome
func (do *DecisionOutcome) Validate() error {
	// Validate required fields
	if do.ID == "" {
		return fmt.Errorf("ID is required")
	}

	if do.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if do.IntentID == "" {
		return fmt.Errorf("intent ID is required")
	}

	// Validate decision
	if !validDecisions[do.Decision] {
		return fmt.Errorf("decision must be one of: BUY, WAIT, BYE")
	}

	// Validate confidence
	if do.Confidence < 0.0 || do.Confidence > 1.0 {
		return fmt.Errorf("confidence must be between 0.0 and 1.0")
	}

	// Validate primary reason
	if do.PrimaryReason == "" {
		return fmt.Errorf("primary reason is required")
	}

	if len(do.PrimaryReason) > 500 {
		return fmt.Errorf("primary reason must not exceed 500 characters")
	}

	// Validate wait period
	if do.WaitPeriod < 0 {
		return fmt.Errorf("wait period must be non-negative")
	}

	if do.WaitPeriod > 365 {
		return fmt.Errorf("wait period must not exceed 365 days")
	}

	// Validate max budget
	if do.MaxBudget < 0 {
		return fmt.Errorf("max budget must be non-negative")
	}

	// Validate processing time
	if do.ProcessingTime < 0 {
		return fmt.Errorf("processing time must be non-negative")
	}

	// Validate factors
	for i, factor := range do.Factors {
		if err := factor.Validate(); err != nil {
			return fmt.Errorf("factor validation failed at index %d: %w", i, err)
		}
	}

	return nil
}

// Validate performs validation on a DecisionFactor
func (df *DecisionFactor) Validate() error {
	// Validate category
	if !validFactorCategories[df.Category] {
		return fmt.Errorf("invalid category: %s. Must be one of: financial, health, practical, timing", df.Category)
	}

	// Validate impact
	if !validImpacts[df.Impact] {
		return fmt.Errorf("invalid impact: %s. Must be one of: positive, negative, neutral", df.Impact)
	}

	// Validate weight
	if df.Weight < 0.0 || df.Weight > 1.0 {
		return fmt.Errorf("weight must be between 0.0 and 1.0")
	}

	// Validate description
	if df.Description == "" {
		return fmt.Errorf("description is required")
	}

	if len(df.Description) > 200 {
		return fmt.Errorf("description must not exceed 200 characters")
	}

	return nil
}

// IsHighConfidence returns true if confidence is above 0.8 threshold
func (do *DecisionOutcome) IsHighConfidence() bool {
	return do.Confidence >= 0.8
}

// ShouldWait returns true if the decision is WAIT
func (do *DecisionOutcome) ShouldWait() bool {
	return do.Decision == "WAIT"
}

// IsBuyRecommendation returns true if the decision is BUY
func (do *DecisionOutcome) IsBuyRecommendation() bool {
	return do.Decision == "BUY"
}

// IsRejection returns true if the decision is BYE
func (do *DecisionOutcome) IsRejection() bool {
	return do.Decision == "BYE"
}

// GetConfidenceLevel returns a human-readable confidence level
func (do *DecisionOutcome) GetConfidenceLevel() string {
	if do.Confidence >= 0.8 {
		return "high"
	} else if do.Confidence >= 0.6 {
		return "medium"
	} else {
		return "low"
	}
}

// HasWaitPeriod returns true if a wait period is specified
func (do *DecisionOutcome) HasWaitPeriod() bool {
	return do.WaitPeriod > 0
}

// HasMaxBudget returns true if a maximum budget recommendation is provided
func (do *DecisionOutcome) HasMaxBudget() bool {
	return do.MaxBudget > 0
}

// GetPositiveFactors returns factors with positive impact
func (do *DecisionOutcome) GetPositiveFactors() []DecisionFactor {
	var positive []DecisionFactor
	for _, factor := range do.Factors {
		if factor.Impact == "positive" {
			positive = append(positive, factor)
		}
	}
	return positive
}

// GetNegativeFactors returns factors with negative impact
func (do *DecisionOutcome) GetNegativeFactors() []DecisionFactor {
	var negative []DecisionFactor
	for _, factor := range do.Factors {
		if factor.Impact == "negative" {
			negative = append(negative, factor)
		}
	}
	return negative
}

// GetFactorsByCategory returns factors filtered by category
func (do *DecisionOutcome) GetFactorsByCategory(category string) []DecisionFactor {
	var filtered []DecisionFactor
	for _, factor := range do.Factors {
		if factor.Category == category {
			filtered = append(filtered, factor)
		}
	}
	return filtered
}

// GetTotalWeight calculates the sum of all factor weights
func (do *DecisionOutcome) GetTotalWeight() float64 {
	total := 0.0
	for _, factor := range do.Factors {
		total += factor.Weight
	}
	return total
}

// HasRecommendations returns true if recommendations are provided
func (do *DecisionOutcome) HasRecommendations() bool {
	return len(do.Recommendations) > 0
}

// String returns a string representation of the decision outcome
func (do *DecisionOutcome) String() string {
	return fmt.Sprintf("DecisionOutcome{ID: %s, Decision: %s, Confidence: %.2f, Reason: %s}",
		do.ID, do.Decision, do.Confidence, do.PrimaryReason)
}