package main

import (
	"context"
	"fmt"
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
		orchestrator.WithService[*DatabaseService](&DatabaseService{host: "localhost", port: 5432})(
			orchestrator.NewFeature("database"),
		).
			WithLifetime(orchestrator.Singleton).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(orchestrator.WithStartFunc[*DatabaseService](func(db *DatabaseService) error {
						return db.Connect()
					})).
					WithStop(orchestrator.WithStopFuncWithApp[*DatabaseService](app, func(db *DatabaseService) error {
						return db.Disconnect()
					})).
					WithHealth(func(ctx context.Context) orchestrator.HealthStatus {
						return orchestrator.HealthStatus{
							Status:  "healthy",
							Message: "Database is connected",
						}
					}),
			),
	)

	// Add cache feature
	app.AddFeature(
		orchestrator.WithService[*CacheService](&CacheService{host: "localhost", port: 6379})(
			orchestrator.NewFeature("cache"),
		).
			WithLifetime(orchestrator.Singleton).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(orchestrator.WithStartFunc[*CacheService](func(cache *CacheService) error {
						return cache.Connect()
					})).
					WithStop(orchestrator.WithStopFuncWithApp[*CacheService](app, func(cache *CacheService) error {
						return cache.Disconnect()
					})).
					WithHealth(func(ctx context.Context) orchestrator.HealthStatus {
						return orchestrator.HealthStatus{
							Status:  "healthy",
							Message: "Cache is connected",
						}
					}),
			),
	)

	// Add API feature that depends on both database and cache
	app.AddFeature(
		orchestrator.WithServiceFactory[*APIService](
			func(ctx context.Context, container *orchestrator.Container) (*APIService, error) {
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
		)(
			orchestrator.NewFeature("api").
				WithDependencies("database", "cache"),
		).
			WithLifetime(orchestrator.Singleton).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(orchestrator.WithStartFunc[*APIService](func(api *APIService) error {
						return api.Start()
					})).
					WithStop(orchestrator.WithStopFuncWithApp[*APIService](app, func(api *APIService) error {
						return api.Stop()
					})).
					WithHealth(orchestrator.WithHealthFunc[*APIService](app, func(api *APIService) orchestrator.HealthStatus {
						return orchestrator.HealthStatus{
							Status:  api.Health(),
							Message: "API server is running",
						}
					})),
			),
	)

	// Add worker feature that depends on database and cache
	app.AddFeature(
		orchestrator.WithServiceFactory[*WorkerService](
			func(ctx context.Context, container *orchestrator.Container) (*WorkerService, error) {
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
		)(
			orchestrator.NewFeature("worker").
				WithDependencies("database", "cache"),
		).
			WithLifetime(orchestrator.Singleton).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(orchestrator.WithStartFunc[*WorkerService](func(worker *WorkerService) error {
						return worker.Start()
					})).
					WithStop(orchestrator.WithStopFuncWithApp[*WorkerService](app, func(worker *WorkerService) error {
						return worker.Stop()
					})).
					WithHealth(orchestrator.WithHealthFunc[*WorkerService](app, func(worker *WorkerService) orchestrator.HealthStatus {
						return orchestrator.HealthStatus{
							Status:  worker.Health(),
							Message: "Worker is running",
						}
					})),
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
