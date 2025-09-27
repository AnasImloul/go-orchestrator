package main

import (
	"fmt"
	"log/slog"
	"os"
)

// This is a demonstration of how to use the Go Orchestrator library
// as an external dependency. Since the public API is still being developed,
// this example shows the intended usage patterns without importing the library.

func main() {
	fmt.Println("🚀 Go Orchestrator External Usage Demo")
	fmt.Println("======================================")

	// Create logger (for demonstration)
	_ = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	fmt.Println("📋 Step 1: Install the library")
	fmt.Println("   go get github.com/AnasImloul/go-orchestrator")

	fmt.Println("\n📦 Step 2: Import the packages")
	fmt.Println("   import (")
	fmt.Println("       \"github.com/AnasImloul/go-orchestrator/pkg/orchestrator\"")
	fmt.Println("       \"github.com/AnasImloul/go-orchestrator/pkg/di\"")
	fmt.Println("       \"github.com/AnasImloul/go-orchestrator/pkg/lifecycle\"")
	fmt.Println("       \"github.com/AnasImloul/go-orchestrator/pkg/logger\"")
	fmt.Println("   )")

	fmt.Println("\n🔧 Step 3: Create your application structure")
	fmt.Println("   your-project/")
	fmt.Println("   ├── cmd/")
	fmt.Println("   │   └── your-app/")
	fmt.Println("   │       └── main.go")
	fmt.Println("   ├── internal/")
	fmt.Println("   │   ├── features/")
	fmt.Println("   │   │   ├── database/")
	fmt.Println("   │   │   ├── cache/")
	fmt.Println("   │   │   └── api/")
	fmt.Println("   │   └── services/")
	fmt.Println("   ├── go.mod")
	fmt.Println("   └── go.sum")

	fmt.Println("\n⚙️ Step 4: Configure the orchestrator")
	fmt.Println("   config := orchestrator.DefaultOrchestratorConfig()")
	fmt.Println("   config.StartupTimeout = 30 * time.Second")
	fmt.Println("   config.ShutdownTimeout = 15 * time.Second")
	fmt.Println("   config.EnableMetrics = true")

	fmt.Println("\n🏗️ Step 5: Create your features")
	fmt.Println("   type DatabaseFeature struct {")
	fmt.Println("       name string")
	fmt.Println("   }")
	fmt.Println("   ")
	fmt.Println("   func (f *DatabaseFeature) GetName() string {")
	fmt.Println("       return f.name")
	fmt.Println("   }")
	fmt.Println("   ")
	fmt.Println("   // ... implement other required methods")

	fmt.Println("\n🔌 Step 6: Implement your components")
	fmt.Println("   type DatabaseComponent struct {")
	fmt.Println("       name string")
	fmt.Println("   }")
	fmt.Println("   ")
	fmt.Println("   func (c *DatabaseComponent) Start(ctx context.Context) error {")
	fmt.Println("       // Initialize your component")
	fmt.Println("       return nil")
	fmt.Println("   }")
	fmt.Println("   ")
	fmt.Println("   // ... implement other required methods")

	fmt.Println("\n🚀 Step 7: Start your application")
	fmt.Println("   orch, err := orchestrator.NewOrchestrator(config, logger)")
	fmt.Println("   if err != nil {")
	fmt.Println("       panic(err)")
	fmt.Println("   }")
	fmt.Println("   ")
	fmt.Println("   // Register features")
	fmt.Println("   orch.RegisterFeature(&DatabaseFeature{name: \"database\"})")
	fmt.Println("   ")
	fmt.Println("   // Start application")
	fmt.Println("   ctx := context.Background()")
	fmt.Println("   if err := orch.Start(ctx); err != nil {")
	fmt.Println("       panic(err)")
	fmt.Println("   }")
	fmt.Println("   ")
	fmt.Println("   // Graceful shutdown")
	fmt.Println("   defer orch.Stop(ctx)")

	fmt.Println("\n📚 Available Documentation:")
	fmt.Println("   - Usage Guide: docs/usage.md")
	fmt.Println("   - External Usage Guide: docs/external-usage.md")
	fmt.Println("   - API Reference: docs/api.md")
	fmt.Println("   - Contributing Guide: CONTRIBUTING.md")

	fmt.Println("\n🎯 Working Examples:")
	fmt.Println("   - Basic: examples/basic/main.go")
	fmt.Println("   - Advanced: examples/advanced/main.go")
	fmt.Println("   - Command-line: cmd/example/main.go")

	fmt.Println("\n⚠️ Current Status:")
	fmt.Println("   The public API is under development.")
	fmt.Println("   The examples in this directory show intended usage patterns.")
	fmt.Println("   For working examples, see the other examples in the repository.")

	fmt.Println("\n✅ Demo completed!")
	fmt.Println("Check the documentation for complete implementation details.")
}
