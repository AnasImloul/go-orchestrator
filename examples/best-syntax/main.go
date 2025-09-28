package main

import (
	"context"
	"fmt"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// DatabaseService interface - now implements the Service interface
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

func (d *databaseService) Start(ctx context.Context) error {
	fmt.Printf("🔌 Database starting connection to %s:%d\n", d.host, d.port)
	return d.Connect()
}

func (d *databaseService) Stop(ctx context.Context) error {
	fmt.Printf("🔌 Database stopping connection to %s:%d\n", d.host, d.port)
	return d.Disconnect()
}

func (d *databaseService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: fmt.Sprintf("Database %s is connected", d.connectionID),
	}
}

func (d *databaseService) Connect() error {
	d.connectionID = fmt.Sprintf("conn_%d", time.Now().UnixNano())
	fmt.Printf("   ✅ Database connected to %s:%d (ID: %s)\n", d.host, d.port, d.connectionID)
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (d *databaseService) Disconnect() error {
	fmt.Printf("   ❌ Database disconnected from %s:%d\n", d.host, d.port)
	time.Sleep(50 * time.Millisecond)
	return nil
}

func (d *databaseService) GetConnectionID() string {
	return d.connectionID
}

// CacheService interface - now implements the Service interface
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

func (c *cacheService) Start(ctx context.Context) error {
	fmt.Printf("💾 Cache starting connection to %s:%d\n", c.host, c.port)
	return c.Connect()
}

func (c *cacheService) Stop(ctx context.Context) error {
	fmt.Printf("💾 Cache stopping connection to %s:%d\n", c.host, c.port)
	return c.Disconnect()
}

func (c *cacheService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: fmt.Sprintf("Cache %s is connected", c.instanceID),
	}
}

func (c *cacheService) Connect() error {
	c.instanceID = fmt.Sprintf("cache_%d", time.Now().UnixNano())
	fmt.Printf("   ✅ Cache connected to %s:%d (ID: %s)\n", c.host, c.port, c.instanceID)
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (c *cacheService) Disconnect() error {
	fmt.Printf("   ❌ Cache disconnected from %s:%d\n", c.host, c.port)
	time.Sleep(50 * time.Millisecond)
	return nil
}

func (c *cacheService) GetInstanceID() string {
	if c.instanceID == "" {
		c.instanceID = fmt.Sprintf("cache_%d", time.Now().UnixNano())
	}
	return c.instanceID
}

// APIService interface - now implements the Service interface
type APIService interface {
	orchestrator.Service
}

// apiService implementation
type apiService struct {
	port  int
	db    DatabaseService
	cache CacheService
}

func (a *apiService) Start(ctx context.Context) error {
	fmt.Printf("🚀 API starting on port %d\n", a.port)
	time.Sleep(150 * time.Millisecond)
	fmt.Printf("   ✅ API server started successfully\n")
	return nil
}

func (a *apiService) Stop(ctx context.Context) error {
	fmt.Printf("🛑 API stopping on port %d\n", a.port)
	time.Sleep(75 * time.Millisecond)
	fmt.Printf("   ✅ API server stopped successfully\n")
	return nil
}

func (a *apiService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: fmt.Sprintf("API server is running on port %d", a.port),
	}
}

func main() {
	fmt.Println("Go Orchestrator - Best Syntax Example")
	fmt.Println("====================================")

	// Create service registry with default configuration
	registry := orchestrator.New()

	// Method 1: Ultra-clean syntax using NewServiceSingleton (automatic name inference)
	// Lifecycle methods are automatically wired!
	registry.Register(
		orchestrator.NewServiceSingleton[DatabaseService](
			&databaseService{host: "localhost", port: 5432},
		),
	)

	// Method 2: Clean syntax using NewServiceSingleton with Transient lifetime
	// Lifecycle methods are automatically wired!
	registry.Register(
		orchestrator.NewServiceSingleton[CacheService](
			&cacheService{host: "localhost", port: 6379},
		),
	)

	// Method 3: Factory-based service with automatic dependency discovery
	// Lifecycle methods are automatically wired!
	registry.Register(
		orchestrator.NewServiceFactory[APIService](
			func(db DatabaseService, cache CacheService) APIService {
				return &apiService{port: 8080, db: db, cache: cache}
			},
			orchestrator.Singleton,
		),
	)

	// Start the service registry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Starting service registry...")
	if err := registry.Start(ctx); err != nil {
		panic(fmt.Errorf("Failed to start service registry: %w", err))
	}
	fmt.Println("Service registry started successfully!")

	// Demonstrate different lifetimes by resolving services multiple times
	fmt.Println("\nDemonstrating service lifetimes:")
	fmt.Println("================================")

	container := registry.Container()

	// Resolve services multiple times to show lifetime behavior
	for i := 0; i < 3; i++ {
		fmt.Printf("\n--- Resolution #%d ---\n", i+1)

		// Singleton: Same instance every time
		db1, _ := orchestrator.ResolveType[DatabaseService](container)
		db2, _ := orchestrator.ResolveType[DatabaseService](container)
		fmt.Printf("Database (Singleton): %s == %s? %t\n",
			db1.GetConnectionID(), db2.GetConnectionID(),
			db1.GetConnectionID() == db2.GetConnectionID())

		// Transient: New instance every time
		cache1, _ := orchestrator.ResolveType[CacheService](container)
		cache2, _ := orchestrator.ResolveType[CacheService](container)
		fmt.Printf("Cache (Transient): %s == %s? %t\n",
			cache1.GetInstanceID(), cache2.GetInstanceID(),
			cache1.GetInstanceID() == cache2.GetInstanceID())

		time.Sleep(100 * time.Millisecond)
	}

	// Check health
	fmt.Println("\nChecking application health...")
	health := registry.Health(ctx)
	for name, status := range health {
		fmt.Printf("  %s: %s - %s\n", name, status.Status, status.Message)
	}

	// Run for a bit
	fmt.Println("\nRunning for 2 seconds...")
	time.Sleep(2 * time.Second)

	// Stop the service registry
	fmt.Println("Stopping service registry...")
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer stopCancel()

	if err := registry.Stop(stopCtx); err != nil {
		panic(fmt.Errorf("Failed to stop service registry: %w", err))
	}
	fmt.Println("Service registry stopped successfully!")
}
