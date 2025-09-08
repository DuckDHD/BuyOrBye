# BuyOrBye Authentication API

Complete authentication system with JWT tokens, CSRF protection, and rate limiting.

## Environment Configuration

Required environment variables in `.env`:

```bash
# Authentication
JWT_SECRET=your-very-secure-32-character-jwt-secret-key-here-2024-buyorbye
BCRYPT_COST=14
ACCESS_TOKEN_TTL=15m
REFRESH_TOKEN_TTL=168h

# CSRF Protection
CSRF_SECRET=your-very-secure-32-character-csrf-secret-key-here-2024-secure

# Database
BLUEPRINT_DB_HOST=mysql_bp
BLUEPRINT_DB_PORT=3306
BLUEPRINT_DB_DATABASE=blueprint
BLUEPRINT_DB_USERNAME=melkey
BLUEPRINT_DB_PASSWORD=password1234
```

## API Endpoints

### Public Authentication Endpoints

#### POST /auth/register
Register a new user account.

**Request:**
```json
{
  "name": "John Doe",
  "email": "john@example.com", 
  "password": "securepassword123"
}
```

**Response (201):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

#### POST /auth/login
Authenticate with email and password.

**Rate Limit:** 5 attempts per 15 minutes per IP

**Request:**
```json
{
  "email": "john@example.com",
  "password": "securepassword123"
}
```

**Response (200):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

#### POST /auth/refresh
Get new access token using refresh token.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

#### GET /auth/csrf
Get CSRF token for SPA applications.

**Response (200):**
```json
{
  "csrf_token": "aMqL8xZkKoE5by9cqkqV5k5T5k5T5k5T"
}
```

### Protected Authentication Endpoints

Require `Authorization: Bearer <access_token>` header.

#### POST /auth/logout
Revoke refresh token and logout.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (200):**
```json
{
  "message": "logged out successfully"
}
```

#### GET /auth/me
Get current user profile information.

**Response (200):**
```json
{
  "user": {
    "id": "user-123",
    "email": "john@example.com",
    "expires_at": 1640995200
  },
  "message": "user profile retrieved successfully"
}
```

### Protected API Endpoints

#### GET /api/protected
Example protected resource.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200):**
```json
{
  "message": "This is a protected resource",
  "user_id": "user-123",
  "accessed_at": "2024-09-08T16:30:00Z"
}
```

### Public Endpoints with Optional Authentication

#### GET /public/info
Public information with optional user context.

**Response (200) - Unauthenticated:**
```json
{
  "message": "This is public information",
  "timestamp": "2024-09-08T16:30:00Z",
  "authenticated": false
}
```

**Response (200) - Authenticated:**
```json
{
  "message": "This is public information", 
  "timestamp": "2024-09-08T16:30:00Z",
  "authenticated": true,
  "user_id": "user-123"
}
```

## Security Features

### JWT Authentication
- **Access tokens:** 15 minute expiration
- **Refresh tokens:** 7 day expiration (168 hours)
- **HS256 signing** with 32+ character secret
- **Automatic token rotation** on refresh

### CSRF Protection
- **SameSite=Strict** cookies
- **HttpOnly** and **Secure** flags in production
- **32-byte minimum** secret requirement
- **JSON error responses** for invalid tokens

### Rate Limiting
- **Login endpoint:** 5 attempts per 15 minutes per IP
- **General API:** 100 requests per minute per IP
- **Rate limit headers** (`X-RateLimit-*`)
- **Health check exemptions**

### Password Security
- **bcrypt hashing** with cost 14
- **Minimum 8 characters** validation
- **Password strength** requirements

## Error Responses

### Authentication Errors (401)
```json
{
  "error": "unauthorized",
  "message": "Access token has expired",
  "code": 401
}
```

### Validation Errors (400)
```json
{
  "error": "validation_error",
  "message": "Validation failed",
  "code": 400,
  "fields": {
    "email": "email must be a valid email address",
    "password": "password must be at least 8 characters"
  }
}
```

### Rate Limit Errors (429)
```json
{
  "error": "rate_limit_exceeded", 
  "message": "Too many requests. Please try again later.",
  "code": 429
}
```

### CSRF Errors (403)
```json
{
  "error": "forbidden",
  "message": "CSRF token invalid or missing",
  "code": 403
}
```

## Usage Examples

### Registration Flow
1. GET `/auth/csrf` - Get CSRF token
2. POST `/auth/register` with CSRF token in header
3. Store access/refresh tokens securely
4. Use access token in Authorization header

### Login Flow  
1. GET `/auth/csrf` - Get CSRF token
2. POST `/auth/login` with CSRF token
3. Store tokens securely
4. Use access token for API calls

### Token Refresh Flow
1. When access token expires (15 min)
2. POST `/auth/refresh` with refresh token
3. Update stored access token
4. Continue API calls with new token

### SPA Integration
```javascript
// Get CSRF token first
const csrfResponse = await fetch('/auth/csrf');
const { csrf_token } = await csrfResponse.json();

// Login with CSRF protection
const loginResponse = await fetch('/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-CSRF-Token': csrf_token
  },
  credentials: 'include', // Include cookies
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'password123'
  })
});

const tokens = await loginResponse.json();
localStorage.setItem('access_token', tokens.access_token);
localStorage.setItem('refresh_token', tokens.refresh_token);
```

## Health Monitoring

### GET /health
Database and system health check.

**Response (200):**
```json
{
  "status": "healthy",
  "database": {
    "status": "up",
    "message": "GORM database connection is healthy",
    "open_connections": "5"
  },
  "timestamp": "2024-09-08T16:30:00Z"
}
```

### GET /ping
Simple ping endpoint for load balancers.

**Response (200):**
```json
{
  "message": "pong"
}
```

## System Architecture

The authentication system follows clean architecture principles:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Middleware    │    │    Handlers     │    │    Services     │
│                 │    │                 │    │                 │
│ • JWT Auth      │───▶│ • Auth Handler  │───▶│ • Auth Service  │
│ • CSRF          │    │ • Validation    │    │ • Password Svc  │
│ • Rate Limiting │    │ • HTTP Logic    │    │ • JWT Service   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                       │
                                              ┌─────────────────┐
                                              │  Repositories   │
                                              │                 │
                                              │ • User Repo     │
                                              │ • Token Repo    │
                                              │ • GORM Models   │
                                              └─────────────────┘
```

**Layer Responsibilities:**
- **Middleware:** Cross-cutting security concerns
- **Handlers:** HTTP transport and DTO conversion
- **Services:** Business logic and orchestration
- **Repositories:** Data persistence with GORM

The system is production-ready with comprehensive security, testing, and monitoring capabilities.