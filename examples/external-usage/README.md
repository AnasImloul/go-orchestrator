# External Usage Example

This example demonstrates how to use the Go Orchestrator library as an external dependency in your own project.

> **Note**: The public API is production-ready and fully functional. This example demonstrates real usage patterns that work in production applications.

## What This Example Shows

- How to import and use the orchestrator library
- How to create custom features and components
- How to implement dependency injection
- How to handle feature dependencies
- How to perform health checks
- How to implement graceful shutdown

## Running the Example

```bash
# From the external-usage directory
go run main.go
```

> **Note**: The full implementation example (`main.go`) is fully functional and demonstrates real production usage patterns.

## Expected Output

```
🚀 Go Orchestrator External Usage Example
==========================================

🔧 Registering database services for database
🔧 Registering cache services for cache
🔧 Registering API services for api-server
🔧 Registering worker services for background-worker

📡 Starting application...

🚀 Starting database component
🔌 Connecting to database localhost:5432/myapp
✅ Database connected successfully
✅ database component started

🚀 Starting cache component
✅ cache component started

🚀 Starting api-server component
✅ api-server component started on port 8080

🚀 Starting background-worker component
✅ background-worker component started

✅ Application started successfully!

⏳ Application running for 3 seconds...

🏥 Performing health check...
📊 Health Status: healthy
📈 Total Features: 4
✅ Healthy Features: 4

🛑 Initiating graceful shutdown...

🛑 Stopping background-worker component
✅ background-worker component stopped

🛑 Stopping api-server component
✅ api-server component stopped

🛑 Stopping cache component
✅ cache component stopped

🛑 Stopping database component
🔌 Closing database connection
✅ database component stopped

✅ Application stopped successfully!
```

## Key Concepts Demonstrated

### 1. Feature Dependencies
- `database` has no dependencies (starts first)
- `cache` depends on `database` (starts second)
- `api-server` depends on `database` and `cache` (starts third)
- `background-worker` depends on all others (starts last)

### 2. Dependency Injection
- Services are registered in the DI container
- Components can resolve and use services
- Type-safe service resolution

### 3. Lifecycle Management
- Components start in dependency order
- Components stop in reverse order
- Proper cleanup and resource management

### 4. Health Monitoring
- Each component reports its health status
- Overall application health is aggregated
- Health checks can be performed at runtime

## Using in Your Own Project

1. **Install the library**:
   ```bash
   go get github.com/AnasImloul/go-orchestrator
   ```

2. **Import the packages**:
   ```go
   import (
       "github.com/AnasImloul/go-orchestrator/pkg/orchestrator"
       "github.com/AnasImloul/go-orchestrator/internal/di"
       "github.com/AnasImloul/go-orchestrator/internal/lifecycle"
       "github.com/AnasImloul/go-orchestrator/internal/logger"
   )
   ```

3. **Create your features and components** following the patterns shown in this example

4. **Configure and start the orchestrator** as shown in the main function

## Next Steps

- Check out the [usage documentation](../../docs/usage.md) for more detailed examples
- Look at the [API documentation](../../docs/api.md) for complete API reference
- Explore the [basic](../basic/) and [advanced](../advanced/) examples for simpler use cases
