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

// Config struct - simple struct, no lifecycle
type Config struct {
	AppName    string
	Version    string
	Debug      bool
	DatabaseURL string
}

// DatabaseConfig struct - simple struct, no lifecycle
type DatabaseConfig struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

func (c *DatabaseConfig) GetConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", c.Username, c.Password, c.Host, c.Port, c.Database)
}

// DatabaseService interface - implements Service interface (has lifecycle)
type DatabaseService interface {
	orchestrator.Service
	Connect() error
	Disconnect() error
	GetConnectionID() string
}

// databaseService implementation - implements Service interface
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
	time.Sleep(100 * time.Millisecond)
	d.logger.Info("Database connected successfully", "connectionString", d.config.GetConnectionString())
	return nil
}

func (d *databaseService) Disconnect() error {
	d.logger.Info("Database disconnecting")
	time.Sleep(50 * time.Millisecond)
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

// APIService interface - implements Service interface (has lifecycle)
type APIService interface {
	orchestrator.Service
	GetPort() int
	GetAddress() string
}

// apiService implementation - implements Service interface
type apiService struct {
	config *Config
	db     DatabaseService
	logger orchestrator.Logger
}

func NewAPIService(config *Config, db DatabaseService, logger orchestrator.Logger) APIService {
	return &apiService{
		config: config,
		db:     db,
		logger: logger.WithComponent("api"),
	}
}

func (a *apiService) GetPort() int {
	return 8080
}

func (a *apiService) GetAddress() string {
	return fmt.Sprintf("0.0.0.0:%d", a.GetPort())
}

func (a *apiService) Start(ctx context.Context) error {
	a.logger.Info("API server starting", "app", a.config.AppName, "version", a.config.Version)
	time.Sleep(200 * time.Millisecond)
	a.logger.Info("API server started successfully", "address", a.GetAddress())
	return nil
}

func (a *apiService) Stop(ctx context.Context) error {
	a.logger.Info("API server stopping")
	time.Sleep(100 * time.Millisecond)
	a.logger.Info("API server stopped successfully")
	return nil
}

func (a *apiService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: "API server is healthy",
	}
}

func main() {
	log.Println("Polymorphic API Example - Go Orchestrator")
	log.Println("==========================================")
	log.Println("This example demonstrates:")
	log.Println("- Polymorphic API: same methods work for structs and services")
	log.Println("- Structs: registered without lifecycle management")
	log.Println("- Services: registered with full lifecycle management")
	log.Println("- Automatic dependency injection for both")
	log.Println("- Mixing structs and services seamlessly")

	// Create service registry
	registry := orchestrator.New()

	// Register structs directly - no lifecycle management
	registry.Register(
		orchestrator.NewStructSingleton(&Config{
			AppName:     "MyApp",
			Version:     "1.0.0",
			Debug:       true,
			DatabaseURL: "postgres://localhost:5432/myapp",
		}),
	)

	registry.Register(
		orchestrator.NewStructSingleton(&DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "myapp",
			Username: "user",
			Password: "password",
		}),
	)

	// Register services - with full lifecycle management
	registry.Register(
		orchestrator.NewAutoServiceFactory[DatabaseService](
			func(config *DatabaseConfig, logger orchestrator.Logger) DatabaseService {
				return NewDatabaseService(config, logger)
			},
			orchestrator.Singleton,
		),
	)

	registry.Register(
		orchestrator.NewAutoServiceFactory[APIService](
			func(config *Config, db DatabaseService, logger orchestrator.Logger) APIService {
				return NewAPIService(config, db, logger)
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
