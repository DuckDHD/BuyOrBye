package repositories

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/DuckDHD/BuyOrBye/internal/services"
)

// PromptLogRepository implements the PromptLogRepository interface using GORM
type PromptLogRepository struct {
	db *gorm.DB
}

// NewPromptLogRepository creates a new instance of PromptLogRepository
func NewPromptLogRepository(db *gorm.DB) services.PromptLogRepository {
	return &PromptLogRepository{
		db: db,
	}
}

// LogPrompt creates an initial AI prompt log entry
func (r *PromptLogRepository) LogPrompt(ctx context.Context, prompt domain.AIPrompt, userID, requestID, intentID string) (string, error) {
	// Create model instance
	record := &models.AIPromptLogModel{}
	
	// Populate from domain prompt
	record.FromDomainPrompt(prompt, userID, requestID, intentID)
	
	// Save to database
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return "", fmt.Errorf("failed to save prompt log: %w", err)
	}
	
	// Return the generated ID as string
	return strconv.FormatUint(uint64(record.ID), 10), nil
}

// UpdateWithResponse updates the prompt log with AI response data
func (r *PromptLogRepository) UpdateWithResponse(ctx context.Context, logID string, response domain.AIResponse, processingTimeMs int64) error {
	// Parse string ID to uint
	id, err := strconv.ParseUint(logID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid log ID format: %w", err)
	}
	
	// Get the existing record
	var record models.AIPromptLogModel
	if err := r.db.WithContext(ctx).First(&record, uint(id)).Error; err != nil {
		return fmt.Errorf("failed to find prompt log: %w", err)
	}
	
	// Update with response data
	record.UpdateFromResponse(response, processingTimeMs)
	
	// Save updated record
	if err := r.db.WithContext(ctx).Save(&record).Error; err != nil {
		return fmt.Errorf("failed to update prompt log with response: %w", err)
	}
	
	return nil
}

// UpdateWithError updates the prompt log with error information
func (r *PromptLogRepository) UpdateWithError(ctx context.Context, logID string, err error, statusCode int, processingTimeMs int64) error {
	// Parse string ID to uint
	id, errParse := strconv.ParseUint(logID, 10, 32)
	if errParse != nil {
		return fmt.Errorf("invalid log ID format: %w", errParse)
	}
	
	// Get the existing record
	var record models.AIPromptLogModel
	if errFind := r.db.WithContext(ctx).First(&record, uint(id)).Error; errFind != nil {
		return fmt.Errorf("failed to find prompt log: %w", errFind)
	}
	
	// Update with error data
	record.UpdateFromError(err, statusCode, processingTimeMs)
	
	// Save updated record
	if errSave := r.db.WithContext(ctx).Save(&record).Error; errSave != nil {
		return fmt.Errorf("failed to update prompt log with error: %w", errSave)
	}
	
	return nil
}

// GeneratePromptHash creates a SHA-256 hash of the prompt content for deduplication
func (r *PromptLogRepository) GeneratePromptHash(prompt domain.AIPrompt) string {
	return prompt.GetContentHash()
}

// GetPromptByHash retrieves a cached AI response by prompt content hash
func (r *PromptLogRepository) GetPromptByHash(ctx context.Context, promptHash string) (*domain.AIResponse, bool, error) {
	// Create a hash column for querying (we'll need to compute hash during save)
	// For now, we'll search by comparing the computed hash with stored content
	var records []models.AIPromptLogModel
	
	// Get successful prompts only (failed requests shouldn't be cached)
	if err := r.db.WithContext(ctx).
		Where("success = ?", true).
		Where("parsed_decision != ''").
		Find(&records).Error; err != nil {
		return nil, false, fmt.Errorf("failed to search for cached prompts: %w", err)
	}
	
	// Check each record's content hash
	for _, rec := range records {
		// Reconstruct prompt from stored data
		prompt := domain.AIPrompt{
			SystemContext:    rec.SystemContext,
			UserContext:      rec.UserContext,
			PurchaseDetails:  rec.PurchaseDetails,
			DecisionCriteria: rec.DecisionCriteria,
			ResponseFormat:   rec.ResponseFormat,
			MaxTokens:        rec.MaxTokens,
			Temperature:      rec.Temperature,
		}
		
		if prompt.GetContentHash() == promptHash {
			// Found matching hash, return cached response
			response := &domain.AIResponse{
				RawResponse: rec.RawResponse,
				Decision:    rec.ParsedDecision,
				Confidence:  rec.ParsedConfidence,
				TokensUsed:  rec.TokensOutput,
			}
			
			return response, true, nil
		}
	}
	
	// No matching hash found
	return nil, false, nil
}

// GetFailedPrompts retrieves recent failed AI requests for debugging
func (r *PromptLogRepository) GetFailedPrompts(ctx context.Context, hoursBack int) (interface{}, error) {
	var records []models.AIPromptLogModel
	
	// Calculate time filter
	cutoffTime := time.Now().Add(-time.Duration(hoursBack) * time.Hour)
	
	if err := r.db.WithContext(ctx).
		Where("success = ? AND created_at > ?", false, cutoffTime).
		Order("created_at DESC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get failed prompts: %w", err)
	}
	
	// Convert to slice of pointers
	result := make([]*models.AIPromptLogModel, len(records))
	for i := range records {
		result[i] = &records[i]
	}
	
	return result, nil
}

// GetRecentPrompts retrieves recent AI prompts for a user
func (r *PromptLogRepository) GetRecentPrompts(ctx context.Context, userID string, hoursBack int) (interface{}, error) {
	var records []models.AIPromptLogModel
	
	// Calculate time filter
	cutoffTime := time.Now().Add(-time.Duration(hoursBack) * time.Hour)
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND created_at > ?", userID, cutoffTime).
		Order("created_at DESC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent prompts: %w", err)
	}
	
	// Convert to slice of pointers
	result := make([]*models.AIPromptLogModel, len(records))
	for i := range records {
		result[i] = &records[i]
	}
	
	return result, nil
}

// GetTokenUsageStats calculates token usage statistics for a user
func (r *PromptLogRepository) GetTokenUsageStats(ctx context.Context, userID string, hoursBack int) (*services.TokenUsageStats, error) {
	// Calculate time filter
	cutoffTime := time.Now().Add(-time.Duration(hoursBack) * time.Hour)
	
	stats := &services.TokenUsageStats{}
	
	// Get aggregate statistics
	var aggregateResult struct {
		TotalInputTokens    int64
		TotalOutputTokens   int64
		TotalTokens         int64
		SuccessfulRequests  int64
		FailedRequests      int64
		EstimatedTotalCost  float64
		AvgResponseTime     float64
	}
	
	if err := r.db.WithContext(ctx).Model(&models.AIPromptLogModel{}).
		Select(`
			SUM(tokens_input) as total_input_tokens,
			SUM(tokens_output) as total_output_tokens,
			SUM(tokens_total) as total_tokens,
			SUM(CASE WHEN success = true THEN 1 ELSE 0 END) as successful_requests,
			SUM(CASE WHEN success = false THEN 1 ELSE 0 END) as failed_requests,
			SUM(estimated_cost_usd) as estimated_total_cost,
			AVG(response_time_ms) as avg_response_time
		`).
		Where("user_id = ? AND created_at > ?", userID, cutoffTime).
		Scan(&aggregateResult).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate token usage stats: %w", err)
	}
	
	stats.TotalInputTokens = aggregateResult.TotalInputTokens
	stats.TotalOutputTokens = aggregateResult.TotalOutputTokens
	stats.TotalTokens = aggregateResult.TotalTokens
	stats.SuccessfulRequests = aggregateResult.SuccessfulRequests
	stats.FailedRequests = aggregateResult.FailedRequests
	stats.EstimatedTotalCost = aggregateResult.EstimatedTotalCost
	stats.AverageResponseTime = aggregateResult.AvgResponseTime
	
	// Calculate derived statistics
	totalRequests := stats.SuccessfulRequests + stats.FailedRequests
	if totalRequests > 0 {
		stats.SuccessRate = float64(stats.SuccessfulRequests) / float64(totalRequests)
		stats.AverageTokensPerReq = float64(stats.TotalTokens) / float64(totalRequests)
	}
	
	return stats, nil
}

// GetPromptsByProvider retrieves AI prompts filtered by provider
func (r *PromptLogRepository) GetPromptsByProvider(ctx context.Context, userID string, provider string, hoursBack int) (interface{}, error) {
	var records []models.AIPromptLogModel
	
	// Calculate time filter
	cutoffTime := time.Now().Add(-time.Duration(hoursBack) * time.Hour)
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND ai_provider = ? AND created_at > ?", userID, provider, cutoffTime).
		Order("created_at DESC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get prompts by provider: %w", err)
	}
	
	// Convert to slice of pointers
	result := make([]*models.AIPromptLogModel, len(records))
	for i := range records {
		result[i] = &records[i]
	}
	
	return result, nil
}

// Additional helper methods for prompt log repository functionality

// GetPromptByID retrieves a single prompt log by ID
func (r *PromptLogRepository) GetPromptByID(ctx context.Context, logID string) (*models.AIPromptLogModel, error) {
	// Parse string ID to uint
	id, err := strconv.ParseUint(logID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid log ID format: %w", err)
	}
	
	var record models.AIPromptLogModel
	if err := r.db.WithContext(ctx).First(&record, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("prompt log not found with ID %s", logID)
		}
		return nil, fmt.Errorf("failed to get prompt log: %w", err)
	}
	
	return &record, nil
}

// GetPromptsByIntentID retrieves all prompt logs for a specific intent
func (r *PromptLogRepository) GetPromptsByIntentID(ctx context.Context, intentID string) ([]*models.AIPromptLogModel, error) {
	var records []models.AIPromptLogModel
	
	if err := r.db.WithContext(ctx).
		Where("intent_id = ?", intentID).
		Order("created_at ASC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get prompts by intent ID: %w", err)
	}
	
	// Convert to slice of pointers
	result := make([]*models.AIPromptLogModel, len(records))
	for i := range records {
		result[i] = &records[i]
	}
	
	return result, nil
}

// GetPromptCount returns the total number of prompts for a user
func (r *PromptLogRepository) GetPromptCount(ctx context.Context, userID string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.AIPromptLogModel{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to get prompt count: %w", err)
	}
	return count, nil
}

// GetLatestPrompt returns the most recent prompt for a user
func (r *PromptLogRepository) GetLatestPrompt(ctx context.Context, userID string) (*models.AIPromptLogModel, error) {
	var record models.AIPromptLogModel
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no prompts found for user %s", userID)
		}
		return nil, fmt.Errorf("failed to get latest prompt: %w", err)
	}
	
	return &record, nil
}

// DeletePromptLog removes a prompt log (for cleanup or testing purposes)
func (r *PromptLogRepository) DeletePromptLog(ctx context.Context, logID string) error {
	// Parse string ID to uint
	id, err := strconv.ParseUint(logID, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid log ID format: %w", err)
	}
	
	result := r.db.WithContext(ctx).Delete(&models.AIPromptLogModel{}, uint(id))
	if result.Error != nil {
		return fmt.Errorf("failed to delete prompt log: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("prompt log not found with ID %s", logID)
	}
	
	return nil
}

// GetExpensivePrompts retrieves prompts that exceeded cost threshold
func (r *PromptLogRepository) GetExpensivePrompts(ctx context.Context, costThreshold float64, hoursBack int) ([]*models.AIPromptLogModel, error) {
	var records []models.AIPromptLogModel
	
	// Calculate time filter
	cutoffTime := time.Now().Add(-time.Duration(hoursBack) * time.Hour)
	
	if err := r.db.WithContext(ctx).
		Where("estimated_cost_usd > ? AND created_at > ?", costThreshold, cutoffTime).
		Order("estimated_cost_usd DESC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get expensive prompts: %w", err)
	}
	
	// Convert to slice of pointers
	result := make([]*models.AIPromptLogModel, len(records))
	for i := range records {
		result[i] = &records[i]
	}
	
	return result, nil
}

// GetSlowPrompts retrieves prompts that took longer than response time threshold
func (r *PromptLogRepository) GetSlowPrompts(ctx context.Context, responseTimeMs int64, hoursBack int) ([]*models.AIPromptLogModel, error) {
	var records []models.AIPromptLogModel
	
	// Calculate time filter
	cutoffTime := time.Now().Add(-time.Duration(hoursBack) * time.Hour)
	
	if err := r.db.WithContext(ctx).
		Where("response_time_ms > ? AND created_at > ?", responseTimeMs, cutoffTime).
		Order("response_time_ms DESC").
		Find(&records).Error; err != nil {
		return nil, fmt.Errorf("failed to get slow prompts: %w", err)
	}
	
	// Convert to slice of pointers
	result := make([]*models.AIPromptLogModel, len(records))
	for i := range records {
		result[i] = &records[i]
	}
	
	return result, nil
}

// GetProviderStats returns usage statistics by AI provider
func (r *PromptLogRepository) GetProviderStats(ctx context.Context, userID string, hoursBack int) (map[string]*services.TokenUsageStats, error) {
	// Calculate time filter
	cutoffTime := time.Now().Add(-time.Duration(hoursBack) * time.Hour)
	
	// Get list of providers used
	var providers []string
	if err := r.db.WithContext(ctx).Model(&models.AIPromptLogModel{}).
		Distinct("ai_provider").
		Where("user_id = ? AND created_at > ?", userID, cutoffTime).
		Pluck("ai_provider", &providers).Error; err != nil {
		return nil, fmt.Errorf("failed to get providers: %w", err)
	}
	
	result := make(map[string]*services.TokenUsageStats)
	
	// Get stats for each provider
	for _, provider := range providers {
		stats, err := r.getProviderTokenStats(ctx, userID, provider, cutoffTime)
		if err != nil {
			return nil, fmt.Errorf("failed to get stats for provider %s: %w", provider, err)
		}
		result[provider] = stats
	}
	
	return result, nil
}

// getProviderTokenStats is a helper method to get token stats for a specific provider
func (r *PromptLogRepository) getProviderTokenStats(ctx context.Context, userID, provider string, cutoffTime time.Time) (*services.TokenUsageStats, error) {
	stats := &services.TokenUsageStats{}
	
	var aggregateResult struct {
		TotalInputTokens    int64
		TotalOutputTokens   int64
		TotalTokens         int64
		SuccessfulRequests  int64
		FailedRequests      int64
		EstimatedTotalCost  float64
		AvgResponseTime     float64
	}
	
	if err := r.db.WithContext(ctx).Model(&models.AIPromptLogModel{}).
		Select(`
			SUM(tokens_input) as total_input_tokens,
			SUM(tokens_output) as total_output_tokens,
			SUM(tokens_total) as total_tokens,
			SUM(CASE WHEN success = true THEN 1 ELSE 0 END) as successful_requests,
			SUM(CASE WHEN success = false THEN 1 ELSE 0 END) as failed_requests,
			SUM(estimated_cost_usd) as estimated_total_cost,
			AVG(response_time_ms) as avg_response_time
		`).
		Where("user_id = ? AND ai_provider = ? AND created_at > ?", userID, provider, cutoffTime).
		Scan(&aggregateResult).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate provider token stats: %w", err)
	}
	
	stats.TotalInputTokens = aggregateResult.TotalInputTokens
	stats.TotalOutputTokens = aggregateResult.TotalOutputTokens
	stats.TotalTokens = aggregateResult.TotalTokens
	stats.SuccessfulRequests = aggregateResult.SuccessfulRequests
	stats.FailedRequests = aggregateResult.FailedRequests
	stats.EstimatedTotalCost = aggregateResult.EstimatedTotalCost
	stats.AverageResponseTime = aggregateResult.AvgResponseTime
	
	// Calculate derived statistics
	totalRequests := stats.SuccessfulRequests + stats.FailedRequests
	if totalRequests > 0 {
		stats.SuccessRate = float64(stats.SuccessfulRequests) / float64(totalRequests)
		stats.AverageTokensPerReq = float64(stats.TotalTokens) / float64(totalRequests)
	}
	
	return stats, nil
}