package main

import (
	"context"
	"fmt"
	"log"
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
}

func (d *databaseService) Connect() error {
	fmt.Printf("Database connecting to %s:%d\n", d.host, d.port)
	time.Sleep(100 * time.Millisecond) // Simulate connection time
	d.connectionID = fmt.Sprintf("db-%d", time.Now().UnixNano())
	fmt.Printf("   Database connected to %s:%d (ID: %s)\n", d.host, d.port, d.connectionID)
	return nil
}

func (d *databaseService) Disconnect() error {
	fmt.Printf("   Database disconnected from %s:%d\n", d.host, d.port)
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
		Message: fmt.Sprintf("Database connection %s is healthy", d.connectionID),
	}
}

// CacheService interface - implements the Service interface
type CacheService interface {
	orchestrator.Service
	Connect() error
	Disconnect() error
	GetInstanceID() string
}

// cacheService implementation
type cacheService struct {
	host       string
	port       int
	instanceID string
}

func (c *cacheService) Connect() error {
	fmt.Printf("Cache connecting to %s:%d\n", c.host, c.port)
	time.Sleep(50 * time.Millisecond) // Simulate connection time
	c.instanceID = fmt.Sprintf("cache-%d", time.Now().UnixNano())
	fmt.Printf("   Cache connected to %s:%d (ID: %s)\n", c.host, c.port, c.instanceID)
	return nil
}

func (c *cacheService) Disconnect() error {
	fmt.Printf("   Cache disconnected from %s:%d\n", c.host, c.port)
	return nil
}

func (c *cacheService) GetInstanceID() string {
	return c.instanceID
}

func (c *cacheService) Start(ctx context.Context) error {
	return c.Connect()
}

func (c *cacheService) Stop(ctx context.Context) error {
	return c.Disconnect()
}

func (c *cacheService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: fmt.Sprintf("Cache instance %s is healthy", c.instanceID),
	}
}

// APIService interface - implements the Service interface
type APIService interface {
	orchestrator.Service
	GetPort() int
}

// apiService implementation
type apiService struct {
	port  int
	db    DatabaseService
	cache CacheService
}

func (a *apiService) start() error {
	fmt.Printf("API starting on port %d\n", a.port)
	time.Sleep(200 * time.Millisecond) // Simulate startup time
	fmt.Printf("   API server started successfully\n")
	return nil
}

func (a *apiService) stop() error {
	fmt.Printf("API stopping on port %d\n", a.port)
	time.Sleep(100 * time.Millisecond) // Simulate shutdown time
	fmt.Printf("   API server stopped successfully\n")
	return nil
}

func (a *apiService) GetPort() int {
	return a.port
}

func (a *apiService) Start(ctx context.Context) error {
	return a.start()
}

func (a *apiService) Stop(ctx context.Context) error {
	return a.stop()
}

func (a *apiService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: fmt.Sprintf("API server on port %d is healthy", a.port),
	}
}

func main() {
	fmt.Println("Basic Example - Go Orchestrator")
	fmt.Println("================================")
	fmt.Println("This example demonstrates:")
	fmt.Println("- Basic service registration")
	fmt.Println("- Service lifecycle management")
	fmt.Println("- Health checking")
	fmt.Println("- Dependency injection")

	// Create service registry
	registry := orchestrator.New()

	// Register services with explicit dependencies
	registry.Register(
		orchestrator.NewServiceSingleton[DatabaseService](
			&databaseService{host: "localhost", port: 5432},
		),
	)

	registry.Register(
		orchestrator.NewServiceSingleton[CacheService](
			&cacheService{host: "localhost", port: 6379},
		),
	)

	registry.Register(
		orchestrator.NewServiceSingleton[APIService](
			&apiService{port: 8080},
		).WithDependencies("DatabaseService", "CacheService"),
	)

	// Start service registry
	ctx := context.Background()
	fmt.Println("\nStarting service registry...")
	if err := registry.Start(ctx); err != nil {
		log.Fatalf("Failed to start service registry: %v", err)
	}

	fmt.Println("\nService registry started successfully!")

	// Check service health
	fmt.Println("\nChecking service health...")
	health := registry.Health(ctx)
	for serviceName, status := range health {
		fmt.Printf("Service %s: %s - %s\n", serviceName, status.Status, status.Message)
	}

	// Run for a few seconds
	fmt.Println("\nRunning for 3 seconds...")
	time.Sleep(3 * time.Second)

	// Stop service registry
	fmt.Println("\nStopping service registry...")
	if err := registry.Stop(ctx); err != nil {
		log.Fatalf("Failed to stop service registry: %v", err)
	}

	fmt.Println("Service registry stopped successfully!")
}
