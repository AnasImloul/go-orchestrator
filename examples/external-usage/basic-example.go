package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/AnasImloul/go-orchestrator/pkg/orchestrator"
)

// This example shows the basic structure for using the Go Orchestrator library
// as an external dependency. Since the public API is still being developed,
// this example demonstrates the intended usage patterns.

func main() {
	fmt.Println("🚀 Go Orchestrator External Usage Example")
	fmt.Println("==========================================")

	// Create logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create orchestrator configuration
	config := orchestrator.DefaultOrchestratorConfig()
	config.StartupTimeout = 15 * time.Second
	config.ShutdownTimeout = 10 * time.Second
	config.EnableMetrics = true
	config.EnableTracing = false

	fmt.Printf("📋 Configuration:\n")
	fmt.Printf("  - Startup Timeout: %v\n", config.StartupTimeout)
	fmt.Printf("  - Shutdown Timeout: %v\n", config.ShutdownTimeout)
	fmt.Printf("  - Health Check Interval: %v\n", config.HealthCheckInterval)
	fmt.Printf("  - Enable Metrics: %v\n", config.EnableMetrics)
	fmt.Printf("  - Enable Tracing: %v\n", config.EnableTracing)

	// In a real external project, you would:
	// 1. Create a logger adapter
	// logger := logger.NewSlogAdapter(logger)
	//
	// 2. Create the orchestrator
	// orch, err := orchestrator.NewOrchestrator(config, logger)
	// if err != nil {
	//     panic(err)
	// }
	//
	// 3. Register your features
	// features := []orchestrator.Feature{
	//     &DatabaseFeature{name: "database"},
	//     &CacheFeature{name: "cache"},
	//     &APIFeature{name: "api-server"},
	// }
	//
	// for _, feature := range features {
	//     if err := orch.RegisterFeature(feature); err != nil {
	//         panic(err)
	//     }
	// }
	//
	// 4. Start the application
	// ctx := context.Background()
	// if err := orch.Start(ctx); err != nil {
	//     panic(err)
	// }
	//
	// 5. Perform health checks
	// health := orch.HealthCheck(ctx)
	// fmt.Printf("Health Status: %s\n", health.Status)
	//
	// 6. Graceful shutdown
	// defer orch.Stop(ctx)

	fmt.Println("\n📦 Example Features:")
	features := []string{
		"database-service (PostgreSQL)",
		"cache-service (Redis)",
		"api-server (HTTP REST API)",
		"background-worker (Job processor)",
	}

	for i, feature := range features {
		fmt.Printf("  %d. %s\n", i+1, feature)
	}

	fmt.Println("\n🔧 How to implement your features:")
	fmt.Println("1. Create a struct that implements orchestrator.Feature")
	fmt.Println("2. Implement all required methods:")
	fmt.Println("   - GetName() string")
	fmt.Println("   - GetDependencies() []string")
	fmt.Println("   - GetPriority() int")
	fmt.Println("   - RegisterServices(container di.Container) error")
	fmt.Println("   - CreateComponent(container di.Container) (lifecycle.Component, error)")
	fmt.Println("   - GetRetryConfig() *lifecycle.RetryConfig")
	fmt.Println("   - GetMetadata() orchestrator.FeatureMetadata")

	fmt.Println("\n3. Create a component that implements lifecycle.Component")
	fmt.Println("4. Register services in the DI container")
	fmt.Println("5. Handle startup, shutdown, and health checks")

	fmt.Println("\n📚 Complete Examples:")
	fmt.Println("- Basic usage: examples/basic/main.go")
	fmt.Println("- Advanced usage: examples/advanced/main.go")
	fmt.Println("- Command-line tool: cmd/example/main.go")

	fmt.Println("\n📖 Documentation:")
	fmt.Println("- Usage Guide: docs/usage.md")
	fmt.Println("- API Reference: docs/api.md")
	fmt.Println("- Contributing: CONTRIBUTING.md")

	fmt.Println("\n✅ This example demonstrates the intended usage patterns!")
	fmt.Println("Check the documentation and other examples for complete implementations.")
}
