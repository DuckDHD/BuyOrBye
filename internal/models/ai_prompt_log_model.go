package models

import (
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"gorm.io/gorm"
)

// AIPromptLogModel represents AI interaction logs for debugging and analytics
type AIPromptLogModel struct {
	gorm.Model
	
	// Request identifiers
	UserID          string `gorm:"type:varchar(255);not null;index:idx_user_ai_logs" json:"user_id"`
	RequestID       string `gorm:"type:varchar(255);not null;index:idx_request_logs" json:"request_id"`
	IntentID        string `gorm:"type:varchar(255);index:idx_intent_logs" json:"intent_id"`
	
	// AI Provider details
	AIProvider      string `gorm:"type:varchar(50);not null;default:'openai';index:idx_provider_logs" json:"ai_provider"`
	AIModel         string `gorm:"type:varchar(100);not null;default:'gpt-4o-mini';column:model" json:"model"`
	Temperature     float64 `gorm:"type:decimal(3,2);default:0.7" json:"temperature"`
	
	// Prompt content
	SystemContext    string `gorm:"type:text" json:"system_context"`
	UserContext      string `gorm:"type:text" json:"user_context"`
	PurchaseDetails  string `gorm:"type:text" json:"purchase_details"`
	DecisionCriteria string `gorm:"type:text" json:"decision_criteria"`
	ResponseFormat   string `gorm:"type:text" json:"response_format"`
	
	// Token usage
	MaxTokens       int `gorm:"type:int;not null" json:"max_tokens"`
	TokensInput     int `gorm:"type:int;default:0" json:"tokens_input"`
	TokensOutput    int `gorm:"type:int;default:0" json:"tokens_output"`
	TokensTotal     int `gorm:"type:int;default:0" json:"tokens_total"`
	
	// AI Response
	RawResponse     string `gorm:"type:text" json:"raw_response"`
	ParsedDecision  string `gorm:"type:varchar(10);index:idx_parsed_decision" json:"parsed_decision"`
	ParsedConfidence float64 `gorm:"type:decimal(3,2)" json:"parsed_confidence"`
	
	// Performance metrics
	ResponseTimeMs  int64 `gorm:"type:bigint;not null" json:"response_time_ms"`
	ProcessingTimeMs int64 `gorm:"type:bigint;not null" json:"processing_time_ms"`
	
	// Status and error tracking
	Success      bool   `gorm:"type:boolean;default:false;index:idx_success_logs" json:"success"`
	ErrorMessage string `gorm:"type:text" json:"error_message"`
	StatusCode   int    `gorm:"type:int;default:0" json:"status_code"`
	
	// Cost tracking (optional, for monitoring)
	EstimatedCostUSD float64 `gorm:"type:decimal(10,6);default:0" json:"estimated_cost_usd"`
	
	// Composite indexes will be created in migration
}

// TableName specifies the table name for GORM
func (AIPromptLogModel) TableName() string {
	return "ai_prompt_logs"
}

// BeforeCreate hook to set defaults and validate
func (m *AIPromptLogModel) BeforeCreate(tx *gorm.DB) error {
	return m.validate()
}

// BeforeUpdate hook to validate updates
func (m *AIPromptLogModel) BeforeUpdate(tx *gorm.DB) error {
	return m.validate()
}

// validate performs model validation
func (m *AIPromptLogModel) validate() error {
	if m.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	
	if m.RequestID == "" {
		return fmt.Errorf("request_id is required")
	}
	
	if m.AIProvider == "" {
		m.AIProvider = "openai" // Default provider
	}
	
	if m.AIModel == "" {
		m.AIModel = "gpt-4o-mini" // Default model
	}
	
	if m.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be greater than 0")
	}
	
	if m.Temperature < 0.0 || m.Temperature > 2.0 {
		return fmt.Errorf("temperature must be between 0.0 and 2.0")
	}
	
	if m.ParsedConfidence < 0.0 || m.ParsedConfidence > 1.0 {
		return fmt.Errorf("parsed_confidence must be between 0.0 and 1.0")
	}
	
	validDecisions := map[string]bool{"BUY": true, "WAIT": true, "BYE": true, "": true}
	if !validDecisions[m.ParsedDecision] {
		return fmt.Errorf("invalid parsed_decision: %s", m.ParsedDecision)
	}
	
	// Calculate total tokens
	m.TokensTotal = m.TokensInput + m.TokensOutput
	
	return nil
}

// FromDomainPrompt creates AIPromptLogModel from domain.AIPrompt
func (m *AIPromptLogModel) FromDomainPrompt(prompt domain.AIPrompt, userID, requestID, intentID string) {
	m.UserID = userID
	m.RequestID = requestID
	m.IntentID = intentID
	m.SystemContext = prompt.SystemContext
	m.UserContext = prompt.UserContext
	m.PurchaseDetails = prompt.PurchaseDetails
	m.DecisionCriteria = prompt.DecisionCriteria
	m.ResponseFormat = prompt.ResponseFormat
	m.MaxTokens = prompt.MaxTokens
	m.Temperature = prompt.Temperature
	
	// Estimate input tokens
	m.TokensInput = prompt.EstimateTokens()
}

// UpdateFromResponse updates the model with AI response data
func (m *AIPromptLogModel) UpdateFromResponse(response domain.AIResponse, processingTimeMs int64) {
	m.RawResponse = response.RawResponse
	m.ParsedDecision = response.Decision
	m.ParsedConfidence = response.Confidence
	m.TokensOutput = response.TokensUsed
	m.TokensTotal = m.TokensInput + m.TokensOutput
	m.ProcessingTimeMs = processingTimeMs
	m.Success = true
	
	// Calculate estimated cost (rough approximation for GPT-4o-mini)
	m.EstimatedCostUSD = m.calculateEstimatedCost()
}

// UpdateFromError updates the model with error information
func (m *AIPromptLogModel) UpdateFromError(err error, statusCode int, processingTimeMs int64) {
	m.ErrorMessage = err.Error()
	m.StatusCode = statusCode
	m.ProcessingTimeMs = processingTimeMs
	m.Success = false
}

// calculateEstimatedCost estimates the cost based on token usage
// Using approximate GPT-4o-mini pricing: $0.15/1M input tokens, $0.60/1M output tokens
func (m *AIPromptLogModel) calculateEstimatedCost() float64 {
	inputCost := float64(m.TokensInput) * 0.15 / 1000000  // $0.15 per 1M tokens
	outputCost := float64(m.TokensOutput) * 0.60 / 1000000 // $0.60 per 1M tokens
	return inputCost + outputCost
}

// GetPromptContent returns the combined prompt content
func (m *AIPromptLogModel) GetPromptContent() string {
	return m.SystemContext + "\n\n" + m.UserContext + "\n\n" + 
		   m.PurchaseDetails + "\n\n" + m.DecisionCriteria + "\n\n" + 
		   m.ResponseFormat
}

// GetTokenEfficiency calculates tokens used per character
func (m *AIPromptLogModel) GetTokenEfficiency() float64 {
	contentLength := len(m.GetPromptContent())
	if contentLength == 0 || m.TokensInput == 0 {
		return 0.0
	}
	return float64(m.TokensInput) / float64(contentLength)
}

// IsSuccessful returns true if the AI call was successful
func (m *AIPromptLogModel) IsSuccessful() bool {
	return m.Success
}

// HasValidDecision returns true if a valid decision was parsed
func (m *AIPromptLogModel) HasValidDecision() bool {
	validDecisions := map[string]bool{"BUY": true, "WAIT": true, "BYE": true}
	return validDecisions[m.ParsedDecision]
}

// IsHighConfidence returns true if parsed confidence is high
func (m *AIPromptLogModel) IsHighConfidence() bool {
	return m.ParsedConfidence >= 0.8
}

// GetResponseTimeSeconds returns response time in seconds
func (m *AIPromptLogModel) GetResponseTimeSeconds() float64 {
	return float64(m.ResponseTimeMs) / 1000.0
}

// GetProcessingTimeSeconds returns processing time in seconds
func (m *AIPromptLogModel) GetProcessingTimeSeconds() float64 {
	return float64(m.ProcessingTimeMs) / 1000.0
}

// IsExpensive returns true if the request was costly
func (m *AIPromptLogModel) IsExpensive(threshold float64) bool {
	return m.EstimatedCostUSD > threshold
}

// GetProvider returns the AI provider used
func (m *AIPromptLogModel) GetProvider() string {
	if m.AIProvider == "" {
		return "unknown"
	}
	return m.AIProvider
}

// GetModelInfo returns formatted model information
func (m *AIPromptLogModel) GetModelInfo() string {
	return fmt.Sprintf("%s/%s", m.AIProvider, m.AIModel)
}

// HasError returns true if there was an error
func (m *AIPromptLogModel) HasError() bool {
	return !m.Success || m.ErrorMessage != ""
}

// GetAgeInHours returns how many hours ago this log was created
func (m *AIPromptLogModel) GetAgeInHours() int {
	return int(time.Since(m.CreatedAt).Hours())
}

// IsRecent returns true if log was created within specified hours
func (m *AIPromptLogModel) IsRecent(hours int) bool {
	return m.GetAgeInHours() <= hours
}

// GetQualityScore returns a quality score based on success, confidence, and response time
func (m *AIPromptLogModel) GetQualityScore() float64 {
	if !m.Success {
		return 0.0
	}
	
	score := m.ParsedConfidence * 0.6 // 60% weight for confidence
	
	// Response time factor (faster is better, penalty for >30s)
	timeScore := 1.0
	if m.ResponseTimeMs > 30000 { // >30 seconds
		timeScore = 0.3
	} else if m.ResponseTimeMs > 10000 { // >10 seconds
		timeScore = 0.7
	}
	score += timeScore * 0.3 // 30% weight for response time
	
	// Valid decision factor
	if m.HasValidDecision() {
		score += 0.1 // 10% bonus for valid decision
	}
	
	return score
}

// GetSummary returns a brief summary of the AI interaction
func (m *AIPromptLogModel) GetSummary() string {
	status := "failed"
	if m.Success {
		status = "success"
	}
	
	return fmt.Sprintf("AI Request: %s - Decision: %s (%.2f confidence) - %s in %dms",
		m.RequestID, m.ParsedDecision, m.ParsedConfidence, status, m.ResponseTimeMs)
}