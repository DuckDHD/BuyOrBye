package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// EnhancedDecisionService provides comprehensive decision making with GPT-4o-mini integration
type EnhancedDecisionService struct {
	// Core dependencies
	contextAggregator   *ContextAggregator
	promptBuilder       *PromptBuilder
	openaiClient        AIClient
	decisionInterpreter *DecisionInterpreter
	recommendationEngine *RecommendationEngine
	
	// Repositories
	decisionRepo  DecisionRepository
	promptLogRepo PromptLogRepository
	
	// Cache for identical queries (1 hour TTL)
	decisionCache map[string]*cachedDecision
}

// cachedDecision represents a cached decision with TTL
type cachedDecision struct {
	decision  *domain.DecisionOutcome
	timestamp time.Time
	ttl       time.Duration
}

// NewEnhancedDecisionService creates a new enhanced decision service
func NewEnhancedDecisionService(
	contextAggregator *ContextAggregator,
	promptBuilder *PromptBuilder,
	openaiClient AIClient,
	decisionInterpreter *DecisionInterpreter,
	recommendationEngine *RecommendationEngine,
	decisionRepo DecisionRepository,
	promptLogRepo PromptLogRepository,
) *EnhancedDecisionService {
	return &EnhancedDecisionService{
		contextAggregator:   contextAggregator,
		promptBuilder:       promptBuilder,
		openaiClient:        openaiClient,
		decisionInterpreter: decisionInterpreter,
		recommendationEngine: recommendationEngine,
		decisionRepo:        decisionRepo,
		promptLogRepo:       promptLogRepo,
		decisionCache:       make(map[string]*cachedDecision),
	}
}

// MakeDecision implements the main decision flow with GPT-4o-mini integration
func (eds *EnhancedDecisionService) MakeDecision(ctx context.Context, intent domain.PurchaseIntent) (*domain.DecisionOutcome, error) {
	startTime := time.Now()
	
	// Validate purchase intent
	if err := intent.Validate(); err != nil {
		return nil, fmt.Errorf("invalid purchase intent: %w", err)
	}
	
	// Step 1: Check cache for identical query (1 hour TTL)
	cacheKey := eds.generateCacheKey(intent)
	if cached := eds.getCachedDecision(cacheKey); cached != nil {
		fmt.Printf("[DecisionService] Cache hit for key: %s\n", cacheKey[:16]+"...")
		return cached, nil
	}
	
	// Step 2: Aggregate context from Finance/Health services
	userContext, err := eds.contextAggregator.AggregateUserContext(ctx, intent.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate user context: %w", err)
	}
	
	fmt.Printf("[DecisionService] Context aggregated: %s\n", eds.contextAggregator.GetContextSummary(userContext))
	
	// Step 3: Try business rules first (never rely solely on AI)
	businessRuleDecision := eds.recommendationEngine.MakeFallbackDecision(intent, *userContext)
	businessRuleWarnings := eds.recommendationEngine.ValidateBusinessRules(intent, *userContext)
	
	// Step 4: Build structured prompt for GPT-4o-mini
	prompt, err := eds.promptBuilder.BuildPrompt(intent, *userContext)
	if err != nil {
		fmt.Printf("[DecisionService] Prompt building failed, using business rules: %v\n", err)
		decision, err := eds.saveAndReturn(ctx, businessRuleDecision, intent, "business_rules_fallback", startTime)
		return decision, err
	}
	
	// Step 5: Log prompt for debugging
	requestID := eds.generateRequestID()
	promptHash := eds.generatePromptHash(*prompt)
	
	// Check for duplicate prompts to avoid redundant API calls
	if existingResponse, _, err := eds.promptLogRepo.GetPromptByHash(ctx, promptHash); err == nil && existingResponse != nil {
		fmt.Printf("[DecisionService] Duplicate prompt detected, reusing previous result\n")
		// Parse the existing AI response
		if existingDecision, parseErr := eds.decisionInterpreter.ParseResponse(*existingResponse, intent); parseErr == nil {
			return existingDecision, nil
		}
	}
	
	// Log initial prompt
	logID, err := eds.promptLogRepo.LogPrompt(ctx, *prompt, intent.UserID, requestID, intent.ID)
	if err != nil {
		fmt.Printf("[DecisionService] Failed to log prompt: %v\n", err)
	}
	
	// Step 6: Call OpenAI GPT-4o-mini with timeout (30 seconds) and retry
	aiCallStart := time.Now()
	aiResponse, err := eds.callOpenAIWithTimeout(ctx, *prompt, 30*time.Second)
	aiCallDuration := time.Since(aiCallStart)
	
	if err != nil {
		// AI failed - use business rules fallback
		fmt.Printf("[DecisionService] AI call failed (%v), falling back to business rules\n", err)
		
		// Log the failure
		if logID != "" {
			eds.promptLogRepo.UpdateWithError(ctx, logID, err, 500, aiCallDuration.Milliseconds())
		}
		
		// Apply business rules validation to AI decision
		businessRuleDecision.PrimaryReason += " (AI unavailable, using business rules)"
		decision, err := eds.saveAndReturn(ctx, businessRuleDecision, intent, "business_rules_ai_failed", startTime)
		return decision, err
	}
	
	fmt.Printf("[DecisionService] AI response received in %v\n", aiCallDuration)
	
	// Step 7: Parse and validate AI response
	aiDecision, err := eds.decisionInterpreter.ParseResponse(*aiResponse, intent)
	if err != nil {
		fmt.Printf("[DecisionService] AI response parsing failed: %v\n", err)
		
		// Log parsing failure
		if logID != "" {
			parseErr := fmt.Errorf("parsing failed: %w", err)
			eds.promptLogRepo.UpdateWithError(ctx, logID, parseErr, 422, aiCallDuration.Milliseconds())
		}
		
		decision, err := eds.saveAndReturn(ctx, businessRuleDecision, intent, "parsing_failed", startTime)
		return decision, err
	}
	
	// Step 8: Apply business rules validation to AI decision
	if len(businessRuleWarnings) > 0 {
		// Business rules override AI in risky situations
		riskScore := eds.recommendationEngine.GetRecommendationStrength(intent, *userContext, aiDecision.Decision)
		businessScore := eds.recommendationEngine.GetRecommendationStrength(intent, *userContext, businessRuleDecision.Decision)
		
		if businessScore > riskScore && (businessRuleDecision.Decision == "BYE" || businessRuleDecision.Decision == "WAIT") {
			fmt.Printf("[DecisionService] Business rules override AI decision due to risk factors: %v\n", businessRuleWarnings)
			
			// Enhance business rule decision with AI insights
			businessRuleDecision.PrimaryReason += fmt.Sprintf(" (AI suggested %s but business rules override due to risk)", aiDecision.Decision)
			if len(aiDecision.Recommendations) > 0 {
				businessRuleDecision.Recommendations = append(businessRuleDecision.Recommendations, "AI insights: "+aiDecision.Recommendations[0])
			}
			
			// Log successful AI call but business override
			if logID != "" {
				eds.promptLogRepo.UpdateWithResponse(ctx, logID, *aiResponse, aiCallDuration.Milliseconds())
			}
			
			decision, err := eds.saveAndReturn(ctx, businessRuleDecision, intent, "business_rules_override", startTime)
			return decision, err
		}
	}
	
	// Step 9: AI decision is acceptable, enhance it with business insights
	if len(businessRuleWarnings) > 0 {
		warningText := fmt.Sprintf("Note: %d business rule warnings detected", len(businessRuleWarnings))
		aiDecision.Recommendations = append([]string{warningText}, aiDecision.Recommendations...)
	}
	
	// Step 10: Log successful AI interaction
	if logID != "" {
		eds.promptLogRepo.UpdateWithResponse(ctx, logID, *aiResponse, aiCallDuration.Milliseconds())
	}
	
	// Step 11: Save decision and return
	finalDecision, err := eds.saveAndReturn(ctx, aiDecision, intent, "ai_assisted", startTime)
	if err != nil {
		return nil, err
	}
	
	// Cache the decision for 1 hour
	eds.cacheDecision(cacheKey, finalDecision)
	
	return finalDecision, nil
}

// callOpenAIWithTimeout calls OpenAI with a specific timeout
func (eds *EnhancedDecisionService) callOpenAIWithTimeout(ctx context.Context, prompt domain.AIPrompt, timeout time.Duration) (*domain.AIResponse, error) {
	// Create context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	// Make the API call
	response, err := eds.openaiClient.GenerateDecision(timeoutCtx, prompt)
	if err != nil {
		return nil, fmt.Errorf("OpenAI call failed: %w", err)
	}
	
	return response, nil
}

// saveAndReturn saves the decision and logs the outcome
func (eds *EnhancedDecisionService) saveAndReturn(ctx context.Context, decision *domain.DecisionOutcome, intent domain.PurchaseIntent, source string, startTime time.Time) (*domain.DecisionOutcome, error) {
	// Set processing time
	decision.ProcessingTime = time.Since(startTime).Milliseconds()
	
	// Save to repository
	if err := eds.decisionRepo.SaveDecision(ctx, *decision, intent); err != nil {
		fmt.Printf("[DecisionService] Failed to save decision: %v\n", err)
		return nil, fmt.Errorf("failed to save decision: %w", err)
	}
	
	fmt.Printf("[DecisionService] Decision completed: %s (%.2f confidence) via %s in %dms\n",
		decision.Decision, decision.Confidence, source, decision.ProcessingTime)
	
	return decision, nil
}

// generateCacheKey creates a cache key for identical queries
func (eds *EnhancedDecisionService) generateCacheKey(intent domain.PurchaseIntent) string {
	data := fmt.Sprintf("%s|%s|%.2f|%s|%s|%s",
		intent.UserID, intent.ItemName, intent.ItemCost,
		intent.Category, intent.Urgency, intent.Frequency)
	
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// getCachedDecision retrieves a cached decision if still valid
func (eds *EnhancedDecisionService) getCachedDecision(key string) *domain.DecisionOutcome {
	cached, exists := eds.decisionCache[key]
	if !exists {
		return nil
	}
	
	// Check if cache entry is expired (1 hour TTL)
	if time.Since(cached.timestamp) > cached.ttl {
		delete(eds.decisionCache, key)
		return nil
	}
	
	return cached.decision
}

// cacheDecision stores a decision in the cache
func (eds *EnhancedDecisionService) cacheDecision(key string, decision *domain.DecisionOutcome) {
	eds.decisionCache[key] = &cachedDecision{
		decision:  decision,
		timestamp: time.Now(),
		ttl:       1 * time.Hour, // 1 hour TTL as specified
	}
}

// generateRequestID creates a unique request ID for logging
func (eds *EnhancedDecisionService) generateRequestID() string {
	return fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond()%1000)
}

// generatePromptHash creates a hash for prompt deduplication
func (eds *EnhancedDecisionService) generatePromptHash(prompt domain.AIPrompt) string {
	content := prompt.GetTotalContent()
	hash := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", hash)[:32] // First 32 characters
}

// GetDecisionHistory retrieves user's decision history
func (eds *EnhancedDecisionService) GetDecisionHistory(ctx context.Context, userID string, limit, offset int) ([]*domain.DecisionOutcome, error) {
	return eds.decisionRepo.GetDecisionHistory(ctx, userID, limit, offset)
}

// GetDecisionStats retrieves decision statistics for the user  
func (eds *EnhancedDecisionService) GetDecisionStats(ctx context.Context, userID string, daysBack int) (*DecisionStats, error) {
	return eds.decisionRepo.GetDecisionStats(ctx, userID, daysBack)
}

// CleanupCache removes expired cache entries
func (eds *EnhancedDecisionService) CleanupCache() {
	now := time.Now()
	for key, cached := range eds.decisionCache {
		if now.Sub(cached.timestamp) > cached.ttl {
			delete(eds.decisionCache, key)
		}
	}
}

// GetCacheStats returns cache statistics
func (eds *EnhancedDecisionService) GetCacheStats() map[string]interface{} {
	return map[string]interface{}{
		"entries":     len(eds.decisionCache),
		"cache_size":  len(eds.decisionCache),
		"last_cleanup": time.Now(),
	}
}

// ValidateAIResponse checks if AI response meets quality standards
func (eds *EnhancedDecisionService) ValidateAIResponse(response *domain.AIResponse, intent domain.PurchaseIntent) error {
	if response.Decision == "" {
		return fmt.Errorf("no decision found in AI response")
	}
	
	validDecisions := map[string]bool{"BUY": true, "WAIT": true, "BYE": true}
	if !validDecisions[response.Decision] {
		return fmt.Errorf("invalid decision: %s", response.Decision)
	}
	
	if response.Confidence < 0.0 || response.Confidence > 1.0 {
		return fmt.Errorf("invalid confidence: %f", response.Confidence)
	}
	
	if len(response.Reasoning) < 10 {
		return fmt.Errorf("reasoning too short: %s", response.Reasoning)
	}
	
	return nil
}

// GetServiceHealth returns the health status of the decision service
func (eds *EnhancedDecisionService) GetServiceHealth(ctx context.Context) map[string]interface{} {
	health := map[string]interface{}{
		"service":       "enhanced_decision_service",
		"status":        "healthy",
		"timestamp":     time.Now(),
		"cache_entries": len(eds.decisionCache),
	}
	
	// Test basic functionality
	testIntent := domain.PurchaseIntent{
		UserID:      "health_check",
		ItemName:    "Test Item",
		ItemCost:    100.0,
		Category:    "other",
		Urgency:     "low",
		Frequency:   "one_time",
		CreatedAt:   time.Now(),
	}
	
	// Test context aggregation
	_, err := eds.contextAggregator.AggregateUserContext(ctx, "health_check")
	if err != nil {
		health["context_aggregator"] = "error: " + err.Error()
		health["status"] = "degraded"
	} else {
		health["context_aggregator"] = "healthy"
	}
	
	// Test prompt building
	defaultContext := &domain.DecisionContext{
		UserID: "health_check",
		FinancialContext: *eds.contextAggregator.getDefaultFinancialSnapshot(),
		HealthContext: *eds.contextAggregator.getDefaultHealthSnapshot(),
		CurrentDate: time.Now(),
	}
	
	_, err = eds.promptBuilder.BuildPrompt(testIntent, *defaultContext)
	if err != nil {
		health["prompt_builder"] = "error: " + err.Error()
		health["status"] = "degraded"
	} else {
		health["prompt_builder"] = "healthy"
	}
	
	return health
}