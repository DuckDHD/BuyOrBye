package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/DuckDHD/BuyOrBye/internal/services"
	"gorm.io/gorm"
)

// userRepository implements the UserRepository interface using GORM
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *gorm.DB) services.UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create saves a new user to the database
// Returns an error if the user already exists or if there's a database error
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil: %w", domain.ErrInvalidUserData)
	}

	// Validate domain user before creating
	if err := user.Validate(); err != nil {
		return fmt.Errorf("user validation failed: %w", domain.ErrInvalidUserData)
	}

	// Convert domain user to GORM model
	userModel := models.UserFromDomain(*user)
	
	// Ensure timestamps are set
	now := time.Now()
	if userModel.CreatedAt.IsZero() {
		userModel.CreatedAt = now
	}
	if userModel.UpdatedAt.IsZero() {
		userModel.UpdatedAt = now
	}

	// Create user in database
	if err := r.db.WithContext(ctx).Create(&userModel).Error; err != nil {
		// Check for unique constraint violation (duplicate email)
		if isDuplicateKeyError(err) {
			return fmt.Errorf("user with email %s already exists: %w", user.Email, domain.ErrUserAlreadyExists)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Convert back to domain and update the original user with generated ID
	domainUser := userModel.ToDomain()
	*user = domainUser

	return nil
}

// GetByEmail retrieves a user by their email address
// Returns domain.ErrUserNotFound if the user doesn't exist
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty: %w", domain.ErrInvalidUserData)
	}

	var userModel models.UserModel
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&userModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with email %s not found: %w", email, domain.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	domainUser := userModel.ToDomain()
	return &domainUser, nil
}

// GetByID retrieves a user by their ID
// Returns domain.ErrUserNotFound if the user doesn't exist
func (r *userRepository) GetByID(ctx context.Context, userID string) (*domain.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty: %w", domain.ErrInvalidUserData)
	}

	var userModel models.UserModel
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&userModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with ID %s not found: %w", userID, domain.ErrUserNotFound)
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	domainUser := userModel.ToDomain()
	return &domainUser, nil
}

// Update modifies an existing user's data
// Returns domain.ErrUserNotFound if the user doesn't exist
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil: %w", domain.ErrInvalidUserData)
	}

	if user.ID == "" {
		return fmt.Errorf("user ID cannot be empty: %w", domain.ErrInvalidUserData)
	}

	// Validate domain user before updating
	if err := user.Validate(); err != nil {
		return fmt.Errorf("user validation failed: %w", domain.ErrInvalidUserData)
	}

	// Convert domain user to GORM model
	userModel := models.UserFromDomain(*user)
	
	// Update timestamp
	userModel.UpdatedAt = time.Now()

	// Update user in database
	// Use Updates with specific fields to avoid zero value issues
	updates := map[string]interface{}{
		"email":         userModel.Email,
		"name":          userModel.Name,
		"password_hash": userModel.PasswordHash,
		"is_active":     userModel.IsActive,
		"updated_at":    userModel.UpdatedAt,
	}
	result := r.db.WithContext(ctx).Model(&userModel).Where("id = ?", user.ID).Updates(updates)
	if result.Error != nil {
		// Check for unique constraint violation (duplicate email)
		if isDuplicateKeyError(result.Error) {
			return fmt.Errorf("user with email %s already exists: %w", user.Email, domain.ErrUserAlreadyExists)
		}
		return fmt.Errorf("failed to update user: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found: %w", user.ID, domain.ErrUserNotFound)
	}

	// Update the user with the new timestamp
	user.UpdatedAt = userModel.UpdatedAt

	return nil
}

// UpdateLastLogin updates the last login timestamp for a user
// Returns domain.ErrUserNotFound if the user doesn't exist
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID string, loginTime time.Time) error {
	if userID == "" {
		return fmt.Errorf("userID cannot be empty: %w", domain.ErrInvalidUserData)
	}

	result := r.db.WithContext(ctx).Model(&models.UserModel{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at": loginTime,
			"updated_at":    time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update last login: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found: %w", userID, domain.ErrUserNotFound)
	}

	return nil
}

// isDuplicateKeyError checks if the error is a duplicate key constraint violation
// This helper function checks for common database-specific error patterns
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	// Check for common duplicate key error patterns
	return contains(errMsg, "duplicate key") ||
		contains(errMsg, "UNIQUE constraint") ||
		contains(errMsg, "Duplicate entry") ||
		contains(errMsg, "duplicate value")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(str, substr string) bool {
	return len(str) >= len(substr) && 
		(str == substr || 
		 len(str) > len(substr) && 
		 (containsAt(str, substr, 0) || contains(str[1:], substr)))
}

// containsAt checks if str contains substr starting at the given position
func containsAt(str, substr string, pos int) bool {
	if pos < 0 || pos > len(str)-len(substr) {
		return false
	}
	for i := 0; i < len(substr); i++ {
		if toLower(str[pos+i]) != toLower(substr[i]) {
			return false
		}
	}
	return true
}

// toLower converts a byte to lowercase
func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}