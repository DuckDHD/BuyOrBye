package types

import "github.com/DuckDHD/BuyOrBye/internal/domain"

/*
Request LoginRequestDTO dto
User authentication request with email and password credentials
*/
type LoginRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// ToDomain converts LoginRequestDTO to domain.Credentials
func (dto LoginRequestDTO) ToDomain() domain.Credentials {
	return domain.Credentials{
		Email:    dto.Email,
		Password: dto.Password,
	}
}

/*
Request RegisterRequestDTO dto
New user registration request with email, name, and password
*/
type RegisterRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=1"`
	Password string `json:"password" validate:"required,min=8"`
}

// ToDomain converts RegisterRequestDTO to domain.User
func (dto RegisterRequestDTO) ToDomain() *domain.User {
	return &domain.User{
		Email: dto.Email,
		Name:  dto.Name,
		// PasswordHash and timestamps will be set by the service layer
		IsActive: true,
	}
}

/*
Request RefreshTokenRequestDTO dto
Token refresh request using valid refresh token
*/
type RefreshTokenRequestDTO struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

/*
Response TokenResponseDTO dto
Successful authentication response containing JWT token pair
*/
type TokenResponseDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// FromDomain converts domain.TokenPair to TokenResponseDTO
func (dto *TokenResponseDTO) FromDomain(tokenPair *domain.TokenPair) {
	dto.AccessToken = tokenPair.AccessToken
	dto.RefreshToken = tokenPair.RefreshToken
	dto.ExpiresIn = tokenPair.ExpiresIn
	dto.TokenType = "Bearer"
}

// FromDomainTokenPair creates TokenResponseDTO from domain.TokenPair
func FromDomainTokenPair(tokenPair *domain.TokenPair) TokenResponseDTO {
	return TokenResponseDTO{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
		TokenType:    "Bearer",
	}
}

// ToDomain converts RefreshTokenRequestDTO to domain refresh token string
func (dto RefreshTokenRequestDTO) ToDomain() string {
	return dto.RefreshToken
}

/*
Page LoginPageDTO dto
Login page data transfer object for full page rendering
*/
type LoginPageDTO struct {
	Layout LayoutDTO         `json:"layout"`
	Error  *ErrorResponseDTO `json:"error,omitempty"`
}

/*
Page RegisterPageDTO dto
Register page data transfer object for full page rendering
*/
type RegisterPageDTO struct {
	Layout LayoutDTO         `json:"layout"`
	Error  *ErrorResponseDTO `json:"error,omitempty"`
}

/*
Page ForgotPasswordPageDTO dto
Forgot password page data transfer object for full page rendering
*/
type ForgotPasswordPageDTO struct {
	Layout LayoutDTO `json:"layout"`
	Error  *ErrorResponseDTO `json:"error,omitempty"`
}

// AuthFormDTO represents authentication form data for templates
type AuthFormDTO struct {
	FormType  string             `json:"form_type"` // "login" or "register"
	Email     string             `json:"email"`
	Password  string             `json:"password"`
	Error     *ErrorResponseDTO  `json:"error,omitempty"`
	CSRFToken string             `json:"csrf_token"`
}