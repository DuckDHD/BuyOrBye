package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID string, loginTime time.Time) error {
	args := m.Called(ctx, userID, loginTime)
	return args.Error(0)
}

// MockTokenRepository is a mock implementation of TokenRepository
type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) SaveRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	args := m.Called(ctx, userID, token, expiresAt)
	return args.Error(0)
}

func (m *MockTokenRepository) GetRefreshToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func (m *MockTokenRepository) RevokeToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockTokenRepository) CleanupExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockPasswordService is a mock implementation of PasswordService
type MockPasswordService struct {
	mock.Mock
}

func (m *MockPasswordService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockPasswordService) CheckPassword(hash, password string) error {
	args := m.Called(hash, password)
	return args.Error(0)
}

// MockJWTService is a mock implementation of JWTService
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

// Helper functions for test setup
func createValidUser() *domain.User {
	return &domain.User{
		ID:           "1",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hashed_password",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func createValidTokenPair() *domain.TokenPair {
	return &domain.TokenPair{
		AccessToken:  "access_token_value",
		RefreshToken: "refresh_token_value",
		ExpiresIn:    900, // 15 minutes
	}
}

func createValidTokenClaims() *domain.TokenClaims {
	return &domain.TokenClaims{
		UserID:    "1",
		Email:     "test@example.com",
		ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
	}
}

func setupAuthServiceMocks() (*MockUserRepository, *MockTokenRepository, *MockPasswordService, *MockJWTService) {
	userRepo := &MockUserRepository{}
	tokenRepo := &MockTokenRepository{}
	passwordService := &MockPasswordService{}
	jwtService := &MockJWTService{}

	return userRepo, tokenRepo, passwordService, jwtService
}

// Test Login with valid credentials returns tokens
func TestAuthService_Login_ValidCredentials_ReturnsTokens(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	credentials := domain.Credentials{
		Email:    "test@example.com",
		Password: "password123",
	}

	user := createValidUser()
	tokenPair := createValidTokenPair()

	userRepo.On("GetByEmail", ctx, credentials.Email).Return(user, nil)
	passwordService.On("CheckPassword", user.PasswordHash, credentials.Password).Return(nil)
	jwtService.On("GenerateTokenPair", user.ID, user.Email).Return(tokenPair, nil)
	tokenRepo.On("SaveRefreshToken", ctx, user.ID, tokenPair.RefreshToken, mock.AnythingOfType("time.Time")).Return(nil)
	userRepo.On("UpdateLastLogin", ctx, user.ID, mock.AnythingOfType("time.Time")).Return(nil)

	// Act
	result, err := service.Login(ctx, credentials)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, tokenPair, result)
	userRepo.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	passwordService.AssertExpectations(t)
	jwtService.AssertExpectations(t)
}

// Test Login with invalid email returns error
func TestAuthService_Login_InvalidEmail_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	credentials := domain.Credentials{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	userRepo.On("GetByEmail", ctx, credentials.Email).Return(nil, domain.ErrUserNotFound)

	// Act
	result, err := service.Login(ctx, credentials)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrInvalidCredentials, err)
	userRepo.AssertExpectations(t)
	tokenRepo.AssertNotCalled(t, "SaveRefreshToken")
	passwordService.AssertNotCalled(t, "CheckPassword")
	jwtService.AssertNotCalled(t, "GenerateTokenPair")
}

// Test Login with wrong password returns error
func TestAuthService_Login_WrongPassword_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	credentials := domain.Credentials{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}

	user := createValidUser()

	userRepo.On("GetByEmail", ctx, credentials.Email).Return(user, nil)
	passwordService.On("CheckPassword", user.PasswordHash, credentials.Password).Return(errors.New("password mismatch"))

	// Act
	result, err := service.Login(ctx, credentials)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrInvalidCredentials, err)
	userRepo.AssertExpectations(t)
	passwordService.AssertExpectations(t)
	tokenRepo.AssertNotCalled(t, "SaveRefreshToken")
	jwtService.AssertNotCalled(t, "GenerateTokenPair")
}

// Test Login with inactive account returns error
func TestAuthService_Login_InactiveAccount_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	credentials := domain.Credentials{
		Email:    "test@example.com",
		Password: "password123",
	}

	user := createValidUser()
	user.IsActive = false

	userRepo.On("GetByEmail", ctx, credentials.Email).Return(user, nil)

	// Act
	result, err := service.Login(ctx, credentials)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrAccountInactive, err)
	userRepo.AssertExpectations(t)
	passwordService.AssertNotCalled(t, "CheckPassword")
	jwtService.AssertNotCalled(t, "GenerateTokenPair")
	tokenRepo.AssertNotCalled(t, "SaveRefreshToken")
}

// Test Register creates user and returns tokens
func TestAuthService_Register_ValidData_CreatesUserAndReturnsTokens(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	user := &domain.User{
		Email: "newuser@example.com",
		Name:  "New User",
	}
	password := "password123"

	hashedPassword := "hashed_password"
	tokenPair := createValidTokenPair()

	userRepo.On("GetByEmail", ctx, user.Email).Return(nil, domain.ErrUserNotFound)
	passwordService.On("HashPassword", password).Return(hashedPassword, nil)
	userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil).Run(func(args mock.Arguments) {
		user := args.Get(1).(*domain.User)
		user.ID = "2" // Simulate ID assignment
	})
	jwtService.On("GenerateTokenPair", "2", user.Email).Return(tokenPair, nil)
	tokenRepo.On("SaveRefreshToken", ctx, "2", tokenPair.RefreshToken, mock.AnythingOfType("time.Time")).Return(nil)

	// Act
	result, err := service.Register(ctx, user, password)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, tokenPair, result)
	userRepo.AssertExpectations(t)
	passwordService.AssertExpectations(t)
	jwtService.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

// Test Register with duplicate email returns error
func TestAuthService_Register_DuplicateEmail_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	user := &domain.User{
		Email: "existing@example.com",
		Name:  "Existing User",
	}
	password := "password123"

	existingUser := createValidUser()
	existingUser.Email = user.Email

	userRepo.On("GetByEmail", ctx, user.Email).Return(existingUser, nil)

	// Act
	result, err := service.Register(ctx, user, password)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrUserAlreadyExists, err)
	userRepo.AssertExpectations(t)
	passwordService.AssertNotCalled(t, "HashPassword")
	userRepo.AssertNotCalled(t, "Create")
	jwtService.AssertNotCalled(t, "GenerateTokenPair")
	tokenRepo.AssertNotCalled(t, "SaveRefreshToken")
}

// Test Register with password hashing failure
func TestAuthService_Register_PasswordHashingFails_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	user := &domain.User{
		Email: "newuser@example.com",
		Name:  "New User",
	}
	password := "password123"

	userRepo.On("GetByEmail", ctx, user.Email).Return(nil, domain.ErrUserNotFound)
	passwordService.On("HashPassword", password).Return("", errors.New("hashing failed"))

	// Act
	result, err := service.Register(ctx, user, password)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to hash password")
	userRepo.AssertExpectations(t)
	passwordService.AssertExpectations(t)
	userRepo.AssertNotCalled(t, "Create")
	jwtService.AssertNotCalled(t, "GenerateTokenPair")
}

// Test RefreshToken with valid token returns new pair
func TestAuthService_RefreshToken_ValidToken_ReturnsNewTokenPair(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	refreshToken := "valid_refresh_token"
	claims := createValidTokenClaims()
	user := createValidUser()
	newTokenPair := createValidTokenPair()

	jwtService.On("ValidateRefreshToken", refreshToken).Return(claims, nil)
	tokenRepo.On("GetRefreshToken", ctx, refreshToken).Return(claims.UserID, nil)
	userRepo.On("GetByID", ctx, claims.UserID).Return(user, nil)
	tokenRepo.On("RevokeToken", ctx, refreshToken).Return(nil)
	jwtService.On("GenerateTokenPair", user.ID, user.Email).Return(newTokenPair, nil)
	tokenRepo.On("SaveRefreshToken", ctx, user.ID, newTokenPair.RefreshToken, mock.AnythingOfType("time.Time")).Return(nil)

	// Act
	result, err := service.RefreshToken(ctx, refreshToken)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, newTokenPair, result)
	jwtService.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

// Test RefreshToken with invalid token returns error
func TestAuthService_RefreshToken_InvalidToken_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	refreshToken := "invalid_refresh_token"

	jwtService.On("ValidateRefreshToken", refreshToken).Return(nil, domain.ErrInvalidToken)

	// Act
	result, err := service.RefreshToken(ctx, refreshToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrInvalidToken, err)
	jwtService.AssertExpectations(t)
	tokenRepo.AssertNotCalled(t, "GetRefreshToken")
	userRepo.AssertNotCalled(t, "GetByID")
}

// Test RefreshToken with expired token returns error
func TestAuthService_RefreshToken_ExpiredToken_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	refreshToken := "expired_refresh_token"

	jwtService.On("ValidateRefreshToken", refreshToken).Return(nil, domain.ErrTokenExpired)

	// Act
	result, err := service.RefreshToken(ctx, refreshToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrTokenExpired, err)
	jwtService.AssertExpectations(t)
	tokenRepo.AssertNotCalled(t, "GetRefreshToken")
}

// Test RefreshToken with revoked token returns error
func TestAuthService_RefreshToken_RevokedToken_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	refreshToken := "valid_but_revoked_token"
	claims := createValidTokenClaims()

	jwtService.On("ValidateRefreshToken", refreshToken).Return(claims, nil)
	tokenRepo.On("GetRefreshToken", ctx, refreshToken).Return("", domain.ErrTokenRevoked)

	// Act
	result, err := service.RefreshToken(ctx, refreshToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrTokenRevoked, err)
	jwtService.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	userRepo.AssertNotCalled(t, "GetByID")
}

// Test Logout revokes tokens successfully
func TestAuthService_Logout_ValidToken_RevokesTokensSuccessfully(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	refreshToken := "valid_refresh_token"
	claims := createValidTokenClaims()

	jwtService.On("ValidateRefreshToken", refreshToken).Return(claims, nil)
	tokenRepo.On("RevokeToken", ctx, refreshToken).Return(nil)

	// Act
	err := service.Logout(ctx, refreshToken)

	// Assert
	assert.NoError(t, err)
	jwtService.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

// Test Logout with invalid token handles gracefully
func TestAuthService_Logout_InvalidToken_HandlesGracefully(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	refreshToken := "invalid_refresh_token"

	jwtService.On("ValidateRefreshToken", refreshToken).Return(nil, domain.ErrInvalidToken)

	// Act
	err := service.Logout(ctx, refreshToken)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidToken, err)
	jwtService.AssertExpectations(t)
	tokenRepo.AssertNotCalled(t, "RevokeToken")
}

// Table-driven test for Login validation scenarios
func TestAuthService_Login_ValidationScenarios(t *testing.T) {
	tests := []struct {
		name        string
		credentials domain.Credentials
		expectError bool
		errorType   error
	}{
		{
			name: "Empty email",
			credentials: domain.Credentials{
				Email:    "",
				Password: "password123",
			},
			expectError: true,
		},
		{
			name: "Empty password",
			credentials: domain.Credentials{
				Email:    "test@example.com",
				Password: "",
			},
			expectError: true,
		},
		{
			name: "Invalid email format",
			credentials: domain.Credentials{
				Email:    "invalid-email",
				Password: "password123",
			},
			expectError: true,
		},
		{
			name: "Short password",
			credentials: domain.Credentials{
				Email:    "test@example.com",
				Password: "short",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
			service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
			ctx := context.Background()

			// Act
			result, err := service.Login(ctx, tt.credentials)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// Test Register validation scenarios
func TestAuthService_Register_ValidationScenarios(t *testing.T) {
	tests := []struct {
		name        string
		user        *domain.User
		password    string
		expectError bool
	}{
		{
			name: "Empty email",
			user: &domain.User{
				Email: "",
				Name:  "Test User",
			},
			password:    "password123",
			expectError: true,
		},
		{
			name: "Empty name",
			user: &domain.User{
				Email: "test@example.com",
				Name:  "",
			},
			password:    "password123",
			expectError: true,
		},
		{
			name: "Invalid email format",
			user: &domain.User{
				Email: "invalid-email",
				Name:  "Test User",
			},
			password:    "password123",
			expectError: true,
		},
		{
			name: "Short password",
			user: &domain.User{
				Email: "test@example.com",
				Name:  "Test User",
			},
			password:    "short",
			expectError: true,
		},
		{
			name: "Empty password",
			user: &domain.User{
				Email: "test@example.com",
				Name:  "Test User",
			},
			password:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
			service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
			ctx := context.Background()

			// For validation tests, we expect early validation failures
			// So we don't need to set up repository expectations if validation fails
			
			// Act
			result, err := service.Register(ctx, tt.user, tt.password)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				// Set up expectations for successful path (not expected in these validation tests)
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

// Test business logic edge cases
func TestAuthService_Login_DatabaseError_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	credentials := domain.Credentials{
		Email:    "test@example.com",
		Password: "password123",
	}

	userRepo.On("GetByEmail", ctx, credentials.Email).Return(nil, errors.New("database connection failed"))

	// Act
	result, err := service.Login(ctx, credentials)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get user")
	userRepo.AssertExpectations(t)
}

// Test concurrent login attempts
func TestAuthService_Login_ConcurrentAttempts_HandledCorrectly(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	credentials := domain.Credentials{
		Email:    "test@example.com",
		Password: "password123",
	}

	user := createValidUser()
	tokenPair := createValidTokenPair()

	// Setup expectations for multiple calls
	userRepo.On("GetByEmail", ctx, credentials.Email).Return(user, nil).Times(3)
	passwordService.On("CheckPassword", user.PasswordHash, credentials.Password).Return(nil).Times(3)
	jwtService.On("GenerateTokenPair", user.ID, user.Email).Return(tokenPair, nil).Times(3)
	tokenRepo.On("SaveRefreshToken", ctx, user.ID, tokenPair.RefreshToken, mock.AnythingOfType("time.Time")).Return(nil).Times(3)
	userRepo.On("UpdateLastLogin", ctx, user.ID, mock.AnythingOfType("time.Time")).Return(nil).Times(3)

	// Act - simulate concurrent requests
	results := make(chan error, 3)
	for i := 0; i < 3; i++ {
		go func() {
			_, err := service.Login(ctx, credentials)
			results <- err
		}()
	}

	// Assert
	for i := 0; i < 3; i++ {
		err := <-results
		assert.NoError(t, err)
	}

	userRepo.AssertExpectations(t)
	passwordService.AssertExpectations(t)
	jwtService.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

// Additional test coverage for edge cases

// Test Register with user creation failure
func TestAuthService_Register_UserCreationFails_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	user := &domain.User{
		Email: "newuser@example.com",
		Name:  "New User",
	}
	password := "password123"

	hashedPassword := "hashed_password"

	userRepo.On("GetByEmail", ctx, user.Email).Return(nil, domain.ErrUserNotFound)
	passwordService.On("HashPassword", password).Return(hashedPassword, nil)
	userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(errors.New("database error"))

	// Act
	result, err := service.Register(ctx, user, password)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create user")
	userRepo.AssertExpectations(t)
	passwordService.AssertExpectations(t)
	jwtService.AssertNotCalled(t, "GenerateTokenPair")
	tokenRepo.AssertNotCalled(t, "SaveRefreshToken")
}

// Test Register with token generation failure
func TestAuthService_Register_TokenGenerationFails_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	user := &domain.User{
		Email: "newuser@example.com",
		Name:  "New User",
	}
	password := "password123"

	hashedPassword := "hashed_password"

	userRepo.On("GetByEmail", ctx, user.Email).Return(nil, domain.ErrUserNotFound)
	passwordService.On("HashPassword", password).Return(hashedPassword, nil)
	userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil).Run(func(args mock.Arguments) {
		user := args.Get(1).(*domain.User)
		user.ID = "2" // Simulate ID assignment
	})
	jwtService.On("GenerateTokenPair", "2", user.Email).Return(nil, errors.New("jwt generation failed"))

	// Act
	result, err := service.Register(ctx, user, password)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to generate tokens")
	userRepo.AssertExpectations(t)
	passwordService.AssertExpectations(t)
	jwtService.AssertExpectations(t)
	tokenRepo.AssertNotCalled(t, "SaveRefreshToken")
}

// Test Register with save refresh token failure
func TestAuthService_Register_SaveRefreshTokenFails_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	user := &domain.User{
		Email: "newuser@example.com",
		Name:  "New User",
	}
	password := "password123"

	hashedPassword := "hashed_password"
	tokenPair := createValidTokenPair()

	userRepo.On("GetByEmail", ctx, user.Email).Return(nil, domain.ErrUserNotFound)
	passwordService.On("HashPassword", password).Return(hashedPassword, nil)
	userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil).Run(func(args mock.Arguments) {
		user := args.Get(1).(*domain.User)
		user.ID = "2" // Simulate ID assignment
	})
	jwtService.On("GenerateTokenPair", "2", user.Email).Return(tokenPair, nil)
	tokenRepo.On("SaveRefreshToken", ctx, "2", tokenPair.RefreshToken, mock.AnythingOfType("time.Time")).Return(errors.New("save token failed"))

	// Act
	result, err := service.Register(ctx, user, password)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save refresh token")
	userRepo.AssertExpectations(t)
	passwordService.AssertExpectations(t)
	jwtService.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

// Test Login with JWT generation failure
func TestAuthService_Login_JWTGenerationFails_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	credentials := domain.Credentials{
		Email:    "test@example.com",
		Password: "password123",
	}

	user := createValidUser()

	userRepo.On("GetByEmail", ctx, credentials.Email).Return(user, nil)
	passwordService.On("CheckPassword", user.PasswordHash, credentials.Password).Return(nil)
	jwtService.On("GenerateTokenPair", user.ID, user.Email).Return(nil, errors.New("jwt generation failed"))

	// Act
	result, err := service.Login(ctx, credentials)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to generate tokens")
	userRepo.AssertExpectations(t)
	passwordService.AssertExpectations(t)
	jwtService.AssertExpectations(t)
	tokenRepo.AssertNotCalled(t, "SaveRefreshToken")
}

// Test Login with save refresh token failure
func TestAuthService_Login_SaveRefreshTokenFails_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	credentials := domain.Credentials{
		Email:    "test@example.com",
		Password: "password123",
	}

	user := createValidUser()
	tokenPair := createValidTokenPair()

	userRepo.On("GetByEmail", ctx, credentials.Email).Return(user, nil)
	passwordService.On("CheckPassword", user.PasswordHash, credentials.Password).Return(nil)
	jwtService.On("GenerateTokenPair", user.ID, user.Email).Return(tokenPair, nil)
	tokenRepo.On("SaveRefreshToken", ctx, user.ID, tokenPair.RefreshToken, mock.AnythingOfType("time.Time")).Return(errors.New("save failed"))

	// Act
	result, err := service.Login(ctx, credentials)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save refresh token")
	userRepo.AssertExpectations(t)
	passwordService.AssertExpectations(t)
	jwtService.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

// Test RefreshToken with user not found
func TestAuthService_RefreshToken_UserNotFound_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	refreshToken := "valid_refresh_token"
	claims := createValidTokenClaims()

	jwtService.On("ValidateRefreshToken", refreshToken).Return(claims, nil)
	tokenRepo.On("GetRefreshToken", ctx, refreshToken).Return(claims.UserID, nil)
	userRepo.On("GetByID", ctx, claims.UserID).Return(nil, domain.ErrUserNotFound)

	// Act
	result, err := service.RefreshToken(ctx, refreshToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrInvalidToken, err)
	jwtService.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

// Test RefreshToken with database error on user lookup
func TestAuthService_RefreshToken_UserLookupFails_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	refreshToken := "valid_refresh_token"
	claims := createValidTokenClaims()

	jwtService.On("ValidateRefreshToken", refreshToken).Return(claims, nil)
	tokenRepo.On("GetRefreshToken", ctx, refreshToken).Return(claims.UserID, nil)
	userRepo.On("GetByID", ctx, claims.UserID).Return(nil, errors.New("database error"))

	// Act
	result, err := service.RefreshToken(ctx, refreshToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get user")
	jwtService.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

// Test RefreshToken with inactive user
func TestAuthService_RefreshToken_InactiveUser_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	refreshToken := "valid_refresh_token"
	claims := createValidTokenClaims()
	user := createValidUser()
	user.IsActive = false

	jwtService.On("ValidateRefreshToken", refreshToken).Return(claims, nil)
	tokenRepo.On("GetRefreshToken", ctx, refreshToken).Return(claims.UserID, nil)
	userRepo.On("GetByID", ctx, claims.UserID).Return(user, nil)

	// Act
	result, err := service.RefreshToken(ctx, refreshToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrAccountInactive, err)
	jwtService.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

// Test RefreshToken with revoke old token failure
func TestAuthService_RefreshToken_RevokeOldTokenFails_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	refreshToken := "valid_refresh_token"
	claims := createValidTokenClaims()
	user := createValidUser()

	jwtService.On("ValidateRefreshToken", refreshToken).Return(claims, nil)
	tokenRepo.On("GetRefreshToken", ctx, refreshToken).Return(claims.UserID, nil)
	userRepo.On("GetByID", ctx, claims.UserID).Return(user, nil)
	tokenRepo.On("RevokeToken", ctx, refreshToken).Return(errors.New("revoke failed"))

	// Act
	result, err := service.RefreshToken(ctx, refreshToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to revoke old token")
	jwtService.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

// Test RefreshToken with new token generation failure
func TestAuthService_RefreshToken_NewTokenGenerationFails_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	refreshToken := "valid_refresh_token"
	claims := createValidTokenClaims()
	user := createValidUser()

	jwtService.On("ValidateRefreshToken", refreshToken).Return(claims, nil)
	tokenRepo.On("GetRefreshToken", ctx, refreshToken).Return(claims.UserID, nil)
	userRepo.On("GetByID", ctx, claims.UserID).Return(user, nil)
	tokenRepo.On("RevokeToken", ctx, refreshToken).Return(nil)
	jwtService.On("GenerateTokenPair", user.ID, user.Email).Return(nil, errors.New("token generation failed"))

	// Act
	result, err := service.RefreshToken(ctx, refreshToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to generate new tokens")
	jwtService.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

// Test RefreshToken with save new token failure
func TestAuthService_RefreshToken_SaveNewTokenFails_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	refreshToken := "valid_refresh_token"
	claims := createValidTokenClaims()
	user := createValidUser()
	newTokenPair := createValidTokenPair()

	jwtService.On("ValidateRefreshToken", refreshToken).Return(claims, nil)
	tokenRepo.On("GetRefreshToken", ctx, refreshToken).Return(claims.UserID, nil)
	userRepo.On("GetByID", ctx, claims.UserID).Return(user, nil)
	tokenRepo.On("RevokeToken", ctx, refreshToken).Return(nil)
	jwtService.On("GenerateTokenPair", user.ID, user.Email).Return(newTokenPair, nil)
	tokenRepo.On("SaveRefreshToken", ctx, user.ID, newTokenPair.RefreshToken, mock.AnythingOfType("time.Time")).Return(errors.New("save failed"))

	// Act
	result, err := service.RefreshToken(ctx, refreshToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save new refresh token")
	jwtService.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

// Test Logout with revoke token failure
func TestAuthService_Logout_RevokeTokenFails_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	refreshToken := "valid_refresh_token"
	claims := createValidTokenClaims()

	jwtService.On("ValidateRefreshToken", refreshToken).Return(claims, nil)
	tokenRepo.On("RevokeToken", ctx, refreshToken).Return(errors.New("revoke failed"))

	// Act
	err := service.Logout(ctx, refreshToken)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to revoke token")
	jwtService.AssertExpectations(t)
	tokenRepo.AssertExpectations(t)
}

// Test Register with nil user
func TestAuthService_Register_NilUser_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	// Act
	result, err := service.Register(ctx, nil, "password123")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "user cannot be nil")
}

// Test RefreshToken with empty token
func TestAuthService_RefreshToken_EmptyToken_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	// Act
	result, err := service.RefreshToken(ctx, "")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain.ErrInvalidToken, err)
}

// Test Logout with empty token
func TestAuthService_Logout_EmptyToken_ReturnsError(t *testing.T) {
	// Arrange
	userRepo, tokenRepo, passwordService, jwtService := setupAuthServiceMocks()
	service := NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	ctx := context.Background()

	// Act
	err := service.Logout(ctx, "")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, domain.ErrInvalidToken, err)
}