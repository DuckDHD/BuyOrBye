package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/config"
	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/DuckDHD/BuyOrBye/internal/services"
	"github.com/DuckDHD/BuyOrBye/templates"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/google/uuid"
)

type Handlers struct {
	decisionService *services.DecisionService
	userService     *services.UserService
	paymentService  *services.PaymentService
	config          *config.Config
	logger          *slog.Logger
}

func New(
	decisionService *services.DecisionService,
	userService *services.UserService,
	paymentService *services.PaymentService,
	config *config.Config,
) *Handlers {
	// Use slog.Default() if caller doesn't inject a logger elsewhere.
	return &Handlers{
		decisionService: decisionService,
		userService:     userService,
		paymentService:  paymentService,
		config:          config,
		logger:          slog.Default(),
	}
}

func (h *Handlers) SetupRoutes(r *chi.Mux) {
	// Core middleware
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(h.loggingMiddleware()) // structured request logs w/ status & duration

	// CORS middleware
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   h.config.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiting
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	// Static files
	fileServer := http.FileServer(http.Dir("./static/"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Health check
	r.Get("/health", h.healthCheck)

	// Main routes
	r.Get("/", h.home)
	r.Post("/decide", h.makeDecision)
	r.Get("/decision/{id}", h.getSharedDecision)

	// Auth routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/magic-link", h.sendMagicLink)
		r.Get("/callback", h.authCallback)
		r.Post("/logout", h.logout)
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Use(h.requireAuth) // Protected routes

		r.Get("/me", h.getCurrentUser)
		r.Get("/me/decisions", h.getUserDecisions)
		r.Get("/me/credits", h.getUserCredits)

		r.Post("/profile", h.updateProfile)
		r.Post("/referrals", h.createReferral)
	})

	// Payment webhooks (no auth required)
	r.Post("/webhooks/paddle", h.handlePaddleWebhook)
}

func (h *Handlers) home(w http.ResponseWriter, r *http.Request) {
	l := h.reqLogger(r)
	l.Info("render home")
	component := templates.Home()
	component.Render(r.Context(), w) // template.Render has no error; log is best-effort
}

func (h *Handlers) makeDecision(w http.ResponseWriter, r *http.Request) {
	l := h.reqLogger(r)

	// Parse form data
	if err := r.ParseForm(); err != nil {
		h.respondError(w, r, http.StatusBadRequest, "Invalid form data", err,
			"raw_body_len", r.ContentLength)
		return
	}

	// Extract decision input
	input := models.DecisionInput{
		ItemName:    strings.TrimSpace(r.FormValue("item_name")),
		Currency:    strings.TrimSpace(r.FormValue("currency")),
		IsNecessity: r.FormValue("is_necessity") == "on",
		Description: r.FormValue("description"),
	}

	// Parse price
	priceStr := strings.TrimSpace(r.FormValue("price"))
	if priceStr == "" {
		h.respondError(w, r, http.StatusBadRequest, "Price is required", nil)
		return
	}
	var price float64
	if _, err := fmt.Sscanf(priceStr, "%f", &price); err != nil {
		h.respondError(w, r, http.StatusBadRequest, "Invalid price format", err, "price_str", priceStr)
		return
	}
	input.Price = price

	// Get or create session
	sessionID := h.getOrCreateSession(r)

	// Get user ID if authenticated (header-based for now)
	var userID *uuid.UUID
	if userIDStr := r.Header.Get("X-User-ID"); userIDStr != "" {
		if id, err := uuid.Parse(userIDStr); err == nil {
			userID = &id
		} else {
			l.Warn("invalid X-User-ID header", "value", userIDStr, "err", err)
		}
	}

	// Log sanitized input (avoid large/PII in description)
	l.Info("make decision request",
		"item", safeTrunc(input.ItemName, 80),
		"currency", input.Currency,
		"is_necessity", input.IsNecessity,
		"price", input.Price,
		"desc_len", len(input.Description),
		"user_id", userID,
		"session_id", sessionID,
	)

	// Make decision
	ctx := r.Context()
	response, err := h.decisionService.MakeDecision(ctx, input, userID, sessionID)
	if err != nil {
		h.respondError(w, r, http.StatusInternalServerError, "Failed to make decision", err,
			"item", safeTrunc(input.ItemName, 80),
			"price", input.Price,
		)
		return
	}

	// Return HTML response for HTMX
	w.Header().Set("Content-Type", "text/html")
	if err := h.renderDecisionResult(w, response); err != nil {
		h.respondError(w, r, http.StatusInternalServerError, "Failed to render decision result", err)
		return
	}

	l.Info("decision made",
		"verdict", response.Decision.Verdict,
		"score", response.Decision.Score,
		"reasons_count", len(response.Decision.Reasons),
		"has_flip_to_yes", response.FlipToYes != nil,
	)
}

func (h *Handlers) renderDecisionResult(w http.ResponseWriter, response *models.DecisionResponse) error {
	decision := response.Decision
	verdictClass := "text-green-600"
	if decision.Verdict == models.VerdictNo {
		verdictClass = "text-red-600"
	}

	html := fmt.Sprintf(`
		<div class="mt-6 p-4 border rounded-lg %s">
			<div class="flex items-center justify-between mb-4">
				<h3 class="text-lg font-semibold %s">%s</h3>
				<span class="text-sm text-gray-500">Score: %.1f</span>
			</div>
			<div class="space-y-2">
				%s
			</div>
			%s
		</div>
	`,
		getBorderClass(decision.Verdict),
		verdictClass,
		decision.Verdict,
		decision.Score,
		formatReasons(decision.Reasons),
		formatFlipToYes(response.FlipToYes),
	)

	if _, err := w.Write([]byte(html)); err != nil {
		return err
	}
	return nil
}

func getBorderClass(verdict models.DecisionVerdict) string {
	if verdict == models.VerdictYes {
		return "border-green-200 bg-green-50"
	}
	return "border-red-200 bg-red-50"
}

func formatReasons(reasons models.ReasonsJSON) string {
	var b strings.Builder
	for _, reason := range reasons {
		b.WriteString(`<p class="text-sm text-gray-700">• `)
		b.WriteString(htmlEscape(reason))
		b.WriteString(`</p>`)
	}
	return b.String()
}

func htmlEscape(s string) string {
	// keep it minimal; you can swap with html/template if needed
	replacer := strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`>`, "&gt;",
		`"`, "&quot;",
		`'`, "&#39;",
	)
	return replacer.Replace(s)
}

func formatFlipToYes(flipToYes *string) string {
	if flipToYes == nil {
		return ""
	}
	return fmt.Sprintf(`
		<div class="mt-4 p-3 bg-blue-50 border border-blue-200 rounded">
			<p class="text-sm text-blue-800">
				<strong>To make this a YES:</strong> %s
			</p>
		</div>
	`, htmlEscape(*flipToYes))
}

func (h *Handlers) getOrCreateSession(r *http.Request) uuid.UUID {
	// Check for existing session cookie
	if cookie, err := r.Cookie("session_id"); err == nil {
		if sessionID, err := uuid.Parse(cookie.Value); err == nil {
			return sessionID
		}
		h.reqLogger(r).Warn("invalid session_id cookie", "value", cookie.Value, "err", err)
	}

	// Create new session (note: consider setting cookie upstream)
	newID := uuid.New()
	h.reqLogger(r).Info("created new session_id", "session_id", newID)
	return newID
}

func (h *Handlers) healthCheck(w http.ResponseWriter, r *http.Request) {
	h.reqLogger(r).Info("healthcheck called")
	response := map[string]string{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

func (h *Handlers) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("X-User-ID")
		if userID == "" {
			h.respondError(w, r, http.StatusUnauthorized, "Authentication required", nil)
			return
		}
		ctx := context.WithValue(r.Context(), "user_id", userID)
		h.reqLogger(r).Info("auth ok", "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

/* ===========================
 * Placeholder handlers (TODO)
 * ===========================
 * Keep basic logs so issues are traceable if these are invoked.
 */

func (h *Handlers) sendMagicLink(w http.ResponseWriter, r *http.Request) {
	h.reqLogger(r).Info("sendMagicLink called (TODO)")
	w.WriteHeader(http.StatusNotImplemented)
}
func (h *Handlers) authCallback(w http.ResponseWriter, r *http.Request) {
	h.reqLogger(r).Info("authCallback called (TODO)")
	w.WriteHeader(http.StatusNotImplemented)
}
func (h *Handlers) logout(w http.ResponseWriter, r *http.Request) {
	h.reqLogger(r).Info("logout called (TODO)")
	w.WriteHeader(http.StatusNotImplemented)
}
func (h *Handlers) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	h.reqLogger(r).Info("getCurrentUser called (TODO)")
	w.WriteHeader(http.StatusNotImplemented)
}
func (h *Handlers) getUserDecisions(w http.ResponseWriter, r *http.Request) {
	h.reqLogger(r).Info("getUserDecisions called (TODO)")
	w.WriteHeader(http.StatusNotImplemented)
}
func (h *Handlers) getUserCredits(w http.ResponseWriter, r *http.Request) {
	h.reqLogger(r).Info("getUserCredits called (TODO)")
	w.WriteHeader(http.StatusNotImplemented)
}
func (h *Handlers) updateProfile(w http.ResponseWriter, r *http.Request) {
	h.reqLogger(r).Info("updateProfile called (TODO)")
	w.WriteHeader(http.StatusNotImplemented)
}
func (h *Handlers) createReferral(w http.ResponseWriter, r *http.Request) {
	h.reqLogger(r).Info("createReferral called (TODO)")
	w.WriteHeader(http.StatusNotImplemented)
}
func (h *Handlers) getSharedDecision(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.reqLogger(r).Info("getSharedDecision called (TODO)", "id", id)
	w.WriteHeader(http.StatusNotImplemented)
}
func (h *Handlers) handlePaddleWebhook(w http.ResponseWriter, r *http.Request) {
	h.reqLogger(r).Info("handlePaddleWebhook called (TODO)")
	w.WriteHeader(http.StatusNotImplemented)
}

/* ===========================
 * Logging helpers
 * ===========================
 */

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
	bytes       int
}

func (sr *statusRecorder) WriteHeader(code int) {
	if !sr.wroteHeader {
		sr.status = code
		sr.wroteHeader = true
		sr.ResponseWriter.WriteHeader(code)
	}
}

func (sr *statusRecorder) Write(b []byte) (int, error) {
	if !sr.wroteHeader {
		sr.WriteHeader(http.StatusOK)
	}
	n, err := sr.ResponseWriter.Write(b)
	sr.bytes += n
	return n, err
}

func (h *Handlers) loggingMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := middleware.GetReqID(r.Context())
			ip := r.Header.Get("X-Forwarded-For")
			if ip == "" {
				ip = r.RemoteAddr
			}

			sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sr, r)

			h.logger.Info("http request",
				"req_id", reqID,
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"status", sr.status,
				"bytes", sr.bytes,
				"duration_ms", time.Since(start).Milliseconds(),
				"ip", ip,
				"user_agent", r.UserAgent(),
				"referer", r.Referer(),
			)
		})
	}
}

func (h *Handlers) reqLogger(r *http.Request) *slog.Logger {
	reqID := middleware.GetReqID(r.Context())
	userID := r.Context().Value("user_id")
	return h.logger.With(
		"req_id", reqID,
		"user_id", userID,
		"path", r.URL.Path,
		"method", r.Method,
	)
}

func (h *Handlers) respondError(w http.ResponseWriter, r *http.Request, status int, publicMsg string, err error, kv ...any) {
	l := h.reqLogger(r)
	args := []any{
		"status", status,
	}
	if err != nil {
		args = append(args, "err", err.Error())
	}
	if len(kv) > 0 {
		args = append(args, kv...)
	}
	l.Error("request error", args...)
	http.Error(w, publicMsg, status)
}

func safeTrunc(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
