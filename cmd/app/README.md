# BuyOrBye Finance API Application

## Overview
This is the main BuyOrBye finance application with full authentication and finance domain integration.

## Features
- **Authentication**: JWT-based auth with login/register/refresh/logout
- **Finance Management**: Income, expense, and loan tracking
- **Validation Middleware**: Comprehensive data validation and normalization
- **Security**: Ownership validation, positive amount checks, request limits
- **Analysis**: Financial summary and affordability calculations

## API Endpoints

### Authentication Routes (`/api/v1/auth`)
- `POST /register` - Register new user account
- `POST /login` - User login
- `POST /refresh` - Refresh JWT tokens
- `POST /logout` - User logout (requires auth)

### Finance Routes (`/api/v1/finance`) - All require authentication
#### Income Management
- `POST /income` - Add new income source
- `GET /income` - Get user's income sources
- `PUT /income/:id` - Update income (owner only)
- `DELETE /income/:id` - Delete income (owner only)

#### Expense Management
- `POST /expense` - Add new expense
- `GET /expenses` - Get user's expenses
- `PUT /expense/:id` - Update expense (owner only)
- `DELETE /expense/:id` - Delete expense (owner only)

#### Loan Management
- `POST /loan` - Add new loan
- `GET /loans` - Get user's loans
- `PUT /loan/:id` - Update loan (owner only)

#### Financial Analysis
- `GET /summary` - Get complete financial summary
- `GET /affordability` - Get maximum affordable purchase amount

## Middleware Stack

### Global Middleware
- **CORS**: Cross-origin resource sharing
- **Logger**: Request logging
- **Recovery**: Panic recovery
- **Request Limits**: Size and content validation (1MB max)

### Finance-Specific Middleware
- **JWT Authentication**: Required for all finance endpoints
- **Ownership Validation**: Users can only access their own data
- **Financial Data Validation**: Positive amounts, required fields
- **Frequency Normalization**: Standardizes frequency values
- **User Ownership Checks**: Resource-level access control

## Environment Variables

### Required Configuration
```env
# Server
PORT=8080

# Database
BLUEPRINT_DB_HOST=localhost
BLUEPRINT_DB_PORT=3306
BLUEPRINT_DB_DATABASE=buyorbye
BLUEPRINT_DB_USERNAME=user
BLUEPRINT_DB_PASSWORD=password

# Authentication
JWT_SECRET=your-32-char-secret
BCRYPT_COST=14
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=168h

# Finance Thresholds
HEALTHY_DTI_RATIO=0.36      # 36% debt-to-income ratio
MIN_SAVINGS_RATE=0.20       # 20% minimum savings rate
EMERGENCY_FUND_MONTHS=6     # 6 months emergency fund
```

## Running the Application

```bash
# Install dependencies
go mod tidy

# Set up environment variables
cp .env.example .env
# Edit .env with your configuration

# Run the application
go run cmd/app/main.go

# Or build and run
go build -o bin/buyorbye cmd/app/main.go
./bin/buyorbye
```

## Architecture

### Dependency Injection Pattern
```go
DB → Repositories → Services → Handlers → Routes
```

### Middleware Chain
```
Request → Global MW → Auth MW → Finance MW → Handler
```

### Error Handling
- Domain-specific errors with proper HTTP status codes
- Validation errors with detailed field information
- Ownership checks with 403 Forbidden responses
- Internal errors with 500 status codes

## Security Features
- JWT-based authentication with refresh tokens
- Password hashing with bcrypt (configurable cost)
- CSRF protection ready
- Request size limits
- User data ownership validation
- Positive amount validation for financial data

## Data Validation
- Required field validation
- Email format validation
- Password strength requirements
- Positive amount validation
- Frequency normalization
- JSON format validation