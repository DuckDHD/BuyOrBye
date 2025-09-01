package fx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	client      *http.Client
	rates       map[string]float64
	lastUpdated time.Time
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{Timeout: 5 * time.Second},
		rates:  make(map[string]float64),
	}
}

func (c *Client) ConvertToUSD(amount float64, fromCurrency string) (float64, error) {
	if fromCurrency == "USD" {
		return amount, nil
	}

	// Update rates if needed
	if time.Since(c.lastUpdated) > 24*time.Hour {
		if err := c.updateRates(); err != nil {
			// Use fallback rates if API fails
			c.setFallbackRates()
		}
	}

	rate, exists := c.rates[fromCurrency]
	if !exists {
		return 0, fmt.Errorf("unsupported currency: %s", fromCurrency)
	}

	return amount * rate, nil
}

func (c *Client) updateRates() error {
	// Use a free exchange rate API
	resp, err := c.client.Get("https://api.exchangerate-api.com/v4/latest/USD")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Rates map[string]float64 `json:"rates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	// Convert to rates against USD
	c.rates = make(map[string]float64)
	for currency, rate := range result.Rates {
		c.rates[currency] = 1 / rate // Convert to rate against USD
	}
	c.rates["USD"] = 1.0
	c.lastUpdated = time.Now()

	return nil
}

func (c *Client) setFallbackRates() {
	c.rates = map[string]float64{
		"USD": 1.0,
		"EUR": 0.85,
		"MAD": 0.095,
		"GBP": 0.75,
	}
	c.lastUpdated = time.Now()
}
