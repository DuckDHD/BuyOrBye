# BuyOrBye Project

## Project Description
Go-based web application using layered architecture with strict separation of concerns. Purchase decision platform with GORM persistence layer.

## Tech Stack
- Language: Go 1.24+
- Web Framework: Gin/Echo (HTTP handlers)
- ORM: GORM v2
- Database: PostgreSQL/MySQL
- Testing: Go testing package + testify
- Build: Go modules

## Architecture Rules - STRICT ENFORCEMENT

### Layer Responsibilities
- **Handlers:** HTTP transport only, use DTOs exclusively
- **Services:** Business logic only, use domain structs exclusively  
- **Repositories:** Persistence with GORM only, use model structs exclusively
- **No cross-layer violations allowed**

### Data Flow (MANDATORY)
```
Middleware → Validation → Handler → Service → Repository → Database
DTO → Domain → Model
```

### GORM Usage Rules
- GORM code ONLY in repository layer
- Never use GORM in services or handlers
- All database operations through repository interfaces

## Project Structure
```
cmd/app/main.go
internal/
  database/          # GORM connection & migrations
  models/            # GORM model structs (DB schema)
  domain/            # Business entities (service layer)
  repositories/      # GORM implementations
  services/          # Business logic
  handlers/          # HTTP transport (DTOs only)
  types/dto.go       # Request/response DTOs
  middleware/        # Cross-cutting concerns
tests/               # All test files
```

## Commands
- `go run cmd/app/main.go` - Start development server
- `go test ./...` - Run all tests
- `go test -v ./internal/services/...` - Test services only
- `go test -race ./...` - Run tests with race detection
- `go mod tidy` - Clean dependencies
- `go build -o bin/app cmd/app/main.go` - Build binary

## Code Conventions

### File Naming
- `*_test.go` for test files
- `*_repo.go` for repository implementations
- `*_service.go` for service implementations
- `*_handler.go` for HTTP handlers

### Interface Definitions
- Define interfaces in consumer layer (services define repo interfaces)
- Use descriptive interface names: `PurchaseRepository`, `UserService`
- Keep interfaces small and focused

### Error Handling
- Always return errors as last parameter
- Use wrapped errors with context: `fmt.Errorf("failed to save purchase: %w", err)`
- Handle errors at every layer boundary

### Testing Patterns
- Use table-driven tests for multiple scenarios
- Mock interfaces using testify/mock
- Test files adjacent to source files
- Separate unit tests from integration tests

## TDD Requirements

### Test Structure
- **Unit Tests:** Test individual functions/methods in isolation
- **Integration Tests:** Test layer interactions (service + repository)
- **Handler Tests:** Test HTTP endpoints with mocked services
- **Repository Tests:** Test against real database (test containers)

### Test Naming Convention
```go
func TestServiceMethodName_Scenario_ExpectedResult(t *testing.T)
func TestPurchaseService_SavePurchase_ReturnsErrorWhenInvalidPrice(t *testing.T)
```

### Required Test Coverage
- Services: >90% coverage (business logic critical)
- Repositories: >85% coverage  
- Handlers: >80% coverage
- Run `go test -cover ./...` to check coverage

### Test Organization
```
internal/
  services/
    purchase_service.go
    purchase_service_test.go    # Unit tests with mocks
  repositories/  
    purchase_repo.go
    purchase_repo_test.go       # Integration tests
  handlers/
    purchase_handler.go
    purchase_handler_test.go    # HTTP tests
tests/
  integration/                  # Cross-layer integration tests
  testutils/                    # Test helpers and mocks
```

## Layer Implementation Rules

### Handlers Layer
- Parse request DTOs only
- Validate input using struct tags or custom validators
- Call service methods with domain structs
- Convert service responses to DTOs for JSON/HTML output
- Never access repositories directly

### Services Layer  
- Accept domain structs as parameters
- Return domain structs or primitive types
- Contain all business logic and rules
- Use repository interfaces only
- Never import GORM or model packages

### Repository Layer
- Implement service-defined interfaces
- Convert between domain structs and model structs
- Use GORM for all database operations
- Return domain structs to services
- Handle database-specific errors

## Data Conversion Patterns

### DTO to Domain (in handlers)
```go
func (dto PurchaseRequestDTO) ToDomain() domain.PurchaseRequest {
    return domain.PurchaseRequest{
        Name:  dto.Name,
        Price: dto.Price,
    }
}
```

### Domain to Model (in repositories)
```go
func (d domain.PurchaseRequest) ToModel() models.PurchaseModel {
    return models.PurchaseModel{
        Name:  d.Name,
        Price: d.Price,
    }
}
```

### Model to Domain (in repositories)
```go
func (m models.PurchaseModel) ToDomain() domain.PurchaseRequest {
    return domain.PurchaseRequest{
        Name:  m.Name,
        Price: m.Price,
    }
}
```

## Testing Implementation

### Service Tests (with mocks)
```go
func TestPurchaseService_SavePurchase_Success(t *testing.T) {
    // Arrange
    mockRepo := &mocks.PurchaseRepository{}
    service := services.NewPurchaseService(mockRepo)
    purchase := domain.PurchaseRequest{Name: "Test", Price: 100.0}
    
    mockRepo.On("Save", purchase).Return(nil)
    
    // Act
    err := service.SavePurchase(purchase)
    
    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

### Repository Integration Tests
```go
func TestPurchaseRepository_Save_Success(t *testing.T) {
    // Arrange
    db := setupTestDB(t) // Use testcontainers or in-memory SQLite
    repo := repositories.NewPurchaseRepository(db)
    ctx := context.Background()
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

## Do Not
- Import GORM in services or handlers
- Use model structs outside repositories
- Use DTOs in services 
- Call repositories from handlers directly
- Write business logic in handlers or repositories
- Commit without running tests
- Skip error handling at any layer
- Write code without corresponding tests

## Database Rules
- Migrations only in `internal/database/migrations/`
- Use AutoMigrate for PoC, proper migrations for production
- Model structs define database schema with GORM tags
- All queries through repository methods only

## Development Workflow
1. Write failing test first (TDD)
2. Implement minimal code to pass test
3. Refactor while keeping tests green
4. Run full test suite before committing
5. Ensure proper layer separation maintained

## Performance Guidelines
- Use GORM preloading for related data: `db.Preload("Purchases").Find(&users)`
- Implement pagination for list endpoints
- Use database indexes for frequently queried fields
- Monitor N+1 query problems

## Performance Guidelines
- Use prepared statements for repeated queries
- Implement pagination for list endpoints using LIMIT/OFFSET
- Use database indexes for frequently queried fields
- Monitor and optimize slow queries
- Use connection pooling effectively
- Minimize database round trips with batch operations

## Security Rules
- Validate all input at handler layer using go-playground/validator
- Use parameterized queries to prevent SQL injection
- Sanitize data before database operations
- Never expose internal model structures in API responses
- Use CORS middleware (gin-contrib/cors) appropriately
- Load environment variables securely with godotenv