package main

import (
	"context"
	"fmt"
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
		orchestrator.WithService[*SimpleService](&SimpleService{name: "basic-service"})(
			orchestrator.NewFeature("basic-service"),
		).
			WithLifetime(orchestrator.Singleton).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(orchestrator.WithStartFunc[*SimpleService](func(service *SimpleService) error {
						return service.Start()
					})).
					WithStop(orchestrator.WithStopFuncWithApp[*SimpleService](app, func(service *SimpleService) error {
						return service.Stop()
					})).
					WithHealth(orchestrator.WithHealthFunc[*SimpleService](app, func(service *SimpleService) orchestrator.HealthStatus {
						return orchestrator.HealthStatus{
							Status:  service.Health(),
							Message: "Service is running",
						}
					})),
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