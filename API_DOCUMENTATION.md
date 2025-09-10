# BuyOrBye Finance API Documentation

## 🚀 Overview

The BuyOrBye Finance API provides comprehensive financial management capabilities with secure authentication, real-time calculations, and intelligent financial analysis.

**Base URL**: `http://localhost:8080/api/v1`
**Authentication**: JWT Bearer Token (required for all finance endpoints)

---

## 🔐 Authentication

### Register User
Create a new user account and receive authentication tokens.

**Endpoint**: `POST /auth/register`
**Authentication**: Not required

#### Request Body
```json
{
  "email": "user@example.com",
  "name": "John Doe", 
  "password": "securePassword123"
}
```

#### Validation Rules
- **Email**: Required, valid email format
- **Name**: Required, minimum 1 character
- **Password**: Required, minimum 8 characters

#### Response
```json
// 201 Created
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "token_type": "bearer"
}

// 409 Conflict - User exists
{
  "error": {
    "status": 409,
    "code": "conflict", 
    "message": "User with this email already exists"
  }
}

// 400 Bad Request - Validation error
{
  "error": {
    "status": 400,
    "code": "validation_error",
    "message": "Validation failed",
    "details": {
      "email": "email must be a valid email address",
      "password": "password must be at least 8 characters"
    }
  }
}
```

### Login User
Authenticate with email and password to receive tokens.

**Endpoint**: `POST /auth/login`
**Authentication**: Not required

#### Request Body
```json
{
  "email": "user@example.com",
  "password": "securePassword123"
}
```

#### Response
```json
// 200 OK
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "token_type": "bearer"
}

// 401 Unauthorized - Invalid credentials
{
  "error": {
    "status": 401,
    "code": "unauthorized",
    "message": "Invalid email or password"
  }
}
```

### Refresh Token
Generate new access token using refresh token.

**Endpoint**: `POST /auth/refresh`
**Authentication**: Not required (uses refresh token)

#### Request Body
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Logout User
Revoke refresh token to invalidate session.

**Endpoint**: `POST /auth/logout`
**Authentication**: Required (Bearer token)

#### Request Body
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

## 💰 Income Management

### Add Income Source
Create a new income source with frequency-based calculations.

**Endpoint**: `POST /finance/income`
**Authentication**: Required

#### Request Body
```json
{
  "source": "Software Engineer Salary",
  "amount": 8333.33,
  "frequency": "monthly",
  "is_active": true,
  "description": "Primary employment income"
}
```

#### Validation Rules
- **Source**: Required, non-empty string
- **Amount**: Required, positive number
- **Frequency**: Required, one of: `daily`, `weekly`, `monthly`, `one-time`
- **Is_active**: Optional, defaults to `true`

#### Frequency Normalization
The API automatically normalizes frequency values:
- `daily`, `day`, `d` → `daily`
- `weekly`, `week`, `w` → `weekly` 
- `monthly`, `month`, `m` → `monthly`
- `one-time`, `onetime`, `once`, `single` → `one-time`

#### Response
```json
// 201 Created
{
  "id": "income-123-456-789",
  "source": "Software Engineer Salary", 
  "amount": 8333.33,
  "frequency": "monthly",
  "is_active": true,
  "description": "Primary employment income",
  "monthly_amount": 8333.33,
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}

// 400 Bad Request - Invalid data
{
  "error": {
    "status": 400,
    "code": "bad_request",
    "message": "Invalid financial data provided"
  }
}
```

### Get User Income Sources
Retrieve all income sources for authenticated user.

**Endpoint**: `GET /finance/income`
**Authentication**: Required

#### Response
```json
// 200 OK
{
  "incomes": [
    {
      "id": "income-123-456-789",
      "source": "Software Engineer Salary",
      "amount": 8333.33,
      "frequency": "monthly", 
      "is_active": true,
      "monthly_amount": 8333.33,
      "created_at": "2025-01-15T10:30:00Z"
    },
    {
      "id": "income-987-654-321", 
      "source": "Freelance Projects",
      "amount": 500.00,
      "frequency": "weekly",
      "is_active": true,
      "monthly_amount": 2166.67,
      "created_at": "2025-01-10T14:20:00Z"
    }
  ],
  "total_monthly_income": 10500.00
}
```

### Update Income Source
Modify existing income source (owner only).

**Endpoint**: `PUT /finance/income/:id`
**Authentication**: Required
**Authorization**: Owner only

#### Request Body
```json
{
  "source": "Senior Software Engineer Salary",
  "amount": 9500.00,
  "frequency": "monthly",
  "is_active": true,
  "description": "Promoted position"
}
```

#### Response
```json
// 200 OK - Updated income object
// 403 Forbidden - Not owner
{
  "error": {
    "status": 403,
    "code": "forbidden", 
    "message": "Access denied: You can only access your own financial records"
  }
}
// 404 Not Found - Income doesn't exist
{
  "error": {
    "status": 404,
    "code": "not_found",
    "message": "Income record not found"
  }
}
```

### Delete Income Source
Remove income source (soft delete, owner only).

**Endpoint**: `DELETE /finance/income/:id`
**Authentication**: Required
**Authorization**: Owner only

#### Response
```json
// 200 OK
{
  "message": "Income source deleted successfully"
}
```

---

## 💸 Expense Management

### Add Expense
Create a new expense with category classification.

**Endpoint**: `POST /finance/expense`
**Authentication**: Required

#### Request Body
```json
{
  "name": "Monthly Rent",
  "amount": 2500.00,
  "category": "Housing",
  "frequency": "monthly",
  "is_fixed": true,
  "priority": "Essential",
  "description": "Apartment rent payment"
}
```

#### Validation Rules
- **Name**: Required, non-empty string
- **Amount**: Required, positive number
- **Category**: Required, one of: `Housing`, `Food`, `Transportation`, `Utilities`, `Healthcare`, `Entertainment`, `Shopping`, `Other`
- **Frequency**: Required, one of: `daily`, `weekly`, `monthly`
- **Is_fixed**: Optional, defaults to `false`
- **Priority**: Optional, one of: `Essential`, `Important`, `Optional`

#### Response
```json
// 201 Created
{
  "id": "expense-123-456-789",
  "name": "Monthly Rent",
  "amount": 2500.00,
  "category": "Housing",
  "frequency": "monthly", 
  "is_fixed": true,
  "priority": "Essential",
  "monthly_amount": 2500.00,
  "created_at": "2025-01-15T10:30:00Z"
}
```

### Get User Expenses
Retrieve all expenses with optional filtering.

**Endpoint**: `GET /finance/expenses`
**Authentication**: Required

#### Query Parameters
- `category` (optional): Filter by expense category
- `fixed` (optional): Filter by fixed expenses (`true`/`false`)
- `priority` (optional): Filter by priority level

#### Response
```json
// 200 OK
{
  "expenses": [
    {
      "id": "expense-123-456-789",
      "name": "Monthly Rent",
      "amount": 2500.00,
      "category": "Housing",
      "frequency": "monthly",
      "is_fixed": true,
      "priority": "Essential",
      "monthly_amount": 2500.00
    }
  ],
  "total_monthly_expenses": 4250.00,
  "breakdown_by_category": {
    "Housing": 2500.00,
    "Food": 800.00,
    "Transportation": 450.00,
    "Utilities": 300.00,
    "Entertainment": 200.00
  }
}
```

---

## 🏦 Loan Management

### Add Loan
Create a new loan record with payment tracking.

**Endpoint**: `POST /finance/loan`
**Authentication**: Required

#### Request Body
```json
{
  "lender": "Chase Bank",
  "loan_type": "Mortgage", 
  "original_amount": 350000.00,
  "remaining_balance": 298500.00,
  "interest_rate": 3.75,
  "monthly_payment": 1820.00,
  "term_months": 360,
  "start_date": "2022-06-15"
}
```

#### Validation Rules
- **Lender**: Required, non-empty string
- **Loan_type**: Required, one of: `Mortgage`, `Auto`, `Student`, `Personal`, `Credit Card`, `Other`
- **Original_amount**: Required, positive number
- **Remaining_balance**: Required, positive number, ≤ original_amount
- **Interest_rate**: Required, positive number (as percentage)
- **Monthly_payment**: Required, positive number
- **Term_months**: Required, positive integer
- **Start_date**: Required, valid date format (YYYY-MM-DD)

#### Response
```json
// 201 Created  
{
  "id": "loan-123-456-789",
  "lender": "Chase Bank",
  "loan_type": "Mortgage",
  "original_amount": 350000.00,
  "remaining_balance": 298500.00,
  "interest_rate": 3.75,
  "monthly_payment": 1820.00,
  "term_months": 360,
  "remaining_months": 296,
  "total_interest_paid": 15420.00,
  "payoff_date": "2047-06-15",
  "created_at": "2025-01-15T10:30:00Z"
}
```

### Get User Loans
Retrieve all loans with calculated metrics.

**Endpoint**: `GET /finance/loans`
**Authentication**: Required

#### Response
```json
// 200 OK
{
  "loans": [
    {
      "id": "loan-123-456-789",
      "lender": "Chase Bank", 
      "loan_type": "Mortgage",
      "remaining_balance": 298500.00,
      "monthly_payment": 1820.00,
      "interest_rate": 3.75,
      "remaining_months": 296
    }
  ],
  "total_debt": 398500.00,
  "total_monthly_payments": 2650.00,
  "average_interest_rate": 4.2,
  "debt_breakdown_by_type": {
    "Mortgage": 298500.00,
    "Auto": 85000.00, 
    "Student": 15000.00
  }
}
```

### Update Loan
Modify existing loan details (owner only).

**Endpoint**: `PUT /finance/loan/:id`
**Authentication**: Required
**Authorization**: Owner only

---

## 📊 Financial Analysis

### Get Financial Summary
Comprehensive financial health analysis with recommendations.

**Endpoint**: `GET /finance/summary`
**Authentication**: Required

#### Response
```json
// 200 OK
{
  "user_id": "user-123-456",
  "monthly_income": 10500.00,
  "monthly_expenses": 4250.00, 
  "monthly_loan_payments": 2650.00,
  "disposable_income": 3600.00,
  "debt_to_income_ratio": 0.253,
  "savings_rate": 0.343,
  "financial_health": "Good",
  "budget_remaining": 3600.00,
  "recommendations": [
    "Your debt-to-income ratio of 25.3% is healthy",
    "Excellent savings rate of 34.3% - keep it up!",
    "Consider building emergency fund to 6 months expenses"
  ],
  "health_score": 82,
  "last_updated": "2025-01-15T10:30:00Z"
}
```

#### Financial Health Scoring
- **Excellent** (90-100): DTI ≤28%, Savings ≥20%
- **Good** (70-89): DTI ≤36%, Balanced finances  
- **Fair** (50-69): DTI ≤50%, Some concerns
- **Poor** (0-49): DTI >50% or overspending

### Get Purchase Affordability
Calculate maximum affordable purchase amount based on financial health.

**Endpoint**: `GET /finance/affordability`
**Authentication**: Required

#### Query Parameters
- `purchase_type` (optional): `emergency`, `luxury`, `necessity` 
- `timeframe` (optional): `immediate`, `3_months`, `6_months`, `1_year`

#### Response
```json
// 200 OK
{
  "user_id": "user-123-456",
  "disposable_income": 3600.00,
  "financial_health": "Good",
  "affordability_multiplier": 3.0,
  "recommendations": {
    "immediate_purchase": {
      "max_amount": 10800.00,
      "confidence": "high",
      "reasoning": "Based on 3x disposable income for Good financial health"
    },
    "planned_purchase_3_months": {
      "max_amount": 21600.00, 
      "confidence": "high",
      "reasoning": "Accumulated savings over 3 months"
    },
    "emergency_fund_needed": {
      "target_amount": 25500.00,
      "current_progress": "60%", 
      "months_to_goal": 4.2
    }
  },
  "risk_factors": [
    "DTI ratio is moderate at 25.3%",
    "Good savings buffer available"
  ],
  "last_calculated": "2025-01-15T10:30:00Z"
}
```

#### Affordability Calculation Rules
- **Excellent Health**: 3.5x disposable income
- **Good Health**: 3.0x disposable income  
- **Fair Health**: 2.0x disposable income
- **Poor Health**: 0.5x disposable income

---

## 🔒 Security Features

### Authentication & Authorization
- **JWT Tokens**: 15-minute access tokens, 7-day refresh tokens
- **User Isolation**: Users can only access their own financial data
- **Route Protection**: All finance endpoints require authentication
- **Ownership Validation**: Update/delete operations verify record ownership

### Input Validation
- **Positive Amounts**: All financial amounts must be positive
- **Required Fields**: Server-side validation for all required fields
- **Data Types**: Strict type checking for numbers, dates, enums
- **Frequency Normalization**: Automatic standardization of frequency values

### Request Security
- **Size Limits**: 1MB maximum request payload
- **Rate Limiting**: Configurable rate limiting per endpoint
- **CORS Protection**: Configurable cross-origin policies
- **Error Sanitization**: No sensitive data exposed in error messages

---

## 📏 Business Rules

### Debt-to-Income (DTI) Ratios
- **Excellent**: ≤28% 
- **Healthy**: ≤36%
- **Concerning**: 36-50%
- **Poor**: >50%

### Savings Rate Targets
- **Excellent**: ≥20%
- **Good**: 15-19%
- **Fair**: 10-14%
- **Poor**: <10%

### 50/30/20 Budget Rule
- **Needs** (50%): Essential expenses (housing, utilities, groceries)
- **Wants** (30%): Discretionary spending (entertainment, dining out)
- **Savings** (20%): Emergency fund, investments, debt paydown

### Frequency Conversion Rates
- **Daily**: × 30 days = Monthly
- **Weekly**: × 4.33 weeks = Monthly  
- **Monthly**: × 1 = Monthly
- **One-time**: Not included in monthly calculations

---

## ⚠️ Error Codes

| Status Code | Error Code | Description |
|-------------|------------|-------------|
| 400 | `bad_request` | Invalid request format or data |
| 400 | `validation_error` | Input validation failed |
| 401 | `unauthorized` | Authentication required or invalid |
| 403 | `forbidden` | Access denied - insufficient permissions |
| 404 | `not_found` | Resource not found |
| 409 | `conflict` | Resource already exists |
| 413 | `payload_too_large` | Request payload exceeds limit |
| 429 | `too_many_requests` | Rate limit exceeded |
| 500 | `internal_error` | Server error occurred |

---

## 🧪 Testing

### Example Complete Flow
```bash
# 1. Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","name":"Test User","password":"password123"}'

# 2. Add income
curl -X POST http://localhost:8080/api/v1/finance/income \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"source":"Salary","amount":5000,"frequency":"monthly"}'

# 3. Add expense  
curl -X POST http://localhost:8080/api/v1/finance/expense \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Rent","amount":1500,"category":"Housing","frequency":"monthly"}'

# 4. Get financial summary
curl -X GET http://localhost:8080/api/v1/finance/summary \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Test Data Sets
The API includes comprehensive test scenarios for:
- New graduates with student loans
- Mid-career professionals with family expenses  
- High earners with lifestyle inflation
- Budget-conscious savers with multiple income streams

This API provides a complete financial management solution with intelligent analysis, security-first design, and comprehensive business rule enforcement.