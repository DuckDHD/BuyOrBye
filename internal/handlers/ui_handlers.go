package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/DuckDHD/BuyOrBye/cmd/web/templates/pages"
	"github.com/DuckDHD/BuyOrBye/internal/dtos"
	"github.com/DuckDHD/BuyOrBye/internal/logging"
)

// UIHandlers contains handlers for frontend UI routes
type UIHandlers struct {
	authService     AuthService
	financeService  FinanceServiceInterface
	healthService   HealthServiceInterface
	decisionService DecisionServiceInterface
}

// NewUIHandlers creates a new UIHandlers instance
func NewUIHandlers(
	authService AuthService,
	financeService FinanceServiceInterface,
	healthService HealthServiceInterface,
	decisionService DecisionServiceInterface,
) *UIHandlers {
	return &UIHandlers{
		authService:     authService,
		financeService:  financeService,
		healthService:   healthService,
		decisionService: decisionService,
	}
}

// PageHandlers - Full page rendering

// DashboardPage renders the main dashboard page
func (h *UIHandlers) DashboardPage(c *gin.Context) {
	h.setCacheHeaders(c, false)

	if !h.isHTMXRequest(c) {
		// Full page request
		data := &dtos.DashboardPageDTO{
			Title: "Dashboard - BuyOrBye",
			User:  h.getCurrentUserDTO(c),
		}

		if err := pages.DashboardPage(data).Render(c.Request.Context(), c.Writer); err != nil {
			h.handleError(c, err, "Failed to render dashboard page")
			return
		}
	} else {
		// HTMX request should get partial
		c.Redirect(http.StatusSeeOther, "/ui/partials/dashboard-content")
	}
}

// FinanceOverviewPage renders the finance overview page
func (h *UIHandlers) FinanceOverviewPage(c *gin.Context) {
	h.setCacheHeaders(c, false)

	if !h.isHTMXRequest(c) {
		userID := h.getUserID(c)

		// Get finance summary
		summary, err := h.financeService.CalculateFinanceSummary(c.Request.Context(), userID)
		if err != nil {
			h.handleError(c, err, "Failed to get finance summary")
			return
		}

		data := &dtos.FinanceOverviewPageDTO{
			Title:   "Finance Overview - BuyOrBye",
			User:    h.getCurrentUserDTO(c),
			Summary: h.convertFinanceSummaryToDTO(summary),
		}

		if err := pages.FinanceOverviewPage(data).Render(c.Request.Context(), c.Writer); err != nil {
			h.handleError(c, err, "Failed to render finance overview page")
			return
		}
	} else {
		// HTMX request should get partial
		c.Redirect(http.StatusSeeOther, "/ui/partials/finance-overview")
	}
}

// HealthProfilePage renders the health profile page
func (h *UIHandlers) HealthProfilePage(c *gin.Context) {
	h.setCacheHeaders(c, false)

	if !h.isHTMXRequest(c) {
		userID := h.getUserID(c)

		// Get health profiles for user
		profiles, err := h.healthService.GetProfilesByUserID(c.Request.Context(), userID)
		if err != nil {
			h.handleError(c, err, "Failed to get health profiles")
			return
		}

		data := &dtos.HealthProfilePageDTO{
			Title:    "Health Profile - BuyOrBye",
			User:     h.getCurrentUserDTO(c),
			Profiles: h.convertHealthProfilesToDTO(profiles),
		}

		if err := pages.HealthProfilePage(data).Render(c.Request.Context(), c.Writer); err != nil {
			h.handleError(c, err, "Failed to render health profile page")
			return
		}
	} else {
		// HTMX request should get partial
		c.Redirect(http.StatusSeeOther, "/ui/partials/health-profile")
	}
}

// DecisionNewPage renders the new decision page
func (h *UIHandlers) DecisionNewPage(c *gin.Context) {
	h.setCacheHeaders(c, false)

	if !h.isHTMXRequest(c) {
		data := &dtos.DecisionNewPageDTO{
			Title: "New Decision - BuyOrBye",
			User:  h.getCurrentUserDTO(c),
		}

		if err := pages.DecisionNewPage(data).Render(c.Request.Context(), c.Writer); err != nil {
			h.handleError(c, err, "Failed to render new decision page")
			return
		}
	} else {
		// HTMX request should get partial
		c.Redirect(http.StatusSeeOther, "/ui/partials/decision-form")
	}
}

// DecisionHistoryPage renders the decision history page
func (h *UIHandlers) DecisionHistoryPage(c *gin.Context) {
	h.setCacheHeaders(c, false)

	if !h.isHTMXRequest(c) {
		// userID := h.getUserID(c)

		// Get decision history
		var decisions []interface{} // This will be properly typed when decision service is implemented
		if h.decisionService != nil {
			// TODO: Implement when decision service is available
			// decisions, err = h.decisionService.GetDecisionHistory(userID)
		}

		data := &dtos.DecisionHistoryPageDTO{
			Title:     "Decision History - BuyOrBye",
			User:      h.getCurrentUserDTO(c),
			Decisions: decisions,
		}

		if err := pages.DecisionHistoryPage(data).Render(c.Request.Context(), c.Writer); err != nil {
			h.handleError(c, err, "Failed to render decision history page")
			return
		}
	} else {
		// HTMX request should get partial
		c.Redirect(http.StatusSeeOther, "/ui/partials/decision-history")
	}
}

// PartialHandlers - Fragment rendering for HTMX

// DashboardContentPartial renders the dashboard content fragment
func (h *UIHandlers) DashboardContentPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	userID := h.getUserID(c)

	// Get summary data for dashboard
	var financeSummary interface{}
	if h.financeService != nil {
		summary, err := h.financeService.CalculateFinanceSummary(c.Request.Context(), userID)
		if err != nil {
			logging.GetLogger().Warn("Failed to get finance summary for dashboard",
				logging.WithError(err),
				zap.String("userID", userID))
		} else {
			financeSummary = h.convertFinanceSummaryToDTO(summary)
		}
	}

	data := &dtos.DashboardContentDTO{
		User:           h.getCurrentUserDTO(c),
		FinanceSummary: financeSummary,
		RecentActivity: []interface{}{}, // TODO: Implement recent activity
	}

	if err := pages.DashboardContentPartial(data).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render dashboard content partial")
		return
	}
}

// FinanceOverviewPartial renders the finance overview fragment
func (h *UIHandlers) FinanceOverviewPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	userID := h.getUserID(c)

	// Get finance summary
	summary, err := h.financeService.CalculateFinanceSummary(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err, "Failed to get finance summary")
		return
	}

	data := &dtos.FinanceOverviewDTO{
		Summary: h.convertFinanceSummaryToDTO(summary),
	}

	if err := pages.FinanceOverviewPartial(data).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render finance overview partial")
		return
	}
}

// IncomeListPartial renders the income list fragment
func (h *UIHandlers) IncomeListPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	userID := h.getUserID(c)

	// Get incomes
	incomes, err := h.financeService.GetUserIncomes(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err, "Failed to get incomes")
		return
	}

	data := &dtos.IncomeListDTO{
		Incomes: h.convertIncomesToDTO(incomes),
	}

	if err := pages.IncomeListPartial(data).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render income list partial")
		return
	}
}

// ExpenseFormPartial renders the expense form fragment
func (h *UIHandlers) ExpenseFormPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	// Check if editing existing expense
	expenseID := c.Query("id")
	var expense interface{}

	if expenseID != "" {
		_, err := strconv.ParseUint(expenseID, 10, 32)
		if err != nil {
			h.handleError(c, err, "Invalid expense ID")
			return
		}

		// TODO: Get specific expense by ID when GetExpenseByID is implemented
		// For now, create a placeholder expense for editing form
		expense = map[string]interface{}{
			"id":          expenseID,
			"description": "Sample expense",
			"amount":      100.0,
			"category":    "Other",
		}
	}

	data := &dtos.ExpenseFormDTO{
		Expense: expense,
		IsEdit:  expense != nil,
	}

	if err := pages.ExpenseFormPartial(data).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render expense form partial")
		return
	}
}

// FinanceSummaryPartial renders the finance summary fragment
func (h *UIHandlers) FinanceSummaryPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	userID := h.getUserID(c)

	// Get finance summary
	summary, err := h.financeService.CalculateFinanceSummary(c.Request.Context(), userID)
	if err != nil {
		h.handleError(c, err, "Failed to get finance summary")
		return
	}

	data := &dtos.FinanceSummaryDTO{
		Summary: h.convertFinanceSummaryToDTO(summary),
	}

	if err := pages.FinanceSummaryPartial(data).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render finance summary partial")
		return
	}
}

// HealthRiskGaugePartial renders the health risk gauge fragment
func (h *UIHandlers) HealthRiskGaugePartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	profileID := c.Param("profileId")
	if profileID == "" {
		h.handleError(c, nil, "Profile ID is required")
		return
	}

	id, err := strconv.ParseUint(profileID, 10, 32)
	if err != nil {
		h.handleError(c, err, "Invalid profile ID")
		return
	}

	// Calculate risk for profile
	riskData, err := h.healthService.CalculateRisk(c.Request.Context(), uint(id))
	if err != nil {
		h.handleError(c, err, "Failed to calculate health risk")
		return
	}

	data := &dtos.HealthRiskGaugeDTO{
		Risk: h.convertRiskDataToDTO(riskData),
	}

	if err := pages.HealthRiskGaugePartial(data).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render health risk gauge partial")
		return
	}
}

// ConditionAddPartial renders the condition add form fragment
func (h *UIHandlers) ConditionAddPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	profileID := c.Query("profileId")

	data := &dtos.ConditionAddDTO{
		ProfileID: profileID,
	}

	if err := pages.ConditionAddPartial(data).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render condition add partial")
		return
	}
}

// InsuranceCardPartial renders the insurance card fragment
func (h *UIHandlers) InsuranceCardPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	policyID := c.Param("policyId")
	if policyID == "" {
		h.handleError(c, nil, "Policy ID is required")
		return
	}

	id, err := strconv.ParseUint(policyID, 10, 32)
	if err != nil {
		h.handleError(c, err, "Invalid policy ID")
		return
	}

	// Get insurance policy
	policy, err := h.healthService.GetPolicyByID(c.Request.Context(), uint(id))
	if err != nil {
		h.handleError(c, err, "Failed to get insurance policy")
		return
	}

	data := &dtos.InsuranceCardDTO{
		Policy: h.convertPolicyToDTO(policy),
	}

	if err := pages.InsuranceCardPartial(data).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render insurance card partial")
		return
	}
}

// DecisionResultPartial renders the decision result fragment
func (h *UIHandlers) DecisionResultPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	// This will be implemented when decision service is available
	data := &dtos.DecisionResultDTO{
		Decision:     "WAIT",
		Confidence:   85,
		Reasoning:    "Decision service not yet implemented",
		Factors:      []interface{}{},
		Alternatives: []interface{}{},
	}

	if err := pages.DecisionResultPartial(data).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render decision result partial")
		return
	}
}

// DecisionFilterPartial renders the decision filter fragment
func (h *UIHandlers) DecisionFilterPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	data := &dtos.DecisionFilterDTO{
		Categories: []string{"Electronics", "Clothing", "Home", "Health", "Other"},
		DateRange:  "last_30_days",
	}

	if err := pages.DecisionFilterPartial(data).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render decision filter partial")
		return
	}
}

// Helper methods

// isHTMXRequest checks if the request is from HTMX
func (h *UIHandlers) isHTMXRequest(c *gin.Context) bool {
	return c.GetHeader("HX-Request") == "true"
}

// setCacheHeaders sets appropriate cache headers
func (h *UIHandlers) setCacheHeaders(c *gin.Context, isStatic bool) {
	if isStatic {
		// Cache static assets for 1 year
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
		c.Header("ETag", `"static-v1"`)
	} else {
		// No cache for dynamic content
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		// ETag for conditional requests
		etag := `"dynamic-` + strconv.FormatInt(time.Now().Unix(), 10) + `"`
		c.Header("ETag", etag)

		// Check If-None-Match header
		if match := c.GetHeader("If-None-Match"); match == etag {
			c.Status(http.StatusNotModified)
			return
		}
	}
}

// handleError handles errors appropriately for HTMX vs full page requests
func (h *UIHandlers) handleError(c *gin.Context, err error, message string) {
	logger := logging.GetLogger()

	if err != nil {
		logger.Error("UI handler error",
			logging.WithError(err),
			zap.String("message", message),
			zap.String("path", c.Request.URL.Path))
	}

	if h.isHTMXRequest(c) || h.acceptsJSON(c) {
		// Return error partial or JSON for HTMX requests
		errorDTO := dtos.NewErrorResponse(500, "internal_error", message)

		if h.acceptsJSON(c) {
			c.JSON(500, errorDTO)
		} else {
			// Render error partial for HTMX
			data := &dtos.ErrorPartialDTO{
				Error:   errorDTO,
				Message: message,
			}

			if renderErr := pages.ErrorPartial(data).Render(c.Request.Context(), c.Writer); renderErr != nil {
				logger.Error("Failed to render error partial", logging.WithError(renderErr))
				c.JSON(500, errorDTO)
			}
		}
	} else {
		// Return error page for full page requests
		data := &dtos.ErrorPageDTO{
			Title:   "Error - BuyOrBye",
			Error:   dtos.NewErrorResponse(500, "internal_error", message),
			Message: message,
		}

		c.Status(500)
		if renderErr := pages.ErrorPage(data).Render(c.Request.Context(), c.Writer); renderErr != nil {
			logger.Error("Failed to render error page", logging.WithError(renderErr))
			c.String(500, "Internal Server Error")
		}
	}
}

// acceptsJSON checks if the client accepts JSON
func (h *UIHandlers) acceptsJSON(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	return accept == "application/json" || accept == "*/*"
}

// getUserID extracts user ID from context (set by auth middleware)
func (h *UIHandlers) getUserID(c *gin.Context) string {
	userID, exists := c.Get("userID")
	if !exists {
		return ""
	}
	return userID.(string)
}

// getCurrentUserDTO gets current user DTO from context
func (h *UIHandlers) getCurrentUserDTO(c *gin.Context) *dtos.UserDTO {
	userID := h.getUserID(c)
	if userID == "" {
		return nil
	}

	// TODO: Get user details from auth service
	return &dtos.UserDTO{
		ID:    userID,
		Email: "user@example.com", // Placeholder
		Name:  "User",             // Placeholder
	}
}

// Conversion helper methods (placeholders - will be implemented based on actual domain models)

func (h *UIHandlers) convertFinanceSummaryToDTO(summary interface{}) interface{} {
	// TODO: Implement actual conversion when finance domain is finalized
	return summary
}

func (h *UIHandlers) convertHealthProfilesToDTO(profiles interface{}) []interface{} {
	// TODO: Implement actual conversion when health domain is finalized
	return []interface{}{}
}

func (h *UIHandlers) convertIncomesToDTO(incomes interface{}) []interface{} {
	// TODO: Implement actual conversion when finance domain is finalized
	return []interface{}{}
}

func (h *UIHandlers) convertExpenseToDTO(expense interface{}) interface{} {
	// TODO: Implement actual conversion when finance domain is finalized
	return expense
}

func (h *UIHandlers) convertRiskDataToDTO(risk interface{}) interface{} {
	// TODO: Implement actual conversion when health domain is finalized
	return risk
}

func (h *UIHandlers) convertPolicyToDTO(policy interface{}) interface{} {
	// TODO: Implement actual conversion when health domain is finalized
	return policy
}
