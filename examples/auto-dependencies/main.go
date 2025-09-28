package main

import (
	"context"
	"fmt"
	"log"

	"github.com/AnasImloul/go-orchestrator"
)

// DatabaseService interface - now implements the Service interface
type DatabaseService interface {
	orchestrator.Service
	Connect() error
	Disconnect() error
	GetInstanceID() string
}

// CacheService interface - now implements the Service interface
type CacheService interface {
	orchestrator.Service
	Connect() error
	Disconnect() error
	GetInstanceID() string
}

// APIService interface - now implements the Service interface
type APIService interface {
	orchestrator.Service
	GetInstanceID() string
}

// databaseService implementation
type databaseService struct {
	host string
	port int
	id   string
}

func (d *databaseService) Start(ctx context.Context) error {
	fmt.Printf("Database starting connection to %s:%d (ID: %s)\n", d.host, d.port, d.id)
	return d.Connect()
}

func (d *databaseService) Stop(ctx context.Context) error {
	fmt.Printf("Database stopping connection to %s:%d (ID: %s)\n", d.host, d.port, d.id)
	return d.Disconnect()
}

func (d *databaseService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: fmt.Sprintf("Database %s is connected", d.id),
	}
}

func (d *databaseService) Connect() error {
	fmt.Printf("   Database connected to %s:%d\n", d.host, d.port)
	return nil
}

func (d *databaseService) Disconnect() error {
	fmt.Printf("   Database disconnected from %s:%d\n", d.host, d.port)
	return nil
}

func (d *databaseService) GetInstanceID() string {
	return d.id
}

// cacheService implementation
type cacheService struct {
	host string
	port int
	id   string
}

func (c *cacheService) Start(ctx context.Context) error {
	fmt.Printf("Cache starting connection to %s:%d (ID: %s)\n", c.host, c.port, c.id)
	return c.Connect()
}

func (c *cacheService) Stop(ctx context.Context) error {
	fmt.Printf("Cache stopping connection to %s:%d (ID: %s)\n", c.host, c.port, c.id)
	return c.Disconnect()
}

func (c *cacheService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: fmt.Sprintf("Cache %s is connected", c.id),
	}
}

func (c *cacheService) Connect() error {
	fmt.Printf("   Cache connected to %s:%d\n", c.host, c.port)
	return nil
}

func (c *cacheService) Disconnect() error {
	fmt.Printf("   Cache disconnected from %s:%d\n", c.host, c.port)
	return nil
}

func (c *cacheService) GetInstanceID() string {
	return c.id
}

// apiService implementation
type apiService struct {
	port  int
	db    DatabaseService
	cache CacheService
	id    string
}

func (a *apiService) Start(ctx context.Context) error {
	fmt.Printf("API starting on port %d (ID: %s)\n", a.port, a.id)
	fmt.Printf("   - Database ID: %s\n", a.db.GetInstanceID())
	fmt.Printf("   - Cache ID: %s\n", a.cache.GetInstanceID())
	fmt.Printf("   API server started successfully\n")
	return nil
}

func (a *apiService) Stop(ctx context.Context) error {
	fmt.Printf("API stopping on port %d (ID: %s)\n", a.port, a.id)
	fmt.Printf("   API server stopped successfully\n")
	return nil
}

func (a *apiService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: fmt.Sprintf("API %s is running on port %d", a.id, a.port),
	}
}

func (a *apiService) GetInstanceID() string {
	return a.id
}

func main() {
	fmt.Println("Automatic Dependency Discovery Example")
	fmt.Println("==========================================")

	// Create service registry
	registry := orchestrator.New()

	// Register services with ZERO boilerplate!
	// Just pass the instance - lifecycle is automatically wired!
	registry.Register(
		orchestrator.NewServiceSingleton[DatabaseService](
			&databaseService{host: "localhost", port: 5432, id: "db-001"},
		),
	)

	registry.Register(
		orchestrator.NewServiceSingleton[CacheService](
			&cacheService{host: "localhost", port: 6379, id: "cache-001"},
		),
	)

	// Factory-based service with automatic dependency discovery AND lifecycle wiring
	registry.Register(
		orchestrator.NewServiceFactory[APIService](
			func(db DatabaseService, cache CacheService) APIService {
				return &apiService{
					port:  8080,
					db:    db,
					cache: cache,
					id:    "api-001",
				}
			},
			orchestrator.Singleton,
		),
	)

	// Start the service registry
	ctx := context.Background()
	fmt.Println("\nStarting service registry...")
	if err := registry.Start(ctx); err != nil {
		log.Fatalf("Failed to start service registry: %v", err)
	}

	// Show that dependencies were automatically resolved
	fmt.Println("\nService registry started successfully!")
	fmt.Println("   - Dependencies were automatically discovered and injected")
	fmt.Println("   - Lifecycle methods were automatically wired")
	fmt.Println("   - No manual configuration needed!")

	// Stop the service registry
	fmt.Println("\nStopping service registry...")
	if err := registry.Stop(ctx); err != nil {
		log.Fatalf("Failed to stop service registry: %v", err)
	}

	fmt.Println("Service registry stopped successfully!")
}
