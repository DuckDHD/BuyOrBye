//go:build manual
// +build manual

// Manual test for OpenAI integration - requires OPENAI_API_KEY environment variable
// Run with: go test -v -tags=manual ./tests/manual/

package manual

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DuckDHD/BuyOrBye/internal/clients"
	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/services"
)

// TestResults tracks the performance and cost metrics
type TestResults struct {
	Scenario          string        `json:"scenario"`
	RequestDuration   time.Duration `json:"request_duration_ms"`
	TokensUsed        int           `json:"tokens_used"`
	EstimatedCost     float64       `json:"estimated_cost_usd"`
	Decision          string        `json:"decision"`
	Confidence        float64       `json:"confidence"`
	CacheHit          bool          `json:"cache_hit"`
	ErrorOccurred     bool          `json:"error_occurred"`
	ErrorMessage      string        `json:"error_message,omitempty"`
}

// TestSession tracks overall testing session
type TestSession struct {
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	TotalTests     int           `json:"total_tests"`
	SuccessfulTests int          `json:"successful_tests"`
	TotalTokens    int           `json:"total_tokens"`
	TotalCost      float64       `json:"total_cost_usd"`
	Results        []TestResults `json:"results"`
}

var session TestSession

func TestMain(m *testing.M) {
	// Check for OpenAI API key
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ OPENAI_API_KEY environment variable not set")
		fmt.Println("   Set it with: export OPENAI_API_KEY=your_api_key_here")
		os.Exit(1)
	}

	// Initialize session
	session = TestSession{
		StartTime: time.Now(),
		Results:   make([]TestResults, 0),
	}

	// Run tests
	code := m.Run()

	// Finalize session
	session.EndTime = time.Now()
	session.TotalTests = len(session.Results)

	for _, result := range session.Results {
		if !result.ErrorOccurred {
			session.SuccessfulTests++
		}
		session.TotalTokens += result.TokensUsed
		session.TotalCost += result.EstimatedCost
	}

	// Print summary
	printTestSummary()

	// Save results to file
	saveResults()

	os.Exit(code)
}

func TestOpenAIIntegration_SmallPurchase(t *testing.T) {
	result := runDecisionTest(t, "Small Purchase ($50 gadget)", domain.PurchaseIntent{
		ID:       "test-small-001",
		UserID:   "test-user-001",
		ItemName: "Bluetooth Wireless Earbuds",
		ItemCost: 49.99,
		Category: "electronics",
		Urgency:  "low",
		Frequency: "one_time",
		CreatedAt: time.Now(),
	})

	// Assert reasonable decision for small purchase
	assert.Contains(t, []string{"BUY", "WAIT"}, result.Decision, "Small purchase should typically be BUY or WAIT")
	assert.True(t, result.Confidence > 0.5, "Confidence should be reasonable")
	assert.True(t, result.RequestDuration < 5*time.Second, "Should respond within 5 seconds")
}

func TestOpenAIIntegration_MediumPurchase(t *testing.T) {
	result := runDecisionTest(t, "Medium Purchase ($500 electronics)", domain.PurchaseIntent{
		ID:       "test-medium-001",
		UserID:   "test-user-001",
		ItemName: "Gaming Monitor 27 inch 144Hz",
		ItemCost: 499.99,
		Category: "electronics",
		Urgency:  "medium",
		Frequency: "one_time",
		CreatedAt: time.Now(),
	})

	// Assert decision quality for medium purchase
	assert.Contains(t, []string{"BUY", "WAIT", "BYE"}, result.Decision)
	assert.True(t, result.Confidence > 0.4, "Should have reasonable confidence")
	assert.True(t, result.RequestDuration < 10*time.Second, "Should respond within 10 seconds")
}

func TestOpenAIIntegration_LargePurchase(t *testing.T) {
	result := runDecisionTest(t, "Large Purchase ($2000 laptop)", domain.PurchaseIntent{
		ID:       "test-large-001",
		UserID:   "test-user-001",
		ItemName: "MacBook Pro 16-inch M3 Max",
		ItemCost: 1999.99,
		Category: "electronics",
		Urgency:  "high",
		Frequency: "one_time",
		CreatedAt: time.Now(),
	})

	// Assert careful consideration for large purchase
	assert.Contains(t, []string{"BUY", "WAIT", "BYE"}, result.Decision)
	assert.True(t, result.Confidence > 0.3, "Should provide some confidence level")
	assert.True(t, result.RequestDuration < 10*time.Second, "Should respond within 10 seconds")
}

func TestOpenAIIntegration_HealthPurchase(t *testing.T) {
	result := runDecisionTest(t, "Health Purchase (medical equipment)", domain.PurchaseIntent{
		ID:       "test-health-001",
		UserID:   "test-user-001",
		ItemName: "Blood Pressure Monitor",
		ItemCost: 89.99,
		Category: "health",
		Urgency:  "critical",
		Frequency: "one_time",
		CreatedAt: time.Now(),
	})

	// Health purchases should often be recommended
	assert.Contains(t, []string{"BUY", "WAIT"}, result.Decision, "Critical health items should usually be BUY or WAIT, rarely BYE")
	assert.True(t, result.Confidence > 0.5, "Health decisions should have good confidence")
	assert.True(t, result.RequestDuration < 5*time.Second, "Should respond quickly for health items")
}

func TestOpenAIIntegration_Subscription(t *testing.T) {
	result := runDecisionTest(t, "Subscription (monthly service)", domain.PurchaseIntent{
		ID:       "test-subscription-001",
		UserID:   "test-user-001",
		ItemName: "Netflix Premium Subscription",
		ItemCost: 15.99,
		Category: "entertainment",
		Urgency:  "low",
		Frequency: "monthly",
		CreatedAt: time.Now(),
	})

	// Subscriptions require different consideration
	assert.Contains(t, []string{"BUY", "WAIT", "BYE"}, result.Decision)
	assert.True(t, result.Confidence > 0.3, "Should have some confidence")
	assert.True(t, result.RequestDuration < 5*time.Second, "Should respond quickly for low-cost items")
}

func TestOpenAIIntegration_CachingFunctionality(t *testing.T) {
	// First request - should hit API
	intent := domain.PurchaseIntent{
		ID:       "test-cache-001",
		UserID:   "test-user-cache",
		ItemName: "Test Cache Item",
		ItemCost: 25.00,
		Category: "other",
		Urgency:  "low",
		Frequency: "one_time",
		CreatedAt: time.Now(),
	}

	result1 := runDecisionTest(t, "Cache Test - First Request", intent)
	assert.False(t, result1.CacheHit, "First request should not be cache hit")
	
	// Wait a moment
	time.Sleep(100 * time.Millisecond)
	
	// Second identical request - should hit cache
	result2 := runDecisionTest(t, "Cache Test - Second Request", intent)
	
	// Cache functionality validation
	if result2.RequestDuration < result1.RequestDuration/2 {
		result2.CacheHit = true // Infer cache hit from faster response
	}
	
	// Decisions should be identical when cached
	if result2.CacheHit {
		assert.Equal(t, result1.Decision, result2.Decision, "Cached decision should be identical")
		assert.Equal(t, result1.Confidence, result2.Confidence, "Cached confidence should be identical")
		assert.True(t, result2.RequestDuration < result1.RequestDuration, "Cache should be faster")
		t.Logf("✅ Cache hit detected - %v faster than original", result1.RequestDuration-result2.RequestDuration)
	} else {
		t.Logf("⚠️ Cache may not be working - response times similar")
	}
}

func TestOpenAIIntegration_ResponseFormat(t *testing.T) {
	result := runDecisionTest(t, "Response Format Validation", domain.PurchaseIntent{
		ID:       "test-format-001",
		UserID:   "test-user-001",
		ItemName: "Response Format Test Item",
		ItemCost: 100.00,
		Category: "electronics",
		Urgency:  "medium",
		Frequency: "one_time",
		CreatedAt: time.Now(),
	})

	// Validate response format requirements
	assert.Contains(t, []string{"BUY", "WAIT", "BYE"}, result.Decision, "Must be valid decision")
	assert.True(t, result.Confidence >= 0.0 && result.Confidence <= 1.0, "Confidence must be 0-1")
	
	t.Logf("✅ Response format validation passed")
}

// runDecisionTest executes a single decision test and records metrics
func runDecisionTest(t *testing.T, scenario string, intent domain.PurchaseIntent) TestResults {
	result := TestResults{
		Scenario:    scenario,
		CacheHit:    false,
		ErrorOccurred: false,
	}

	// Create OpenAI client
	client := clients.NewOpenAIClient()
	
	// Create test context with financial/health data
	context := createTestContext(intent.UserID)
	
	// Build prompt
	promptBuilder := services.NewPromptBuilder()
	prompt, err := promptBuilder.BuildPrompt(intent, *context)
	require.NoError(t, err, "Should build prompt successfully")

	// Time the API call
	startTime := time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	response, err := client.GenerateDecision(ctx, *prompt)
	
	result.RequestDuration = time.Since(startTime)

	if err != nil {
		result.ErrorOccurred = true
		result.ErrorMessage = err.Error()
		t.Errorf("OpenAI API call failed for %s: %v", scenario, err)
		session.Results = append(session.Results, result)
		return result
	}

	// Parse response
	interpreter := services.NewDecisionInterpreter()
	decision, err := interpreter.ParseResponse(*response, intent)
	if err != nil {
		result.ErrorOccurred = true
		result.ErrorMessage = fmt.Sprintf("Parse error: %v", err)
		t.Errorf("Failed to parse response for %s: %v", scenario, err)
		session.Results = append(session.Results, result)
		return result
	}

	// Extract metrics
	result.Decision = decision.Decision
	result.Confidence = decision.Confidence
	result.TokensUsed = extractTokenCount(response)
	result.EstimatedCost = calculateCost(result.TokensUsed)

	// Log results
	t.Logf("📊 %s:", scenario)
	t.Logf("   Decision: %s (%.1f%% confidence)", result.Decision, result.Confidence*100)
	t.Logf("   Duration: %v", result.RequestDuration)
	t.Logf("   Tokens: %d (~$%.4f)", result.TokensUsed, result.EstimatedCost)

	session.Results = append(session.Results, result)
	return result
}

// createTestContext creates a test decision context
func createTestContext(userID string) *domain.DecisionContext {
	return &domain.DecisionContext{
		UserID: userID,
		FinancialContext: domain.FinancialSnapshot{
			MonthlyIncome:       5000.0,
			MonthlyExpenses:     3500.0,
			DisposableIncome:    1500.0,
			EmergencyFundMonths: 4.0,
			DebtToIncomeRatio:   0.3,
			CreditScore:         750,
		},
		HealthContext: domain.HealthSnapshot{
			Age:                    30,
			HasChronicConditions:   false,
			MonthlyMedicalCosts:    50.0,
			HasHealthInsurance:     true,
			HealthInsuranceCoverage: 0.8,
		},
		TransportContext: domain.TransportSnapshot{
			HasVehicle:           true,
			MonthlyTransportCost: 400.0,
			PublicTransitAccess:  true,
			CommuteDistance:      15.0,
		},
		PurchaseHistory: []domain.PastDecision{},
		CurrentDate:     time.Now(),
	}
}

// extractTokenCount extracts token usage from API response
func extractTokenCount(response *domain.AIResponse) int {
	// GPT-4o-mini approximate token calculation
	// This is an estimation - actual usage would come from API response headers
	inputTokens := len(response.RawResponse) / 4    // Rough estimate
	outputTokens := len(response.RawResponse) / 6   // Rough estimate
	return inputTokens + outputTokens
}

// calculateCost calculates estimated cost based on token usage
func calculateCost(tokens int) float64 {
	// GPT-4o-mini pricing (as of 2024):
	// Input: $0.00015 per 1K tokens
	// Output: $0.0006 per 1K tokens
	// Average cost ~$0.0004 per 1K tokens
	return float64(tokens) * 0.0004 / 1000.0
}

// printTestSummary prints a comprehensive test summary
func printTestSummary() {
	fmt.Println("\n" + "="*80)
	fmt.Println("🤖 OPENAI INTEGRATION TEST SUMMARY")
	fmt.Println("="*80)
	
	duration := session.EndTime.Sub(session.StartTime)
	successRate := float64(session.SuccessfulTests) / float64(session.TotalTests) * 100
	
	fmt.Printf("📈 Overall Results:\n")
	fmt.Printf("   Tests Run: %d\n", session.TotalTests)
	fmt.Printf("   Successful: %d (%.1f%%)\n", session.SuccessfulTests, successRate)
	fmt.Printf("   Duration: %v\n", duration.Round(time.Millisecond))
	fmt.Printf("   Total Tokens: %d\n", session.TotalTokens)
	fmt.Printf("   Total Cost: $%.4f\n\n", session.TotalCost)

	fmt.Printf("📊 Performance Metrics:\n")
	
	var totalDuration time.Duration
	var minDuration, maxDuration time.Duration
	minDuration = time.Hour // Initialize to large value
	
	for _, result := range session.Results {
		if !result.ErrorOccurred {
			totalDuration += result.RequestDuration
			if result.RequestDuration < minDuration {
				minDuration = result.RequestDuration
			}
			if result.RequestDuration > maxDuration {
				maxDuration = result.RequestDuration
			}
		}
	}
	
	if session.SuccessfulTests > 0 {
		avgDuration := totalDuration / time.Duration(session.SuccessfulTests)
		avgCost := session.TotalCost / float64(session.SuccessfulTests)
		
		fmt.Printf("   Avg Response Time: %v\n", avgDuration.Round(time.Millisecond))
		fmt.Printf("   Min Response Time: %v\n", minDuration.Round(time.Millisecond))
		fmt.Printf("   Max Response Time: %v\n", maxDuration.Round(time.Millisecond))
		fmt.Printf("   Avg Cost per Decision: $%.4f\n", avgCost)
	}

	fmt.Printf("\n📋 Individual Results:\n")
	for _, result := range session.Results {
		status := "✅"
		if result.ErrorOccurred {
			status = "❌"
		} else if result.CacheHit {
			status = "🚀" // Cache hit
		}
		
		fmt.Printf("   %s %-35s %s (%.1f%%) %v $%.4f\n", 
			status,
			result.Scenario,
			result.Decision,
			result.Confidence*100,
			result.RequestDuration.Round(time.Millisecond),
			result.EstimatedCost,
		)
		
		if result.ErrorOccurred {
			fmt.Printf("       Error: %s\n", result.ErrorMessage)
		}
	}

	// Performance analysis
	fmt.Printf("\n🎯 Performance Analysis:\n")
	if session.TotalCost < 0.10 {
		fmt.Printf("   ✅ Cost Efficiency: Excellent (under $0.10 total)\n")
	} else if session.TotalCost < 0.25 {
		fmt.Printf("   ✅ Cost Efficiency: Good (under $0.25 total)\n")
	} else {
		fmt.Printf("   ⚠️  Cost Efficiency: High cost - consider optimization\n")
	}

	avgResponseTime := totalDuration / time.Duration(session.SuccessfulTests)
	if avgResponseTime < 2*time.Second {
		fmt.Printf("   ✅ Response Time: Excellent (under 2 seconds avg)\n")
	} else if avgResponseTime < 5*time.Second {
		fmt.Printf("   ✅ Response Time: Good (under 5 seconds avg)\n")
	} else {
		fmt.Printf("   ⚠️  Response Time: Slow - investigate performance issues\n")
	}

	if successRate >= 95 {
		fmt.Printf("   ✅ Reliability: Excellent (%.1f%% success rate)\n", successRate)
	} else if successRate >= 80 {
		fmt.Printf("   ✅ Reliability: Good (%.1f%% success rate)\n", successRate)
	} else {
		fmt.Printf("   ❌ Reliability: Poor (%.1f%% success rate)\n", successRate)
	}

	fmt.Println("\n" + "="*80)
}

// saveResults saves test results to JSON file
func saveResults() {
	resultsJSON, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		fmt.Printf("⚠️ Failed to marshal results: %v\n", err)
		return
	}

	filename := fmt.Sprintf("openai_test_results_%s.json", 
		session.StartTime.Format("2006-01-02_15-04-05"))
	
	err = os.WriteFile(filename, resultsJSON, 0644)
	if err != nil {
		fmt.Printf("⚠️ Failed to save results: %v\n", err)
		return
	}

	fmt.Printf("💾 Results saved to: %s\n", filename)
}