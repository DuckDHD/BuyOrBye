package database

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/wait"
)

func mustStartMySQLContainer() (func(context.Context, ...testcontainers.TerminateOption) error, error) {
	var (
		dbName = "testdb"
		dbPwd  = "testpassword"
		dbUser = "testuser"
	)

	dbContainer, err := mysql.Run(context.Background(),
		"mysql:8.0.36",
		mysql.WithDatabase(dbName),
		mysql.WithUsername(dbUser),
		mysql.WithPassword(dbPwd),
		testcontainers.WithWaitStrategy(wait.ForLog("port: 3306  MySQL Community Server - GPL").WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, err
	}

	dbHost, err := dbContainer.Host(context.Background())
	if err != nil {
		return dbContainer.Terminate, err
	}

	dbPort, err := dbContainer.MappedPort(context.Background(), "3306/tcp")
	if err != nil {
		return dbContainer.Terminate, err
	}

	// Set environment variables for GormService
	os.Setenv("BLUEPRINT_DB_DATABASE", dbName)
	os.Setenv("BLUEPRINT_DB_PASSWORD", dbPwd)
	os.Setenv("BLUEPRINT_DB_USERNAME", dbUser)
	os.Setenv("BLUEPRINT_DB_HOST", dbHost)
	os.Setenv("BLUEPRINT_DB_PORT", dbPort.Port())

	return dbContainer.Terminate, err
}

func TestMain(m *testing.M) {
	teardown, err := mustStartMySQLContainer()
	if err != nil {
		log.Fatalf("could not start mysql container: %v", err)
	}

	m.Run()

	if teardown != nil && teardown(context.Background()) != nil {
		log.Fatalf("could not teardown mysql container: %v", err)
	}
}

func TestNewGormService(t *testing.T) {
	srv, err := NewGormService()
	if err != nil {
		t.Fatalf("NewGormService() returned error: %v", err)
	}
	if srv == nil {
		t.Fatal("NewGormService() returned nil")
	}
	defer srv.Close()
}

func TestGormService_GetDB(t *testing.T) {
	srv, err := NewGormService()
	if err != nil {
		t.Fatalf("NewGormService() returned error: %v", err)
	}
	defer srv.Close()

	db := srv.GetDB()
	if db == nil {
		t.Fatal("GetDB() returned nil")
	}
}

func TestGormService_Health(t *testing.T) {
	srv, err := NewGormService()
	if err != nil {
		t.Fatalf("NewGormService() returned error: %v", err)
	}
	defer srv.Close()

	stats := srv.Health()

	if stats["status"] != "up" {
		t.Fatalf("expected status to be up, got %s", stats["status"])
	}

	if _, ok := stats["error"]; ok {
		t.Fatalf("expected error not to be present, but got: %s", stats["error"])
	}

	if stats["message"] != "GORM database connection is healthy" {
		t.Fatalf("expected message to be 'GORM database connection is healthy', got %s", stats["message"])
	}

	// Check connection pool stats are present
	if _, ok := stats["open_connections"]; !ok {
		t.Fatal("expected open_connections to be present in health stats")
	}
	if _, ok := stats["in_use"]; !ok {
		t.Fatal("expected in_use to be present in health stats")
	}
	if _, ok := stats["idle"]; !ok {
		t.Fatal("expected idle to be present in health stats")
	}
}

func TestGormService_Close(t *testing.T) {
	srv, err := NewGormService()
	if err != nil {
		t.Fatalf("NewGormService() returned error: %v", err)
	}

	if err := srv.Close(); err != nil {
		t.Fatalf("expected Close() to return nil, got: %v", err)
	}

	// After closing, health check should fail
	stats := srv.Health()
	if stats["status"] != "down" {
		t.Fatalf("expected status to be down after close, got %s", stats["status"])
	}
}
