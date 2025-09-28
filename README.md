# Go Orchestrator

A powerful Go library for service registry, dependency injection, and lifecycle management.

## Features

- **Service Registry**: Type-safe service registration with support for singletons, scoped, and transient lifetimes
- **Lifecycle Management**: Automatic startup/shutdown ordering based on dependencies
- **Dependency Injection**: Coordinate multiple services with proper dependency resolution
- **Health Checking**: Built-in health monitoring for all services
- **Generic Helpers**: Type-safe resolution with `orchestrator.ResolveType[T]()` syntax
- **DAG-based Dependencies**: Directed Acyclic Graph for proper dependency ordering
- **Automatic Discovery**: Auto-discover dependencies from factory function parameters
- **Extensible**: Plugin architecture for custom services and lifecycle management

## Project Structure

This project follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout):

```
go-orchestrator/
├── internal/         # Private implementation (not importable)
│   ├── di/          # Dependency injection container
│   ├── lifecycle/   # Lifecycle management
│   └── logger/      # Logging interface
├── examples/        # Usage examples
│   ├── auto-dependencies/ # Automatic dependency discovery example
│   └── best-syntax/      # Clean API usage example
├── orchestrator.go # Single entry point for the library
├── .gitignore
├── LICENSE
├── CHANGELOG.md
├── CONTRIBUTING.md
├── Makefile
├── go.mod
└── README.md
```

## Public API

The library provides a **single entry point** with a clean, declarative API:

- **`github.com/AnasImloul/go-orchestrator`** - Single entry point with all functionality

**Note**: The `internal/` packages are not accessible to external projects and should not be imported.

## Installation

```bash
go get github.com/AnasImloul/go-orchestrator
```

### Using Specific Version

```bash
# Get latest version
go get github.com/AnasImloul/go-orchestrator@latest

# Get specific version
go get github.com/AnasImloul/go-orchestrator@v1.0.0
```

## Quick Start

```go
package main

import (
    "context"
    
    "github.com/AnasImloul/go-orchestrator"
)

func main() {
    // Create service registry
    registry := orchestrator.New()
    
    // Register service definitions declaratively
    registry.Register(
        orchestrator.WithLifecycleFor[DatabaseService](
            orchestrator.NewServiceWithInstance("database",
                DatabaseService(&databaseService{host: "localhost", port: 5432}),
                orchestrator.Singleton,
            ),
            registry,
        ).
            WithStartFor(func(db DatabaseService) error { return db.Connect() }).
            WithStopFor(func(db DatabaseService) error { return db.Disconnect() }).
            Build(),
    )
    
    // Start service registry
    ctx := context.Background()
    if err := registry.Start(ctx); err != nil {
        panic(err)
    }
    
    // Service registry is now running...
    
    // Graceful shutdown
    registry.Stop(ctx)
}
```

## Key Features

### Declarative API
- **Single entry point**: Import only `github.com/AnasImloul/go-orchestrator`
- **Fluent interface**: Chain method calls for clean, readable code
- **Type-safe**: Generic helpers for service resolution
- **Interface-based DI**: Enforces dependency on interfaces, not concrete types
- **Parallel execution**: Independent components start simultaneously for better performance
- **Less verbose**: Minimal boilerplate code

### Service Registry & Dependency Injection
```go
// Define interfaces (best practice)
type DatabaseService interface {
    Connect() error
    Disconnect() error
}

// Register service definitions declaratively

// Approach 1: Using service instance (factory-based)
registry.Register(
    orchestrator.NewServiceWithInstance("database",
        DatabaseService(&databaseService{host: "localhost", port: 5432}),
        orchestrator.Singleton,
    ),
)

// Approach 2: Using service factory (recommended for complex dependencies)
registry.Register(
    orchestrator.NewServiceWithFactory("database",
        func(ctx context.Context, container *orchestrator.Container) (DatabaseService, error) {
            return &databaseService{host: "localhost", port: 5432}, nil
        },
        orchestrator.Singleton,
    ),
)

// Resolve services by interface (enforced by library)
service, err := orchestrator.ResolveType[DatabaseService](container)

// ❌ This will fail at runtime - concrete types not allowed
// service, err := orchestrator.ResolveType[*databaseService](container)
```

### Lifecycle Management
```go
// Automatic startup/shutdown ordering based on dependencies
// Independent services start in parallel for better performance
registry.Register(
    orchestrator.NewServiceWithInstance("database",
        DatabaseService(&databaseService{host: "localhost", port: 5432}),
        orchestrator.Singleton,
    ).
        WithDependencies("config"). // Depends on config
        WithLifecycle(
            orchestrator.NewLifecycle().
                WithStart(startFunc).
                WithStop(stopFunc).
                WithHealth(healthFunc),
        ),
)
```

### Parallel Execution

The service registry automatically groups services by dependency level and starts independent services in parallel:

```go
// These three services have no dependencies - they start in parallel at level 0
registry.Register(orchestrator.NewServiceWithInstance("cache", CacheService(&cacheService{}), orchestrator.Singleton))
registry.Register(orchestrator.NewServiceWithInstance("metrics", MetricsService(&metricsService{}), orchestrator.Singleton)) 
registry.Register(orchestrator.NewServiceWithInstance("logging", LoggingService(&loggingService{}), orchestrator.Singleton))

// This service depends on all three - it starts at level 1 after they're ready
registry.Register(
    orchestrator.NewServiceWithFactory("api",
        func(ctx context.Context, container *orchestrator.Container) (APIService, error) {
            cache, _ := orchestrator.ResolveType[CacheService](container)
            metrics, _ := orchestrator.ResolveType[MetricsService](container)
            logging, _ := orchestrator.ResolveType[LoggingService](container)
            return &apiService{cache: cache, metrics: metrics, logging: logging}, nil
        },
        orchestrator.Singleton,
    ).WithDependencies("cache", "metrics", "logging"),
)
```

**Execution Flow:**
- **Level 0**: cache, metrics, logging start simultaneously
- **Level 1**: api starts after all dependencies are ready

This provides significant performance improvements over sequential startup.

## Service Lifetimes

The library supports three service lifetimes:

- **Singleton**: One instance for the entire application lifecycle
- **Scoped**: One instance per scope (request/operation) - requires proper scope management
- **Transient**: New instance created every time the service is resolved

### Lifetime Behavior

- **Singleton**: Same instance returned every time, shared across the entire application
- **Scoped**: Same instance within a scope, different instances across scopes
- **Transient**: New instance created for each resolution, even within the same scope

### Service Registration Methods

#### New Ultra-Clean API Benefits

The new API provides several key improvements:

- **🎯 Type Safety**: Full compile-time type checking with generics
- **📝 Declarative**: Clean, readable syntax that expresses intent clearly
- **🔧 Less Verbose**: 60% reduction in boilerplate code
- **⚡ Intuitive**: Natural fluent API that's easy to understand
- **🛡️ Robust**: Runtime type verification and error handling
- **🏭 Factory-Based**: All services use factories for consistent lifetime management
- **🔄 Backward Compatible**: All existing code continues to work

#### Service Registration (Best Syntax - Recommended)

**Ultra-Clean Syntax (Factory-Based):**
```go
// One-liner for simple services with type-safe component lifecycle
app.AddFeature(
    orchestrator.WithComponentFor[DatabaseService](
        orchestrator.NewFeatureWithInstance("database", DatabaseService(&databaseService{host: "localhost", port: 5432}), orchestrator.Singleton),
        app,
    ).
        WithStartFor(func(db DatabaseService) error { return db.Connect() }).
        WithStopFor(func(db DatabaseService) error { return db.Disconnect() }).
        Build(),
)
```

**Factory-Based Services with Dependencies:**
```go
// Clean factory syntax with automatic dependency resolution
app.AddFeature(
    orchestrator.WithComponentFor[APIService](
        orchestrator.NewFeatureWithFactory("api", 
            func(ctx context.Context, container *orchestrator.Container) (APIService, error) {
                db, _ := orchestrator.ResolveType[DatabaseService](container)
                cache, _ := orchestrator.ResolveType[CacheService](container)
                return &apiService{port: 8080, db: db, cache: cache}, nil
            }, 
            orchestrator.Singleton,
        ).WithDependencies("database", "cache"),
        app,
    ).
        WithStartFor(func(api APIService) error { return api.Start() }).
        WithStopFor(func(api APIService) error { return api.Stop() }).
        WithHealthFor(func(api APIService) orchestrator.HealthStatus {
            return orchestrator.HealthStatus{Status: api.Health(), Message: "API server is running"}
        }).
        Build(),
)
```

**Automatic Dependency Discovery (NEW!):**
```go
// Ultra-clean syntax with automatic dependency injection AND discovery
app.AddFeature(
    orchestrator.NewFeatureWithAutoFactory[APIService]("api",
        func(db DatabaseService, cache CacheService) APIService {
            return &apiService{port: 8080, db: db, cache: cache}
        },
        orchestrator.Singleton,
    ).
        WithComponent(
            orchestrator.NewComponent().
                WithStart(func(ctx context.Context, container *orchestrator.Container) error {
                    api, _ := orchestrator.ResolveType[APIService](container)
                    return api.Start()
                }).
                WithStop(func(ctx context.Context) error {
                    api, _ := orchestrator.ResolveType[APIService](app.Container())
                    return api.Stop()
                }),
        ),
)
```

#### Automatic Dependency Discovery Benefits

The new `NewFeatureWithAutoFactory` provides:

- **🎯 Zero Boilerplate**: No manual dependency resolution OR declaration needed
- **🔍 Type-Safe**: Compile-time verification of dependency types
- **⚡ Reflection-Based**: Automatic parameter scanning and injection
- **🛡️ Error Handling**: Clear error messages for missing dependencies
- **📝 Clean Syntax**: Factory functions only need to declare their dependencies
- **🔗 Auto-Discovery**: Dependencies are automatically discovered from function parameters
- **🚀 Zero Configuration**: No need for `WithDependencies()` calls

#### Factory-Only Approach Benefits

All services now use factories internally, which provides:

- **Consistent Lifetime Management**: Transient services create new instances via factory calls
- **No Cloning Complexity**: Eliminates the need for deep cloning logic
- **Simplified Architecture**: Single code path for all service creation
- **Better Performance**: Factory calls are more efficient than cloning
- **Clearer Intent**: Factory functions make service creation explicit

#### Legacy Syntax (Still Supported)
```go
// Original verbose syntax - still works but not recommended for new code
app.AddFeature(
    orchestrator.WithServiceFactory[DatabaseService](func(ctx context.Context, container *orchestrator.Container) (DatabaseService, error) {
        return &databaseService{host: "localhost", port: 5432}, nil
    })(
        orchestrator.NewFeature("database"),
    ).
        WithLifetime(orchestrator.Singleton).
        WithComponent(
            orchestrator.NewComponent().
                WithStart(orchestrator.WithStartFunc[DatabaseService](func(db DatabaseService) error {
                    return db.Connect()
                })).
                WithStop(orchestrator.WithStopFuncWithApp[DatabaseService](app, func(db DatabaseService) error {
                    return db.Disconnect()
                })),
        ),
)
```

#### Factory Registration (For Complex Dependencies)
```go
// Register a service factory for services with dependencies
app.AddFeature(
    orchestrator.WithServiceFactory[APIService](
        func(ctx context.Context, container *orchestrator.Container) (APIService, error) {
            db, err := orchestrator.ResolveType[DatabaseService](container)
            if err != nil {
                return nil, err
            }
            return &apiService{port: 8080, db: db}, nil
        },
    )(
        orchestrator.NewFeature("api").
            WithDependencies("database"),
    ).
        WithLifetime(orchestrator.Singleton),
)
```

**When to use each approach:**
- **`WithService[T]()`**: For simple services without dependencies (clean, declarative)
- **`WithServiceFactory[T]()`**: For services with dependencies that need to be resolved from the container (type-safe, generic)

**Helper Functions for Reduced Verbosity:**
- **`WithStartFunc[T]()`**: Eliminates repetitive service resolution in start methods
- **`WithStopFuncWithApp[T]()`**: Eliminates repetitive service resolution in stop methods  
- **`WithHealthFunc[T]()`**: Eliminates repetitive service resolution in health methods

#### Named Services
```go
// Register multiple services with the same interface type using names
app.AddFeature(
    orchestrator.NewFeature("databases").
        WithNamedService(
            "primary-db",
            reflect.TypeOf((*DatabaseService)(nil)).Elem(),
            func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
                return &databaseService{host: "primary", port: 5432}, nil
            },
            orchestrator.Singleton,
        ).
        WithNamedService(
            "secondary-db",
            reflect.TypeOf((*DatabaseService)(nil)).Elem(),
            func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
                return &databaseService{host: "secondary", port: 5433}, nil
            },
            orchestrator.Singleton,
        ),
)

// Resolve named services
primaryDB, _ := orchestrator.ResolveNamedType[DatabaseService](container, "primary-db")
secondaryDB, _ := orchestrator.ResolveNamedType[DatabaseService](container, "secondary-db")
```

### Scope Management

```go
// Create a scope for scoped services
scopedContainer := container.CreateScope()
defer scopedContainer.Dispose()

// Resolve services within the scope
service, _ := orchestrator.ResolveType[MyService](scopedContainer)
```

### Limitations and Best Practices

1. **Type Collision**: You cannot register multiple services with the same interface type without using named services
2. **Transient Cloning**: Transient services use deep cloning for instance registration, which may not work for all types
3. **Factory Registration**: Use factory registration for better control over service creation and lifetime behavior
4. **Scope Lifecycle**: Always dispose scopes when done to prevent memory leaks
5. **Interface Enforcement**: The library enforces that services are resolved as interfaces, not concrete types

## Using as External Dependency

### 1. Install the Library

```bash
go get github.com/AnasImloul/go-orchestrator
```

### 2. Import the Library

```go
import "github.com/AnasImloul/go-orchestrator"
```

> **Note**: The library provides a single entry point with all functionality. No need to import multiple packages.

### 3. Create Your Service Registry

```go
package main

import (
    "context"
    "github.com/AnasImloul/go-orchestrator"
)

// Define your service interfaces
type DatabaseService interface {
    Connect() error
    Disconnect() error
}

type APIService interface {
    Start() error
    Stop() error
}

// Implement your services
type databaseService struct {
    host string
    port int
}

func (d *databaseService) Connect() error {
    // Implementation
    return nil
}

func (d *databaseService) Disconnect() error {
    // Implementation
    return nil
}

type apiService struct {
    port int
    db   DatabaseService
}

func (a *apiService) Start() error {
    // Implementation
    return nil
}

func (a *apiService) Stop() error {
    // Implementation
    return nil
}

func main() {
    // Create service registry
    registry := orchestrator.New()
    
    // Register services
    registry.Register(
        orchestrator.NewServiceWithInstance("database",
            DatabaseService(&databaseService{host: "localhost", port: 5432}),
            orchestrator.Singleton,
        ),
    )
    
    registry.Register(
        orchestrator.NewServiceWithFactory("api",
            func(ctx context.Context, container *orchestrator.Container) (APIService, error) {
                db, err := orchestrator.ResolveType[DatabaseService](container)
                if err != nil {
                    return nil, err
                }
                return &apiService{port: 8080, db: db}, nil
            },
            orchestrator.Singleton,
        ).WithDependencies("database"),
    )
    
    // Start the service registry
    ctx := context.Background()
    if err := registry.Start(ctx); err != nil {
        panic(err)
    }
    
    // Your application is now running...
    
    // Graceful shutdown
    registry.Stop(ctx)
}
```

## Examples

### Automatic Dependency Discovery

The `auto-dependencies` example demonstrates how the library automatically discovers and injects dependencies based on factory function parameters:

```bash
go run examples/auto-dependencies/main.go
```

### Best Syntax Example

The `best-syntax` example shows the cleanest API usage patterns with different service lifetimes:

```bash
go run examples/best-syntax/main.go
```

### Running Examples

```bash
# Automatic dependency discovery example
go run examples/auto-dependencies/main.go

# Best syntax example (recommended)
go run examples/best-syntax/main.go
```

## Development

### Building

```bash
# Build the example application
make build

# Run tests
make test

# Run with coverage
make test-coverage

# Format code
make fmt

# Lint code
make lint

# Run all checks
make check
```

### Project Layout

This project follows Go standard project layout conventions:

- **`pkg/`**: Public API that can be imported by external projects
- **`internal/`**: Private implementation details that cannot be imported externally
- **`cmd/`**: Example applications and command-line tools
- **`examples/`**: Usage examples and tutorials
- **`docs/`**: Documentation and API references

## API Documentation

See [docs/api.md](docs/api.md) for comprehensive API documentation.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for contribution guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
# Test commit for avatar verification
