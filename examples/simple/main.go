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
	time.Sleep(100 * time.Millisecond)
	fmt.Println("Database connected successfully")
	return nil
}

func (d *DatabaseService) Disconnect() error {
	fmt.Println("Disconnecting from database")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("Database disconnected")
	return nil
}

// APIService represents an API service
type APIService struct {
	port int
	db   *DatabaseService
}

func (a *APIService) Start() error {
	fmt.Printf("Starting API server on port %d\n", a.port)
	time.Sleep(100 * time.Millisecond)
	fmt.Println("API server started successfully")
	return nil
}

func (a *APIService) Stop() error {
	fmt.Println("Stopping API server")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("API server stopped")
	return nil
}

func (a *APIService) Health() string {
	return "healthy"
}

func main() {
	fmt.Println("Go Orchestrator - Simple Declarative Example")
	fmt.Println("=============================================")

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

	// Add API feature that depends on database
	app.AddFeature(
		orchestrator.NewFeature("api").
			WithDependencies("database").
			WithPriority(20). // Start after database
			WithService(
				reflect.TypeOf((*APIService)(nil)),
				func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
					db, err := orchestrator.ResolveType[*DatabaseService](container)
					if err != nil {
						return nil, err
					}
					return &APIService{port: 8080, db: db}, nil
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

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Starting application...")
	if err := app.Start(ctx); err != nil {
		panic(fmt.Sprintf("Failed to start application: %v", err))
	}

	fmt.Println("Application started successfully!")

	// Check health
	fmt.Println("Checking application health...")
	health := app.Health(ctx)
	for name, status := range health {
		fmt.Printf("  %s: %s - %s\n", name, status.Status, status.Message)
	}

	// Run for a bit
	fmt.Println("Running for 2 seconds...")
	time.Sleep(2 * time.Second)

	// Stop the application
	fmt.Println("Stopping application...")
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()

	if err := app.Stop(stopCtx); err != nil {
		panic(fmt.Sprintf("Failed to stop application: %v", err))
	}

	fmt.Println("Application stopped successfully!")
}
