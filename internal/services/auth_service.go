package services

import (
	"context"
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// AuthService defines the interface for authentication operations
// This interface contains all business logic for user authentication
type AuthService interface {
	// Login authenticates a user with email and password
	// Returns a token pair on successful authentication
	// Returns domain.ErrInvalidCredentials if credentials are invalid
	// Returns domain.ErrAccountInactive if account is inactive
	Login(ctx context.Context, credentials domain.Credentials) (*domain.TokenPair, error)

	// Register creates a new user account and returns authentication tokens
	// Returns domain.ErrUserAlreadyExists if user already exists
	// Returns domain.ErrInvalidUserData if user data validation fails
	Register(ctx context.Context, user *domain.User, password string) (*domain.TokenPair, error)

	// RefreshToken generates a new token pair using a valid refresh token
	// Returns domain.ErrInvalidToken if token is invalid
	// Returns domain.ErrTokenExpired if token has expired
	// Returns domain.ErrTokenRevoked if token has been revoked
	RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error)

	// Logout revokes a user's refresh token
	// Returns domain.ErrInvalidToken if token is invalid
	Logout(ctx context.Context, refreshToken string) error
}

// authService implements AuthService with dependency injection
type authService struct {
	userRepo        UserRepository
	tokenRepo       TokenRepository
	passwordService PasswordService
	jwtService      JWTService
}

// NewAuthService creates a new authentication service instance
// Requires all dependencies to be injected following dependency inversion principle
func NewAuthService(
	userRepo UserRepository,
	tokenRepo TokenRepository,
	passwordService PasswordService,
	jwtService JWTService,
) AuthService {
	return &authService{
		userRepo:        userRepo,
		tokenRepo:       tokenRepo,
		passwordService: passwordService,
		jwtService:      jwtService,
	}
}

// Login authenticates a user with email and password
func (a *authService) Login(ctx context.Context, credentials domain.Credentials) (*domain.TokenPair, error) {
	// Validate credentials
	if err := credentials.Validate(); err != nil {
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	// Get user by email
	user, err := a.userRepo.GetByEmail(ctx, credentials.Email)
	if err != nil {
		if err == domain.ErrUserNotFound {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if account is active
	if !user.IsActive {
		return nil, domain.ErrAccountInactive
	}

	// Verify password
	if err := a.passwordService.CheckPassword(user.PasswordHash, credentials.Password); err != nil {
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
		// In production, you would use a proper logger here
		_ = err // Ignore error for now
	}

	return tokenPair, nil
}

// Register creates a new user account and returns authentication tokens
func (a *authService) Register(ctx context.Context, user *domain.User, password string) (*domain.TokenPair, error) {
	// Validate input parameters
	if user == nil {
		return nil, fmt.Errorf("user cannot be nil")
	}

	// Validate name
	if user.Name == "" {
		return nil, fmt.Errorf("invalid registration data: name is required")
	}

	// Validate credentials format
	credentials := domain.Credentials{
		Email:    user.Email,
		Password: password,
	}
	if err := credentials.Validate(); err != nil {
		return nil, fmt.Errorf("invalid registration data: %w", err)
	}

	// Check if user already exists
	existingUser, err := a.userRepo.GetByEmail(ctx, user.Email)
	if err != nil && err != domain.ErrUserNotFound {
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
		if err == domain.ErrUserNotFound {
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