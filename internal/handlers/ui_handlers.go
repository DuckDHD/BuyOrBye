package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/DuckDHD/BuyOrBye/cmd/web/templates/pages"
	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/logging"
	"github.com/DuckDHD/BuyOrBye/internal/middleware"
	"github.com/DuckDHD/BuyOrBye/internal/services"
	"github.com/DuckDHD/BuyOrBye/internal/types"
)

// UIHandlers contains handlers for frontend UI routes
type UIHandlers struct {
	authService     services.AuthService
	financeService  services.FinanceService
	healthService   services.HealthService
	decisionService DecisionServiceInterface // Keep this for now since decision service is not implemented
}

// NewUIHandlers creates a new UIHandlers instance
func NewUIHandlers(
	authService services.AuthService,
	financeService services.FinanceService,
	healthService services.HealthService,
	decisionService DecisionServiceInterface,
) *UIHandlers {
	return &UIHandlers{
		authService:     authService,
		financeService:  financeService,
		healthService:   healthService,
		decisionService: decisionService,
	}
}

// ================================
// PAGE HANDLERS - Full page rendering with layouts
// ================================

// LoginPage renders the login page (GET /auth/login)
func (h *UIHandlers) LoginPage(c *gin.Context) {
	h.setCacheHeaders(c, false)

	csrfToken := c.GetString("csrf_token")
	dto := types.LoginPageDTO{
		Layout: types.LayoutDTO{
			Title:     "Login",
			CSRFToken: csrfToken,
		},
	}

	// Handle error parameters from redirects
	if errorParam := c.Query("error"); errorParam != "" {
		var errorMsg string
		switch errorParam {
		case "invalid_data":
			errorMsg = "Invalid form data. Please check your input and try again."
		case "invalid_credentials":
			errorMsg = "Invalid email or password. Please try again."
		case "session_expired":
			errorMsg = "Your session has expired. Please sign in again."
		case "unauthorized":
			errorMsg = "You need to sign in to access this page."
		default:
			errorMsg = "An error occurred during login. Please try again."
		}

		dto.Error = &types.ErrorResponseDTO{
			Message: errorMsg,
		}
	}

	if err := pages.LoginPage(dto).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render login page")
		return
	}
}

// RegisterPage renders the register page (GET /auth/register)
func (h *UIHandlers) RegisterPage(c *gin.Context) {
	h.setCacheHeaders(c, false)

	csrfToken := c.GetString(middleware.CSRFContextKey)
	dto := types.RegisterPageDTO{
		Layout: types.LayoutDTO{
			Title:     "Register",
			CSRFToken: csrfToken,
		},
	}

	// Handle error parameters from redirects
	if errorParam := c.Query("error"); errorParam != "" {
		var errorMsg string
		switch errorParam {
		case "invalid_data":
			errorMsg = "Invalid form data. Please check your input and try again."
		case "registration_failed":
			errorMsg = "Registration failed. Please try again."
		default:
			errorMsg = "An error occurred during registration. Please try again."
		}

		dto.Error = &types.ErrorResponseDTO{
			Message: errorMsg,
		}
	}

	if err := pages.RegisterPage(dto).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render register page")
		return
	}
}

// DashboardPage renders the main dashboard page (GET /dashboard)
func (h *UIHandlers) DashboardPage(c *gin.Context) {
	h.setCacheHeaders(c, false)

	user := h.getCurrentUserDTO(c)
	csrfToken := c.GetString("csrf_token")

	dto := types.DashboardPageDTO{
		Layout: types.LayoutDTO{
			Title:     "Dashboard",
			CSRFToken: csrfToken,
			User:      user,
		},
	}

	if err := pages.DashboardPage(dto).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render dashboard page")
		return
	}
}

// FinanceOverviewPage renders the finance overview page (GET /finance)
func (h *UIHandlers) FinanceOverviewPage(c *gin.Context) {
	h.setCacheHeaders(c, false)

	user := h.getCurrentUserDTO(c)
	csrfToken := c.GetString("csrf_token")

	dto := types.FinanceOverviewPageDTO{
		Layout: types.LayoutDTO{
			Title:     "Finance Overview",
			CSRFToken: csrfToken,
			User:      user,
		},
	}

	if err := pages.FinanceOverviewPage(dto).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render finance overview page")
		return
	}
}

// HealthProfilePage renders the health profile page (GET /health)
func (h *UIHandlers) HealthProfilePage(c *gin.Context) {
	h.setCacheHeaders(c, false)

	user := h.getCurrentUserDTO(c)
	csrfToken := c.GetString("csrf_token")

	dto := types.HealthProfilePageDTO{
		Layout: types.LayoutDTO{
			Title:     "Health Profile",
			CSRFToken: csrfToken,
			User:      user,
		},
	}

	if err := pages.HealthProfilePage(dto).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render health profile page")
		return
	}
}

// DecisionNewPage renders the new decision page (GET /decisions/new)
func (h *UIHandlers) DecisionNewPage(c *gin.Context) {
	h.setCacheHeaders(c, false)

	user := h.getCurrentUserDTO(c)
	csrfToken := c.GetString("csrf_token")

	dto := types.DecisionNewPageDTO{
		Layout: types.LayoutDTO{
			Title:     "New Decision",
			CSRFToken: csrfToken,
			User:      user,
		},
	}

	if err := pages.DecisionNewPage(dto).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render new decision page")
		return
	}
}

// DecisionHistoryPage renders the decision history page (GET /decisions/history)
func (h *UIHandlers) DecisionHistoryPage(c *gin.Context) {
	h.setCacheHeaders(c, false)

	user := h.getCurrentUserDTO(c)
	csrfToken := c.GetString("csrf_token")

	dto := types.DecisionHistoryPageDTO{
		Layout: types.LayoutDTO{
			Title:     "Decision History",
			CSRFToken: csrfToken,
			User:      user,
		},
	}

	if err := pages.DecisionHistoryPage(dto).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render decision history page")
		return
	}
}

// ChatPage renders the chat landing page (GET /)
func (h *UIHandlers) ChatPage(c *gin.Context) {
	h.setCacheHeaders(c, false)

	user := h.getCurrentUserDTO(c)
	csrfToken := c.GetString("csrf_token")

	dto := types.NewChatPageDTO(csrfToken, user)

	if err := pages.ChatPage(dto).Render(c.Request.Context(), c.Writer); err != nil {
		h.handleError(c, err, "Failed to render chat page")
		return
	}
}

// ================================
// PARTIAL HANDLERS - Fragment rendering for HTMX (no layouts)
// ================================

// DashboardContentPartial renders the dashboard content fragment (GET /ui/partials/dashboard/content)
func (h *UIHandlers) DashboardContentPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	// For now, render a simple message until DashboardContentPartial is implemented
	c.String(200, "Dashboard content loading...")
}

// FinanceOverviewPartial renders the finance overview fragment (GET /ui/partials/finance/overview)
func (h *UIHandlers) FinanceOverviewPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	// For now, render a simple message until FinanceOverviewPartial is implemented
	c.String(200, "Finance overview loading...")
}

// FinanceSummaryPartial renders the finance summary fragment (GET /ui/partials/finance/summary)
func (h *UIHandlers) FinanceSummaryPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	// For now, render a simple message until FinanceSummaryPartial is implemented
	c.String(200, "Finance summary loading...")
}

// IncomeListPartial renders the income list fragment (GET /ui/partials/finance/income/list)
func (h *UIHandlers) IncomeListPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	// For now, render a simple message until IncomeListPartial is implemented
	c.String(200, "Income list loading...")
}

// ExpenseFormPartial renders the expense form fragment (GET /ui/partials/finance/expense/form)
func (h *UIHandlers) ExpenseFormPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	// For now, render a simple message until ExpenseFormPartial is implemented
	c.String(200, "Expense form loading...")
}

// HealthRiskGaugePartial renders the health risk gauge fragment (GET /ui/partials/health/risk-gauge/:profileId)
func (h *UIHandlers) HealthRiskGaugePartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	// For now, render a simple message until HealthRiskGaugePartial is implemented
	c.String(200, "Health risk gauge loading...")
}

// ConditionAddPartial renders the condition add form fragment (GET /ui/partials/health/condition/add)
func (h *UIHandlers) ConditionAddPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	// For now, render a simple message until ConditionAddPartial is implemented
	c.String(200, "Condition add form loading...")
}

// InsuranceCardPartial renders the insurance card fragment (GET /ui/partials/health/insurance/card/:policyId)
func (h *UIHandlers) InsuranceCardPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	// For now, render a simple message until InsuranceCardPartial is implemented
	c.String(200, "Insurance card loading...")
}

// DecisionResultPartial renders the decision result fragment (GET /ui/partials/decision/result)
func (h *UIHandlers) DecisionResultPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	// This will be implemented when decision service is available
	// For now, render a simple message until DecisionResultPartial is implemented
	c.String(200, "Decision result loading...")
}

// DecisionFilterPartial renders the decision filter fragment (GET /ui/partials/decision/filter)
func (h *UIHandlers) DecisionFilterPartial(c *gin.Context) {
	h.setCacheHeaders(c, false)

	// For now, render a simple message until DecisionFilterPartial is implemented
	c.String(200, "Decision filter loading...")
}

// ================================
// ACTION HANDLERS - Form submissions that redirect after processing
// ================================

// LoginAction handles login form submission (POST /auth/login)
func (h *UIHandlers) LoginAction(c *gin.Context) {
	logger := logging.ContextLogger(c).With(logging.WithOperation("login_action"))

	// Parse form data for debugging
	c.Request.ParseForm()
	logger.Info("Login form data received",
		zap.Any("form_values", c.Request.Form),
		zap.String("content_type", c.ContentType()))

	var request types.LoginRequestDTO

	if err := c.ShouldBind(&request); err != nil {
		logger.Warn("Login form binding failed",
			logging.WithError(err),
			zap.Any("raw_form", c.Request.Form))
		h.handleLoginError(c, "Invalid form data. Please check your input and try again.")
		return
	}

	// Log bound request for debugging
	logger.Info("Login form bound successfully",
		zap.String("email", request.Email),
		zap.Int("password_length", len(request.Password)))

	// Basic validation
	if strings.TrimSpace(request.Email) == "" {
		h.handleLoginError(c, "Email address is required.")
		return
	}
	if request.Password == "" {
		h.handleLoginError(c, "Password is required.")
		return
	}

	// Transform DTO to domain object
	credentials := request.ToDomain()

	// Call auth service to authenticate with domain object
	logger.Info("Attempting user login", logging.WithUserID(request.Email))
	tokenPair, err := h.authService.Login(c.Request.Context(), credentials)
	if err != nil {
		logger.Error("Login failed", logging.WithUserID(request.Email), logging.WithError(err))
		h.handleLoginError(c, "Invalid email or password. Please try again.")
		return
	}

	// Set authentication cookies/session
	h.setAuthCookies(c, tokenPair)

	logger.Info("Login successful", logging.WithUserID(request.Email))

	// Handle success differently for HTMX vs regular form submissions
	if h.isHTMXRequest(c) {
		// For HTMX requests, return success partial or redirect header
		c.Header("HX-Redirect", "/dashboard")
		c.Status(http.StatusOK)
		c.String(http.StatusOK, `<div class="rounded-md bg-green-50 dark:bg-green-900/20 p-4 border border-green-200 dark:border-green-800">
			<div class="flex">
				<div class="flex-shrink-0">
					<svg class="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.236 4.53L8.53 10.53a.75.75 0 00-1.06 1.06l2.25 2.25a.75.75 0 001.14-.094l3.75-5.25z" clip-rule="evenodd"/>
					</svg>
				</div>
				<div class="ml-3">
					<h3 class="text-sm font-medium text-green-800 dark:text-green-200">Login successful!</h3>
					<div class="mt-2 text-sm text-green-700 dark:text-green-300">
						Redirecting to dashboard...
					</div>
				</div>
			</div>
		</div>`)
	} else {
		// Regular redirect for non-HTMX requests
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}

// RegisterAction handles registration form submission (POST /auth/register)
func (h *UIHandlers) RegisterAction(c *gin.Context) {
	logger := logging.ContextLogger(c).With(logging.WithOperation("register_action"))

	// Parse form data for debugging
	c.Request.ParseForm()
	logger.Info("Registration form data received",
		zap.Any("form_values", c.Request.Form),
		zap.String("content_type", c.ContentType()))

	var request types.RegisterRequestDTO

	if err := c.ShouldBind(&request); err != nil {
		logger.Warn("Registration form binding failed",
			logging.WithError(err),
			zap.Any("raw_form", c.Request.Form))
		h.handleRegistrationError(c, "Invalid form data. Please check your input and try again.")
		return
	}

	// Log bound request for debugging
	logger.Info("Registration form bound successfully",
		zap.String("name", request.Name),
		zap.String("email", request.Email),
		zap.Int("password_length", len(request.Password)))

	// Validate request fields
	if strings.TrimSpace(request.Name) == "" {
		h.handleRegistrationError(c, "Full name is required.")
		return
	}
	if strings.TrimSpace(request.Email) == "" {
		h.handleRegistrationError(c, "Email address is required.")
		return
	}
	if request.Password == "" {
		h.handleRegistrationError(c, "Password is required.")
		return
	}
	if len(request.Password) < 8 {
		h.handleRegistrationError(c, "Password must be at least 8 characters long.")
		return
	}
	if request.ConfirmPassword == "" {
		h.handleRegistrationError(c, "Password confirmation is required.")
		return
	}
	if request.Password != request.ConfirmPassword {
		h.handleRegistrationError(c, "Passwords do not match. Please ensure both password fields are identical.")
		return
	}
	if request.Terms == "" {
		h.handleRegistrationError(c, "You must accept the terms and conditions to register.")
		return
	}

	// Transform DTO to domain object
	user := request.ToDomain()

	// Call auth service to register with domain object and password
	logger.Info("Attempting user registration", logging.WithUserID(request.Email))
	tokenPair, err := h.authService.Register(c.Request.Context(), user, request.Password)
	if err != nil {
		logger.Error("Registration failed", logging.WithUserID(request.Email), logging.WithError(err))

		// Handle specific registration errors
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "already exists"):
			h.handleRegistrationError(c, "An account with this email address already exists. Please use a different email or try logging in.")
		case strings.Contains(errMsg, "invalid email"):
			h.handleRegistrationError(c, "Please enter a valid email address.")
		case strings.Contains(errMsg, "weak password"):
			h.handleRegistrationError(c, "Password is too weak. Please use a stronger password with uppercase, lowercase, numbers, and special characters.")
		default:
			h.handleRegistrationError(c, "Registration failed due to a server error. Please try again later.")
		}
		return
	}

	// Set authentication cookies/session
	h.setAuthCookies(c, tokenPair)

	logger.Info("Registration successful", logging.WithUserID(request.Email))

	// Handle success differently for HTMX vs regular form submissions
	if h.isHTMXRequest(c) {
		// For HTMX requests, return success partial or redirect header
		c.Header("HX-Redirect", "/dashboard")
		c.Status(http.StatusOK)
		c.String(http.StatusOK, `<div class="rounded-md bg-green-50 dark:bg-green-900/20 p-4 border border-green-200 dark:border-green-800">
			<div class="flex">
				<div class="flex-shrink-0">
					<svg class="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.857-9.809a.75.75 0 00-1.214-.882l-3.236 4.53L8.53 10.53a.75.75 0 00-1.06 1.06l2.25 2.25a.75.75 0 001.14-.094l3.75-5.25z" clip-rule="evenodd"/>
					</svg>
				</div>
				<div class="ml-3">
					<h3 class="text-sm font-medium text-green-800 dark:text-green-200">Registration successful!</h3>
					<div class="mt-2 text-sm text-green-700 dark:text-green-300">
						Your account has been created. Redirecting to dashboard...
					</div>
				</div>
			</div>
		</div>`)
	} else {
		// Regular redirect for non-HTMX requests
		c.Redirect(http.StatusSeeOther, "/dashboard")
	}
}

// LogoutAction handles logout form submission (POST /auth/logout)
func (h *UIHandlers) LogoutAction(c *gin.Context) {
	// Clear authentication cookies/session
	// TODO: Implement session clearing

	// Redirect to login page
	c.Redirect(http.StatusSeeOther, "/auth/login")
}

// FinanceCreateIncomeAction handles income creation (POST /finance/income)
func (h *UIHandlers) FinanceCreateIncomeAction(c *gin.Context) {
	// TODO: Implement income creation logic
	// After processing, redirect to finance page
	c.Redirect(http.StatusSeeOther, "/finance")
}

// FinanceUpdateIncomeAction handles income updates (PUT /finance/income/:id)
func (h *UIHandlers) FinanceUpdateIncomeAction(c *gin.Context) {
	// TODO: Implement income update logic
	// After processing, redirect to finance page
	c.Redirect(http.StatusSeeOther, "/finance")
}

// FinanceDeleteIncomeAction handles income deletion (DELETE /finance/income/:id)
func (h *UIHandlers) FinanceDeleteIncomeAction(c *gin.Context) {
	// TODO: Implement income deletion logic
	// After processing, redirect to finance page
	c.Redirect(http.StatusSeeOther, "/finance")
}

// FinanceCreateExpenseAction handles expense creation (POST /finance/expense)
func (h *UIHandlers) FinanceCreateExpenseAction(c *gin.Context) {
	// TODO: Implement expense creation logic
	// After processing, redirect to finance page
	c.Redirect(http.StatusSeeOther, "/finance")
}

// FinanceUpdateExpenseAction handles expense updates (PUT /finance/expense/:id)
func (h *UIHandlers) FinanceUpdateExpenseAction(c *gin.Context) {
	// TODO: Implement expense update logic
	// After processing, redirect to finance page
	c.Redirect(http.StatusSeeOther, "/finance")
}

// FinanceDeleteExpenseAction handles expense deletion (DELETE /finance/expense/:id)
func (h *UIHandlers) FinanceDeleteExpenseAction(c *gin.Context) {
	// TODO: Implement expense deletion logic
	// After processing, redirect to finance page
	c.Redirect(http.StatusSeeOther, "/finance")
}

// HealthCreateProfileAction handles health profile creation (POST /health/profile)
func (h *UIHandlers) HealthCreateProfileAction(c *gin.Context) {
	// TODO: Implement health profile creation logic
	// After processing, redirect to health page
	c.Redirect(http.StatusSeeOther, "/health")
}

// HealthUpdateProfileAction handles health profile updates (PUT /health/profile/:id)
func (h *UIHandlers) HealthUpdateProfileAction(c *gin.Context) {
	// TODO: Implement health profile update logic
	// After processing, redirect to health page
	c.Redirect(http.StatusSeeOther, "/health")
}

// HealthDeleteProfileAction handles health profile deletion (DELETE /health/profile/:id)
func (h *UIHandlers) HealthDeleteProfileAction(c *gin.Context) {
	// TODO: Implement health profile deletion logic
	// After processing, redirect to health page
	c.Redirect(http.StatusSeeOther, "/health")
}

// DecisionCreateAction handles decision creation (POST /decisions)
func (h *UIHandlers) DecisionCreateAction(c *gin.Context) {
	// TODO: Implement decision creation logic when decision service is available
	// After processing, redirect to decision result page
	c.Redirect(http.StatusSeeOther, "/decisions/history")
}

// ================================
// HELPER METHODS
// ================================

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
		errorDTO := types.NewErrorResponse(500, "internal_error", message)

		if h.acceptsJSON(c) {
			c.JSON(500, errorDTO)
		} else {
			// For now, return JSON for HTMX until ErrorPartial is implemented
			c.JSON(500, errorDTO)
		}
	} else {
		// For now, return simple error message until ErrorPage is implemented
		c.Status(500)
		c.String(500, "Internal Server Error: "+message)
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
func (h *UIHandlers) getCurrentUserDTO(c *gin.Context) *types.UserResponseDTO {
	userID := h.getUserID(c)
	if userID == "" {
		return nil
	}

	// TODO: Get user details from auth service
	return &types.UserResponseDTO{
		ID:       userID,
		Email:    "user@example.com", // Placeholder
		Name:     "User",             // Placeholder
		IsActive: true,
	}
}

// convertToUserDTO converts UserResponseDTO to UserDTO for UI templates
func convertToUserDTO(userResponse *types.UserResponseDTO) *types.UserDTO {
	if userResponse == nil {
		return nil
	}
	return &types.UserDTO{
		ID:    userResponse.ID,
		Email: userResponse.Email,
		Name:  userResponse.Name,
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

// handleLoginError handles login errors for both HTMX and regular requests
func (h *UIHandlers) handleLoginError(c *gin.Context, message string) {
	if h.isHTMXRequest(c) {
		// For HTMX requests, return error HTML fragment
		errorHTML := `<div class="rounded-md bg-red-50 dark:bg-red-900/20 p-4 border border-red-200 dark:border-red-800">
			<div class="flex">
				<div class="flex-shrink-0">
					<svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd"/>
					</svg>
				</div>
				<div class="ml-3">
					<h3 class="text-sm font-medium text-red-800 dark:text-red-200">Login failed</h3>
					<div class="mt-2 text-sm text-red-700 dark:text-red-300">` + message + `</div>
				</div>
			</div>
		</div>`

		c.Header("Content-Type", "text/html")
		c.String(http.StatusBadRequest, errorHTML)
	} else {
		// For regular requests, redirect with error parameter
		c.Redirect(http.StatusSeeOther, "/auth/login?error=invalid_credentials")
	}
}

// handleRegistrationError handles registration errors for both HTMX and regular requests
func (h *UIHandlers) handleRegistrationError(c *gin.Context, message string) {
	if h.isHTMXRequest(c) {
		// For HTMX requests, return error HTML fragment
		errorHTML := `<div class="rounded-md bg-red-50 dark:bg-red-900/20 p-4 border border-red-200 dark:border-red-800">
			<div class="flex">
				<div class="flex-shrink-0">
					<svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
						<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.28 7.22a.75.75 0 00-1.06 1.06L8.94 10l-1.72 1.72a.75.75 0 101.06 1.06L10 11.06l1.72 1.72a.75.75 0 101.06-1.06L11.06 10l1.72-1.72a.75.75 0 00-1.06-1.06L10 8.94 8.28 7.22z" clip-rule="evenodd"/>
					</svg>
				</div>
				<div class="ml-3">
					<h3 class="text-sm font-medium text-red-800 dark:text-red-200">Registration failed</h3>
					<div class="mt-2 text-sm text-red-700 dark:text-red-300">` + message + `</div>
				</div>
			</div>
		</div>`

		c.Header("Content-Type", "text/html")
		c.String(http.StatusBadRequest, errorHTML)
	} else {
		// For regular requests, redirect with error parameter
		c.Redirect(http.StatusSeeOther, "/auth/register?error=registration_failed")
	}
}

// setAuthCookies sets secure HTTP-only cookies for authentication tokens
func (h *UIHandlers) setAuthCookies(c *gin.Context, tokenPair *domain.TokenPair) {
	if tokenPair == nil {
		return
	}

	// Set access token cookie (1 hour)
	c.SetCookie(
		"access_token",      // name
		tokenPair.AccessToken, // value
		3600,               // maxAge (1 hour)
		"/",                // path
		"",                 // domain
		false,              // secure (set true in production)
		true,               // httpOnly
	)

	// Set refresh token cookie (7 days)
	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		604800,             // maxAge (7 days)
		"/auth",            // path
		"",                 // domain
		false,              // secure
		true,               // httpOnly
	)
}
