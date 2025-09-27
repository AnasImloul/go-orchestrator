package main

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// DatabaseService represents a database service
type DatabaseService struct {
	host string
	port int
}

func (d *DatabaseService) Connect() error {
	fmt.Printf("Connecting to database at %s:%d\n", d.host, d.port)
	time.Sleep(200 * time.Millisecond)
	fmt.Println("Database connected successfully")
	return nil
}

func (d *DatabaseService) Disconnect() error {
	fmt.Println("Disconnecting from database")
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Database disconnected")
	return nil
}

// CacheService represents a cache service
type CacheService struct {
	host string
	port int
}

func (c *CacheService) Connect() error {
	fmt.Printf("Connecting to cache at %s:%d\n", c.host, c.port)
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Cache connected successfully")
	return nil
}

func (c *CacheService) Disconnect() error {
	fmt.Println("Disconnecting from cache")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("Cache disconnected")
	return nil
}

// APIService represents an API service
type APIService struct {
	port  int
	db    *DatabaseService
	cache *CacheService
}

func (a *APIService) Start() error {
	fmt.Printf("Starting API server on port %d\n", a.port)
	time.Sleep(150 * time.Millisecond)
	fmt.Println("API server started successfully")
	return nil
}

func (a *APIService) Stop() error {
	fmt.Println("Stopping API server")
	time.Sleep(75 * time.Millisecond)
	fmt.Println("API server stopped")
	return nil
}

func (a *APIService) Health() string {
	return "healthy"
}

// WorkerService represents a background worker service
type WorkerService struct {
	db    *DatabaseService
	cache *CacheService
}

func (w *WorkerService) Start() error {
	fmt.Println("Starting background worker")
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Background worker started")
	return nil
}

func (w *WorkerService) Stop() error {
	fmt.Println("Stopping background worker")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("Background worker stopped")
	return nil
}

func (w *WorkerService) Health() string {
	return "healthy"
}

func main() {
	fmt.Println("Go Orchestrator - Advanced Example")
	fmt.Println("==================================")

	// Create application with default configuration
	app := orchestrator.New()

	// Add database feature
	app.AddFeature(
		orchestrator.NewFeature("database").
			WithPriority(10). // Start first
			WithServiceInstance(
				reflect.TypeOf((*DatabaseService)(nil)),
				&DatabaseService{host: "localhost", port: 5432},
			).
			WithComponent(
				func(ctx context.Context, container *orchestrator.Container) error {
					db, err := orchestrator.ResolveType[*DatabaseService](container)
					if err != nil {
						return err
					}
					return db.Connect()
				},
				func(ctx context.Context) error {
					db, err := orchestrator.ResolveType[*DatabaseService](app.Container())
					if err != nil {
						return err
					}
					return db.Disconnect()
				},
				func(ctx context.Context) orchestrator.HealthStatus {
					return orchestrator.HealthStatus{
						Status:  "healthy",
						Message: "Database is connected",
					}
				},
			),
	)

	// Add cache feature
	app.AddFeature(
		orchestrator.NewFeature("cache").
			WithPriority(15). // Start after database
			WithServiceInstance(
				reflect.TypeOf((*CacheService)(nil)),
				&CacheService{host: "localhost", port: 6379},
			).
			WithComponent(
				func(ctx context.Context, container *orchestrator.Container) error {
					cache, err := orchestrator.ResolveType[*CacheService](container)
					if err != nil {
						return err
					}
					return cache.Connect()
				},
				func(ctx context.Context) error {
					cache, err := orchestrator.ResolveType[*CacheService](app.Container())
					if err != nil {
						return err
					}
					return cache.Disconnect()
				},
				func(ctx context.Context) orchestrator.HealthStatus {
					return orchestrator.HealthStatus{
						Status:  "healthy",
						Message: "Cache is connected",
					}
				},
			),
	)

	// Add API feature that depends on both database and cache
	app.AddFeature(
		orchestrator.NewFeature("api").
			WithDependencies("database", "cache").
			WithPriority(20). // Start after database and cache
			WithService(
				reflect.TypeOf((*APIService)(nil)),
				func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
					db, err := orchestrator.ResolveType[*DatabaseService](container)
					if err != nil {
						return nil, err
					}
					cache, err := orchestrator.ResolveType[*CacheService](container)
					if err != nil {
						return nil, err
					}
					return &APIService{port: 8080, db: db, cache: cache}, nil
				},
				orchestrator.Singleton,
			).
			WithComponent(
				func(ctx context.Context, container *orchestrator.Container) error {
					api, err := orchestrator.ResolveType[*APIService](container)
					if err != nil {
						return err
					}
					return api.Start()
				},
				func(ctx context.Context) error {
					api, err := orchestrator.ResolveType[*APIService](app.Container())
					if err != nil {
						return err
					}
					return api.Stop()
				},
				func(ctx context.Context) orchestrator.HealthStatus {
					api, err := orchestrator.ResolveType[*APIService](app.Container())
					if err != nil {
						return orchestrator.HealthStatus{
							Status:  "unhealthy",
							Message: "Failed to resolve API service",
						}
					}
					return orchestrator.HealthStatus{
						Status:  api.Health(),
						Message: "API server is running",
					}
				},
			),
	)

	// Add worker feature that depends on database and cache
	app.AddFeature(
		orchestrator.NewFeature("worker").
			WithDependencies("database", "cache").
			WithPriority(25). // Start after API
			WithService(
				reflect.TypeOf((*WorkerService)(nil)),
				func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
					db, err := orchestrator.ResolveType[*DatabaseService](container)
					if err != nil {
						return nil, err
					}
					cache, err := orchestrator.ResolveType[*CacheService](container)
					if err != nil {
						return nil, err
					}
					return &WorkerService{db: db, cache: cache}, nil
				},
				orchestrator.Singleton,
			).
			WithComponent(
				func(ctx context.Context, container *orchestrator.Container) error {
					worker, err := orchestrator.ResolveType[*WorkerService](container)
					if err != nil {
						return err
					}
					return worker.Start()
				},
				func(ctx context.Context) error {
					worker, err := orchestrator.ResolveType[*WorkerService](app.Container())
					if err != nil {
						return err
					}
					return worker.Stop()
				},
				func(ctx context.Context) orchestrator.HealthStatus {
					worker, err := orchestrator.ResolveType[*WorkerService](app.Container())
					if err != nil {
						return orchestrator.HealthStatus{
							Status:  "unhealthy",
							Message: "Failed to resolve worker service",
						}
					}
					return orchestrator.HealthStatus{
						Status:  worker.Health(),
						Message: "Worker is running",
					}
				},
			),
	)

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Starting application...")
	if err := app.Start(ctx); err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
		return
	}

	fmt.Println("Application started successfully!")

	// Check health
	fmt.Println("Checking application health...")
	health := app.Health(ctx)
	for name, status := range health {
		fmt.Printf("  %s: %s - %s\n", name, status.Status, status.Message)
	}

	// Run for a bit
	fmt.Println("Running for 3 seconds...")
	time.Sleep(3 * time.Second)

	// Stop the application
	fmt.Println("Stopping application...")
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer stopCancel()

	if err := app.Stop(stopCtx); err != nil {
		fmt.Printf("Failed to stop application: %v\n", err)
		return
	}

	fmt.Println("Application stopped successfully!")
}