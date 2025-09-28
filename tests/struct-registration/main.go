package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// DatabaseConfig represents a simple configuration struct
type DatabaseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// GetConnectionString returns the database connection string
func (c *DatabaseConfig) GetConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", c.Username, c.Password, c.Host, c.Port, c.Database)
}

// APIConfig represents API configuration
type APIConfig struct {
	Port        int
	Host        string
	Environment string
	Version     string
}

// GetAddress returns the API server address
func (c *APIConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// DatabaseService interface - implements the Service interface
type DatabaseService interface {
	orchestrator.Service
	Connect() error
	Disconnect() error
	GetConnectionID() string
}

// databaseService implementation
type databaseService struct {
	config *DatabaseConfig
	logger orchestrator.Logger
}

func NewDatabaseService(config *DatabaseConfig, logger orchestrator.Logger) DatabaseService {
	return &databaseService{
		config: config,
		logger: logger.WithComponent("database"),
	}
}

func (d *databaseService) Connect() error {
	d.logger.Info("Database connecting", "host", d.config.Host, "port", d.config.Port)
	time.Sleep(100 * time.Millisecond) // Simulate connection time
	d.logger.Info("Database connected successfully", "connectionString", d.config.GetConnectionString())
	return nil
}

func (d *databaseService) Disconnect() error {
	d.logger.Info("Database disconnecting")
	time.Sleep(50 * time.Millisecond) // Simulate disconnection time
	d.logger.Info("Database disconnected successfully")
	return nil
}

func (d *databaseService) GetConnectionID() string {
	return "db-12345"
}

func (d *databaseService) Start(ctx context.Context) error {
	return d.Connect()
}

func (d *databaseService) Stop(ctx context.Context) error {
	return d.Disconnect()
}

func (d *databaseService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: "Database is healthy",
	}
}

func main() {
	log.Println("Struct Registration Example - Go Orchestrator")
	log.Println("=============================================")
	log.Println("This example demonstrates:")
	log.Println("- Registering struct instances directly")
	log.Println("- Using structs as dependencies")
	log.Println("- Mixing structs and services")
	log.Println("- Automatic dependency injection for structs")

	// Create service registry
	registry := orchestrator.New()

	// Register struct instances directly
	registry.Register(
		orchestrator.NewStructSingleton(&DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "myapp",
			Username: "user",
			Password: "password",
		}),
	)

	registry.Register(
		orchestrator.NewStructSingleton(&APIConfig{
			Port:        8080,
			Host:        "0.0.0.0",
			Environment: "development",
			Version:     "1.0.0",
		}),
	)

	// Register a service that depends on the struct
	registry.Register(
		orchestrator.NewAutoServiceFactory[DatabaseService](
			func(config *DatabaseConfig, logger orchestrator.Logger) DatabaseService {
				return NewDatabaseService(config, logger)
			},
			orchestrator.Singleton,
		),
	)

	// Set up signal handling
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Use the convenience function for graceful shutdown
	shutdownTimeout := 30 * time.Second
	if err := orchestrator.RunWithGracefulShutdown(registry, ctx, shutdownTimeout); err != nil {
		log.Printf("Application error: %v", err)
		os.Exit(1)
	}

	log.Println("Application shutdown completed successfully")
}
