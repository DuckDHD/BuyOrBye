//go:build integration
// +build integration

package testutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/DuckDHD/BuyOrBye/internal/dtos"
)

// HTTPClient represents a test HTTP client with authentication support
type HTTPClient struct {
	BaseURL     string
	AccessToken string
	Client      *http.Client
}

// NewHTTPClient creates a new test HTTP client
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		BaseURL: baseURL,
		Client:  &http.Client{},
	}
}

// SetAccessToken sets the access token for authenticated requests
func (c *HTTPClient) SetAccessToken(token string) {
	c.AccessToken = token
}

// makeRequest makes an HTTP request with optional authentication
func (c *HTTPClient) makeRequest(t *testing.T, method, path string, body interface{}, headers map[string]string) (*http.Response, []byte) {
	var reqBody io.Reader
	
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		require.NoError(t, err, "Failed to marshal request body")
		reqBody = bytes.NewBuffer(jsonBytes)
	}
	
	url := c.BaseURL + path
	req, err := http.NewRequest(method, url, reqBody)
	require.NoError(t, err, "Failed to create request")
	
	// Set default headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	// Set authentication header if token is available
	if c.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.AccessToken)
	}
	
	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	
	resp, err := c.Client.Do(req)
	require.NoError(t, err, "Failed to make request")
	
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body")
	resp.Body.Close()
	
	return resp, respBody
}

// GET makes a GET request
func (c *HTTPClient) GET(t *testing.T, path string) (*http.Response, []byte) {
	return c.makeRequest(t, "GET", path, nil, nil)
}

// POST makes a POST request
func (c *HTTPClient) POST(t *testing.T, path string, body interface{}) (*http.Response, []byte) {
	return c.makeRequest(t, "POST", path, body, nil)
}

// PUT makes a PUT request
func (c *HTTPClient) PUT(t *testing.T, path string, body interface{}) (*http.Response, []byte) {
	return c.makeRequest(t, "PUT", path, body, nil)
}

// DELETE makes a DELETE request
func (c *HTTPClient) DELETE(t *testing.T, path string) (*http.Response, []byte) {
	return c.makeRequest(t, "DELETE", path, nil, nil)
}

// TestUser represents a test user for integration tests
type TestUser struct {
	Email    string
	Name     string
	Password string
	Token    string
}

// NewTestUser creates a new test user
func NewTestUser(email, name, password string) *TestUser {
	return &TestUser{
		Email:    email,
		Name:     name,
		Password: password,
	}
}

// Register registers the test user and stores the access token
func (u *TestUser) Register(t *testing.T, client *HTTPClient) {
	reqBody := dtos.RegisterRequestDTO{
		Email:    u.Email,
		Name:     u.Name,
		Password: u.Password,
	}
	
	resp, body := client.POST(t, "/api/v1/auth/register", reqBody)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Registration failed: %s", string(body))
	
	var tokenResponse dtos.TokenResponseDTO
	err := json.Unmarshal(body, &tokenResponse)
	require.NoError(t, err, "Failed to unmarshal token response")
	
	u.Token = tokenResponse.AccessToken
	client.SetAccessToken(u.Token)
}

// Login authenticates the test user and stores the access token
func (u *TestUser) Login(t *testing.T, client *HTTPClient) {
	reqBody := dtos.LoginRequestDTO{
		Email:    u.Email,
		Password: u.Password,
	}
	
	resp, body := client.POST(t, "/api/v1/auth/login", reqBody)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Login failed: %s", string(body))
	
	var tokenResponse dtos.TokenResponseDTO
	err := json.Unmarshal(body, &tokenResponse)
	require.NoError(t, err, "Failed to unmarshal token response")
	
	u.Token = tokenResponse.AccessToken
	client.SetAccessToken(u.Token)
}

// FinanceTestData represents test data for financial scenarios
type FinanceTestData struct {
	Incomes  []dtos.AddIncomeDTO
	Expenses []dtos.AddExpenseDTO
	Loans    []dtos.AddLoanDTO
}

// NewBasicFinanceData creates basic financial data for testing
func NewBasicFinanceData() *FinanceTestData {
	return &FinanceTestData{
		Incomes: []dtos.AddIncomeDTO{
			{
				Source:    "Software Engineer Salary",
				Amount:    8000.00,
				Frequency: "monthly",
			},
			{
				Source:    "Freelance Work",
				Amount:    1200.00,
				Frequency: "monthly",
			},
		},
		Expenses: []dtos.AddExpenseDTO{
			{
				Category:  "housing",
				Name:      "Monthly Rent",
				Amount:    2500.00,
				Frequency: "monthly",
				IsFixed:   true,
				Priority:  1,
			},
			{
				Category:  "food",
				Name:      "Groceries",
				Amount:    600.00,
				Frequency: "monthly",
				IsFixed:   false,
				Priority:  1,
			},
			{
				Category:  "transport",
				Name:      "Gas",
				Amount:    200.00,
				Frequency: "monthly",
				IsFixed:   false,
				Priority:  2,
			},
			{
				Category:  "entertainment",
				Name:      "Netflix",
				Amount:    15.99,
				Frequency: "monthly",
				IsFixed:   true,
				Priority:  3,
			},
		},
		Loans: []dtos.AddLoanDTO{
			{
				Lender:           "Chase Bank",
				Type:             "mortgage",
				PrincipalAmount:  450000.00,
				RemainingBalance: 420000.00,
				MonthlyPayment:   2200.00,
				InterestRate:     4.5,
				EndDate:          mustParseTime("2054-01-15T00:00:00Z"),
			},
		},
	}
}

// NewExcellentFinanceData creates data for excellent financial health
func NewExcellentFinanceData() *FinanceTestData {
	return &FinanceTestData{
		Incomes: []dtos.AddIncomeDTO{
			{
				Source:    "Senior Software Engineer",
				Amount:    12000.00,
				Frequency: "monthly",
			},
			{
				Source:    "Investment Income",
				Amount:    800.00,
				Frequency: "monthly",
			},
		},
		Expenses: []dtos.AddExpenseDTO{
			{
				Category:  "housing",
				Name:      "Mortgage",
				Amount:    2800.00,
				Frequency: "monthly",
				IsFixed:   true,
				Priority:  1,
			},
			{
				Category:  "food",
				Name:      "Groceries",
				Amount:    500.00,
				Frequency: "monthly",
				IsFixed:   false,
				Priority:  1,
			},
			{
				Category:  "transport",
				Name:      "Car Payment",
				Amount:    400.00,
				Frequency: "monthly",
				IsFixed:   true,
				Priority:  2,
			},
		},
		Loans: []dtos.AddLoanDTO{
			{
				Lender:           "Bank of America",
				Type:             "mortgage",
				PrincipalAmount:  350000.00,
				RemainingBalance: 320000.00,
				MonthlyPayment:   1800.00,
				InterestRate:     3.5,
				EndDate:          mustParseTime("2050-01-15T00:00:00Z"),
			},
		},
	}
}

// NewPoorFinanceData creates data for poor financial health
func NewPoorFinanceData() *FinanceTestData {
	return &FinanceTestData{
		Incomes: []dtos.AddIncomeDTO{
			{
				Source:    "Part-time Job",
				Amount:    2800.00,
				Frequency: "monthly",
			},
		},
		Expenses: []dtos.AddExpenseDTO{
			{
				Category:  "housing",
				Name:      "Rent",
				Amount:    1800.00,
				Frequency: "monthly",
				IsFixed:   true,
				Priority:  1,
			},
			{
				Category:  "food",
				Name:      "Food",
				Amount:    600.00,
				Frequency: "monthly",
				IsFixed:   false,
				Priority:  1,
			},
			{
				Category:  "transport",
				Name:      "Car Payment",
				Amount:    450.00,
				Frequency: "monthly",
				IsFixed:   true,
				Priority:  2,
			},
			{
				Category:  "utilities",
				Name:      "Phone Bill",
				Amount:    120.00,
				Frequency: "monthly",
				IsFixed:   true,
				Priority:  2,
			},
		},
		Loans: []dtos.AddLoanDTO{
			{
				Lender:           "Credit Union",
				Type:             "personal",
				PrincipalAmount:  15000.00,
				RemainingBalance: 12000.00,
				MonthlyPayment:   350.00,
				InterestRate:     8.5,
				EndDate:          mustParseTime("2027-06-15T00:00:00Z"),
			},
			{
				Lender:           "Student Loan Corp",
				Type:             "student",
				PrincipalAmount:  45000.00,
				RemainingBalance: 38000.00,
				MonthlyPayment:   420.00,
				InterestRate:     6.2,
				EndDate:          mustParseTime("2035-08-15T00:00:00Z"),
			},
		},
	}
}

// AddFinanceData adds all financial data for a user
func (fd *FinanceTestData) AddFinanceData(t *testing.T, client *HTTPClient) {
	// Add incomes
	for _, income := range fd.Incomes {
		resp, body := client.POST(t, "/api/v1/finance/income", income)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to add income: %s", string(body))
	}
	
	// Add expenses
	for _, expense := range fd.Expenses {
		resp, body := client.POST(t, "/api/v1/finance/expense", expense)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to add expense: %s", string(body))
	}
	
	// Add loans
	for _, loan := range fd.Loans {
		resp, body := client.POST(t, "/api/v1/finance/loan", loan)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to add loan: %s", string(body))
	}
}

// GetFinanceSummary retrieves the financial summary for the user
func (c *HTTPClient) GetFinanceSummary(t *testing.T) *dtos.FinanceSummaryResponseDTO {
	resp, body := c.GET(t, "/api/v1/finance/summary")
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get finance summary: %s", string(body))
	
	var summary dtos.FinanceSummaryResponseDTO
	err := json.Unmarshal(body, &summary)
	require.NoError(t, err, "Failed to unmarshal finance summary")
	
	return &summary
}

// GetAffordability retrieves the affordability calculation for the user
func (c *HTTPClient) GetAffordability(t *testing.T) float64 {
	resp, body := c.GET(t, "/api/v1/finance/affordability")
	require.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get affordability: %s", string(body))
	
	var affordabilityResponse map[string]interface{}
	err := json.Unmarshal(body, &affordabilityResponse)
	require.NoError(t, err, "Failed to unmarshal affordability response")
	
	maxAffordable, ok := affordabilityResponse["max_affordable_amount"].(float64)
	require.True(t, ok, "max_affordable_amount not found or not a number")
	
	return maxAffordable
}

// AssertValidationError asserts that the response contains validation errors
func AssertValidationError(t *testing.T, resp *http.Response, body []byte, expectedField string) {
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, "Expected validation error")
	
	var errorResponse dtos.ValidationErrorResponseDTO
	err := json.Unmarshal(body, &errorResponse)
	require.NoError(t, err, "Failed to unmarshal validation error response")
	
	require.Equal(t, "validation_error", errorResponse.Error)
	require.Contains(t, errorResponse.Fields, expectedField, "Expected field validation error not found")
}

// AssertErrorResponse asserts that the response contains the expected error
func AssertErrorResponse(t *testing.T, expectedStatus int, expectedError string, resp *http.Response, body []byte) {
	require.Equal(t, expectedStatus, resp.StatusCode, "Unexpected status code")
	
	var errorResponse dtos.ErrorResponseDTO
	err := json.Unmarshal(body, &errorResponse)
	require.NoError(t, err, "Failed to unmarshal error response")
	
	require.Equal(t, expectedError, errorResponse.Error)
}

// mustParseTime parses time or panics (for test data initialization)
func mustParseTime(timeStr string) time.Time {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse time %s: %v", timeStr, err))
	}
	return t
}