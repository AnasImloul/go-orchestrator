package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// Simple struct - no Service interface
type Config struct {
	AppName string
	Version string
	Debug   bool
}

// Service that implements Service interface
type DatabaseService interface {
	orchestrator.Service
	Connect() error
}

type databaseService struct {
	config *Config
	logger orchestrator.Logger
}

func NewDatabaseService(config *Config, logger orchestrator.Logger) DatabaseService {
	return &databaseService{
		config: config,
		logger: logger.WithComponent("database"),
	}
}

func (d *databaseService) Connect() error {
	d.logger.Info("Database connecting", "app", d.config.AppName)
	time.Sleep(100 * time.Millisecond)
	d.logger.Info("Database connected successfully")
	return nil
}

func (d *databaseService) Start(ctx context.Context) error {
	return d.Connect()
}

func (d *databaseService) Stop(ctx context.Context) error {
	d.logger.Info("Database disconnecting")
	time.Sleep(50 * time.Millisecond)
	d.logger.Info("Database disconnected successfully")
	return nil
}

func (d *databaseService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: "Database is healthy",
	}
}

func main() {
	log.Println("Simple Struct Test - Go Orchestrator")
	log.Println("====================================")
	log.Println("This example demonstrates:")
	log.Println("- NewAutoServiceFactory works with structs (no Service interface)")
	log.Println("- NewAutoServiceFactory works with services (implements Service interface)")
	log.Println("- Automatic dependency injection for both")

	// Create service registry
	registry := orchestrator.New()

	// Register struct using NewAutoServiceFactory - this works!
	registry.Register(
		orchestrator.NewAutoServiceFactory[*Config](
			func() *Config {
				return &Config{
					AppName: "MyApp",
					Version: "1.0.0",
					Debug:   true,
				}
			},
			orchestrator.Singleton,
		),
	)

	// Register service using NewAutoServiceFactory - this also works!
	registry.Register(
		orchestrator.NewAutoServiceFactory[DatabaseService](
			func(config *Config, logger orchestrator.Logger) DatabaseService {
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
