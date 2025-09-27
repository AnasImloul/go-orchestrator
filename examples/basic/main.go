package main

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// SimpleService represents a basic service
type SimpleService struct {
	name string
}

func (s *SimpleService) Start() error {
	fmt.Printf("Starting %s\n", s.name)
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("%s started\n", s.name)
	return nil
}

func (s *SimpleService) Stop() error {
	fmt.Printf("Stopping %s\n", s.name)
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("%s stopped\n", s.name)
	return nil
}

func (s *SimpleService) Health() string {
	return "healthy"
}

func main() {
	fmt.Println("Go Orchestrator - Basic Example")
	fmt.Println("===============================")

	// Create application with default configuration
	app := orchestrator.New()

	// Add a simple feature
	app.AddFeature(
		orchestrator.NewFeature("basic-service").
			WithServiceInstance(
				reflect.TypeOf((*SimpleService)(nil)),
				&SimpleService{name: "basic-service"},
			).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(func(ctx context.Context, container *orchestrator.Container) error {
						service, err := orchestrator.ResolveType[*SimpleService](container)
						if err != nil {
							return err
						}
						return service.Start()
					}).
					WithStop(func(ctx context.Context) error {
						service, err := orchestrator.ResolveType[*SimpleService](app.Container())
						if err != nil {
							return err
						}
						return service.Stop()
					}).
					WithHealth(func(ctx context.Context) orchestrator.HealthStatus {
						service, err := orchestrator.ResolveType[*SimpleService](app.Container())
						if err != nil {
							return orchestrator.HealthStatus{
								Status:  "unhealthy",
								Message: "Failed to resolve service",
							}
						}
						return orchestrator.HealthStatus{
							Status:  service.Health(),
							Message: "Service is running",
						}
					}),
			),
	)

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("Starting application...")
	if err := app.Start(ctx); err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
		return
	}

	fmt.Println("Application started successfully!")

	// Run for a short time
	time.Sleep(1 * time.Second)

	// Stop gracefully
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	fmt.Println("Stopping application...")
	if err := app.Stop(stopCtx); err != nil {
		fmt.Printf("Failed to stop application: %v\n", err)
		return
	}

	fmt.Println("Application stopped successfully!")
}