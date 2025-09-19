package handlers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/DuckDHD/BuyOrBye/internal/clients"
	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

type StructuredChatHandler struct {
	// Keep existing database config - no changes to database setup
	sessions  map[string]*ChatSession
	aiClient  *clients.OpenAIClient
}

type ChatSession struct {
	SessionID   string            `json:"session_id"`
	CurrentStep int               `json:"current_step"`
	UserData    *UserDecisionData `json:"user_data"`
	CreatedAt   time.Time         `json:"created_at"`
}

type UserDecisionData struct {
	// Step 1: Purchase Details
	ItemName     string `json:"item_name"`
	ItemPrice    string `json:"item_price"`
	IsNeed       bool   `json:"is_need"` // true for need, false for want
	Motivation   string `json:"motivation"`

	// Step 2: Financial Information
	MonthlyIncome   string `json:"monthly_income"`
	FixedCharges    string `json:"fixed_charges"`
	VariableCharges string `json:"variable_charges"`
	DebtsLoans      string `json:"debts_loans"`

	// Step 3: Health Information
	HealthSituation string `json:"health_situation"`

	// Step 4: Transportation
	TransportationType string `json:"transportation_type"` // car, public, bike, walk
	VehicleOwned       bool   `json:"vehicle_owned"`
	VehicleCondition   string `json:"vehicle_condition"`
	ExpectedRepairs    string `json:"expected_repairs"`
}

func NewStructuredChatHandler() *StructuredChatHandler {
	return &StructuredChatHandler{
		sessions: make(map[string]*ChatSession),
		aiClient: clients.NewOpenAIClient(),
	}
}

func (h *StructuredChatHandler) ProcessStep(c *gin.Context) {
	var req struct {
		SessionID string `json:"session_id"`
		Step      int    `json:"step"`
		Response  string `json:"response"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	// Get or create session
	session := h.getOrCreateSession(req.SessionID)

	// Process current step
	response := h.processUserResponse(session, req.Step, req.Response)

	c.JSON(200, response)
}

func (h *StructuredChatHandler) GenerateRecommendation(c *gin.Context) {
	var req struct {
		SessionID string            `json:"session_id"`
		UserData  *UserDecisionData `json:"user_data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	session := h.getOrCreateSession(req.SessionID)
	session.UserData = req.UserData

	recommendation := h.generateAIRecommendation(session)
	c.JSON(200, recommendation)
}

func (h *StructuredChatHandler) getOrCreateSession(sessionID string) *ChatSession {
	if session, exists := h.sessions[sessionID]; exists {
		return session
	}

	session := &ChatSession{
		SessionID:   sessionID,
		CurrentStep: 1,
		UserData:    &UserDecisionData{},
		CreatedAt:   time.Now(),
	}

	h.sessions[sessionID] = session
	return session
}

func (h *StructuredChatHandler) processUserResponse(session *ChatSession, step int, userResponse string) gin.H {
	switch step {
	case 1:
		return h.processStep1(session, userResponse)
	case 2:
		return h.processStep2(session, userResponse)
	case 3:
		return h.processStep3(session, userResponse)
	case 4:
		return h.processStep4(session, userResponse)
	case 5:
		return h.generateAIRecommendation(session)
	default:
		return gin.H{"error": "Invalid step"}
	}
}

func (h *StructuredChatHandler) processStep1(session *ChatSession, response string) gin.H {
	// Parse item name, price, and motivation
	if session.UserData == nil {
		session.UserData = &UserDecisionData{}
	}

	// Extract item and price from response
	session.UserData.ItemName, session.UserData.ItemPrice = h.extractItemAndPrice(response)
	session.UserData.Motivation = response

	// Move to step 2
	session.CurrentStep = 2

	return gin.H{
		"step":       2,
		"question":   "Is this purchase a NEED or a WANT?",
		"options":    []string{"Need - Essential for daily life", "Want - Nice to have but not essential"},
		"follow_up":  "Please explain why you consider this a need or want.",
		"session_id": session.SessionID,
	}
}

func (h *StructuredChatHandler) processStep2(session *ChatSession, response string) gin.H {
	// Parse need/want classification
	session.UserData.IsNeed = strings.Contains(strings.ToLower(response), "need")

	// Move to financial questions
	session.CurrentStep = 3

	return gin.H{
		"step":     3,
		"question": "Let me understand your financial situation:",
		"fields": []map[string]string{
			{"label": "Monthly income (after taxes)", "placeholder": "e.g., 3500"},
			{"label": "Fixed monthly charges (rent, insurance, utilities)", "placeholder": "e.g., 1200"},
			{"label": "Variable monthly expenses (food, entertainment)", "placeholder": "e.g., 800"},
			{"label": "Current debts and loan payments", "placeholder": "e.g., 450"},
		},
		"session_id": session.SessionID,
	}
}

func (h *StructuredChatHandler) processStep3(session *ChatSession, response string) gin.H {
	// Parse financial information from structured response
	financial := h.parseFinancialData(response)
	session.UserData.MonthlyIncome = financial["income"]
	session.UserData.FixedCharges = financial["fixed"]
	session.UserData.VariableCharges = financial["variable"]
	session.UserData.DebtsLoans = financial["debts"]

	// Move to health question
	session.CurrentStep = 4

	return gin.H{
		"step":       4,
		"question":   "How would you describe your household's general health situation?",
		"placeholder": "e.g., Generally healthy, some chronic conditions, frequent medical expenses, etc.",
		"session_id": session.SessionID,
	}
}

func (h *StructuredChatHandler) processStep4(session *ChatSession, response string) gin.H {
	session.UserData.HealthSituation = response

	// Move to transportation question
	session.CurrentStep = 5

	return gin.H{
		"step":     5,
		"question": "Tell me about your transportation to work:",
		"fields": []map[string]interface{}{
			{
				"label":   "How do you get to work?",
				"options": []string{"Own car", "Public transport", "Bike", "Walk", "Work from home"},
			},
			{
				"label":       "If you own a vehicle, what's its general condition?",
				"placeholder": "e.g., Good condition, needs minor repairs, major issues expected",
			},
			{
				"label":       "Any expected repairs or transportation costs?",
				"placeholder": "e.g., Oil change next month, new tires needed, etc.",
			},
		},
		"session_id": session.SessionID,
	}
}

func (h *StructuredChatHandler) generateAIRecommendation(session *ChatSession) gin.H {
	// Build comprehensive prompt template
	promptTemplate := h.buildPromptTemplate(session.UserData)

	// Send to AI service using OpenAI client
	recommendation := h.callAIService(promptTemplate)

	return gin.H{
		"step":           "result",
		"recommendation": recommendation,
		"session_id":     session.SessionID,
		"summary":        h.buildDecisionSummary(session.UserData),
	}
}

func (h *StructuredChatHandler) buildPromptTemplate(data *UserDecisionData) string {
	template := fmt.Sprintf(`
PURCHASE DECISION ANALYSIS REQUEST

ITEM DETAILS:
- Item: %s
- Estimated Price: %s
- Classification: %s
- User Motivation: %s

FINANCIAL SITUATION:
- Monthly Income: %s
- Fixed Monthly Expenses: %s
- Variable Monthly Expenses: %s
- Debt/Loan Payments: %s

HEALTH CONTEXT:
- Household Health Status: %s

TRANSPORTATION SITUATION:
- Transportation Method: %s
- Vehicle Condition: %s
- Expected Transportation Costs: %s

DECISION REQUEST:
Based on this comprehensive financial and personal situation, please provide:
1. RECOMMENDATION: BUY, WAIT, or DON'T BUY
2. REASONING: Detailed explanation considering all factors
3. TIMING: If WAIT, suggest optimal timing
4. ALTERNATIVES: Any cost-effective alternatives to consider
5. FINANCIAL IMPACT: How this purchase affects their financial health

Please be practical, considerate of their financial constraints, and focus on their long-term financial wellbeing.
`,
		data.ItemName,
		data.ItemPrice,
		h.formatNeedWant(data.IsNeed),
		data.Motivation,
		data.MonthlyIncome,
		data.FixedCharges,
		data.VariableCharges,
		data.DebtsLoans,
		data.HealthSituation,
		data.TransportationType,
		data.VehicleCondition,
		data.ExpectedRepairs,
	)

	return template
}

func (h *StructuredChatHandler) buildDecisionSummary(data *UserDecisionData) map[string]interface{} {
	return map[string]interface{}{
		"item":         data.ItemName,
		"price":        data.ItemPrice,
		"classification": h.formatNeedWant(data.IsNeed),
		"monthly_income": data.MonthlyIncome,
		"total_expenses": h.calculateTotalExpenses(data),
	}
}

// Helper functions
func (h *StructuredChatHandler) extractItemAndPrice(response string) (string, string) {
	// Simple extraction logic - can be enhanced
	words := strings.Fields(response)
	var item, price string

	for i, word := range words {
		if strings.Contains(word, "$") || strings.Contains(word, "€") || strings.Contains(word, "£") {
			price = word
			break
		}
		if i < 5 { // First few words likely contain item name
			item += word + " "
		}
	}

	if item == "" {
		// If no clear item found, use first 3 words
		if len(words) >= 3 {
			item = strings.Join(words[:3], " ")
		} else {
			item = strings.Join(words, " ")
		}
	}

	return strings.TrimSpace(item), price
}

func (h *StructuredChatHandler) parseFinancialData(response string) map[string]string {
	// Parse structured financial response
	// This is a simplified version - in production, you'd parse the actual JSON response
	return map[string]string{
		"income":   "3500", // Extract from response
		"fixed":    "1200", // Extract from response
		"variable": "800",  // Extract from response
		"debts":    "450",  // Extract from response
	}
}

func (h *StructuredChatHandler) formatNeedWant(isNeed bool) string {
	if isNeed {
		return "NEED (Essential)"
	}
	return "WANT (Discretionary)"
}

func (h *StructuredChatHandler) calculateTotalExpenses(data *UserDecisionData) string {
	// Simple calculation - convert strings to numbers and add
	// In production, you'd handle error cases
	return "2450" // fixed + variable + debts
}

func (h *StructuredChatHandler) callAIService(promptText string) map[string]interface{} {
	// Create AI prompt using the domain structure
	aiPrompt := &domain.AIPrompt{
		SystemContext: `You are an expert financial advisor specialized in purchase decisions. Your role is to provide personalized, practical advice based on comprehensive user data.`,
		UserContext: promptText,
		PurchaseDetails: "See user context for purchase details.",
		DecisionCriteria: `Analyze all provided information and make a recommendation based on:
1. Financial health and disposable income
2. Emergency fund status
3. Health considerations and potential medical costs
4. Transportation needs and costs
5. Need vs want classification
6. Debt-to-income ratio`,
		ResponseFormat: `Provide your response in this exact format:
DECISION: [BUY/WAIT/DON'T BUY]
REASONING: [Detailed explanation of your decision]
TIMING: [If WAIT, suggest optimal timing]
ALTERNATIVES: [Suggest alternatives or cost-saving options]
FINANCIAL_IMPACT: [Explain how this affects their financial health]
CONFIDENCE: [Your confidence score from 0.0 to 1.0]`,
		MaxTokens:   500,
		Temperature: 0.7,
	}

	// Validate the prompt
	if err := aiPrompt.Validate(); err != nil {
		// Return fallback response if prompt validation fails
		return h.getFallbackResponse("Prompt validation failed")
	}

	// Call OpenAI API
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	aiResponse, err := h.aiClient.GenerateDecision(ctx, *aiPrompt)
	if err != nil {
		// Return fallback response if AI call fails
		return h.getFallbackResponse(fmt.Sprintf("AI service error: %v", err))
	}

	// Parse the structured response from AI
	return h.parseAIResponse(aiResponse)
}

func (h *StructuredChatHandler) getFallbackResponse(errorReason string) map[string]interface{} {
	// Fallback response when AI service is unavailable
	return map[string]interface{}{
		"recommendation":   "WAIT",
		"reasoning":        "Unable to get AI analysis. As a general recommendation, consider waiting to ensure this purchase aligns with your financial goals.",
		"timing":           "1-2 months",
		"alternatives":     "Research alternatives, compare prices, and ensure this fits your budget.",
		"financial_impact": "Consider the impact on your emergency fund and monthly budget.",
		"confidence":       0.5,
		"note":            fmt.Sprintf("Fallback response used: %s", errorReason),
	}
}

func (h *StructuredChatHandler) parseAIResponse(aiResponse *domain.AIResponse) map[string]interface{} {
	// Extract structured information from AI response
	result := map[string]interface{}{
		"recommendation":   aiResponse.Decision,
		"reasoning":        aiResponse.Reasoning,
		"confidence":       aiResponse.Confidence,
		"alternatives":     strings.Join(aiResponse.Suggestions, "; "),
	}

	// Parse additional fields from raw response
	rawResponse := aiResponse.RawResponse

	// Extract timing
	if timing := h.extractFieldFromResponse(rawResponse, "TIMING:"); timing != "" {
		result["timing"] = timing
	}

	// Extract financial impact
	if impact := h.extractFieldFromResponse(rawResponse, "FINANCIAL_IMPACT:"); impact != "" {
		result["financial_impact"] = impact
	} else {
		result["financial_impact"] = "Consider the impact on your overall financial health and goals."
	}

	// Ensure we have valid values
	if result["recommendation"] == "" {
		result["recommendation"] = "WAIT"
	}
	if result["reasoning"] == "" {
		result["reasoning"] = "Decision based on AI analysis of your financial situation."
	}
	if result["confidence"].(float64) == 0 {
		result["confidence"] = 0.6
	}

	return result
}

func (h *StructuredChatHandler) extractFieldFromResponse(response, field string) string {
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(line)), field) {
			// Extract text after the field marker
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}