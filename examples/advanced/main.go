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
	GetInstanceID() string
}

// databaseService implementation
type databaseService struct {
	host string
	port int
	id   string
}

func (d *databaseService) Connect() error {
	fmt.Printf("Database starting connection to %s:%d (ID: %s)\n", d.host, d.port, d.id)
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("   Database connected successfully\n")
	return nil
}

func (d *databaseService) Disconnect() error {
	fmt.Printf("Database stopping connection to %s:%d (ID: %s)\n", d.host, d.port, d.id)
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("   Database disconnected successfully\n")
	return nil
}

func (d *databaseService) GetInstanceID() string {
	return d.id
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
		Message: fmt.Sprintf("Database %s is healthy", d.id),
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
	host string
	port int
	id   string
}

func (c *cacheService) Connect() error {
	fmt.Printf("Cache starting connection to %s:%d (ID: %s)\n", c.host, c.port, c.id)
	time.Sleep(75 * time.Millisecond)
	fmt.Printf("   Cache connected successfully\n")
	return nil
}

func (c *cacheService) Disconnect() error {
	fmt.Printf("Cache stopping connection to %s:%d (ID: %s)\n", c.host, c.port, c.id)
	time.Sleep(25 * time.Millisecond)
	fmt.Printf("   Cache disconnected successfully\n")
	return nil
}

func (c *cacheService) GetInstanceID() string {
	return c.id
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
		Message: fmt.Sprintf("Cache %s is healthy", c.id),
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
	id    string
}

func (a *apiService) start() error {
	fmt.Printf("API starting on port %d (ID: %s)\n", a.port, a.id)
	time.Sleep(150 * time.Millisecond)
	fmt.Printf("   API server started successfully\n")
	return nil
}

func (a *apiService) stop() error {
	fmt.Printf("API stopping on port %d (ID: %s)\n", a.port, a.id)
	time.Sleep(100 * time.Millisecond)
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
		Message: fmt.Sprintf("API server %s on port %d is healthy", a.id, a.port),
	}
}

func main() {
	fmt.Println("Advanced Example - Go Orchestrator")
	fmt.Println("===================================")
	fmt.Println("This example demonstrates:")
	fmt.Println("- Automatic dependency discovery")
	fmt.Println("- Factory-based service creation")
	fmt.Println("- Type-safe dependency injection")
	fmt.Println("- Advanced service lifecycle management")

	// Create service registry
	registry := orchestrator.New()

	// Register services using factories with automatic dependency discovery
	registry.Register(
		orchestrator.NewServiceFactory[DatabaseService](
			func(ctx context.Context, container *orchestrator.Container) (DatabaseService, error) {
				return &databaseService{
					host: "localhost",
					port: 5432,
					id:   fmt.Sprintf("db-%d", time.Now().UnixNano()),
				}, nil
			},
			orchestrator.Singleton,
		),
	)

	registry.Register(
		orchestrator.NewServiceFactory[CacheService](
			func(ctx context.Context, container *orchestrator.Container) (CacheService, error) {
				return &cacheService{
					host: "localhost",
					port: 6379,
					id:   fmt.Sprintf("cache-%d", time.Now().UnixNano()),
				}, nil
			},
			orchestrator.Singleton,
		),
	)

	// API service with automatic dependency injection
	registry.Register(
		orchestrator.NewServiceFactory[APIService](
			func(ctx context.Context, container *orchestrator.Container) (APIService, error) {
				// Dependencies are automatically injected
				db, err := orchestrator.ResolveType[DatabaseService](container)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve database service: %w", err)
				}

				cache, err := orchestrator.ResolveType[CacheService](container)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve cache service: %w", err)
				}

				return &apiService{
					port:  8080,
					db:    db,
					cache: cache,
					id:    fmt.Sprintf("api-%d", time.Now().UnixNano()),
				}, nil
			},
			orchestrator.Singleton,
		),
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
	fmt.Println("\nRunning for 2 seconds...")
	time.Sleep(2 * time.Second)

	// Stop service registry
	fmt.Println("\nStopping service registry...")
	if err := registry.Stop(ctx); err != nil {
		log.Fatalf("Failed to stop service registry: %v", err)
	}

	fmt.Println("Service registry stopped successfully!")
}
