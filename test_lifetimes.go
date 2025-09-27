package main

import (
	"context"
	"fmt"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// TestService interface
type TestService interface {
	GetID() string
}

// testService implementation
type testService struct {
	id string
}

func (t *testService) GetID() string {
	return t.id
}

func (t *testService) Connect() error {
	t.id = fmt.Sprintf("test_%d", time.Now().UnixNano())
	fmt.Printf("Test service connected with ID: %s\n", t.id)
	return nil
}

func main() {
	fmt.Println("Testing Service Lifetimes")
	fmt.Println("========================")

	// Create application
	app := orchestrator.New()

	// Add services with different lifetimes
	app.AddFeature(
		orchestrator.WithServiceInstanceGeneric[TestService](
			&testService{},
			orchestrator.Singleton,
		)(orchestrator.NewFeature("singleton")).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(func(ctx context.Context, container *orchestrator.Container) error {
						service, err := orchestrator.ResolveType[TestService](container)
						if err != nil {
							return err
						}
						return service.Connect()
					}),
			),
	)

	app.AddFeature(
		orchestrator.WithServiceInstanceGeneric[TestService](
			&testService{},
			orchestrator.Transient,
		)(orchestrator.NewFeature("transient")).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(func(ctx context.Context, container *orchestrator.Container) error {
						service, err := orchestrator.ResolveType[TestService](container)
						if err != nil {
							return err
						}
						return service.Connect()
					}),
			),
	)

	// Start the application
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("Starting application...")
	if err := app.Start(ctx); err != nil {
		panic(fmt.Errorf("Failed to start application: %w", err))
	}
	fmt.Println("Application started successfully!")

	// Test service resolution after startup
	fmt.Println("\nTesting service resolution after startup:")
	fmt.Println("========================================")

	container := app.Container()

	// Resolve services multiple times
	for i := 0; i < 3; i++ {
		fmt.Printf("\n--- Resolution #%d ---\n", i+1)

		// Resolve singleton service
		singleton1, err := orchestrator.ResolveType[TestService](container)
		if err != nil {
			fmt.Printf("Error resolving singleton: %v\n", err)
			continue
		}
		singleton2, err := orchestrator.ResolveType[TestService](container)
		if err != nil {
			fmt.Printf("Error resolving singleton: %v\n", err)
			continue
		}
		fmt.Printf("Singleton: %s == %s? %t\n", 
			singleton1.GetID(), singleton2.GetID(), 
			singleton1.GetID() == singleton2.GetID())

		// Resolve transient service
		transient1, err := orchestrator.ResolveType[TestService](container)
		if err != nil {
			fmt.Printf("Error resolving transient: %v\n", err)
			continue
		}
		transient2, err := orchestrator.ResolveType[TestService](container)
		if err != nil {
			fmt.Printf("Error resolving transient: %v\n", err)
			continue
		}
		fmt.Printf("Transient: %s == %s? %t\n", 
			transient1.GetID(), transient2.GetID(), 
			transient1.GetID() == transient2.GetID())

		time.Sleep(100 * time.Millisecond)
	}

	// Stop the application
	fmt.Println("\nStopping application...")
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer stopCancel()

	if err := app.Stop(stopCtx); err != nil {
		panic(fmt.Errorf("Failed to stop application: %w", err))
	}
	fmt.Println("Application stopped successfully!")
}
