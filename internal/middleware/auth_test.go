package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/types"
)

// MockJWTService is a mock implementation of JWTService for testing
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateTokenPair(userID, email string) (*domain.TokenPair, error) {
	args := m.Called(userID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TokenPair), args.Error(1)
}

func (m *MockJWTService) ValidateAccessToken(tokenString string) (*domain.TokenClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TokenClaims), args.Error(1)
}

func (m *MockJWTService) ValidateRefreshToken(tokenString string) (*domain.TokenClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TokenClaims), args.Error(1)
}

func setupTestRouter(middleware gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	// Add the middleware to a test route
	r.GET("/protected", middleware, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "protected resource accessed"})
	})
	
	return r
}

func createValidTokenClaims() *domain.TokenClaims {
	return &domain.TokenClaims{
		UserID:    "user-123",
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
	}
}

func createExpiredTokenClaims() *domain.TokenClaims {
	return &domain.TokenClaims{
		UserID:    "user-123",
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
	}
}

func TestJWTAuthMiddleware_RequireAuth_ValidToken_AllowsAccess(t *testing.T) {
	// Arrange
	mockJWTService := new(MockJWTService)
	middleware := NewJWTAuthMiddleware(mockJWTService)
	router := setupTestRouter(middleware.RequireAuth())
	
	validClaims := createValidTokenClaims()
	mockJWTService.On("ValidateAccessToken", "valid_token").Return(validClaims, nil)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "protected resource accessed", response["message"])
	
	mockJWTService.AssertExpectations(t)
}

func TestJWTAuthMiddleware_RequireAuth_MissingAuthHeader_Returns401(t *testing.T) {
	// Arrange
	mockJWTService := new(MockJWTService)
	middleware := NewJWTAuthMiddleware(mockJWTService)
	router := setupTestRouter(middleware.RequireAuth())
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	// No Authorization header
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response types.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Equal(t, "unauthorized", response.Error)
	assert.Contains(t, response.Message, "Authorization header is required")
	
	// Should not call JWT service
	mockJWTService.AssertNotCalled(t, "ValidateAccessToken")
}

func TestJWTAuthMiddleware_RequireAuth_InvalidAuthHeaderFormat_Returns401(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
	}{
		{
			name:       "Missing Bearer prefix",
			authHeader: "valid_token",
		},
		{
			name:       "Wrong prefix",
			authHeader: "Basic valid_token",
		},
		{
			name:       "Extra spaces",
			authHeader: "Bearer  valid_token",
		},
		{
			name:       "Missing token after Bearer",
			authHeader: "Bearer",
		},
		{
			name:       "Empty token",
			authHeader: "Bearer ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockJWTService := new(MockJWTService)
			middleware := NewJWTAuthMiddleware(mockJWTService)
			router := setupTestRouter(middleware.RequireAuth())
			
			// Act
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tt.authHeader)
			router.ServeHTTP(w, req)
			
			// Assert
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			
			var response types.ErrorResponseDTO
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			assert.Equal(t, http.StatusUnauthorized, response.Code)
			assert.Equal(t, "unauthorized", response.Error)
			
			// Should not call JWT service for malformed headers
			mockJWTService.AssertNotCalled(t, "ValidateAccessToken")
		})
	}
}

func TestJWTAuthMiddleware_RequireAuth_InvalidToken_Returns401(t *testing.T) {
	// Arrange
	mockJWTService := new(MockJWTService)
	middleware := NewJWTAuthMiddleware(mockJWTService)
	router := setupTestRouter(middleware.RequireAuth())
	
	mockJWTService.On("ValidateAccessToken", "invalid_token").
		Return(nil, fmt.Errorf("invalid token"))
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response types.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Equal(t, "unauthorized", response.Error)
	assert.Equal(t, "Invalid access token", response.Message)
	
	mockJWTService.AssertExpectations(t)
}

func TestJWTAuthMiddleware_RequireAuth_ExpiredToken_Returns401(t *testing.T) {
	// Arrange
	mockJWTService := new(MockJWTService)
	middleware := NewJWTAuthMiddleware(mockJWTService)
	router := setupTestRouter(middleware.RequireAuth())
	
	expiredClaims := createExpiredTokenClaims()
	mockJWTService.On("ValidateAccessToken", "expired_token").Return(expiredClaims, nil)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer expired_token")
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var response types.ErrorResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Equal(t, "unauthorized", response.Error)
	assert.Contains(t, response.Message, "Access token has expired")
	
	mockJWTService.AssertExpectations(t)
}

func TestJWTAuthMiddleware_RequireAuth_TokenErrorTypes_ReturnsAppropriateMessages(t *testing.T) {
	tests := []struct {
		name           string
		tokenError     string
		expectedMessage string
	}{
		{
			name:           "Expired token error",
			tokenError:     "access token is expired",
			expectedMessage: "Access token has expired",
		},
		{
			name:           "Signature error",
			tokenError:     "signature is invalid",
			expectedMessage: "Invalid token signature",
		},
		{
			name:           "Malformed token error",
			tokenError:     "failed to parse access token: malformed",
			expectedMessage: "Malformed access token",
		},
		{
			name:           "Generic error",
			tokenError:     "some other error",
			expectedMessage: "Invalid access token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockJWTService := new(MockJWTService)
			middleware := NewJWTAuthMiddleware(mockJWTService)
			router := setupTestRouter(middleware.RequireAuth())
			
			mockJWTService.On("ValidateAccessToken", "error_token").
				Return(nil, fmt.Errorf("%s", tt.tokenError))
			
			// Act
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", "Bearer error_token")
			router.ServeHTTP(w, req)
			
			// Assert
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			
			var response types.ErrorResponseDTO
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			
			assert.Equal(t, tt.expectedMessage, response.Message)
			
			mockJWTService.AssertExpectations(t)
		})
	}
}

func TestJWTAuthMiddleware_OptionalAuth_ValidToken_SetsContext(t *testing.T) {
	// Arrange
	mockJWTService := new(MockJWTService)
	middleware := NewJWTAuthMiddleware(mockJWTService)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/optional", middleware.OptionalAuth(), func(c *gin.Context) {
		userID := GetUserID(c)
		email := GetUserEmail(c)
		c.JSON(http.StatusOK, gin.H{
			"userID": userID,
			"email":  email,
			"authenticated": IsAuthenticated(c),
		})
	})
	
	validClaims := createValidTokenClaims()
	mockJWTService.On("ValidateAccessToken", "valid_token").Return(validClaims, nil)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/optional", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	r.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "user-123", response["userID"])
	assert.Equal(t, "test@example.com", response["email"])
	assert.Equal(t, true, response["authenticated"])
	
	mockJWTService.AssertExpectations(t)
}

func TestJWTAuthMiddleware_OptionalAuth_NoToken_ContinuesWithoutAuth(t *testing.T) {
	// Arrange
	mockJWTService := new(MockJWTService)
	middleware := NewJWTAuthMiddleware(mockJWTService)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/optional", middleware.OptionalAuth(), func(c *gin.Context) {
		userID := GetUserID(c)
		email := GetUserEmail(c)
		c.JSON(http.StatusOK, gin.H{
			"userID": userID,
			"email":  email,
			"authenticated": IsAuthenticated(c),
		})
	})
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/optional", nil)
	// No Authorization header
	r.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "", response["userID"])
	assert.Equal(t, "", response["email"])
	assert.Equal(t, false, response["authenticated"])
	
	// Should not call JWT service
	mockJWTService.AssertNotCalled(t, "ValidateAccessToken")
}

func TestJWTAuthMiddleware_OptionalAuth_InvalidToken_ContinuesWithoutAuth(t *testing.T) {
	// Arrange
	mockJWTService := new(MockJWTService)
	middleware := NewJWTAuthMiddleware(mockJWTService)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/optional", middleware.OptionalAuth(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"authenticated": IsAuthenticated(c),
		})
	})
	
	mockJWTService.On("ValidateAccessToken", "invalid_token").
		Return(nil, fmt.Errorf("invalid token"))
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/optional", nil)
	req.Header.Set("Authorization", "Bearer invalid_token")
	r.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, false, response["authenticated"])
	
	mockJWTService.AssertExpectations(t)
}

func TestJWTAuthMiddleware_ContextHelpers_NoAuth_ReturnEmptyValues(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		claims := GetUserClaims(c)
		userID := GetUserID(c)
		email := GetUserEmail(c)
		authenticated := IsAuthenticated(c)
		
		c.JSON(http.StatusOK, gin.H{
			"claims":        claims,
			"userID":        userID,
			"email":         email,
			"authenticated": authenticated,
		})
	})
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Nil(t, response["claims"])
	assert.Equal(t, "", response["userID"])
	assert.Equal(t, "", response["email"])
	assert.Equal(t, false, response["authenticated"])
}

func TestJWTAuthMiddleware_ContextHelpers_WithAuth_ReturnValues(t *testing.T) {
	// Arrange
	mockJWTService := new(MockJWTService)
	middleware := NewJWTAuthMiddleware(mockJWTService)
	
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/test", middleware.RequireAuth(), func(c *gin.Context) {
		claims := GetUserClaims(c)
		userID := GetUserID(c)
		email := GetUserEmail(c)
		authenticated := IsAuthenticated(c)
		
		c.JSON(http.StatusOK, gin.H{
			"hasValidClaims": claims != nil && claims.UserID == "user-123",
			"userID":         userID,
			"email":          email,
			"authenticated":  authenticated,
		})
	})
	
	validClaims := createValidTokenClaims()
	mockJWTService.On("ValidateAccessToken", "valid_token").Return(validClaims, nil)
	
	// Act
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid_token")
	r.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, true, response["hasValidClaims"])
	assert.Equal(t, "user-123", response["userID"])
	assert.Equal(t, "test@example.com", response["email"])
	assert.Equal(t, true, response["authenticated"])
	
	mockJWTService.AssertExpectations(t)
}

func TestNewJWTAuthMiddleware_CreatesInstance(t *testing.T) {
	// Arrange
	mockJWTService := new(MockJWTService)
	
	// Act
	middleware := NewJWTAuthMiddleware(mockJWTService)
	
	// Assert
	assert.NotNil(t, middleware)
	assert.Equal(t, mockJWTService, middleware.jwtService)
}