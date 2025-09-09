package domain

import "errors"

// User-related errors
var (
	// ErrUserNotFound is returned when a user cannot be found
	ErrUserNotFound = errors.New("user not found")

	// ErrUserAlreadyExists is returned when trying to create a user that already exists
	ErrUserAlreadyExists = errors.New("user already exists")

	// ErrInvalidUserData is returned when user data validation fails
	ErrInvalidUserData = errors.New("invalid user data")
)

// Token-related errors
var (
	// ErrTokenNotFound is returned when a token cannot be found
	ErrTokenNotFound = errors.New("token not found")

	// ErrTokenExpired is returned when a token has expired
	ErrTokenExpired = errors.New("token has expired")

	// ErrTokenRevoked is returned when a token has been revoked
	ErrTokenRevoked = errors.New("token has been revoked")

	// ErrInvalidToken is returned when a token is malformed or invalid
	ErrInvalidToken = errors.New("invalid token")
)

// Authentication-related errors
var (
	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")

	// ErrAccountInactive is returned when trying to authenticate with an inactive account
	ErrAccountInactive = errors.New("account is inactive")
)

// Finance-related errors
var (
	// ErrFinanceSummaryNotFound is returned when a finance summary cannot be found
	ErrFinanceSummaryNotFound = errors.New("finance summary not found")

	// ErrIncomeNotFound is returned when an income record cannot be found
	ErrIncomeNotFound = errors.New("income not found")

	// ErrExpenseNotFound is returned when an expense record cannot be found
	ErrExpenseNotFound = errors.New("expense not found")

	// ErrLoanNotFound is returned when a loan record cannot be found
	ErrLoanNotFound = errors.New("loan not found")

	// ErrInvalidFinanceData is returned when finance data validation fails
	ErrInvalidFinanceData = errors.New("invalid finance data")

	// ErrUnauthorizedAccess is returned when user tries to access data they don't own
	ErrUnauthorizedAccess = errors.New("unauthorized access")
)