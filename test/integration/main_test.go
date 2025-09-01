package integration

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DuckDHD/BuyOrBye/internal/config"
	"github.com/DuckDHD/BuyOrBye/internal/database"
	"github.com/DuckDHD/BuyOrBye/internal/handlers"
	"github.com/DuckDHD/BuyOrBye/internal/services"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	server *httptest.Server
	config *config.Config
}

func (suite *IntegrationTestSuite) SetupSuite() {
	// Setup test database
	suite.config = &config.Config{
		DatabaseURL: os.Getenv("TEST_DATABASE_URL"),
		Environment: "test",
	}

	if suite.config.DatabaseURL == "" {
		suite.T().Skip("TEST_DATABASE_URL not set")
	}

	db, err := database.Connect(suite.config.DatabaseURL)
	suite.Require().NoError(err)

	// Initialize services
	decisionService := services.NewDecisionService(db, suite.config)
	userService := services.NewUserService(db)
	paymentService := services.NewPaymentService(suite.config)

	// Initialize handlers
	h := handlers.New(decisionService, userService, paymentService, suite.config)

	// Create test server
	r := chi.NewRouter()
	h.SetupRoutes(r)
	suite.server = httptest.NewServer(r)
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	suite.server.Close()
}

func (suite *IntegrationTestSuite) TestHealthCheck() {
	resp, err := http.Get(suite.server.URL + "/health")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func (suite *IntegrationTestSuite) TestHomePageLoads() {
	resp, err := http.Get(suite.server.URL + "/")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
