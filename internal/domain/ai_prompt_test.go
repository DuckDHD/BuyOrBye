package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAIPrompt_Validate_ValidData_Success(t *testing.T) {
	prompt := AIPrompt{
		SystemContext:    "You are a financial advisor AI for BuyOrBye app...",
		UserContext:      "User has monthly income of $5000, emergency fund of 6 months...",
		PurchaseDetails:  "User wants to buy a laptop for $1200 in electronics category...",
		DecisionCriteria: "Consider financial stability, health risks, emergency fund...",
		ResponseFormat:   "Respond with JSON: {\"decision\": \"BUY|WAIT|BYE\", \"confidence\": 0.85, \"reasoning\": \"...\"}",
		MaxTokens:        500,
		Temperature:      0.7,
	}

	err := prompt.Validate()
	assert.NoError(t, err)
}

func TestAIPrompt_Validate_EmptySystemContext_ReturnsError(t *testing.T) {
	prompt := AIPrompt{
		SystemContext:    "", // Empty system context should fail
		UserContext:      "User context",
		PurchaseDetails:  "Purchase details",
		DecisionCriteria: "Decision criteria",
		ResponseFormat:   "Response format",
		MaxTokens:        500,
		Temperature:      0.7,
	}

	err := prompt.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "system context is required")
}

func TestAIPrompt_Validate_EmptyUserContext_ReturnsError(t *testing.T) {
	prompt := AIPrompt{
		SystemContext:    "System context",
		UserContext:      "", // Empty user context should fail
		PurchaseDetails:  "Purchase details",
		DecisionCriteria: "Decision criteria",
		ResponseFormat:   "Response format",
		MaxTokens:        500,
		Temperature:      0.7,
	}

	err := prompt.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user context is required")
}

func TestAIPrompt_Validate_EmptyPurchaseDetails_ReturnsError(t *testing.T) {
	prompt := AIPrompt{
		SystemContext:    "System context",
		UserContext:      "User context",
		PurchaseDetails:  "", // Empty purchase details should fail
		DecisionCriteria: "Decision criteria",
		ResponseFormat:   "Response format",
		MaxTokens:        500,
		Temperature:      0.7,
	}

	err := prompt.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "purchase details is required")
}

func TestAIPrompt_Validate_EmptyDecisionCriteria_ReturnsError(t *testing.T) {
	prompt := AIPrompt{
		SystemContext:    "System context",
		UserContext:      "User context",
		PurchaseDetails:  "Purchase details",
		DecisionCriteria: "", // Empty decision criteria should fail
		ResponseFormat:   "Response format",
		MaxTokens:        500,
		Temperature:      0.7,
	}

	err := prompt.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decision criteria is required")
}

func TestAIPrompt_Validate_EmptyResponseFormat_ReturnsError(t *testing.T) {
	prompt := AIPrompt{
		SystemContext:    "System context",
		UserContext:      "User context",
		PurchaseDetails:  "Purchase details",
		DecisionCriteria: "Decision criteria",
		ResponseFormat:   "", // Empty response format should fail
		MaxTokens:        500,
		Temperature:      0.7,
	}

	err := prompt.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "response format is required")
}

func TestAIPrompt_Validate_InvalidMaxTokens_Zero_ReturnsError(t *testing.T) {
	prompt := AIPrompt{
		SystemContext:    "System context",
		UserContext:      "User context",
		PurchaseDetails:  "Purchase details",
		DecisionCriteria: "Decision criteria",
		ResponseFormat:   "Response format",
		MaxTokens:        0, // Zero tokens should fail
		Temperature:      0.7,
	}

	err := prompt.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max tokens must be between 10 and 4000")
}

func TestAIPrompt_Validate_InvalidMaxTokens_TooLow_ReturnsError(t *testing.T) {
	prompt := AIPrompt{
		SystemContext:    "System context",
		UserContext:      "User context",
		PurchaseDetails:  "Purchase details",
		DecisionCriteria: "Decision criteria",
		ResponseFormat:   "Response format",
		MaxTokens:        5, // Too low, below minimum 10
		Temperature:      0.7,
	}

	err := prompt.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max tokens must be between 10 and 4000")
}

func TestAIPrompt_Validate_InvalidMaxTokens_TooHigh_ReturnsError(t *testing.T) {
	prompt := AIPrompt{
		SystemContext:    "System context",
		UserContext:      "User context",
		PurchaseDetails:  "Purchase details",
		DecisionCriteria: "Decision criteria",
		ResponseFormat:   "Response format",
		MaxTokens:        5000, // Too high, above maximum 4000
		Temperature:      0.7,
	}

	err := prompt.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max tokens must be between 10 and 4000")
}

func TestAIPrompt_Validate_InvalidTemperature_TooLow_ReturnsError(t *testing.T) {
	prompt := AIPrompt{
		SystemContext:    "System context",
		UserContext:      "User context",
		PurchaseDetails:  "Purchase details",
		DecisionCriteria: "Decision criteria",
		ResponseFormat:   "Response format",
		MaxTokens:        500,
		Temperature:      -0.1, // Below 0.0 should fail
	}

	err := prompt.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "temperature must be between 0.0 and 2.0")
}

func TestAIPrompt_Validate_InvalidTemperature_TooHigh_ReturnsError(t *testing.T) {
	prompt := AIPrompt{
		SystemContext:    "System context",
		UserContext:      "User context",
		PurchaseDetails:  "Purchase details",
		DecisionCriteria: "Decision criteria",
		ResponseFormat:   "Response format",
		MaxTokens:        500,
		Temperature:      2.1, // Above 2.0 should fail
	}

	err := prompt.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "temperature must be between 0.0 and 2.0")
}

func TestAIPrompt_GetTotalContent_CombinesAllFields(t *testing.T) {
	prompt := AIPrompt{
		SystemContext:    "System context",
		UserContext:      "User context",
		PurchaseDetails:  "Purchase details",
		DecisionCriteria: "Decision criteria",
		ResponseFormat:   "Response format",
		MaxTokens:        500,
		Temperature:      0.7,
	}

	content := prompt.GetTotalContent()
	
	// Check that all components are included
	assert.Contains(t, content, "System context")
	assert.Contains(t, content, "User context")
	assert.Contains(t, content, "Purchase details")
	assert.Contains(t, content, "Decision criteria")
	assert.Contains(t, content, "Response format")
}

func TestAIPrompt_EstimateTokens_ReturnsReasonableEstimate(t *testing.T) {
	prompt := AIPrompt{
		SystemContext:    "This is a short system context",
		UserContext:      "This is user context",
		PurchaseDetails:  "Purchase details here",
		DecisionCriteria: "Decision criteria",
		ResponseFormat:   "Response format",
	}

	tokenCount := prompt.EstimateTokens()
	
	// Rough estimate: should be reasonable for the content length
	// Typically 1 token ≈ 0.75 words, so for ~20 words we expect ~15-25 tokens
	assert.Greater(t, tokenCount, 10)
	assert.Less(t, tokenCount, 100)
}

func TestAIPrompt_EstimateTokens_LargeContent_ReturnsHigherCount(t *testing.T) {
	// Create large content (more than 1000 characters)
	largeText := strings.Repeat("This is a large amount of text that should result in a higher token count. ", 20)
	
	prompt := AIPrompt{
		SystemContext:    largeText,
		UserContext:      largeText,
		PurchaseDetails:  largeText,
		DecisionCriteria: largeText,
		ResponseFormat:   largeText,
	}

	tokenCount := prompt.EstimateTokens()
	
	// Should be much higher for large content
	assert.Greater(t, tokenCount, 200)
}

func TestAIPrompt_IsWithinTokenLimit_UnderLimit_ReturnsTrue(t *testing.T) {
	prompt := AIPrompt{
		SystemContext:    "Short context",
		UserContext:      "Short context",
		PurchaseDetails:  "Short details",
		DecisionCriteria: "Short criteria",
		ResponseFormat:   "Short format",
		MaxTokens:        1000, // High limit
	}

	withinLimit := prompt.IsWithinTokenLimit()
	assert.True(t, withinLimit)
}

func TestAIPrompt_IsWithinTokenLimit_OverLimit_ReturnsFalse(t *testing.T) {
	// Create content that will exceed the token limit
	largeText := strings.Repeat("This is a very large amount of text that will definitely exceed our token limit when combined with all other fields. ", 50)
	
	prompt := AIPrompt{
		SystemContext:    largeText,
		UserContext:      largeText,
		PurchaseDetails:  largeText,
		DecisionCriteria: largeText,
		ResponseFormat:   largeText,
		MaxTokens:        50, // Very low limit
	}

	withinLimit := prompt.IsWithinTokenLimit()
	assert.False(t, withinLimit)
}

func TestAIPrompt_SanitizeForLogging_RemovesSensitiveInfo(t *testing.T) {
	prompt := AIPrompt{
		SystemContext:    "System context",
		UserContext:      "User earns $5000 per month and has SSN 123-45-6789",
		PurchaseDetails:  "Credit card 4532-1234-5678-9012 purchase",
		DecisionCriteria: "Phone number 555-123-4567 criteria",
		ResponseFormat:   "Response format",
	}

	sanitized := prompt.SanitizeForLogging()
	
	// Should remove sensitive information
	assert.NotContains(t, sanitized, "123-45-6789") // SSN
	assert.NotContains(t, sanitized, "4532-1234-5678-9012") // Credit card
	assert.NotContains(t, sanitized, "555-123-4567") // Phone
	assert.Contains(t, sanitized, "[REDACTED]") // Should contain redaction markers
}

func TestAIPrompt_GetContentHash_SameContent_SameHash(t *testing.T) {
	prompt1 := AIPrompt{
		SystemContext:    "Same context",
		UserContext:      "Same user context",
		PurchaseDetails:  "Same purchase details",
		DecisionCriteria: "Same criteria",
		ResponseFormat:   "Same format",
	}

	prompt2 := AIPrompt{
		SystemContext:    "Same context",
		UserContext:      "Same user context",
		PurchaseDetails:  "Same purchase details",
		DecisionCriteria: "Same criteria",
		ResponseFormat:   "Same format",
	}

	hash1 := prompt1.GetContentHash()
	hash2 := prompt2.GetContentHash()
	
	assert.Equal(t, hash1, hash2)
	assert.NotEmpty(t, hash1)
}

func TestAIPrompt_GetContentHash_DifferentContent_DifferentHash(t *testing.T) {
	prompt1 := AIPrompt{
		SystemContext: "Different context",
		UserContext:   "User context",
	}

	prompt2 := AIPrompt{
		SystemContext: "Another context",
		UserContext:   "User context",
	}

	hash1 := prompt1.GetContentHash()
	hash2 := prompt2.GetContentHash()
	
	assert.NotEqual(t, hash1, hash2)
}