package domain

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"
)

// AIPrompt represents a structured prompt for AI decision making
type AIPrompt struct {
	SystemContext    string  `json:"system_context"`    // System instructions and role definition
	UserContext      string  `json:"user_context"`      // User's financial, health, and personal context
	PurchaseDetails  string  `json:"purchase_details"`  // Specific purchase information
	DecisionCriteria string  `json:"decision_criteria"` // Decision-making framework
	ResponseFormat   string  `json:"response_format"`   // Required response structure
	MaxTokens        int     `json:"max_tokens"`        // Maximum tokens allowed
	Temperature      float64 `json:"temperature"`       // AI creativity/randomness setting
}

// AIResponse represents the AI's response to a decision prompt
type AIResponse struct {
	RawResponse string   `json:"raw_response"` // Full AI response text
	Decision    string   `json:"decision"`     // Extracted decision (BUY/WAIT/BYE)
	Confidence  float64  `json:"confidence"`   // Extracted confidence score
	Reasoning   string   `json:"reasoning"`    // Extracted reasoning/explanation
	Factors     []string `json:"factors"`      // Key factors mentioned
	Suggestions []string `json:"suggestions"`  // Actionable recommendations
	TokensUsed  int      `json:"tokens_used"`  // Actual tokens consumed
}

// Regular expressions for sensitive data detection
var (
	SSNPattern         = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)           // SSN pattern
	CreditCardPattern  = regexp.MustCompile(`\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`) // Credit card pattern
	PhonePattern       = regexp.MustCompile(`\b\d{3}[-.\s]?\d{3}[-.\s]?\d{4}\b`) // Phone pattern
	EmailPattern       = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`) // Email pattern

	// Keep old names for backward compatibility
	aiSsnRegex    = SSNPattern
	aiCcRegex     = CreditCardPattern
	aiPhoneRegex  = PhonePattern
	aiEmailRegex  = EmailPattern

)

// Validate performs comprehensive validation of the AI prompt
func (ap *AIPrompt) Validate() error {
	// Validate required fields
	if ap.SystemContext == "" {
		return fmt.Errorf("system context is required")
	}

	if ap.UserContext == "" {
		return fmt.Errorf("user context is required")
	}

	if ap.PurchaseDetails == "" {
		return fmt.Errorf("purchase details is required")
	}

	if ap.DecisionCriteria == "" {
		return fmt.Errorf("decision criteria is required")
	}

	if ap.ResponseFormat == "" {
		return fmt.Errorf("response format is required")
	}

	// Validate max tokens
	if ap.MaxTokens < 10 || ap.MaxTokens > 4000 {
		return fmt.Errorf("max tokens must be between 10 and 4000")
	}

	// Validate temperature
	if ap.Temperature < 0.0 || ap.Temperature > 2.0 {
		return fmt.Errorf("temperature must be between 0.0 and 2.0")
	}

	return nil
}

// GetTotalContent combines all prompt fields into a single string
func (ap *AIPrompt) GetTotalContent() string {
	var builder strings.Builder
	
	builder.WriteString("System Context:\n")
	builder.WriteString(ap.SystemContext)
	builder.WriteString("\n\nUser Context:\n")
	builder.WriteString(ap.UserContext)
	builder.WriteString("\n\nPurchase Details:\n")
	builder.WriteString(ap.PurchaseDetails)
	builder.WriteString("\n\nDecision Criteria:\n")
	builder.WriteString(ap.DecisionCriteria)
	builder.WriteString("\n\nResponse Format:\n")
	builder.WriteString(ap.ResponseFormat)
	
	return builder.String()
}

// EstimateTokens provides a rough estimate of token count for the prompt
// Using approximation: 1 token ≈ 0.75 words, 1 word ≈ 4 characters
func (ap *AIPrompt) EstimateTokens() int {
	totalContent := ap.GetTotalContent()
	
	// Rough estimation: 4 characters per token (including spaces)
	charCount := len(totalContent)
	estimatedTokens := charCount / 4
	
	// Add some buffer for formatting and structure
	estimatedTokens = int(float64(estimatedTokens) * 1.2)
	
	return estimatedTokens
}

// IsWithinTokenLimit checks if the estimated token count is within the specified limit
func (ap *AIPrompt) IsWithinTokenLimit() bool {
	estimated := ap.EstimateTokens()
	
	// Reserve 20% of max tokens for response
	inputLimit := int(float64(ap.MaxTokens) * 0.8)
	
	return estimated <= inputLimit
}

// SanitizeForLogging returns a version of the prompt with sensitive information redacted
func (ap *AIPrompt) SanitizeForLogging() string {
	totalContent := ap.GetTotalContent()
	
	// Replace sensitive patterns with [REDACTED]
	sanitized := aiSsnRegex.ReplaceAllString(totalContent, "[REDACTED]")
	sanitized = aiCcRegex.ReplaceAllString(sanitized, "[REDACTED]")
	sanitized = aiPhoneRegex.ReplaceAllString(sanitized, "[REDACTED]")
	sanitized = aiEmailRegex.ReplaceAllString(sanitized, "[REDACTED]")
	
	return sanitized
}

// GetContentHash generates a SHA-256 hash of the prompt content for deduplication
func (ap *AIPrompt) GetContentHash() string {
	content := ap.GetTotalContent()
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// TruncateToLimit truncates the prompt content to fit within token limits
func (ap *AIPrompt) TruncateToLimit() error {
	if ap.IsWithinTokenLimit() {
		return nil // Already within limit
	}
	
	// Calculate how much we need to reduce
	currentTokens := ap.EstimateTokens()
	inputLimit := int(float64(ap.MaxTokens) * 0.8)
	excessTokens := currentTokens - inputLimit
	
	// Convert excess tokens to character count (rough estimate)
	excessChars := excessTokens * 4
	
	// Prioritize truncation: UserContext > PurchaseDetails > others stay intact
	// Start with user context (least critical for decision)
	if len(ap.UserContext) > excessChars {
		ap.UserContext = ap.UserContext[:len(ap.UserContext)-excessChars] + "... [truncated]"
		return nil
	}
	
	// If user context isn't enough, also truncate purchase details
	remaining := excessChars - len(ap.UserContext)
	ap.UserContext = "[truncated due to length]"
	
	if len(ap.PurchaseDetails) > remaining {
		ap.PurchaseDetails = ap.PurchaseDetails[:len(ap.PurchaseDetails)-remaining] + "... [truncated]"
		return nil
	}
	
	// If still too long, this prompt is fundamentally too large
	return fmt.Errorf("prompt content too large to fit within token limit even after truncation")
}

// Validate performs validation on AIResponse
func (ar *AIResponse) Validate() error {
	if ar.RawResponse == "" {
		return fmt.Errorf("raw response is required")
	}
	
	if ar.Decision != "" && !validDecisions[ar.Decision] {
		return fmt.Errorf("invalid decision: %s", ar.Decision)
	}
	
	if ar.Confidence < 0.0 || ar.Confidence > 1.0 {
		return fmt.Errorf("confidence must be between 0.0 and 1.0")
	}
	
	if ar.TokensUsed < 0 {
		return fmt.Errorf("tokens used must be non-negative")
	}
	
	return nil
}

// HasDecision returns true if a valid decision was extracted
func (ar *AIResponse) HasDecision() bool {
	return validDecisions[ar.Decision]
}

// IsHighConfidence returns true if confidence is above threshold
func (ar *AIResponse) IsHighConfidence() bool {
	return ar.Confidence >= 0.8
}

// HasFactors returns true if factors were extracted
func (ar *AIResponse) HasFactors() bool {
	return len(ar.Factors) > 0
}

// HasSuggestions returns true if suggestions were provided
func (ar *AIResponse) HasSuggestions() bool {
	return len(ar.Suggestions) > 0
}

// GetTokenEfficiency calculates tokens used per character of response
func (ar *AIResponse) GetTokenEfficiency() float64 {
	if ar.TokensUsed == 0 || len(ar.RawResponse) == 0 {
		return 0.0
	}
	return float64(ar.TokensUsed) / float64(len(ar.RawResponse))
}

// String returns a string representation of the AI prompt
func (ap *AIPrompt) String() string {
	return fmt.Sprintf("AIPrompt{MaxTokens: %d, Temperature: %.2f, EstimatedTokens: %d}",
		ap.MaxTokens, ap.Temperature, ap.EstimateTokens())
}