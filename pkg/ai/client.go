package ai

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/models"
)

type Client struct {
	apiKey string
	client *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) ExplainDecision(ctx context.Context, input models.DecisionInput, verdict models.DecisionVerdict, reasons []string) (string, error) {
	if c.apiKey == "" {
		return "", nil // AI disabled
	}

	prompt := fmt.Sprintf(`
		Provide a brief, friendly explanation for this purchase decision:
		
		Item: %s
		Price: %.2f %s
		Decision: %s
		Reasons: %v
		
		Respond in 1-2 sentences, be encouraging if YES, helpful if NO.
	`, input.ItemName, input.Price, input.Currency, verdict, reasons)

	// This is a placeholder - implement with your preferred AI service
	// For OpenAI, Anthropic, or similar
	return c.callAIService(ctx, prompt)
}

func (c *Client) callAIService(ctx context.Context, prompt string) (string, error) {
	// Placeholder implementation
	// Replace with actual AI service integration
	if len(prompt) > 100 {
		return "This purchase aligns well with your financial situation.", nil
	}
	return "", nil
}
