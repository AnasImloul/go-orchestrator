package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/AnasImloul/go-orchestrator/pkg/orchestrator"
)

// This is a simplified example showing how to use the Go Orchestrator library
// as an external dependency. This example focuses on the basic usage patterns
// without complex dependency injection.

func main() {
	fmt.Println("🚀 Go Orchestrator Simple External Usage Example")
	fmt.Println("================================================")

	// Create a simple logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create orchestrator with default configuration
	config := orchestrator.DefaultOrchestratorConfig()
	config.StartupTimeout = 10 * time.Second
	config.ShutdownTimeout = 5 * time.Second

	// Note: In a real external project, you would use:
	// logger := logger.NewSlogAdapter(slog.Default())
	// orch, err := orchestrator.NewOrchestrator(config, logger)
	// 
	// For this example, we'll show the basic structure without
	// the full implementation since the public API is still being developed.

	fmt.Println("📋 Configuration:")
	fmt.Printf("  - Startup Timeout: %v\n", config.StartupTimeout)
	fmt.Printf("  - Shutdown Timeout: %v\n", config.ShutdownTimeout)
	fmt.Printf("  - Health Check Interval: %v\n", config.HealthCheckInterval)
	fmt.Printf("  - Enable Metrics: %v\n", config.EnableMetrics)
	fmt.Printf("  - Enable Tracing: %v\n", config.EnableTracing)

	fmt.Println("\n📦 Available Features:")
	features := []string{
		"database-service",
		"cache-service", 
		"api-server",
		"background-worker",
	}

	for i, feature := range features {
		fmt.Printf("  %d. %s\n", i+1, feature)
	}

	fmt.Println("\n🔧 How to use in your project:")
	fmt.Println("1. Install the library:")
	fmt.Println("   go get github.com/AnasImloul/go-orchestrator")
	fmt.Println("")
	fmt.Println("2. Import the packages:")
	fmt.Println("   import (")
	fmt.Println("       \"github.com/AnasImloul/go-orchestrator/pkg/orchestrator\"")
	fmt.Println("       \"github.com/AnasImloul/go-orchestrator/pkg/di\"")
	fmt.Println("       \"github.com/AnasImloul/go-orchestrator/pkg/lifecycle\"")
	fmt.Println("       \"github.com/AnasImloul/go-orchestrator/pkg/logger\"")
	fmt.Println("   )")
	fmt.Println("")
	fmt.Println("3. Create your features and components")
	fmt.Println("4. Register them with the orchestrator")
	fmt.Println("5. Start the application")

	fmt.Println("\n📚 Documentation:")
	fmt.Println("- Usage Guide: docs/usage.md")
	fmt.Println("- API Reference: docs/api.md")
	fmt.Println("- Examples: examples/ directory")

	fmt.Println("\n✅ Example completed successfully!")
	fmt.Println("Check the documentation for complete implementation examples.")
}
