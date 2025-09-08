# DOMAIN_AUTHENTICATION.md

## Authentication Domain for BuyOrBye Project
*Following strict layered architecture with GORM persistence*

---

## 🏗️ Architecture Alignment

### Layer Responsibilities (STRICT)
- **Handlers:** HTTP transport only, JWT extraction, use DTOs exclusively
- **Services:** Authentication logic, password hashing, token generation, use domain structs exclusively
- **Repositories:** User persistence with GORM only, use model structs exclusively
- **No cross-layer violations allowed**

### Authentication Data Flow
```
Request → CSRF Middleware → Auth Middleware → Handler → Service → Repository → Database
         ↓                  ↓                  ↓          ↓           ↓          ↓
    [CSRF Token]      [JWT Validation]    [LoginDTO] [Credentials] [UserModel] [SQL]
```

---

## 📁 Project Structure Integration

```
internal/
  database/
    migrations/
      001_create_users_table.go
      002_create_refresh_tokens_table.go
  models/                        # GORM models (DB schema)
    user_model.go               # GORM User model
    refresh_token_model.go      # GORM RefreshToken model
  domain/                        # Business entities
    user.go                     # User domain entity
    credentials.go              # Login credentials
    token.go                    # Token domain entity
  repositories/                  # GORM implementations
    user_repo.go                # User repository with GORM
    token_repo.go               # Token repository with GORM
  services/                      # Business logic
    auth_service.go             # Authentication logic
    jwt_service.go              # JWT generation/validation
    password_service.go         # Password hashing
  handlers/                      # HTTP transport
    auth_handler.go             # Login/Register/Refresh endpoints
  types/
    dto.go                      # Existing DTOs
    auth_dto.go                 # Auth-specific DTOs
  middleware/
    auth.go                     # JWT validation middleware
    csrf.go                     # CSRF protection
    rate_limit.go               # Brute force protection
tests/
  integration/
    auth_integration_test.go
  testutils/
    auth_helpers.go
```

---

## 🔐 Security Implementation

### Password Security (Service Layer)
```go
// internal/services/password_service.go
const bcryptCost = 14 // Balanced security/performance

type PasswordService interface {
    HashPassword(password string) (string, error)
    CheckPassword(password, hash string) error
}

// Claude prompt:
"Implement password service with bcrypt cost 14 in services layer, return domain errors"
```

### JWT Configuration (Service Layer)
```go
// internal/domain/token.go
type TokenPair struct {
    AccessToken  string
    RefreshToken string
    ExpiresIn    int64
}

type TokenClaims struct {
    UserID   string
    Email    string
    ExpireAt time.Time
}

// internal/services/jwt_service.go
type JWTService interface {
    GenerateTokenPair(user *domain.User) (*domain.TokenPair, error)
    ValidateAccessToken(token string) (*domain.TokenClaims, error)
    ValidateRefreshToken(token string) (*domain.TokenClaims, error)
}

// Configuration
const (
    AccessTokenTTL  = 15 * time.Minute
    RefreshTokenTTL = 7 * 24 * time.Hour
)
```

### CSRF Protection (Middleware Layer)
```go
// internal/middleware/csrf.go
import "github.com/gorilla/csrf"

func CSRFMiddleware(secret []byte) gin.HandlerFunc {
    return func(c *gin.Context) {
        csrf.Protect(
            secret,
            csrf.Secure(true),
            csrf.HttpOnly(true),
            csrf.SameSite(csrf.SameSiteStrictMode),
        )(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            c.Request = r
            c.Next()
        })).ServeHTTP(c.Writer, c.Request)
    }
}
```

---

## 📋 Data Conversion Patterns

### Auth DTOs (Handler Layer)
```go
// internal/types/auth_dto.go
type LoginRequestDTO struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

type RegisterRequestDTO struct {
    Email           string `json:"email" binding:"required,email"`
    Password        string `json:"password" binding:"required,min=8"`
    PasswordConfirm string `json:"password_confirm" binding:"required,eqfield=Password"`
    Name            string `json:"name" binding:"required"`
}

type TokenResponseDTO struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
}

// Conversion methods
func (dto LoginRequestDTO) ToDomain() domain.Credentials {
    return domain.Credentials{
        Email:    dto.Email,
        Password: dto.Password,
    }
}
```

### Domain Entities (Service Layer)
```go
// internal/domain/user.go
type User struct {
    ID           string
    Email        string
    Name         string
    PasswordHash string // Only used internally
    IsActive     bool
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// internal/domain/credentials.go
type Credentials struct {
    Email    string
    Password string
}
```

### GORM Models (Repository Layer)
```go
// internal/models/user_model.go
type UserModel struct {
    gorm.Model
    Email        string `gorm:"uniqueIndex;not null"`
    Name         string `gorm:"not null"`
    PasswordHash string `gorm:"not null"`
    IsActive     bool   `gorm:"default:true"`
    LastLoginAt  *time.Time
}

// Conversion methods
func (m UserModel) ToDomain() domain.User {
    return domain.User{
        ID:           strconv.FormatUint(uint64(m.ID), 10),
        Email:        m.Email,
        Name:         m.Name,
        PasswordHash: m.PasswordHash,
        IsActive:     m.IsActive,
        CreatedAt:    m.CreatedAt,
        UpdatedAt:    m.UpdatedAt,
    }
}

func UserFromDomain(d domain.User) UserModel {
    id, _ := strconv.ParseUint(d.ID, 10, 32)
    return UserModel{
        Model:        gorm.Model{ID: uint(id)},
        Email:        d.Email,
        Name:         d.Name,
        PasswordHash: d.PasswordHash,
        IsActive:     d.IsActive,
    }
}
```

---

## 🧪 TDD Test Structure

### Service Tests (Unit Tests with Mocks)
```go
// internal/services/auth_service_test.go
func TestAuthService_Login_Success(t *testing.T) {
    // Arrange
    mockUserRepo := &mocks.UserRepository{}
    mockTokenRepo := &mocks.TokenRepository{}
    mockJWTService := &mocks.JWTService{}
    mockPasswordService := &mocks.PasswordService{}
    
    service := services.NewAuthService(
        mockUserRepo,
        mockTokenRepo,
        mockJWTService,
        mockPasswordService,
    )
    
    credentials := domain.Credentials{
        Email:    "test@example.com",
        Password: "password123",
    }
    
    user := &domain.User{
        ID:           "1",
        Email:        "test@example.com",
        PasswordHash: "hashed",
    }
    
    tokens := &domain.TokenPair{
        AccessToken:  "access",
        RefreshToken: "refresh",
    }
    
    mockUserRepo.On("GetByEmail", credentials.Email).Return(user, nil)
    mockPasswordService.On("CheckPassword", credentials.Password, user.PasswordHash).Return(nil)
    mockJWTService.On("GenerateTokenPair", user).Return(tokens, nil)
    mockTokenRepo.On("SaveRefreshToken", user.ID, "refresh").Return(nil)
    
    // Act
    result, err := service.Login(credentials)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, tokens, result)
    mockUserRepo.AssertExpectations(t)
    mockPasswordService.AssertExpectations(t)
    mockJWTService.AssertExpectations(t)
    mockTokenRepo.AssertExpectations(t)
}
```

### Repository Integration Tests
```go
// internal/repositories/user_repo_test.go
func TestUserRepository_Create_Success(t *testing.T) {
    // Arrange
    db := setupTestDB(t) // Use test container or in-memory SQLite
    repo := repositories.NewUserRepository(db)
    
    user := domain.User{
        Email:        "test@example.com",
        Name:         "Test User",
        PasswordHash: "hashed_password",
        IsActive:     true,
    }
    
    // Act
    err := repo.Create(&user)
    
    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
    
    // Verify in database
    var model models.UserModel
    err = db.Where("email = ?", user.Email).First(&model).Error
    assert.NoError(t, err)
    assert.Equal(t, user.Email, model.Email)
}
```

### Handler Tests
```go
// internal/handlers/auth_handler_test.go
func TestAuthHandler_Login_Success(t *testing.T) {
    // Arrange
    mockAuthService := &mocks.AuthService{}
    handler := handlers.NewAuthHandler(mockAuthService)
    
    router := gin.New()
    router.POST("/login", handler.Login)
    
    loginDTO := types.LoginRequestDTO{
        Email:    "test@example.com",
        Password: "password123",
    }
    
    tokens := &domain.TokenPair{
        AccessToken:  "access_token",
        RefreshToken: "refresh_token",
        ExpiresIn:    900,
    }
    
    mockAuthService.On("Login", loginDTO.ToDomain()).Return(tokens, nil)
    
    body, _ := json.Marshal(loginDTO)
    req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()
    
    // Act
    router.ServeHTTP(rec, req)
    
    // Assert
    assert.Equal(t, http.StatusOK, rec.Code)
    
    var response types.TokenResponseDTO
    json.Unmarshal(rec.Body.Bytes(), &response)
    assert.Equal(t, tokens.AccessToken, response.AccessToken)
    mockAuthService.AssertExpectations(t)
}
```

---

## 🚀 Claude Code Commands

### Slash Commands (.claude/commands/)

**`/auth-init`**
```markdown
Initialize authentication following BuyOrBye layered architecture:
1. Create models: UserModel, RefreshTokenModel (GORM structs)
2. Create domain: User, Credentials, TokenPair
3. Create DTOs: LoginRequestDTO, RegisterRequestDTO, TokenResponseDTO
4. Create repository interfaces in services
5. Implement repositories with GORM
6. Create service interfaces and implementations
7. Create handlers with Gin/Echo
8. Add middleware: auth, csrf, rate_limit
9. Generate migrations
```

**`/auth-test-tdd`**
```markdown
Generate TDD tests following BuyOrBye testing patterns:
1. Service unit tests with mocks (90% coverage target)
2. Repository integration tests with test DB (85% coverage)
3. Handler HTTP tests (80% coverage)
4. Use table-driven tests
5. Follow naming: TestServiceMethodName_Scenario_ExpectedResult
```

**`/auth-security-audit`**
```markdown
Audit authentication security:
1. Verify bcrypt cost = 14
2. Check CSRF on state-changing endpoints
3. Validate JWT secret strength (32+ bytes from env)
4. Verify parameterized queries in repositories
5. Check rate limiting on login endpoint
6. Ensure DTOs don't expose internal structures
7. Validate go-playground/validator on all inputs
```

---

## 🔄 Development Workflow

```bash
# Step 1: Write failing test
claude "Write test for user login in auth_service_test.go"

# Step 2: Implement repository interface in service
claude "Define UserRepository interface in auth_service.go"

# Step 3: Implement GORM repository
claude "Implement UserRepository with GORM in user_repo.go"

# Step 4: Implement service logic
claude "Implement Login method in AuthService"

# Step 5: Create handler
claude "Create Login handler using LoginRequestDTO"

# Step 6: Add middleware
claude "Add JWT validation middleware"

# Step 7: Run tests
go test -v ./internal/services/...
go test -v ./internal/repositories/...
go test -v ./internal/handlers/...

# Step 8: Check coverage
go test -cover ./...
```

---

## ✅ Implementation Checklist

- [ ] Models created (UserModel, RefreshTokenModel) with GORM tags
- [ ] Domain entities created (User, Credentials, TokenPair)
- [ ] DTOs created with validation tags
- [ ] Repository interfaces defined in services
- [ ] GORM repositories implemented
- [ ] Services implemented with business logic
- [ ] Handlers created with DTO conversions
- [ ] Middleware added (auth, csrf, rate_limit)
- [ ] Migrations created in database/migrations
- [ ] Unit tests written (services >90% coverage)
- [ ] Integration tests written (repositories >85%)
- [ ] Handler tests written (>80% coverage)
- [ ] Security audit passed
- [ ] Environment variables configured (JWT_SECRET, etc.)
- [ ] No GORM imports in services or handlers
- [ ] No business logic in handlers or repositories

---

## 🔍 Common Authentication Pitfalls

| Issue | Solution | Verification |
|-------|----------|--------------|
| GORM in services | Move to repository | `grep -r "gorm" internal/services/` should be empty |
| DTOs in services | Use domain structs | Services should only import domain package |
| Business logic in handlers | Move to services | Handlers should only validate and convert |
| Missing password hashing | Use bcrypt cost 14 | Check password_service.go |
| JWT in cookies | Use Authorization header | Prevents CSRF on tokens |
| No rate limiting | Add middleware | Test with multiple failed attempts |

---

## 📚 References

- [BuyOrBye Architecture](./CLAUDE.md) - Project structure and rules
- [Gorilla CSRF](https://github.com/gorilla/csrf) - CSRF middleware
- [JWT-Go](https://github.com/golang-jwt/jwt) - JWT implementation
- [Bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) - Password hashing
- [Go Validator](https://github.com/go-playground/validator) - Struct validation

---

*This authentication domain follows BuyOrBye's strict layered architecture with complete separation of concerns.*