package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// OpenAIClient implements AI client for GPT-4o-mini integration
type OpenAIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	model      string
	maxTokens  int
	temperature float64
	maxRetries int
}

// OpenAIRequest represents the request structure for OpenAI API
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
	TopP        float64   `json:"top_p,omitempty"`
	Stream      bool      `json:"stream"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response from OpenAI API
type OpenAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
	Error   *APIError `json:"error,omitempty"`
}

// Choice represents a response choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// APIError represents an OpenAI API error
type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// NewOpenAIClient creates a new OpenAI client with GPT-4o-mini configuration
func NewOpenAIClient() *OpenAIClient {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		// Return client with empty key - will fail gracefully with clear error message
	}

	return &OpenAIClient{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		model:       "gpt-4o-mini",
		maxTokens:   500,
		temperature: 0.7,
		maxRetries:  3,
	}
}

// GenerateDecision implements the AI client interface for decision generation
func (c *OpenAIClient) GenerateDecision(ctx context.Context, prompt domain.AIPrompt) (*domain.AIResponse, error) {
	startTime := time.Now()

	if c.apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	// Build the request
	request := c.buildRequest(prompt)

	// Make the API call with retry logic
	response, err := c.callAPIWithRetry(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	// Parse and return the response
	aiResponse := &domain.AIResponse{
		RawResponse: response.Choices[0].Message.Content,
		TokensUsed:  response.Usage.TotalTokens,
	}

	// Extract decision, confidence, and other fields from the response
	c.parseResponseFields(aiResponse)

	// Calculate response time
	responseTime := time.Since(startTime)
	fmt.Printf("[OpenAI] Request completed in %v, used %d tokens\n", responseTime, aiResponse.TokensUsed)

	return aiResponse, nil
}

// buildRequest creates an OpenAI API request from the domain prompt
func (c *OpenAIClient) buildRequest(prompt domain.AIPrompt) *OpenAIRequest {
	// Build the complete prompt content
	systemPrompt := prompt.SystemContext
	userPrompt := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s",
		prompt.UserContext,
		prompt.PurchaseDetails,
		prompt.DecisionCriteria,
		prompt.ResponseFormat)

	messages := []Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
		{
			Role:    "user",  
			Content: userPrompt,
		},
	}

	return &OpenAIRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   c.maxTokens,
		Temperature: c.temperature,
		TopP:        1.0,
		Stream:      false,
	}
}

// callAPIWithRetry makes the API call with exponential backoff retry logic
func (c *OpenAIClient) callAPIWithRetry(ctx context.Context, request *OpenAIRequest) (*OpenAIResponse, error) {
	var lastErr error

	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s
			backoffDuration := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			fmt.Printf("[OpenAI] Retry attempt %d after %v\n", attempt+1, backoffDuration)
			
			select {
			case <-time.After(backoffDuration):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		response, err := c.callAPI(ctx, request)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Check if error is retryable
		if !c.isRetryableError(err) {
			break
		}

		fmt.Printf("[OpenAI] Retryable error on attempt %d: %v\n", attempt+1, err)
	}

	return nil, fmt.Errorf("OpenAI API failed after %d attempts: %w", c.maxRetries, lastErr)
}

// callAPI makes a single API call to OpenAI
func (c *OpenAIClient) callAPI(ctx context.Context, request *OpenAIRequest) (*OpenAIResponse, error) {
	// Marshal request to JSON
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse response
	var openaiResponse OpenAIResponse
	if err := json.Unmarshal(responseBody, &openaiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for API errors
	if openaiResponse.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s (type: %s)", openaiResponse.Error.Message, openaiResponse.Error.Type)
	}

	// Validate response structure
	if len(openaiResponse.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenAI API")
	}

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(responseBody))
	}

	return &openaiResponse, nil
}

// isRetryableError determines if an error should trigger a retry
func (c *OpenAIClient) isRetryableError(err error) bool {
	errStr := strings.ToLower(err.Error())
	
	// Retry on rate limits, server errors, and timeouts
	retryableErrors := []string{
		"rate limit",
		"too many requests",
		"server error",
		"internal server error",
		"bad gateway",
		"service unavailable",
		"gateway timeout",
		"timeout",
		"connection reset",
		"connection refused",
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(errStr, retryable) {
			return true
		}
	}

	return false
}

// parseResponseFields extracts structured information from the raw response
func (c *OpenAIClient) parseResponseFields(response *domain.AIResponse) {
	content := response.RawResponse

	// Extract decision using regex patterns
	response.Decision = c.extractDecision(content)
	
	// Extract confidence
	response.Confidence = c.extractConfidence(content)
	
	// Extract reasoning
	response.Reasoning = c.extractReasoning(content)
	
	// Extract recommendations
	response.Suggestions = c.extractRecommendations(content)
}

// extractDecision extracts the decision from the response content
func (c *OpenAIClient) extractDecision(content string) string {
	content = strings.ToUpper(content)
	
	// Look for explicit decision statements
	if strings.Contains(content, "DECISION: BUY") || strings.Contains(content, "DECISION:BUY") {
		return "BUY"
	}
	if strings.Contains(content, "DECISION: WAIT") || strings.Contains(content, "DECISION:WAIT") {
		return "WAIT"
	}
	if strings.Contains(content, "DECISION: BYE") || strings.Contains(content, "DECISION:BYE") {
		return "BYE"
	}
	
	// Look for decision words in content
	if strings.Contains(content, " BUY ") || strings.HasPrefix(content, "BUY ") || strings.HasSuffix(content, " BUY") {
		return "BUY"
	}
	if strings.Contains(content, " WAIT ") || strings.HasPrefix(content, "WAIT ") || strings.HasSuffix(content, " WAIT") {
		return "WAIT"
	}
	if strings.Contains(content, " BYE ") || strings.HasPrefix(content, "BYE ") || strings.HasSuffix(content, " BYE") {
		return "BYE"
	}
	
	// Default to BYE if no clear decision found
	return "BYE"
}

// extractConfidence extracts confidence score from the response
func (c *OpenAIClient) extractConfidence(content string) float64 {
	// Look for confidence patterns
	patterns := []string{
		"CONFIDENCE: ",
		"CONFIDENCE:",
		"confidence: ",
		"confidence:",
	}
	
	for _, pattern := range patterns {
		if idx := strings.Index(content, pattern); idx != -1 {
			// Extract the number after the pattern
			start := idx + len(pattern)
			end := start
			for end < len(content) && (content[end] == '.' || content[end] == '0' || content[end] == '1' || content[end] == '2' || content[end] == '3' || content[end] == '4' || content[end] == '5' || content[end] == '6' || content[end] == '7' || content[end] == '8' || content[end] == '9') {
				end++
			}
			
			if end > start {
				confidenceStr := content[start:end]
				if conf := c.parseFloat(confidenceStr); conf >= 0 && conf <= 1 {
					return conf
				}
			}
		}
	}
	
	// Default confidence based on decision keywords
	upperContent := strings.ToUpper(content)
	if strings.Contains(upperContent, "HIGHLY") || strings.Contains(upperContent, "STRONGLY") {
		return 0.9
	}
	if strings.Contains(upperContent, "RECOMMEND") || strings.Contains(upperContent, "SUGGEST") {
		return 0.7
	}
	if strings.Contains(upperContent, "MAYBE") || strings.Contains(upperContent, "CONSIDER") {
		return 0.5
	}
	
	return 0.6 // Default medium confidence
}

// extractReasoning extracts the reasoning from the response
func (c *OpenAIClient) extractReasoning(content string) string {
	// Look for reason patterns
	patterns := []string{
		"REASON: ",
		"REASON:",
		"reason: ",
		"reason:",
		"because ",
		"Because ",
	}
	
	for _, pattern := range patterns {
		if idx := strings.Index(content, pattern); idx != -1 {
			start := idx + len(pattern)
			// Find the end of the sentence (period, newline, or end of content)
			end := start
			for end < len(content) && content[end] != '.' && content[end] != '\n' && content[end] != '\r' {
				end++
			}
			
			if end > start {
				reason := strings.TrimSpace(content[start:end])
				if len(reason) > 0 {
					return reason
				}
			}
		}
	}
	
	// Extract first substantial sentence as reasoning
	sentences := strings.Split(content, ".")
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 20 { // Substantial sentence
			return sentence
		}
	}
	
	return "Decision based on AI analysis"
}

// extractRecommendations extracts recommendations from the response
func (c *OpenAIClient) extractRecommendations(content string) []string {
	var recommendations []string
	
	// Look for bullet points or numbered lists
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Check for bullet points
		if strings.HasPrefix(line, "•") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
			rec := strings.TrimSpace(line[1:])
			if len(rec) > 5 {
				recommendations = append(recommendations, rec)
			}
		}
		
		// Check for numbered items
		if len(line) > 3 && line[0] >= '1' && line[0] <= '9' && (line[1] == '.' || line[1] == ')') {
			rec := strings.TrimSpace(line[2:])
			if len(rec) > 5 {
				recommendations = append(recommendations, rec)
			}
		}
	}
	
	// Look for recommendations section
	if idx := strings.Index(strings.ToUpper(content), "RECOMMENDATIONS:"); idx != -1 {
		recSection := content[idx+len("RECOMMENDATIONS:"):]
		recLines := strings.Split(recSection, "\n")
		for _, line := range recLines {
			line = strings.TrimSpace(line)
			if len(line) > 10 && !strings.Contains(strings.ToUpper(line), "DECISION") && !strings.Contains(strings.ToUpper(line), "CONFIDENCE") {
				recommendations = append(recommendations, line)
			}
		}
	}
	
	// Default recommendations if none found
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Review your financial situation regularly")
		recommendations = append(recommendations, "Consider alternatives and compare options")
	}
	
	return recommendations
}

// parseFloat safely parses a string to float64
func (c *OpenAIClient) parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	
	// Handle simple cases manually to avoid import
	if s == "0" || s == "0." || s == "0.0" {
		return 0.0
	}
	if s == "1" || s == "1." || s == "1.0" {
		return 1.0
	}
	
	// Parse decimal values
	if strings.Contains(s, ".") {
		parts := strings.Split(s, ".")
		if len(parts) == 2 {
			// Simple decimal parsing for confidence values
			if parts[0] == "0" {
				switch parts[1] {
				case "1": return 0.1
				case "2": return 0.2
				case "3": return 0.3
				case "4": return 0.4
				case "5": return 0.5
				case "6": return 0.6
				case "7": return 0.7
				case "8": return 0.8
				case "9": return 0.9
				}
			}
		}
	}
	
	return 0.6 // Default fallback
}

// GetModel returns the model being used
func (c *OpenAIClient) GetModel() string {
	return c.model
}

// SetModel allows changing the model (for testing or different use cases)
func (c *OpenAIClient) SetModel(model string) {
	c.model = model
}

// GetMaxTokens returns the max tokens configuration
func (c *OpenAIClient) GetMaxTokens() int {
	return c.maxTokens
}

// SetMaxTokens allows changing the max tokens
func (c *OpenAIClient) SetMaxTokens(maxTokens int) {
	c.maxTokens = maxTokens
}