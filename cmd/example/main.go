package main

import (
	"context"
	"fmt"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// ExampleService interface
type ExampleService interface {
	GetName() string
	Start() error
	Stop() error
	Health() string
}

// exampleService represents a simple service implementation
type exampleService struct {
	name string
}

func (s *exampleService) GetName() string {
	return s.name
}

func (s *exampleService) Start() error {
	fmt.Printf("Starting service: %s\n", s.name)
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("Service started: %s\n", s.name)
	return nil
}

func (s *exampleService) Stop() error {
	fmt.Printf("Stopping service: %s\n", s.name)
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("Service stopped: %s\n", s.name)
	return nil
}

func (s *exampleService) Health() string {
	return "healthy"
}

func main() {
	fmt.Println("Go Orchestrator - Example Application")
	fmt.Println("=====================================")

	// Create service registry with default configuration
	registry := orchestrator.New()

	// Register a simple service definition
	registry.Register(
		orchestrator.NewServiceSingleton[ExampleService](
			&exampleService{name: "example-service"},
		).
			WithLifecycle(
				orchestrator.NewLifecycle().
					WithStart(func(ctx context.Context, container *orchestrator.Container) error {
						service, err := orchestrator.ResolveType[ExampleService](container)
						if err != nil {
							return err
						}
						return service.Start()
					}).
					WithStop(func(ctx context.Context) error {
						service, err := orchestrator.ResolveType[ExampleService](registry.Container())
						if err != nil {
							return err
						}
						return service.Stop()
					}).
					WithHealth(func(ctx context.Context) orchestrator.HealthStatus {
						service, err := orchestrator.ResolveType[ExampleService](registry.Container())
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

	// Start the service registry
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Starting service registry...")
	if err := registry.Start(ctx); err != nil {
		fmt.Printf("Failed to start service registry: %v\n", err)
		return
	}

	fmt.Println("Service registry started successfully!")

	// Check health
	fmt.Println("Checking service registry health...")
	health := registry.Health(ctx)
	for name, status := range health {
		fmt.Printf("  %s: %s - %s\n", name, status.Status, status.Message)
	}

	// Run for a bit
	fmt.Println("Running for 2 seconds...")
	time.Sleep(2 * time.Second)

	// Stop the service registry
	fmt.Println("Stopping service registry...")
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()

	if err := registry.Stop(stopCtx); err != nil {
		fmt.Printf("Failed to stop service registry: %v\n", err)
		return
	}

	fmt.Println("Service registry stopped successfully!")
}
