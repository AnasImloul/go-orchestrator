package main

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// TestService interface
type TestService interface {
	GetID() string
	Connect() error
}

// testService implementation
type testService struct {
	id   string
	name string
}

func (t *testService) GetID() string {
	return t.id
}

func (t *testService) Connect() error {
	t.id = fmt.Sprintf("%s_%d", t.name, time.Now().UnixNano())
	fmt.Printf("Test service '%s' connected with ID: %s\n", t.name, t.id)
	return nil
}

func main() {
	fmt.Println("Go Orchestrator - Named Services Example")
	fmt.Println("========================================")

	// Create application
	app := orchestrator.New()

	// Add services with different names but same interface type
	app.AddFeature(
		orchestrator.NewFeature("services").
			WithNamedService(
				"primary-db",
				reflect.TypeOf((*TestService)(nil)).Elem(),
				func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
					return &testService{name: "primary-db"}, nil
				},
				orchestrator.Singleton,
			).
			WithNamedService(
				"secondary-db",
				reflect.TypeOf((*TestService)(nil)).Elem(),
				func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
					return &testService{name: "secondary-db"}, nil
				},
				orchestrator.Singleton,
			).
			WithNamedService(
				"cache-service",
				reflect.TypeOf((*TestService)(nil)).Elem(),
				func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
					return &testService{name: "cache-service"}, nil
				},
				orchestrator.Transient,
			).
			WithComponent(
				orchestrator.NewComponent().
					WithStart(func(ctx context.Context, container *orchestrator.Container) error {
						// Initialize all services
						primaryDB, err := orchestrator.ResolveNamedType[TestService](container, "primary-db")
						if err != nil {
							return err
						}
						primaryDB.Connect()

						secondaryDB, err := orchestrator.ResolveNamedType[TestService](container, "secondary-db")
						if err != nil {
							return err
						}
						secondaryDB.Connect()

						cacheService, err := orchestrator.ResolveNamedType[TestService](container, "cache-service")
						if err != nil {
							return err
						}
						cacheService.Connect()

						return nil
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

	// Test named service resolution
	fmt.Println("\nTesting named service resolution:")
	fmt.Println("=================================")

	container := app.Container()

	// Resolve services multiple times
	for i := 0; i < 3; i++ {
		fmt.Printf("\n--- Resolution #%d ---\n", i+1)

		// Resolve primary database (Singleton)
		primary1, _ := orchestrator.ResolveNamedType[TestService](container, "primary-db")
		primary2, _ := orchestrator.ResolveNamedType[TestService](container, "primary-db")
		fmt.Printf("Primary DB (Singleton): %s == %s? %t\n",
			primary1.GetID(), primary2.GetID(),
			primary1.GetID() == primary2.GetID())

		// Resolve secondary database (Singleton)
		secondary1, _ := orchestrator.ResolveNamedType[TestService](container, "secondary-db")
		secondary2, _ := orchestrator.ResolveNamedType[TestService](container, "secondary-db")
		fmt.Printf("Secondary DB (Singleton): %s == %s? %t\n",
			secondary1.GetID(), secondary2.GetID(),
			secondary1.GetID() == secondary2.GetID())

		// Resolve cache service (Transient)
		cache1, _ := orchestrator.ResolveNamedType[TestService](container, "cache-service")
		cache2, _ := orchestrator.ResolveNamedType[TestService](container, "cache-service")
		fmt.Printf("Cache Service (Transient): %s == %s? %t\n",
			cache1.GetID(), cache2.GetID(),
			cache1.GetID() == cache2.GetID())

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
