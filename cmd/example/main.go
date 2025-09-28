package main

import (
	"context"
	"fmt"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// ExampleService represents a simple service
type ExampleService struct {
	name string
}

func (s *ExampleService) GetName() string {
	return s.name
}

func (s *ExampleService) Start() error {
	fmt.Printf("Starting service: %s\n", s.name)
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("Service started: %s\n", s.name)
	return nil
}

func (s *ExampleService) Stop() error {
	fmt.Printf("Stopping service: %s\n", s.name)
	time.Sleep(50 * time.Millisecond)
	fmt.Printf("Service stopped: %s\n", s.name)
	return nil
}

func (s *ExampleService) Health() string {
	return "healthy"
}

func main() {
	fmt.Println("Go Orchestrator - Example Application")
	fmt.Println("=====================================")

	// Create application with default configuration
	app := orchestrator.New()

	// Add a simple feature
	app.AddFeature(
		orchestrator.WithComponentFor[*ExampleService](
			orchestrator.NewFeatureWithInstance("example-service", &ExampleService{name: "example-service"}, orchestrator.Singleton),
			app,
		).
			WithStartFor(func(service *ExampleService) error { return service.Start() }).
			WithStopFor(func(service *ExampleService) error { return service.Stop() }).
			WithHealthFor(func(service *ExampleService) orchestrator.HealthStatus {
				return orchestrator.HealthStatus{
					Status:  service.Health(),
					Message: "Service is running",
				}
			}).
			Build(),
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
	fmt.Println("Running for 2 seconds...")
	time.Sleep(2 * time.Second)

	// Stop the application
	fmt.Println("Stopping application...")
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer stopCancel()

	if err := app.Stop(stopCtx); err != nil {
		fmt.Printf("Failed to stop application: %v\n", err)
		return
	}

	fmt.Println("Application stopped successfully!")
}
