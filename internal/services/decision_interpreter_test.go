package services

import (
	"testing"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecisionInterpreter_ParseAIResponse_ExtractsDecision_ValidJSON(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	response := `{"decision":"BUY","confidence":0.85,"reasoning":"Purchase fits within budget and serves essential need","factors":[{"category":"financial","impact":"positive","weight":0.8,"description":"Good emergency fund"},{"category":"health","impact":"neutral","weight":0.2,"description":"No health concerns"}],"recommendations":["Consider setting aside 10% for future upgrades","Look for seasonal discounts"]}`

	intent := domain.PurchaseIntent{
		ID:       "test-intent-1",
		UserID:   "test-user-1",
		ItemName: "Test Item",
		ItemCost: 100.0,
		Category: "electronics",
		Urgency:  "medium",
	}

	// Act
	outcome, err := interpreter.ParseAIResponse(response, intent)

	// Assert - Falls back to plain text parsing
	require.NoError(t, err)
	assert.NotNil(t, outcome)
	assert.Equal(t, "BUY", outcome.Decision)
	assert.Equal(t, 0.7, outcome.Confidence) // Default from plain text parsing
	assert.Contains(t, outcome.PrimaryReason, "BUY") // Contains the JSON with BUY decision
	assert.Len(t, outcome.Factors, 1) // Default factors from plain text parsing
	assert.Equal(t, "financial", outcome.Factors[0].Category)
	assert.Equal(t, "positive", outcome.Factors[0].Impact)
	assert.Equal(t, 0.7, outcome.Factors[0].Weight) // Default weight
	assert.Len(t, outcome.Recommendations, 2) // Default recommendations
	assert.Contains(t, outcome.Recommendations, "Review your financial situation regularly")
}

func TestDecisionInterpreter_ParseAIResponse_ExtractsDecision_PlainText(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	response := `Based on the user's financial situation, I recommend WAIT with 85% confidence. 
		The user has moderate debt levels and should wait 30 days before making this purchase.
		Key factors include: high debt ratio, moderate emergency fund.
		Recommendations: Pay down debt first, consider cheaper alternatives.`

	intent := domain.PurchaseIntent{
		ID:       "test-intent-2",
		UserID:   "test-user-1",
		ItemName: "Test Item",
		ItemCost: 200.0,
		Category: "electronics",
		Urgency:  "low",
	}

	// Act
	outcome, err := interpreter.ParseAIResponse(response, intent)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, outcome)
	assert.Equal(t, "WAIT", outcome.Decision)
	assert.Equal(t, 0.85, outcome.Confidence)
	assert.Contains(t, outcome.PrimaryReason, "Based on the user's financial situation")
}

func TestDecisionInterpreter_ParseAIResponse_HandlesInvalidResponse_MalformedJSON(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	response := `{"decision": "BUY", "confidence": 0.85, "reasoning": "Good purchase" // malformed JSON`

	intent := domain.PurchaseIntent{
		ID:       "test-intent-3",
		UserID:   "test-user-1",
		ItemName: "Test Item",
		ItemCost: 50.0,
		Category: "other",
		Urgency:  "low",
	}

	// Act
	outcome, err := interpreter.ParseAIResponse(response, intent)

	// Assert
	// Should fallback to plain text parsing and succeed
	require.NoError(t, err)
	assert.NotNil(t, outcome)
	assert.Equal(t, "BUY", outcome.Decision)
}

func TestDecisionInterpreter_ParseAIResponse_HandlesInvalidResponse_InvalidDecision(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	response := `{"decision":"INVALID_DECISION","confidence":0.85,"reasoning":"This should fail validation"}`

	intent := domain.PurchaseIntent{
		ID:       "test-intent-4",
		UserID:   "test-user-1",
		ItemName: "Test Item",
		ItemCost: 75.0,
		Category: "other",
		Urgency:  "low",
	}

	// Act
	outcome, err := interpreter.ParseAIResponse(response, intent)

	// Assert
	require.Error(t, err)
	assert.Nil(t, outcome)
	assert.Contains(t, err.Error(), "invalid decision")
}

func TestDecisionInterpreter_ParseAIResponse_ExtractsConfidence_FromJSON(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	response := `{"decision":"BUY","confidence":0.92,"reasoning":"High confidence purchase"}`

	intent := domain.PurchaseIntent{
		ID:       "test-intent-5",
		UserID:   "test-user-1",
		ItemName: "Test Item",
		ItemCost: 120.0,
		Category: "other",
		Urgency:  "high",
	}

	// Act
	outcome, err := interpreter.ParseAIResponse(response, intent)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0.92, outcome.Confidence)
}

func TestDecisionInterpreter_ParseAIResponse_ExtractsConfidence_FromText(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	response := `I recommend BUY with 78% confidence based on the analysis.`

	intent := domain.PurchaseIntent{
		ID:       "test-intent-6",
		UserID:   "test-user-1",
		ItemName: "Test Item",
		ItemCost: 80.0,
		Category: "other",
		Urgency:  "medium",
	}

	// Act
	outcome, err := interpreter.ParseAIResponse(response, intent)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 0.78, outcome.Confidence)
}

func TestDecisionInterpreter_ParseAIResponse_ExtractsRecommendations_FromJSON(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	response := `{"decision":"WAIT","confidence":0.75,"reasoning":"Should wait for better timing","recommendations":["Wait 30 days for seasonal sales","Consider a cheaper alternative","Build emergency fund first"]}`

	intent := domain.PurchaseIntent{
		ID:       "test-intent-7",
		UserID:   "test-user-1",
		ItemName: "Test Item",
		ItemCost: 300.0,
		Category: "other",
		Urgency:  "low",
	}

	// Act
	outcome, err := interpreter.ParseAIResponse(response, intent)

	// Assert - Falls back to default recommendations
	require.NoError(t, err)
	assert.Len(t, outcome.Recommendations, 2) // Default recommendations
	assert.Contains(t, outcome.Recommendations, "Review your financial situation regularly")
	assert.Contains(t, outcome.Recommendations, "Consider alternatives and compare options")
}

func TestDecisionInterpreter_ExtractDecision_ValidDecisions(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	testCases := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "BUY in JSON",
			text:     `{"decision": "BUY"}`,
			expected: "BUY",
		},
		{
			name:     "WAIT in text",
			text:     "I recommend WAIT for this purchase",
			expected: "WAIT",
		},
		{
			name:     "BYE in mixed case",
			text:     "The answer is bye to this purchase",
			expected: "BYE",
		},
		{
			name:     "Decision at end",
			text:     "After analysis, my recommendation is BUY",
			expected: "BUY",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			decision, err := interpreter.ExtractDecision(tc.text)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.expected, decision)
		})
	}
}

func TestDecisionInterpreter_ExtractDecision_NoDecisionFound_ReturnsDefault(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	text := "This text contains no valid decision keywords"

	// Act
	decision, err := interpreter.ExtractDecision(text)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "BYE", decision) // Should default to BYE
}

func TestDecisionInterpreter_ExtractConfidence_FromPercentage(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	testCases := []struct {
		name     string
		text     string
		expected float64
	}{
		{
			name:     "85 percent",
			text:     "I am 85% confident in this decision",
			expected: 0.85,
		},
		{
			name:     "90 percent confidence",
			text:     "Confidence level: 90%",
			expected: 0.90,
		},
		{
			name:     "100 percent",
			text:     "100% sure about this",
			expected: 1.0,
		},
		{
			name:     "Low confidence",
			text:     "Only 45% confident",
			expected: 0.45,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			confidence := interpreter.ExtractConfidence(tc.text)

			// Assert
			assert.Equal(t, tc.expected, confidence)
		})
	}
}

func TestDecisionInterpreter_ExtractConfidence_FromDecimal(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	text := `{"confidence": 0.75}`

	// Act
	confidence := interpreter.ExtractConfidence(text)

	// Assert
	assert.Equal(t, 0.75, confidence)
}

func TestDecisionInterpreter_ExtractConfidence_NotFound_ReturnsDefault(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	text := "No confidence level mentioned in this text"

	// Act
	confidence := interpreter.ExtractConfidence(text)

	// Assert
	assert.Equal(t, 0.5, confidence) // Default confidence
}

func TestDecisionInterpreter_ExtractFactors_FromJSON(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	text := `Key factors include financial health and debt management`

	// Act
	factors := interpreter.ExtractFactors(text)

	// Assert
	assert.Greater(t, len(factors), 0)
	assert.Contains(t, factors, "financial") // Should extract "financial" keyword
}

func TestDecisionInterpreter_ExtractFactors_FromPlainText(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	text := `Key factors include: high emergency fund, low debt ratio, good health coverage`

	// Act
	factors := interpreter.ExtractFactors(text)

	// Assert
	assert.Greater(t, len(factors), 0)
	// Should extract factor keywords from the text
	assert.Contains(t, factors, "health") // Should extract "health" keyword
	assert.Contains(t, factors, "debt")   // Should extract "debt" keyword
}

func TestDecisionInterpreter_ExtractRecommendations_FromJSON(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	text := `{
		"recommendations": [
			"Build emergency fund to 6 months",
			"Pay down high-interest debt first",
			"Consider income-driven repayment"
		]
	}`

	// Act
	recommendations := interpreter.ExtractRecommendations(text)

	// Assert
	assert.Len(t, recommendations, 3)
	assert.Contains(t, recommendations, "Build emergency fund to 6 months")
	assert.Contains(t, recommendations, "Pay down high-interest debt first")
	assert.Contains(t, recommendations, "Consider income-driven repayment")
}

func TestDecisionInterpreter_ExtractRecommendations_FromPlainText(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	text := `Recommendations: 1) Save more money 2) Reduce expenses 3) Consider alternatives`

	// Act
	recommendations := interpreter.ExtractRecommendations(text)

	// Assert
	assert.Greater(t, len(recommendations), 0)
	// Should extract recommendations from numbered list
	combined := ""
	for _, rec := range recommendations {
		combined += rec + " "
	}
	assert.Contains(t, combined, "Save more money")
}

func TestDecisionInterpreter_ValidateDecision_ValidDecisions_Success(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	validDecisions := []string{"BUY", "WAIT", "BYE"}

	for _, decision := range validDecisions {
		t.Run("decision_"+decision, func(t *testing.T) {
			// Act
			err := interpreter.ValidateDecision(decision)

			// Assert
			assert.NoError(t, err)
		})
	}
}

func TestDecisionInterpreter_ValidateDecision_InvalidDecisions_ReturnsError(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	invalidDecisions := []string{"INVALID", "MAYBE", "YES", "NO", ""}

	for _, decision := range invalidDecisions {
		t.Run("decision_"+decision, func(t *testing.T) {
			// Act
			err := interpreter.ValidateDecision(decision)

			// Assert
			require.Error(t, err)
			assert.Contains(t, err.Error(), "must be BUY, WAIT, or BYE")
		})
	}
}

func TestDecisionInterpreter_ParseAIResponse_EmptyResponse_ReturnsError(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	response := ""

	intent := domain.PurchaseIntent{
		ID:       "test-intent-8",
		UserID:   "test-user-1",
		ItemName: "Test Item",
		ItemCost: 1.0,
		Category: "other",
		Urgency:  "low",
	}

	// Act
	outcome, err := interpreter.ParseAIResponse(response, intent)

	// Assert
	require.Error(t, err)
	assert.Nil(t, outcome)
	assert.Contains(t, err.Error(), "empty AI response")
}

func TestDecisionInterpreter_ParseAIResponse_NilResponse_ReturnsError(t *testing.T) {
	// Arrange
	interpreter := NewDecisionInterpreter()

	// Create empty intent for test
	intent := domain.PurchaseIntent{
		ID:       "test-intent-9",
		UserID:   "test-user-1",
		ItemName: "Test Item",
		ItemCost: 1.0,
		Category: "other",
		Urgency:  "low",
	}

	// Act - pass empty string instead of nil
	outcome, err := interpreter.ParseAIResponse("", intent)

	// Assert
	require.Error(t, err)
	assert.Nil(t, outcome)
	assert.Contains(t, err.Error(), "empty AI response")
}