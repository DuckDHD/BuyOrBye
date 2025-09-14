package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecisionOutcome_Validate_ValidData_Success(t *testing.T) {
	factors := []DecisionFactor{
		{
			Category:    "financial",
			Impact:      "positive",
			Weight:      0.8,
			Description: "Good emergency fund",
		},
		{
			Category:    "health",
			Impact:      "neutral",
			Weight:      0.2,
			Description: "No health concerns",
		},
	}

	recommendations := []string{
		"Consider setting aside 10% for future upgrades",
		"Look for seasonal discounts",
	}

	outcome := DecisionOutcome{
		ID:              "decision-123",
		UserID:          "user-456",
		IntentID:        "intent-789",
		Decision:        "BUY",
		Confidence:      0.85,
		PrimaryReason:   "Purchase fits within budget and serves essential need",
		Factors:         factors,
		Recommendations: recommendations,
		WaitPeriod:      0,
		MaxBudget:       1200.0,
		CreatedAt:       time.Now(),
		ProcessingTime:  1500,
	}

	err := outcome.Validate()
	assert.NoError(t, err)
}

func TestDecisionOutcome_Validate_EmptyID_ReturnsError(t *testing.T) {
	outcome := DecisionOutcome{
		ID:              "", // Empty ID should fail
		UserID:          "user-456",
		IntentID:        "intent-789",
		Decision:        "BUY",
		Confidence:      0.85,
		PrimaryReason:   "Good reason",
		Factors:         []DecisionFactor{},
		Recommendations: []string{},
		CreatedAt:       time.Now(),
		ProcessingTime:  1500,
	}

	err := outcome.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ID is required")
}

func TestDecisionOutcome_Validate_EmptyUserID_ReturnsError(t *testing.T) {
	outcome := DecisionOutcome{
		ID:              "decision-123",
		UserID:          "", // Empty UserID should fail
		IntentID:        "intent-789",
		Decision:        "BUY",
		Confidence:      0.85,
		PrimaryReason:   "Good reason",
		Factors:         []DecisionFactor{},
		Recommendations: []string{},
		CreatedAt:       time.Now(),
		ProcessingTime:  1500,
	}

	err := outcome.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user ID is required")
}

func TestDecisionOutcome_Validate_EmptyIntentID_ReturnsError(t *testing.T) {
	outcome := DecisionOutcome{
		ID:              "decision-123",
		UserID:          "user-456",
		IntentID:        "", // Empty IntentID should fail
		Decision:        "BUY",
		Confidence:      0.85,
		PrimaryReason:   "Good reason",
		Factors:         []DecisionFactor{},
		Recommendations: []string{},
		CreatedAt:       time.Now(),
		ProcessingTime:  1500,
	}

	err := outcome.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "intent ID is required")
}

func TestDecisionOutcome_Validate_InvalidDecision_ReturnsError(t *testing.T) {
	outcome := DecisionOutcome{
		ID:              "decision-123",
		UserID:          "user-456",
		IntentID:        "intent-789",
		Decision:        "INVALID", // Invalid decision should fail
		Confidence:      0.85,
		PrimaryReason:   "Good reason",
		Factors:         []DecisionFactor{},
		Recommendations: []string{},
		CreatedAt:       time.Now(),
		ProcessingTime:  1500,
	}

	err := outcome.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decision must be one of: BUY, WAIT, BYE")
}

func TestDecisionOutcome_Validate_ValidDecisions_Success(t *testing.T) {
	validDecisions := []string{"BUY", "WAIT", "BYE"}

	for _, decision := range validDecisions {
		t.Run("decision_"+decision, func(t *testing.T) {
			outcome := DecisionOutcome{
				ID:              "decision-123",
				UserID:          "user-456",
				IntentID:        "intent-789",
				Decision:        decision,
				Confidence:      0.85,
				PrimaryReason:   "Good reason",
				Factors:         []DecisionFactor{},
				Recommendations: []string{},
				CreatedAt:       time.Now(),
				ProcessingTime:  1500,
			}

			err := outcome.Validate()
			assert.NoError(t, err, "Decision %s should be valid", decision)
		})
	}
}

func TestDecisionOutcome_Validate_ConfidenceTooLow_ReturnsError(t *testing.T) {
	outcome := DecisionOutcome{
		ID:              "decision-123",
		UserID:          "user-456",
		IntentID:        "intent-789",
		Decision:        "BUY",
		Confidence:      -0.1, // Below 0.0 should fail
		PrimaryReason:   "Good reason",
		Factors:         []DecisionFactor{},
		Recommendations: []string{},
		CreatedAt:       time.Now(),
		ProcessingTime:  1500,
	}

	err := outcome.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "confidence must be between 0.0 and 1.0")
}

func TestDecisionOutcome_Validate_ConfidenceTooHigh_ReturnsError(t *testing.T) {
	outcome := DecisionOutcome{
		ID:              "decision-123",
		UserID:          "user-456",
		IntentID:        "intent-789",
		Decision:        "BUY",
		Confidence:      1.1, // Above 1.0 should fail
		PrimaryReason:   "Good reason",
		Factors:         []DecisionFactor{},
		Recommendations: []string{},
		CreatedAt:       time.Now(),
		ProcessingTime:  1500,
	}

	err := outcome.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "confidence must be between 0.0 and 1.0")
}

func TestDecisionOutcome_Validate_EmptyPrimaryReason_ReturnsError(t *testing.T) {
	outcome := DecisionOutcome{
		ID:              "decision-123",
		UserID:          "user-456",
		IntentID:        "intent-789",
		Decision:        "BUY",
		Confidence:      0.85,
		PrimaryReason:   "", // Empty primary reason should fail
		Factors:         []DecisionFactor{},
		Recommendations: []string{},
		CreatedAt:       time.Now(),
		ProcessingTime:  1500,
	}

	err := outcome.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "primary reason is required")
}

func TestDecisionOutcome_Validate_PrimaryReasonTooLong_ReturnsError(t *testing.T) {
	longReason := string(make([]byte, 501)) // 501 characters, exceeds 500 limit
	for i := range longReason {
		longReason = string(longReason[:i]) + "a" + string(longReason[i+1:])
	}

	outcome := DecisionOutcome{
		ID:              "decision-123",
		UserID:          "user-456",
		IntentID:        "intent-789",
		Decision:        "BUY",
		Confidence:      0.85,
		PrimaryReason:   longReason,
		Factors:         []DecisionFactor{},
		Recommendations: []string{},
		CreatedAt:       time.Now(),
		ProcessingTime:  1500,
	}

	err := outcome.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "primary reason must not exceed 500 characters")
}

func TestDecisionOutcome_Validate_NegativeWaitPeriod_ReturnsError(t *testing.T) {
	outcome := DecisionOutcome{
		ID:              "decision-123",
		UserID:          "user-456",
		IntentID:        "intent-789",
		Decision:        "WAIT",
		Confidence:      0.85,
		PrimaryReason:   "Good reason",
		Factors:         []DecisionFactor{},
		Recommendations: []string{},
		WaitPeriod:      -1, // Negative wait period should fail
		CreatedAt:       time.Now(),
		ProcessingTime:  1500,
	}

	err := outcome.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wait period must be non-negative")
}

func TestDecisionOutcome_Validate_WaitPeriodTooLong_ReturnsError(t *testing.T) {
	outcome := DecisionOutcome{
		ID:              "decision-123",
		UserID:          "user-456",
		IntentID:        "intent-789",
		Decision:        "WAIT",
		Confidence:      0.85,
		PrimaryReason:   "Good reason",
		Factors:         []DecisionFactor{},
		Recommendations: []string{},
		WaitPeriod:      366, // Exceeds 365 days limit
		CreatedAt:       time.Now(),
		ProcessingTime:  1500,
	}

	err := outcome.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "wait period must not exceed 365 days")
}

func TestDecisionOutcome_Validate_NegativeMaxBudget_ReturnsError(t *testing.T) {
	outcome := DecisionOutcome{
		ID:              "decision-123",
		UserID:          "user-456",
		IntentID:        "intent-789",
		Decision:        "BUY",
		Confidence:      0.85,
		PrimaryReason:   "Good reason",
		Factors:         []DecisionFactor{},
		Recommendations: []string{},
		MaxBudget:       -100.0, // Negative budget should fail
		CreatedAt:       time.Now(),
		ProcessingTime:  1500,
	}

	err := outcome.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max budget must be non-negative")
}

func TestDecisionOutcome_Validate_NegativeProcessingTime_ReturnsError(t *testing.T) {
	outcome := DecisionOutcome{
		ID:              "decision-123",
		UserID:          "user-456",
		IntentID:        "intent-789",
		Decision:        "BUY",
		Confidence:      0.85,
		PrimaryReason:   "Good reason",
		Factors:         []DecisionFactor{},
		Recommendations: []string{},
		ProcessingTime:  -100, // Negative processing time should fail
		CreatedAt:       time.Now(),
	}

	err := outcome.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "processing time must be non-negative")
}

func TestDecisionOutcome_Validate_InvalidFactors_ReturnsError(t *testing.T) {
	invalidFactors := []DecisionFactor{
		{
			Category:    "invalid-category", // Invalid category
			Impact:      "positive",
			Weight:      0.5,
			Description: "Test factor",
		},
	}

	outcome := DecisionOutcome{
		ID:              "decision-123",
		UserID:          "user-456",
		IntentID:        "intent-789",
		Decision:        "BUY",
		Confidence:      0.85,
		PrimaryReason:   "Good reason",
		Factors:         invalidFactors,
		Recommendations: []string{},
		CreatedAt:       time.Now(),
		ProcessingTime:  1500,
	}

	err := outcome.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "factor validation failed")
}

func TestDecisionOutcome_IsHighConfidence_AboveThreshold_ReturnsTrue(t *testing.T) {
	outcome := DecisionOutcome{
		Confidence: 0.85,
	}

	assert.True(t, outcome.IsHighConfidence())
}

func TestDecisionOutcome_IsHighConfidence_BelowThreshold_ReturnsFalse(t *testing.T) {
	outcome := DecisionOutcome{
		Confidence: 0.75,
	}

	assert.False(t, outcome.IsHighConfidence())
}

func TestDecisionOutcome_IsHighConfidence_AtThreshold_ReturnsTrue(t *testing.T) {
	outcome := DecisionOutcome{
		Confidence: 0.8,
	}

	assert.True(t, outcome.IsHighConfidence())
}

func TestDecisionOutcome_ShouldWait_WaitDecision_ReturnsTrue(t *testing.T) {
	outcome := DecisionOutcome{
		Decision: "WAIT",
	}

	assert.True(t, outcome.ShouldWait())
}

func TestDecisionOutcome_ShouldWait_BuyDecision_ReturnsFalse(t *testing.T) {
	outcome := DecisionOutcome{
		Decision: "BUY",
	}

	assert.False(t, outcome.ShouldWait())
}

func TestDecisionOutcome_ShouldWait_ByeDecision_ReturnsFalse(t *testing.T) {
	outcome := DecisionOutcome{
		Decision: "BYE",
	}

	assert.False(t, outcome.ShouldWait())
}

func TestDecisionOutcome_GetConfidenceLevel_HighConfidence_ReturnsHigh(t *testing.T) {
	outcome := DecisionOutcome{
		Confidence: 0.9,
	}

	level := outcome.GetConfidenceLevel()
	assert.Equal(t, "high", level)
}

func TestDecisionOutcome_GetConfidenceLevel_MediumConfidence_ReturnsMedium(t *testing.T) {
	outcome := DecisionOutcome{
		Confidence: 0.7,
	}

	level := outcome.GetConfidenceLevel()
	assert.Equal(t, "medium", level)
}

func TestDecisionOutcome_GetConfidenceLevel_LowConfidence_ReturnsLow(t *testing.T) {
	outcome := DecisionOutcome{
		Confidence: 0.5,
	}

	level := outcome.GetConfidenceLevel()
	assert.Equal(t, "low", level)
}