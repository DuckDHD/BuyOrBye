package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to connect to test database")

	// Auto-migrate the tables
	err = db.AutoMigrate(&models.UserModel{}, &models.RefreshTokenModel{})
	require.NoError(t, err, "Failed to migrate test database")

	return db
}

// createTestUser returns a valid test user
func createTestUser() *domain.User {
	return &domain.User{
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "hashed_password_123",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func TestUserRepository_Create_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()
	
	user := createTestUser()

	// Act
	err := repo.Create(ctx, user)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, user.ID, "User ID should be set after creation")

	// Verify in database
	var model models.UserModel
	err = db.Where("email = ?", user.Email).First(&model).Error
	assert.NoError(t, err)
	assert.Equal(t, user.Email, model.Email)
	assert.Equal(t, user.Name, model.Name)
	assert.Equal(t, user.PasswordHash, model.PasswordHash)
	assert.Equal(t, user.IsActive, model.IsActive)
}

func TestUserRepository_Create_DuplicateEmail_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()
	
	user1 := createTestUser()
	user2 := createTestUser()
	user2.Name = "Another User" // Different name but same email

	// Act
	err1 := repo.Create(ctx, user1)
	err2 := repo.Create(ctx, user2)

	// Assert
	assert.NoError(t, err1)
	assert.Error(t, err2)
	assert.ErrorIs(t, err2, domain.ErrUserAlreadyExists)
}

func TestUserRepository_Create_InvalidUser_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()
	
	testCases := []struct {
		name     string
		user     *domain.User
		expected error
	}{
		{
			name:     "Empty email",
			user:     &domain.User{Email: "", Name: "Test", PasswordHash: "hash"},
			expected: domain.ErrInvalidUserData,
		},
		{
			name:     "Empty name",
			user:     &domain.User{Email: "test@example.com", Name: "", PasswordHash: "hash"},
			expected: domain.ErrInvalidUserData,
		},
		{
			name:     "Empty password hash",
			user:     &domain.User{Email: "test@example.com", Name: "Test", PasswordHash: ""},
			expected: domain.ErrInvalidUserData,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := repo.Create(ctx, tc.user)

			// Assert
			assert.Error(t, err)
			assert.ErrorIs(t, err, tc.expected)
		})
	}
}

func TestUserRepository_GetByEmail_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()
	
	originalUser := createTestUser()
	err := repo.Create(ctx, originalUser)
	require.NoError(t, err)

	// Act
	foundUser, err := repo.GetByEmail(ctx, originalUser.Email)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, originalUser.Email, foundUser.Email)
	assert.Equal(t, originalUser.Name, foundUser.Name)
	assert.Equal(t, originalUser.PasswordHash, foundUser.PasswordHash)
	assert.Equal(t, originalUser.IsActive, foundUser.IsActive)
	assert.Equal(t, originalUser.ID, foundUser.ID)
}

func TestUserRepository_GetByEmail_NotFound_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Act
	foundUser, err := repo.GetByEmail(ctx, "nonexistent@example.com")

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	assert.Nil(t, foundUser)
}

func TestUserRepository_GetByID_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()
	
	originalUser := createTestUser()
	err := repo.Create(ctx, originalUser)
	require.NoError(t, err)

	// Act
	foundUser, err := repo.GetByID(ctx, originalUser.ID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, originalUser.ID, foundUser.ID)
	assert.Equal(t, originalUser.Email, foundUser.Email)
	assert.Equal(t, originalUser.Name, foundUser.Name)
}

func TestUserRepository_GetByID_NotFound_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Act
	foundUser, err := repo.GetByID(ctx, "999999")

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
	assert.Nil(t, foundUser)
}

func TestUserRepository_Update_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()
	
	originalUser := createTestUser()
	err := repo.Create(ctx, originalUser)
	require.NoError(t, err)

	// Modify user data
	originalUser.Name = "Updated Name"
	originalUser.IsActive = false

	// Act
	err = repo.Update(ctx, originalUser)

	// Assert
	assert.NoError(t, err)

	// Verify changes in database
	updatedUser, err := repo.GetByID(ctx, originalUser.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedUser.Name)
	assert.False(t, updatedUser.IsActive)
}

func TestUserRepository_Update_NotFound_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()
	
	nonExistentUser := createTestUser()
	nonExistentUser.ID = "999999"

	// Act
	err := repo.Update(ctx, nonExistentUser)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestUserRepository_UpdateLastLogin_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()
	
	user := createTestUser()
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	loginTime := time.Now().Truncate(time.Second)

	// Act
	err = repo.UpdateLastLogin(ctx, user.ID, loginTime)

	// Assert
	assert.NoError(t, err)

	// Verify in database
	var model models.UserModel
	err = db.Where("id = ?", user.ID).First(&model).Error
	require.NoError(t, err)
	assert.NotNil(t, model.LastLoginAt)
	assert.Equal(t, loginTime.Unix(), model.LastLoginAt.Unix())
}

func TestUserRepository_UpdateLastLogin_NotFound_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()
	
	loginTime := time.Now()

	// Act
	err := repo.UpdateLastLogin(ctx, "999999", loginTime)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestUserRepository_Create_NilUser_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Act
	err := repo.Create(ctx, nil)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidUserData)
}

func TestUserRepository_GetByEmail_EmptyEmail_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Act
	foundUser, err := repo.GetByEmail(ctx, "")

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidUserData)
	assert.Nil(t, foundUser)
}

func TestUserRepository_GetByID_EmptyID_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Act
	foundUser, err := repo.GetByID(ctx, "")

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidUserData)
	assert.Nil(t, foundUser)
}

func TestUserRepository_Update_NilUser_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Act
	err := repo.Update(ctx, nil)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidUserData)
}

func TestUserRepository_Update_EmptyID_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()
	
	user := createTestUser()
	user.ID = ""

	// Act
	err := repo.Update(ctx, user)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidUserData)
}

func TestUserRepository_UpdateLastLogin_EmptyUserID_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()
	
	loginTime := time.Now()

	// Act
	err := repo.UpdateLastLogin(ctx, "", loginTime)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidUserData)
}

func TestUserRepository_Update_DuplicateEmail_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()
	
	// Create two users
	user1 := createTestUser()
	user2 := createTestUser()
	user2.Email = "user2@example.com"
	
	err := repo.Create(ctx, user1)
	require.NoError(t, err)
	err = repo.Create(ctx, user2)
	require.NoError(t, err)

	// Try to update user2 with user1's email
	user2.Email = user1.Email

	// Act
	err = repo.Update(ctx, user2)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserAlreadyExists)
}