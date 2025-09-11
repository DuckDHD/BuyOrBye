package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/dtos"
	"github.com/DuckDHD/BuyOrBye/internal/logging"
)

// MockAuthService is a mock implementation of AuthService for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, credentials domain.Credentials) (*domain.TokenPair, error) {
	args := m.Called(ctx, credentials)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TokenPair), args.Error(1)
}

func (m *MockAuthService) Register(ctx context.Context, user *domain.User, password string) (*domain.TokenPair, error) {
	args := m.Called(ctx, user, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TokenPair), args.Error(1)
}

func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	args := m.Called(ctx, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TokenPair), args.Error(1)
}

func (m *MockAuthService) Logout(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func setupTestRouter(authService AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	
	// Initialize logging for tests
	config := logging.LogConfig{
		Environment: "test",
		Level:       "debug",
	}
	logging.InitLogger(config)
	
	r := gin.New()
	
	// Add logging middleware for tests
	r.Use(logging.HTTPLoggingMiddleware(logging.DefaultHTTPLoggingConfig()))
	
	handler := NewAuthHandler(authService)
	
	// Set up auth routes
	auth := r.Group("/api/auth")
	{
		auth.POST("/login", handler.Login)
		auth.POST("/register", handler.Register)
		auth.POST("/refresh", handler.RefreshToken)
		auth.POST("/logout", handler.Logout)
	}
	
	return r
}

func createValidTokenPair() *domain.TokenPair {
	return &domain.TokenPair{
		AccessToken:  "valid_access_token",
		RefreshToken: "valid_refresh_token",
		ExpiresIn:    3600,
	}
}

func TestAuthHandler_Login_ValidCredentials_Returns200AndTokens(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	loginRequest := dtos.LoginRequestDTO{
		Email:    "test@example.com",
		Password: "password123",
	}
	expectedTokenPair := createValidTokenPair()
	
	mockAuthService.On("Login", mock.Anything, mock.MatchedBy(func(creds domain.Credentials) bool {
		return creds.Email == loginRequest.Email && creds.Password == loginRequest.Password
	})).Return(expectedTokenPair, nil)
	
	requestBody, _ := json.Marshal(loginRequest)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response dtos.TokenResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, expectedTokenPair.AccessToken, response.AccessToken)
	assert.Equal(t, expectedTokenPair.RefreshToken, response.RefreshToken)
	assert.Equal(t, expectedTokenPair.ExpiresIn, response.ExpiresIn)
	assert.Equal(t, "Bearer", response.TokenType)
	
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials_Returns401(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	loginRequest := dtos.LoginRequestDTO{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	
	mockAuthService.On("Login", mock.Anything, mock.AnythingOfType("domain.Credentials")).
		Return(nil, domain.ErrInvalidCredentials)
	
	requestBody, _ := json.Marshal(loginRequest)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response dtos.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Equal(t, "unauthorized", response.Error)
	
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidEmailFormat_Returns400(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	loginRequest := dtos.LoginRequestDTO{
		Email:    "invalid-email",
		Password: "password123",
	}
	
	requestBody, _ := json.Marshal(loginRequest)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response dtos.ValidationErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "validation_error", response.Error)
	
	// Should not call service if validation fails
	mockAuthService.AssertNotCalled(t, "Login")
}

func TestAuthHandler_Login_MalformedJSON_Returns400(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response dtos.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "bad_request", response.Error)
	
	// Should not call service if JSON is malformed
	mockAuthService.AssertNotCalled(t, "Login")
}

func TestAuthHandler_Register_ValidData_Returns201AndTokens(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	registerRequest := dtos.RegisterRequestDTO{
		Email:    "newuser@example.com",
		Name:     "New User",
		Password: "password123",
	}
	expectedTokenPair := createValidTokenPair()
	
	mockAuthService.On("Register", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
		return user.Email == registerRequest.Email && user.Name == registerRequest.Name
	}), registerRequest.Password).Return(expectedTokenPair, nil)
	
	requestBody, _ := json.Marshal(registerRequest)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response dtos.TokenResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, expectedTokenPair.AccessToken, response.AccessToken)
	assert.Equal(t, expectedTokenPair.RefreshToken, response.RefreshToken)
	assert.Equal(t, expectedTokenPair.ExpiresIn, response.ExpiresIn)
	assert.Equal(t, "Bearer", response.TokenType)
	
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_Register_DuplicateEmail_Returns409(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	registerRequest := dtos.RegisterRequestDTO{
		Email:    "existing@example.com",
		Name:     "User",
		Password: "password123",
	}
	
	mockAuthService.On("Register", mock.Anything, mock.AnythingOfType("*domain.User"), mock.AnythingOfType("string")).
		Return(nil, domain.ErrUserAlreadyExists)
	
	requestBody, _ := json.Marshal(registerRequest)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusConflict, w.Code)
	
	var response dtos.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusConflict, response.Code)
	assert.Equal(t, "conflict", response.Error)
	
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_Register_InvalidData_Returns400(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	registerRequest := dtos.RegisterRequestDTO{
		Email:    "invalid-email",
		Name:     "",
		Password: "short",
	}
	
	requestBody, _ := json.Marshal(registerRequest)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response dtos.ValidationErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "validation_error", response.Error)
	
	// Should not call service if validation fails
	mockAuthService.AssertNotCalled(t, "Register")
}

func TestAuthHandler_RefreshToken_ValidToken_Returns200AndNewTokens(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	refreshRequest := dtos.RefreshTokenRequestDTO{
		RefreshToken: "valid_refresh_token",
	}
	expectedTokenPair := createValidTokenPair()
	
	mockAuthService.On("RefreshToken", mock.Anything, refreshRequest.RefreshToken).
		Return(expectedTokenPair, nil)
	
	requestBody, _ := json.Marshal(refreshRequest)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response dtos.TokenResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, expectedTokenPair.AccessToken, response.AccessToken)
	assert.Equal(t, expectedTokenPair.RefreshToken, response.RefreshToken)
	assert.Equal(t, expectedTokenPair.ExpiresIn, response.ExpiresIn)
	assert.Equal(t, "Bearer", response.TokenType)
	
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_InvalidToken_Returns401(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	refreshRequest := dtos.RefreshTokenRequestDTO{
		RefreshToken: "invalid_token",
	}
	
	mockAuthService.On("RefreshToken", mock.Anything, refreshRequest.RefreshToken).
		Return(nil, domain.ErrInvalidToken)
	
	requestBody, _ := json.Marshal(refreshRequest)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response dtos.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Equal(t, "unauthorized", response.Error)
	
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_MissingToken_Returns400(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	refreshRequest := dtos.RefreshTokenRequestDTO{
		RefreshToken: "",
	}
	
	requestBody, _ := json.Marshal(refreshRequest)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response dtos.ValidationErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "validation_error", response.Error)
	
	// Should not call service if validation fails
	mockAuthService.AssertNotCalled(t, "RefreshToken")
}

func TestAuthHandler_Logout_ValidToken_Returns200(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	refreshRequest := dtos.RefreshTokenRequestDTO{
		RefreshToken: "valid_refresh_token",
	}
	
	mockAuthService.On("Logout", mock.Anything, refreshRequest.RefreshToken).
		Return(nil)
	
	requestBody, _ := json.Marshal(refreshRequest)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/logout", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "logged out successfully", response["message"])
	
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_Logout_InvalidToken_Returns401(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	refreshRequest := dtos.RefreshTokenRequestDTO{
		RefreshToken: "invalid_token",
	}
	
	mockAuthService.On("Logout", mock.Anything, refreshRequest.RefreshToken).
		Return(domain.ErrInvalidToken)
	
	requestBody, _ := json.Marshal(refreshRequest)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/logout", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response dtos.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Equal(t, "unauthorized", response.Error)
	
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_Login_AccountInactive_Returns401(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	loginRequest := dtos.LoginRequestDTO{
		Email:    "inactive@example.com",
		Password: "password123",
	}
	
	mockAuthService.On("Login", mock.Anything, mock.AnythingOfType("domain.Credentials")).
		Return(nil, domain.ErrAccountInactive)
	
	requestBody, _ := json.Marshal(loginRequest)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response dtos.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Equal(t, "unauthorized", response.Error)
	assert.Contains(t, response.Message, "account is inactive")
	
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_RefreshToken_TokenExpired_Returns401(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	refreshRequest := dtos.RefreshTokenRequestDTO{
		RefreshToken: "expired_token",
	}
	
	mockAuthService.On("RefreshToken", mock.Anything, refreshRequest.RefreshToken).
		Return(nil, domain.ErrTokenExpired)
	
	requestBody, _ := json.Marshal(refreshRequest)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response dtos.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Equal(t, "unauthorized", response.Error)
	assert.Contains(t, response.Message, "token has expired")
	
	mockAuthService.AssertExpectations(t)
}

func TestAuthHandler_Register_MalformedJSON_Returns400(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response dtos.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "bad_request", response.Error)
	
	// Should not call service if JSON is malformed
	mockAuthService.AssertNotCalled(t, "Register")
}

func TestAuthHandler_RefreshToken_MalformedJSON_Returns400(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response dtos.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "bad_request", response.Error)
	
	// Should not call service if JSON is malformed
	mockAuthService.AssertNotCalled(t, "RefreshToken")
}

func TestAuthHandler_Logout_MalformedJSON_Returns400(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/logout", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response dtos.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "bad_request", response.Error)
	
	// Should not call service if JSON is malformed
	mockAuthService.AssertNotCalled(t, "Logout")
}

func TestAuthHandler_Logout_EmptyToken_Returns400(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	refreshRequest := dtos.RefreshTokenRequestDTO{
		RefreshToken: "",
	}
	
	requestBody, _ := json.Marshal(refreshRequest)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/logout", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response dtos.ValidationErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "validation_error", response.Error)
	
	// Should not call service if validation fails
	mockAuthService.AssertNotCalled(t, "Logout")
}

func TestAuthHandler_Login_RequiredFieldsMissing_Returns400(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	loginRequest := dtos.LoginRequestDTO{
		Email:    "",
		Password: "",
	}
	
	requestBody, _ := json.Marshal(loginRequest)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response dtos.ValidationErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "validation_error", response.Error)
	assert.Contains(t, response.Fields, "Email")
	assert.Contains(t, response.Fields, "Password")
	
	// Should not call service if validation fails
	mockAuthService.AssertNotCalled(t, "Login")
}

func TestAuthHandler_Login_PasswordTooShort_Returns400(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	loginRequest := dtos.LoginRequestDTO{
		Email:    "test@example.com",
		Password: "short",
	}
	
	requestBody, _ := json.Marshal(loginRequest)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response dtos.ValidationErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "validation_error", response.Error)
	assert.Contains(t, response.Fields, "Password")
	
	// Should not call service if validation fails
	mockAuthService.AssertNotCalled(t, "Login")
}

func TestAuthHandler_Register_InternalServerError_Returns500(t *testing.T) {
	// Arrange
	mockAuthService := new(MockAuthService)
	router := setupTestRouter(mockAuthService)
	
	registerRequest := dtos.RegisterRequestDTO{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	
	// Return an unexpected error (not one of the predefined domain errors)
	unexpectedErr := fmt.Errorf("database connection failed")
	mockAuthService.On("Register", mock.Anything, mock.AnythingOfType("*domain.User"), mock.AnythingOfType("string")).
		Return(nil, unexpectedErr)
	
	requestBody, _ := json.Marshal(registerRequest)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response dtos.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "internal_error", response.Error)
	
	mockAuthService.AssertExpectations(t)
}