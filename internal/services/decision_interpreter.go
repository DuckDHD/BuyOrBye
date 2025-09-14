package services

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

type DecisionInterpreter struct {
	validDecisions map[string]bool
}

func NewDecisionInterpreter() *DecisionInterpreter {
	return &DecisionInterpreter{
		validDecisions: map[string]bool{
			"BUY":  true,
			"WAIT": true,
			"BYE":  true,
		},
	}
}

type AIDecisionResponse struct {
	Decision        string                  `json:"decision"`
	Confidence      float64                 `json:"confidence"`
	Reasoning       string                  `json:"reasoning"`
	Factors         []AIDecisionFactor      `json:"factors"`
	Recommendations []string                `json:"recommendations"`
	WaitPeriod      int                     `json:"wait_period"`
	MaxBudget       float64                 `json:"max_budget"`
}

type AIDecisionFactor struct {
	Category    string  `json:"category"`
	Impact      string  `json:"impact"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
}

func (di *DecisionInterpreter) ParseResponse(aiResponse domain.AIResponse, intent domain.PurchaseIntent) (*domain.DecisionOutcome, error) {
	if aiResponse.RawResponse == "" {
		return nil, fmt.Errorf("empty AI response")
	}

	// Try JSON parsing first
	outcome, err := di.parseJSONResponse(aiResponse.RawResponse, intent)
	if err == nil {
		return outcome, nil
	}

	// Fallback to plain text parsing
	outcome, err = di.parsePlainTextResponse(aiResponse.RawResponse, intent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return outcome, nil
}

func (di *DecisionInterpreter) parseJSONResponse(rawResponse string, intent domain.PurchaseIntent) (*domain.DecisionOutcome, error) {
	// Clean up response (remove markdown code blocks if present)
	cleanResponse := strings.TrimSpace(rawResponse)
	cleanResponse = strings.TrimPrefix(cleanResponse, "```json")
	cleanResponse = strings.TrimPrefix(cleanResponse, "```")
	cleanResponse = strings.TrimSuffix(cleanResponse, "```")
	cleanResponse = strings.TrimSpace(cleanResponse)

	var aiResp AIDecisionResponse
	if err := json.Unmarshal([]byte(cleanResponse), &aiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Validate and convert to domain object
	return di.convertAIResponseToDomain(aiResp, intent)
}

func (di *DecisionInterpreter) parsePlainTextResponse(rawResponse string, intent domain.PurchaseIntent) (*domain.DecisionOutcome, error) {
	// Extract decision
	decision, err := di.extractDecision(rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to extract decision: %w", err)
	}

	// Extract confidence
	confidence := di.extractConfidence(rawResponse)

	// Extract reasoning
	reasoning := di.extractReasoning(rawResponse)

	// Extract recommendations
	recommendations := di.extractRecommendations(rawResponse)

	// Create basic factors based on decision
	factors := di.createBasicFactors(decision, confidence)

	// Extract wait period and budget
	waitPeriod := di.extractWaitPeriod(rawResponse)
	maxBudget := di.extractMaxBudget(rawResponse, intent.ItemCost)

	outcome := &domain.DecisionOutcome{
		UserID:          intent.UserID,
		IntentID:        intent.ID,
		Decision:        decision,
		Confidence:      confidence,
		PrimaryReason:   reasoning,
		Factors:         factors,
		Recommendations: recommendations,
		WaitPeriod:      waitPeriod,
		MaxBudget:       maxBudget,
		CreatedAt:       time.Now(),
	}

	return outcome, nil
}

func (di *DecisionInterpreter) convertAIResponseToDomain(aiResp AIDecisionResponse, intent domain.PurchaseIntent) (*domain.DecisionOutcome, error) {
	// Validate decision
	if !di.validDecisions[strings.ToUpper(aiResp.Decision)] {
		return nil, fmt.Errorf("invalid decision: %s", aiResp.Decision)
	}

	// Validate confidence
	if aiResp.Confidence < 0.0 || aiResp.Confidence > 1.0 {
		return nil, fmt.Errorf("invalid confidence: %f", aiResp.Confidence)
	}

	// Convert factors
	var factors []domain.DecisionFactor
	for _, factor := range aiResp.Factors {
		if len(factor.Description) > 200 {
			factor.Description = factor.Description[:197] + "..."
		}
		
		factors = append(factors, domain.DecisionFactor{
			Category:    factor.Category,
			Impact:      factor.Impact,
			Weight:      factor.Weight,
			Description: factor.Description,
		})
	}

	// Validate recommendations length
	var recommendations []string
	for _, rec := range aiResp.Recommendations {
		if len(rec) > 200 {
			rec = rec[:197] + "..."
		}
		recommendations = append(recommendations, rec)
	}

	// Validate primary reason length
	primaryReason := aiResp.Reasoning
	if len(primaryReason) > 500 {
		primaryReason = primaryReason[:497] + "..."
	}

	outcome := &domain.DecisionOutcome{
		UserID:          intent.UserID,
		IntentID:        intent.ID,
		Decision:        strings.ToUpper(aiResp.Decision),
		Confidence:      aiResp.Confidence,
		PrimaryReason:   primaryReason,
		Factors:         factors,
		Recommendations: recommendations,
		WaitPeriod:      aiResp.WaitPeriod,
		MaxBudget:       aiResp.MaxBudget,
		CreatedAt:       time.Now(),
	}

	// Validate the outcome
	if err := outcome.Validate(); err != nil {
		return nil, fmt.Errorf("invalid decision outcome: %w", err)
	}

	return outcome, nil
}

func (di *DecisionInterpreter) extractDecision(text string) (string, error) {
	// Look for decision patterns
	decisionPatterns := []string{
		`(?i)decision[:\s]+(\w+)`,
		`(?i)recommendation[:\s]+(\w+)`,
		`(?i)(BUY|WAIT|BYE)`,
		`(?i)I recommend[:\s]+(\w+)`,
	}

	for _, pattern := range decisionPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			decision := strings.ToUpper(matches[1])
			if di.validDecisions[decision] {
				return decision, nil
			}
		}
	}

	// Default to BYE if parsing fails (conservative approach)
	return "BYE", nil
}

func (di *DecisionInterpreter) extractConfidence(text string) float64 {
	// Look for confidence patterns
	patterns := []string{
		`(?i)confidence[:\s]+(\d+(?:\.\d+)?)%?`,
		`(?i)(\d+(?:\.\d+)?)%?\s+confidence`,
		`(?i)certainty[:\s]+(\d+(?:\.\d+)?)%?`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			if conf, err := strconv.ParseFloat(matches[1], 64); err == nil {
				// If it's a percentage over 1.0, convert to decimal
				if conf > 1.0 {
					conf = conf / 100.0
				}
				if conf >= 0.0 && conf <= 1.0 {
					return conf
				}
			}
		}
	}

	// Default confidence based on keywords
	text = strings.ToLower(text)
	if strings.Contains(text, "highly recommend") || strings.Contains(text, "strongly") {
		return 0.9
	}
	if strings.Contains(text, "recommend") || strings.Contains(text, "suggest") {
		return 0.7
	}
	if strings.Contains(text, "maybe") || strings.Contains(text, "consider") {
		return 0.5
	}

	return 0.6 // Default medium confidence
}

func (di *DecisionInterpreter) extractReasoning(text string) string {
	// Look for reasoning patterns
	patterns := []string{
		`(?i)reasoning[:\s]+([^.]+(?:\.[^.]*){0,2})`,
		`(?i)because[:\s]+([^.]+(?:\.[^.]*){0,2})`,
		`(?i)rationale[:\s]+([^.]+(?:\.[^.]*){0,2})`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			reasoning := strings.TrimSpace(matches[1])
			if len(reasoning) > 0 {
				if len(reasoning) > 500 {
					reasoning = reasoning[:497] + "..."
				}
				return reasoning
			}
		}
	}

	// Extract first substantial sentence as reasoning
	sentences := strings.Split(text, ". ")
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 20 { // Substantial sentence
			if len(sentence) > 500 {
				sentence = sentence[:497] + "..."
			}
			return sentence
		}
	}

	return "AI analysis based on financial and personal factors"
}

func (di *DecisionInterpreter) extractRecommendations(text string) []string {
	var recommendations []string

	// Look for bulleted or numbered lists
	bulletPatterns := []string{
		`(?m)^[\s]*[-•*]\s*([^\n]+)`,
		`(?m)^[\s]*\d+[.)]\s*([^\n]+)`,
		`(?i)recommend[^:]*:\s*([^\n]+)`,
		`(?i)suggest[^:]*:\s*([^\n]+)`,
	}

	for _, pattern := range bulletPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				rec := strings.TrimSpace(match[1])
				if len(rec) > 10 { // Substantial recommendation
					if len(rec) > 200 {
						rec = rec[:197] + "..."
					}
					recommendations = append(recommendations, rec)
				}
			}
		}
	}

	// If no structured recommendations found, provide defaults based on decision
	if len(recommendations) == 0 {
		recommendations = []string{
			"Review your financial situation regularly",
			"Consider alternatives and compare options",
		}
	}

	return recommendations
}

func (di *DecisionInterpreter) extractWaitPeriod(text string) int {
	// Look for wait period patterns
	patterns := []string{
		`(?i)wait[^0-9]*(\d+)\s*days?`,
		`(?i)defer[^0-9]*(\d+)\s*days?`,
		`(?i)postpone[^0-9]*(\d+)\s*days?`,
		`(?i)(\d+)\s*days?\s*wait`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			if days, err := strconv.Atoi(matches[1]); err == nil {
				if days >= 0 && days <= 365 {
					return days
				}
			}
		}
	}

	// Look for month patterns
	monthPattern := `(?i)(\d+)\s*months?`
	re := regexp.MustCompile(monthPattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) > 1 {
		if months, err := strconv.Atoi(matches[1]); err == nil {
			days := months * 30
			if days >= 0 && days <= 365 {
				return days
			}
		}
	}

	return 0 // No wait period
}

func (di *DecisionInterpreter) extractMaxBudget(text string, originalCost float64) float64 {
	// Look for budget patterns
	patterns := []string{
		`(?i)budget[^$0-9]*\$?([0-9,]+(?:\.\d{2})?)`,
		`(?i)maximum[^$0-9]*\$?([0-9,]+(?:\.\d{2})?)`,
		`(?i)limit[^$0-9]*\$?([0-9,]+(?:\.\d{2})?)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			// Remove commas and parse
			budgetStr := strings.ReplaceAll(matches[1], ",", "")
			if budget, err := strconv.ParseFloat(budgetStr, 64); err == nil {
				if budget > 0 && budget <= 1000000 {
					return budget
				}
			}
		}
	}

	// Default: 10% above original cost for BUY decisions
	return originalCost * 1.1
}

func (di *DecisionInterpreter) createBasicFactors(decision string, confidence float64) []domain.DecisionFactor {
	var factors []domain.DecisionFactor

	switch decision {
	case "BUY":
		factors = append(factors, domain.DecisionFactor{
			Category:    "financial",
			Impact:      "positive",
			Weight:      confidence,
			Description: "Purchase aligns with financial capacity",
		})
	case "WAIT":
		factors = append(factors, domain.DecisionFactor{
			Category:    "timing",
			Impact:      "negative",
			Weight:      confidence,
			Description: "Timing is not optimal for this purchase",
		})
	case "BYE":
		factors = append(factors, domain.DecisionFactor{
			Category:    "financial",
			Impact:      "negative",
			Weight:      confidence,
			Description: "Purchase exceeds financial constraints",
		})
	}

	return factors
}

func (di *DecisionInterpreter) ValidateDecision(decision string) error {
	if !di.validDecisions[strings.ToUpper(decision)] {
		return fmt.Errorf("invalid decision: %s (must be BUY, WAIT, or BYE)", decision)
	}
	return nil
}

func (di *DecisionInterpreter) ValidateConfidence(confidence float64) error {
	if confidence < 0.0 || confidence > 1.0 {
		return fmt.Errorf("invalid confidence: %f (must be between 0.0 and 1.0)", confidence)
	}
	return nil
}

// Public wrapper methods for testing
func (di *DecisionInterpreter) ParseAIResponse(response string, intent domain.PurchaseIntent) (*domain.DecisionOutcome, error) {
	aiResponse := domain.AIResponse{RawResponse: response}
	return di.ParseResponse(aiResponse, intent)
}

func (di *DecisionInterpreter) ExtractDecision(text string) (string, error) {
	return di.extractDecision(text)
}

func (di *DecisionInterpreter) ExtractConfidence(text string) float64 {
	return di.extractConfidence(text)
}

func (di *DecisionInterpreter) ExtractFactors(text string) []string {
	// Simple factor extraction for testing
	var factors []string
	
	// Look for common factor keywords
	keywords := []string{"financial", "health", "timing", "practical", "affordability", "emergency", "debt"}
	text = strings.ToLower(text)
	
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			factors = append(factors, keyword)
		}
	}
	
	if len(factors) == 0 {
		factors = append(factors, "general")
	}
	
	return factors
}

func (di *DecisionInterpreter) ExtractRecommendations(text string) []string {
	return di.extractRecommendations(text)
}