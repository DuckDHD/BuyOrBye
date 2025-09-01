package services

import (
	"context"
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserService struct {
	db *pgxpool.Pool
}

func NewUserService(db *pgxpool.Pool) *UserService {
	return &UserService{db: db}
}

func (s *UserService) CreateUser(ctx context.Context, email, locale, currency string) (*models.User, error) {
	user := &models.User{
		ID:        uuid.New(),
		Email:     email,
		CreatedAt: time.Now(),
		Locale:    locale,
		Currency:  currency,
		IsPro:     false,
	}

	query := `
		INSERT INTO users (id, email, created_at, locale, currency, is_pro)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := s.db.Exec(ctx, query,
		user.ID, user.Email, user.CreatedAt,
		user.Locale, user.Currency, user.IsPro,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, created_at, locale, currency, is_pro, last_login_at
		FROM users WHERE id = $1
	`

	err := s.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.CreatedAt,
		&user.Locale, &user.Currency, &user.IsPro,
		&user.LastLoginAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT id, email, created_at, locale, currency, is_pro, last_login_at
		FROM users WHERE email = $1
	`

	err := s.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.CreatedAt,
		&user.Locale, &user.Currency, &user.IsPro,
		&user.LastLoginAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (s *UserService) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET last_login_at = NOW() WHERE id = $1`
	_, err := s.db.Exec(ctx, query, userID)
	return err
}
