# BuyOrBye Project

## Project Description
Go-based web application using layered architecture with strict separation of concerns. Purchase decision platform with GORM persistence layer and Templ templating.

See @go.mod for exact dependency versions and @README.md for project overview

## Tech Stack
- Language: Go 1.24.4
- Web Framework: Gin v1.10.1
- ORM: GORM v1.30.4
- Database: MySQL v1.9.3 / SQLite v1.6.0 (development)
- Templating: Templ v0.3.943
- Testing: Go testing package + testify v1.10.0 + testcontainers v0.38.0
- Validation: go-playground/validator v10.27.0
- JWT: golang-jwt/jwt/v5 v5.3.0
- Build: Go modules

## Environment Setup
- Go 1.24.4+ required
- MySQL for production, SQLite for development/testing
- Use `go mod download` for dependencies
- Environment variables via .env file with godotenv v1.5.1
- Docker required for testcontainers integration tests
- Use `go install github.com/a-h/templ/cmd/templ@latest` for template generation

## Commands
- `go run cmd/app/main.go` - Start development server
- `templ generate` - Generate Go code from templ templates
- `go test ./...` - Run all tests
- `go test -v ./internal/services/...` - Test services only
- `go test -v ./internal/repositories/...` - Test repositories only
- `go test -v ./internal/handlers/...` - Test handlers only
- `go test -race ./...` - Run tests with race detection
- `go test -cover ./...` - Run tests with coverage report
- `go mod tidy` - Clean dependencies
- `go build -o bin/app cmd/app/main.go` - Build binary
- `go vet ./...` - Run Go vet for static analysis

## Architecture Rules - STRICT ENFORCEMENT

### Layer Responsibilities
- **Handlers:** HTTP transport only, use DTOs exclusively, render Templ templates
- **Services:** Business logic only, use domain structs exclusively  
- **Repositories:** Persistence with GORM only, use model structs exclusively
- **Templates:** Templ components for HTML rendering only
- **No cross-layer violations allowed**

### Data Flow (MANDATORY)
```
Middleware → Validation → Handler → Service → Repository → Database
DTO → Domain → Model
Templ Templates ← Handler (DTOs only)
```

### GORM Usage Rules
- GORM code ONLY in repository layer
- Never use GORM in services or handlers
- All database operations through repository interfaces
- Use GORM v1.30.4 features appropriately

### Templ Usage Rules
- Templ templates ONLY render DTOs from handlers
- Never pass domain or model structs to templates
- Generate templates with `templ generate` before building

## Project Structure
- `cmd/app/main.go`: Application entry point with Gin router setup
- `internal/database/`: GORM connection, migrations, and database config only
- `internal/models/`: GORM model structs (repository layer only) - DB schema
- `internal/domain/`: Business entities (service layer only) - Pure business logic
- `internal/repositories/`: GORM implementations only - Data persistence
- `internal/services/`: Business logic only - Core application logic
- `internal/handlers/`: HTTP transport using DTOs only - Web layer
- `internal/types/dto.go`: Request/response DTOs for handlers and templates
- `internal/middleware/`: Cross-cutting concerns (CORS, JWT, validation)
- `templates/`: Templ template files (.templ) - HTML generation
- `tests/`: All test files organized by type
- `tests/integration/`: Cross-layer integration tests
- `tests/testutils/`: Test helpers, mocks, and test containers setup

## Code Conventions

### File Naming
- `*_test.go` for test files
- `*_repo.go` for repository implementations
- `*_service.go` for service implementations
- `*_handler.go` for HTTP handlers
- `*.templ` for template files
- `*_templ.go` for generated template code (auto-generated)

### Interface Definitions
- Define interfaces in consumer layer (services define repo interfaces)
- Use descriptive interface names: `PurchaseRepository`, `UserService`
- Keep interfaces small and focused
- Place interfaces in same package as consumer

### Error Handling
- Always return errors as last parameter
- Use wrapped errors with context: `fmt.Errorf("failed to save purchase: %w", err)`
- Handle errors at every layer boundary
- Use appropriate HTTP status codes in handlers
- Log errors appropriately with context

### Testing Patterns
- Use table-driven tests for multiple scenarios
- Mock interfaces using testify/mock v1.10.0
- Test files adjacent to source files
- Separate unit tests from integration tests
- Use testcontainers for database integration tests

### Import Organization
- Standard library imports first
- Third-party imports second (gin, gorm, testify, etc.)
- Local project imports last
- Group imports with blank lines

## TDD Requirements

### Test Structure
- **Unit Tests:** Test individual functions/methods in isolation with mocks
- **Integration Tests:** Test layer interactions (service + repository) with real DB
- **Handler Tests:** Test HTTP endpoints with mocked services using gin/httptest
- **Repository Tests:** Test against real database using testcontainers/mysql
- **Template Tests:** Test templ template rendering with sample DTOs

### Test Naming Convention
```go
func TestServiceMethodName_Scenario_ExpectedResult(t *testing.T)
func TestPurchaseService_SavePurchase_ReturnsErrorWhenInvalidPrice(t *testing.T)
func TestPurchaseRepository_Save_WithValidData_Success(t *testing.T)
func TestPurchaseHandler_CreatePurchase_WithMissingFields_ReturnsBadRequest(t *testing.T)
```

### Required Test Coverage
- Services: >90% coverage (business logic critical)
- Repositories: >85% coverage  
- Handlers: >80% coverage
- Domain logic: >95% coverage
- Run `go test -cover ./...` to check coverage

### Test Organization
```
internal/
  services/
    purchase_service.go
    purchase_service_test.go    # Unit tests with mocks
  repositories/  
    purchase_repo.go
    purchase_repo_test.go       # Integration tests with testcontainers
  handlers/
    purchase_handler.go
    purchase_handler_test.go    # HTTP tests with gin/httptest
tests/
  integration/                  # Cross-layer integration tests
  testutils/                    # Test helpers, mocks, and testcontainer setup
```

## Layer Implementation Rules

### Handlers Layer
- Parse request DTOs only using gin.Context.ShouldBindJSON()
- Validate input using go-playground/validator struct tags
- Call service methods with domain structs converted from DTOs
- Convert service responses to DTOs for JSON output or Templ rendering
- Never access repositories directly
- Use appropriate HTTP status codes (gin.Context.JSON())
- Handle CORS using gin-contrib/cors middleware

### Services Layer  
- Accept domain structs as parameters
- Return domain structs or primitive types
- Contain all business logic and validation rules
- Use repository interfaces only - never concrete implementations
- Never import GORM, models, handlers, or DTOs packages
- Handle business-level error scenarios

### Repository Layer
- Implement service-defined interfaces
- Convert between domain structs and GORM model structs
- Use GORM v1.30.4 for all database operations
- Return domain structs to services (convert from models)
- Handle database-specific errors appropriately
- Use proper GORM features: transactions, preloading, etc.

### Template Layer (Templ)
- Render DTOs only - never domain or model structs
- Use templ v0.3.943 syntax and features
- Keep templates focused on presentation logic only
- Generate templates before building: `templ generate`

## Data Conversion Patterns

### DTO to Domain (in handlers)
```go
func (dto PurchaseRequestDTO) ToDomain() domain.PurchaseRequest {
    return domain.PurchaseRequest{
        Name:  dto.Name,
        Price: dto.Price,
        Tags:  dto.Tags,
    }
}
```

### Domain to Model (in repositories)
```go
func (d domain.PurchaseRequest) ToModel() models.PurchaseModel {
    return models.PurchaseModel{
        Name:      d.Name,
        Price:     d.Price,
        Tags:      d.Tags,
        CreatedAt: time.Now(),
    }
}
```

### Model to Domain (in repositories)
```go
func (m models.PurchaseModel) ToDomain() domain.PurchaseRequest {
    return domain.PurchaseRequest{
        ID:    m.ID,
        Name:  m.Name,
        Price: m.Price,
        Tags:  m.Tags,
    }
}
```

### Domain to DTO (in handlers)
```go
func (d domain.PurchaseRequest) ToDTO() PurchaseResponseDTO {
    return PurchaseResponseDTO{
        ID:    d.ID,
        Name:  d.Name,
        Price: d.Price,
        Tags:  d.Tags,
    }
}
```

## Testing Implementation Examples

### Service Tests (with mocks using testify/mock)
```go
func TestPurchaseService_SavePurchase_Success(t *testing.T) {
    // Arrange
    mockRepo := &mocks.PurchaseRepository{}
    service := services.NewPurchaseService(mockRepo)
    purchase := domain.PurchaseRequest{Name: "Test", Price: 100.0}
    
    mockRepo.On("Save", mock.MatchedBy(func(p domain.PurchaseRequest) bool {
        return p.Name == "Test" && p.Price == 100.0
    })).Return(nil)
    
    // Act
    err := service.SavePurchase(purchase)
    
    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

### Repository Integration Tests (with testcontainers)
```go
func TestPurchaseRepository_Save_Success(t *testing.T) {
    // Arrange
    ctx := context.Background()
    db := setupTestContainerDB(t) // Use testcontainers/mysql
    repo := repositories.NewPurchaseRepository(db)
    purchase := domain.PurchaseRequest{Name: "Test", Price: 100.0}
    
    // Act
    err := repo.Save(ctx, purchase)
    
    // Assert
    assert.NoError(t, err)
    
    // Verify in database
    var count int64
    db.Model(&models.PurchaseModel{}).Where("name = ?", "Test").Count(&count)
    assert.Equal(t, int64(1), count)
}
```

### Handler Tests (with gin/httptest)
```go
func TestPurchaseHandler_CreatePurchase_Success(t *testing.T) {
    // Arrange
    gin.SetMode(gin.TestMode)
    mockService := &mocks.PurchaseService{}
    handler := handlers.NewPurchaseHandler(mockService)
    
    router := gin.Default()
    router.POST("/purchases", handler.CreatePurchase)
    
    reqBody := `{"name":"Test","price":100.0}`
    mockService.On("SavePurchase", mock.AnythingOfType("domain.PurchaseRequest")).Return(nil)
    
    // Act
    w := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/purchases", strings.NewReader(reqBody))
    req.Header.Set("Content-Type", "application/json")
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
    mockService.AssertExpectations(t)
}
```

## Claude Instructions
- Use "think hard" when designing new layer interactions or complex business logic
- Use "ultrathink" for complex architectural decisions or when refactoring across multiple layers
- Always verify strict layer separation before suggesting any code changes
- When implementing new features, always start with tests (TDD approach)
- Consider both MySQL and SQLite compatibility when writing GORM code

## Context Management
- Clear conversation when switching between different layers (handler → service → repository)
- Start fresh conversation when implementing completely new features
- Keep conversations focused on single layer or architectural concern
- Use /clear when moving from testing to implementation or vice versa

## Do Not
- Import GORM packages in services or handlers
- Use model structs outside repository layer
- Use DTOs in services - only domain structs allowed
- Pass domain structs to Templ templates - use DTOs from `internal/dtos/` only
- Call repositories from handlers directly - must go through services
- Write business logic in handlers or repositories - belongs in services only
- Commit without running full test suite including race detection
- Skip error handling at any layer boundary
- Write code without corresponding tests (TDD required)
- Use direct SQL when GORM methods are available
- Expose internal model structures in API responses
- Forget to run `templ generate` before building
- Import concrete implementations in service constructors - use interfaces only
- Mix MySQL and SQLite specific code without compatibility considerations
- Create DTOs without proper structured comments in the required format
- Place DTOs outside the `internal/dtos/` directory
- Mix DTOs from different domains in the same file
- Skip validation struct tags on DTO fields

## Database Rules
- Migrations only in `internal/database/migrations/`
- Use GORM AutoMigrate for development, proper migrations for production
- Model structs define database schema with appropriate GORM tags
- All queries through repository methods only
- Use MySQL for production, SQLite for testing
- Leverage testcontainers for integration testing with real database
- Use connection pooling and proper GORM configuration

## Development Workflow
1. Write failing test first (TDD approach)
2. Implement minimal code to pass test
3. Refactor while keeping tests green
4. Run `templ generate` if templates were modified
5. Run full test suite with race detection before committing
6. Ensure proper layer separation is maintained
7. Verify test coverage meets requirements
8. Validate configuration loads correctly in different environments
9. Check that all logging uses structured Zap fields
10. Ensure no fmt.Print statements remain in code

## Performance Guidelines
- Use GORM preloading for related data: `db.Preload("Purchases").Find(&users)`
- Implement pagination for list endpoints using GORM's `Limit()` and `Offset()`
- Use database indexes for frequently queried fields
- Monitor and optimize N+1 query problems with GORM
- Use GORM's batch operations for bulk inserts/updates
- Leverage MySQL connection pooling effectively
- Cache frequently accessed data appropriately

## Security Rules
- Validate all input at handler layer using go-playground/validator v10.27.0
- Use parameterized queries through GORM to prevent SQL injection
- Sanitize data before database operations
- Never expose internal model structures in API responses
- Use gin-contrib/cors v1.7.6 middleware appropriately
- Implement proper JWT authentication using golang-jwt/jwt/v5 v5.3.0
- Load environment variables securely with godotenv v1.5.1
- Use golang.org/x/crypto for password hashing
- Implement proper session management and CSRF protection