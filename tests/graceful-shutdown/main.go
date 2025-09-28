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

// DatabaseService interface - implements the Service interface
type DatabaseService interface {
	orchestrator.Service
	Connect() error
	Disconnect() error
	GetConnectionID() string
}

// databaseService implementation
type databaseService struct {
	host         string
	port         int
	connectionID string
	logger       orchestrator.Logger
}

func NewDatabaseService(host string, port int, logger orchestrator.Logger) DatabaseService {
	return &databaseService{
		host:   host,
		port:   port,
		logger: logger.WithComponent("database"),
	}
}

func (d *databaseService) Connect() error {
	d.logger.Info("Database connecting", "host", d.host, "port", d.port)
	time.Sleep(100 * time.Millisecond) // Simulate connection time
	d.connectionID = "db-12345"
	d.logger.Info("Database connected successfully", "connectionID", d.connectionID)
	return nil
}

func (d *databaseService) Disconnect() error {
	d.logger.Info("Database disconnecting", "connectionID", d.connectionID)
	time.Sleep(50 * time.Millisecond) // Simulate disconnection time
	d.logger.Info("Database disconnected successfully")
	return nil
}

func (d *databaseService) GetConnectionID() string {
	return d.connectionID
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
	log.Println("Graceful Shutdown Test - Go Orchestrator")
	log.Println("========================================")
	log.Println("This test demonstrates:")
	log.Println("- Using RunWithGracefulShutdown convenience function")
	log.Println("- Automatic signal handling with graceful shutdown")
	log.Println("- Proper public API exposure")

	// Create service registry
	registry := orchestrator.New()

	// Register services with automatic logger injection
	registry.Register(
		orchestrator.NewAutoServiceFactory[DatabaseService](
			func(logger orchestrator.Logger) DatabaseService {
				return NewDatabaseService("localhost", 5432, logger)
			},
			orchestrator.Singleton,
		),
	)

	// Set up signal handling
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Use the convenience function for graceful shutdown
	// This handles: start -> wait for signal -> graceful shutdown
	shutdownTimeout := 30 * time.Second
	if err := orchestrator.RunWithGracefulShutdown(registry, ctx, shutdownTimeout); err != nil {
		log.Printf("Application error: %v", err)
		os.Exit(1)
	}

	log.Println("Application shutdown completed successfully")
}
