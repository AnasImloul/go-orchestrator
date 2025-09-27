package main

import (
	"context"
	"fmt"

	"github.com/AnasImloul/go-orchestrator"
)

// TestService interface
type TestService interface {
	DoSomething() error
}

// testService concrete implementation
type testService struct{}

func (t *testService) DoSomething() error {
	return nil
}

func main() {
	app := orchestrator.New()

	// Register the service as an interface
	app.AddFeature(
		orchestrator.WithService[TestService](&testService{})(
			orchestrator.NewFeature("test"),
		).
			WithLifetime(orchestrator.Singleton),
	)

	// Start the app to register services
	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		panic(err)
	}
	defer app.Stop(ctx)

	// This should work - resolving by interface
	fmt.Println("Testing interface resolution...")
	service, err := orchestrator.ResolveType[TestService](app.Container())
	if err != nil {
		fmt.Printf("❌ Interface resolution failed: %v\n", err)
	} else {
		fmt.Printf("✅ Interface resolution succeeded: %T\n", service)
	}

	// This should fail - resolving by concrete type
	fmt.Println("Testing concrete type resolution (should fail)...")
	_, err = orchestrator.ResolveType[*testService](app.Container())
	if err != nil {
		fmt.Printf("✅ Concrete type resolution correctly failed: %v\n", err)
	} else {
		fmt.Println("❌ Concrete type resolution should have failed but didn't!")
	}
}
