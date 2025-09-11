package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/logging"
)

// authService implements the AuthService interface defined in handlers package
// Following the consumer-defined interface principle

// authService implements AuthService with dependency injection
type authService struct {
	userRepo        UserRepository
	tokenRepo       TokenRepository
	passwordService PasswordService
	jwtService      JWTService
}

// NewAuthService creates a new authentication service instance
// Returns concrete type that implements AuthService interface defined in handlers package
func NewAuthService(
	userRepo UserRepository,
	tokenRepo TokenRepository,
	passwordService PasswordService,
	jwtService JWTService,
) *authService {
	return &authService{
		userRepo:        userRepo,
		tokenRepo:       tokenRepo,
		passwordService: passwordService,
		jwtService:      jwtService,
	}
}

// Login authenticates a user with email and password
func (a *authService) Login(ctx context.Context, credentials domain.Credentials) (*domain.TokenPair, error) {
	logger := logging.ServiceLogger().With(logging.WithOperation("login"), logging.WithUserID(credentials.Email))
	logger.Info("Starting user authentication")

	// Validate credentials
	if err := credentials.Validate(); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Get user by email
	user, err := a.userRepo.GetByEmail(ctx, credentials.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if account is active
	if !user.IsActive {
		return nil, domain.ErrAccountInactive
	}

	// Verify password
	logger.Debug("Verifying user password")
	if err := a.passwordService.CheckPassword(user.PasswordHash, credentials.Password); err != nil {
		logger.Warn("Password verification failed")
		return nil, domain.ErrInvalidCredentials
	}

	// Generate token pair
	tokenPair, err := a.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Save refresh token
	refreshExpiry := time.Now().Add(7 * 24 * time.Hour) // 7 days
	if err := a.tokenRepo.SaveRefreshToken(ctx, user.ID, tokenPair.RefreshToken, refreshExpiry); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	// Update last login time
	if err := a.userRepo.UpdateLastLogin(ctx, user.ID, time.Now()); err != nil {
		// Log error but don't fail the login process
		logger.Warn("Failed to update last login time", logging.WithError(err))
	}

	logger.Info("User authentication successful")
	return tokenPair, nil
}

// Register creates a new user account and returns authentication tokens
func (a *authService) Register(ctx context.Context, user *domain.User, password string) (*domain.TokenPair, error) {
	// Validate input parameters
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil: %w", domain.ErrInvalidUserData)
	}

	// Validate name
	if user.Name == "" {
		return nil, domain.ErrInvalidUserData
	}

	// Validate credentials format
	credentials := domain.Credentials{
		Email:    user.Email,
		Password: password,
	}
	if err := credentials.Validate(); err != nil {
		return nil, domain.ErrInvalidUserData
	}

	// Check if user already exists
	existingUser, err := a.userRepo.GetByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := a.passwordService.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set up user with hashed password and timestamps
	now := time.Now()
	user.PasswordHash = hashedPassword
	user.IsActive = true
	user.CreatedAt = now
	user.UpdatedAt = now

	// Create user
	if err := a.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate token pair
	tokenPair, err := a.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Save refresh token
	refreshExpiry := time.Now().Add(7 * 24 * time.Hour) // 7 days
	if err := a.tokenRepo.SaveRefreshToken(ctx, user.ID, tokenPair.RefreshToken, refreshExpiry); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return tokenPair, nil
}

// RefreshToken generates a new token pair using a valid refresh token
func (a *authService) RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	if refreshToken == "" {
		return nil, domain.ErrInvalidToken
	}

	// Validate refresh token format and expiry
	_, err := a.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err // Pass through the specific error (expired, invalid, etc.)
	}

	// Check if token exists and is not revoked in database
	userID, err := a.tokenRepo.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, err // Pass through the specific error (not found, revoked, etc.)
	}

	// Verify user still exists and is active
	user, err := a.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if !user.IsActive {
		return nil, domain.ErrAccountInactive
	}

	// Revoke old refresh token
	if err := a.tokenRepo.RevokeToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to revoke old token: %w", err)
	}

	// Generate new token pair
	newTokenPair, err := a.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	// Save new refresh token
	refreshExpiry := time.Now().Add(7 * 24 * time.Hour) // 7 days
	if err := a.tokenRepo.SaveRefreshToken(ctx, user.ID, newTokenPair.RefreshToken, refreshExpiry); err != nil {
		return nil, fmt.Errorf("failed to save new refresh token: %w", err)
	}

	return newTokenPair, nil
}

// Logout revokes a user's refresh token
func (a *authService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return domain.ErrInvalidToken
	}

	// Validate refresh token format
	_, err := a.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return err // Pass through the specific error (expired, invalid, etc.)
	}

	// Revoke the token
	if err := a.tokenRepo.RevokeToken(ctx, refreshToken); err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}
