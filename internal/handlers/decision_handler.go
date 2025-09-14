package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/types"
)

// DecisionHandler handles HTTP requests for purchase decision operations
type DecisionHandler struct {
	service   DecisionServiceInterface
	validator *validator.Validate
}

// NewDecisionHandler creates a new DecisionHandler
func NewDecisionHandler(service DecisionServiceInterface) *DecisionHandler {
	return &DecisionHandler{
		service:   service,
		validator: validator.New(),
	}
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error     string      `json:"error"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// MakeDecision handles POST /decision/evaluate - evaluates a purchase intent and returns a decision
func (h *DecisionHandler) MakeDecision(c *gin.Context) {
	// Extract userID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		h.respondWithError(c, http.StatusUnauthorized, "missing_user_context", "User ID not found in request context", nil)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, "invalid_user_context", "Invalid user ID in request context", nil)
		return
	}

	// Parse request body
	var intentDTO types.PurchaseIntentDTO
	if err := c.ShouldBindJSON(&intentDTO); err != nil {
		h.respondWithError(c, http.StatusBadRequest, "invalid_request", "Failed to parse request body", err.Error())
		return
	}

	// Convert DTO to domain object and validate
	domainIntent, err := intentDTO.ToDomain()
	if err != nil {
		// Extract validation details
		var validationDetails interface{}
		if validationErr, ok := err.(*validator.ValidationErrors); ok {
			validationDetails = h.formatValidationErrors(*validationErr)
		} else {
			validationDetails = err.Error()
		}
		h.respondWithError(c, http.StatusBadRequest, "validation_error", "Request validation failed", validationDetails)
		return
	}

	// Set user ID in domain intent
	domainIntent.UserID = userIDStr

	// Call service to make decision
	decision, err := h.service.MakeDecision(c.Request.Context(), *domainIntent)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "service_failure", "Unable to process decision request", nil)
		return
	}

	// Convert domain decision to response DTO
	var responseDTO types.DecisionResponseDTO
	responseDTO.FromDomain(*decision)

	c.JSON(http.StatusOK, responseDTO)
}

// GetDecisionHistory handles GET /decision/history - retrieves user's decision history with pagination
func (h *DecisionHandler) GetDecisionHistory(c *gin.Context) {
	// Extract userID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		h.respondWithError(c, http.StatusUnauthorized, "missing_user_context", "User ID not found in request context", nil)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, "invalid_user_context", "Invalid user ID in request context", nil)
		return
	}

	// Parse pagination parameters
	limit := h.getIntQueryParam(c, "limit", 20)   // Default limit 20
	offset := h.getIntQueryParam(c, "offset", 0)  // Default offset 0
	days := h.getIntQueryParam(c, "days", 30)     // Default 30 days

	// Validate parameters
	if limit <= 0 || limit > 100 {
		h.respondWithError(c, http.StatusBadRequest, "invalid_parameter", "Limit must be between 1 and 100", nil)
		return
	}

	if offset < 0 {
		h.respondWithError(c, http.StatusBadRequest, "invalid_parameter", "Offset must be non-negative", nil)
		return
	}

	if days <= 0 || days > 365 {
		h.respondWithError(c, http.StatusBadRequest, "invalid_parameter", "Days must be between 1 and 365", nil)
		return
	}

	// Get decision history from service
	allDecisions, err := h.service.GetDecisionHistory(c.Request.Context(), userIDStr, days)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "service_failure", "Unable to retrieve decision history", nil)
		return
	}

	// Store total count before pagination
	totalDecisions := len(allDecisions)

	// Apply pagination to results
	var paginatedDecisions []domain.PastDecision
	start := offset
	end := offset + limit
	if start >= len(allDecisions) {
		paginatedDecisions = []domain.PastDecision{} // Empty slice if offset beyond results
	} else {
		if end > len(allDecisions) {
			end = len(allDecisions)
		}
		paginatedDecisions = allDecisions[start:end]
	}

	// Convert to response DTO
	var historyDTO types.DecisionHistoryDTO
	historyDTO.FromDomainHistory(userIDStr, paginatedDecisions, fmt.Sprintf("last_%d_days", days))
	
	// Override total with original count (before pagination)
	historyDTO.TotalDecisions = totalDecisions

	c.JSON(http.StatusOK, historyDTO)
}

// GetDecisionStats handles GET /decision/stats - retrieves decision statistics for a user
func (h *DecisionHandler) GetDecisionStats(c *gin.Context) {
	// Extract userID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		h.respondWithError(c, http.StatusUnauthorized, "missing_user_context", "User ID not found in request context", nil)
		return
	}

	userIDStr, ok := userID.(string)
	if !ok {
		h.respondWithError(c, http.StatusUnauthorized, "invalid_user_context", "Invalid user ID in request context", nil)
		return
	}

	// Parse days parameter
	days := h.getIntQueryParam(c, "days", 30) // Default 30 days

	// Validate days parameter
	if days <= 0 || days > 365 {
		h.respondWithError(c, http.StatusBadRequest, "invalid_parameter", "Days must be between 1 and 365", nil)
		return
	}

	// Get decision history from service
	decisions, err := h.service.GetDecisionHistory(c.Request.Context(), userIDStr, days)
	if err != nil {
		h.respondWithError(c, http.StatusInternalServerError, "service_failure", "Unable to retrieve decision statistics", nil)
		return
	}

	// Calculate statistics
	stats := h.calculateDecisionStats(userIDStr, decisions, days)

	c.JSON(http.StatusOK, stats)
}

// calculateDecisionStats processes decision history to generate statistics
func (h *DecisionHandler) calculateDecisionStats(userID string, decisions []domain.PastDecision, days int) map[string]interface{} {
	totalDecisions := len(decisions)
	totalSpending := 0.0
	decisionPattern := map[string]int{"BUY": 0, "WAIT": 0, "BYE": 0}
	categoryStats := make(map[string]map[string]interface{})

	// Process each decision
	for _, decision := range decisions {
		// Count decision types
		decisionPattern[decision.Decision]++

		// Calculate spending (only for BUY decisions)
		if decision.Decision == "BUY" {
			totalSpending += decision.ItemCost
		}

		// Track category statistics
		if _, exists := categoryStats[decision.Category]; !exists {
			categoryStats[decision.Category] = map[string]interface{}{
				"total_items": 0,
				"buy_count":   0,
				"wait_count":  0,
				"bye_count":   0,
				"total_spent": 0.0,
			}
		}

		stats := categoryStats[decision.Category]
		stats["total_items"] = stats["total_items"].(int) + 1

		switch decision.Decision {
		case "BUY":
			stats["buy_count"] = stats["buy_count"].(int) + 1
			stats["total_spent"] = stats["total_spent"].(float64) + decision.ItemCost
		case "WAIT":
			stats["wait_count"] = stats["wait_count"].(int) + 1
		case "BYE":
			stats["bye_count"] = stats["bye_count"].(int) + 1
		}
	}

	// Calculate buy rates for categories
	for category, stats := range categoryStats {
		totalItems := stats["total_items"].(int)
		buyCount := stats["buy_count"].(int)
		if totalItems > 0 {
			stats["buy_rate"] = float64(buyCount) / float64(totalItems)
		} else {
			stats["buy_rate"] = 0.0
		}

		// Calculate average cost per purchase
		if buyCount > 0 {
			stats["average_cost"] = stats["total_spent"].(float64) / float64(buyCount)
		} else {
			stats["average_cost"] = 0.0
		}

		categoryStats[category] = stats
	}

	// Build response
	response := map[string]interface{}{
		"user_id":          userID,
		"period_days":      days,
		"total_decisions":  totalDecisions,
		"decision_pattern": decisionPattern,
		"total_spending":   totalSpending,
		"category_stats":   categoryStats,
		"generated_at":     time.Now().Format(time.RFC3339),
	}

	// Add percentage breakdowns
	if totalDecisions > 0 {
		response["buy_percentage"] = float64(decisionPattern["BUY"]) / float64(totalDecisions) * 100
		response["wait_percentage"] = float64(decisionPattern["WAIT"]) / float64(totalDecisions) * 100
		response["bye_percentage"] = float64(decisionPattern["BYE"]) / float64(totalDecisions) * 100

		// Average spending per decision (only counting BUY decisions)
		if decisionPattern["BUY"] > 0 {
			response["average_purchase_amount"] = totalSpending / float64(decisionPattern["BUY"])
		} else {
			response["average_purchase_amount"] = 0.0
		}
	} else {
		response["buy_percentage"] = 0.0
		response["wait_percentage"] = 0.0
		response["bye_percentage"] = 0.0
		response["average_purchase_amount"] = 0.0
	}

	return response
}

// getIntQueryParam safely extracts an integer query parameter with a default value
func (h *DecisionHandler) getIntQueryParam(c *gin.Context, key string, defaultValue int) int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// respondWithError sends a standardized error response
func (h *DecisionHandler) respondWithError(c *gin.Context, statusCode int, errorCode, message string, details interface{}) {
	response := ErrorResponse{
		Error:     errorCode,
		Message:   message,
		Details:   details,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	c.JSON(statusCode, response)
}

// formatValidationErrors converts validator.ValidationErrors to a more readable format
func (h *DecisionHandler) formatValidationErrors(errors validator.ValidationErrors) map[string]string {
	formatted := make(map[string]string)

	for _, err := range errors {
		field := strings.ToLower(err.Field())
		tag := err.Tag()
		param := err.Param()

		switch tag {
		case "required":
			formatted[field] = fmt.Sprintf("%s is required", field)
		case "min":
			formatted[field] = fmt.Sprintf("%s must be at least %s characters", field, param)
		case "max":
			formatted[field] = fmt.Sprintf("%s must be at most %s characters", field, param)
		case "gt":
			formatted[field] = fmt.Sprintf("%s must be greater than %s", field, param)
		case "oneof":
			formatted[field] = fmt.Sprintf("%s must be one of: %s", field, param)
		default:
			formatted[field] = fmt.Sprintf("%s is invalid", field)
		}
	}

	return formatted
}