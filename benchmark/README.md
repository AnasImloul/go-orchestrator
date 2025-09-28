# Go Orchestrator Performance Benchmarks

This directory contains performance tests and benchmarks for the Go Orchestrator library. The benchmarks use metaprogramming to generate large DAGs (Directed Acyclic Graphs) of services and measure start/stop times.

## Files

- `performance_benchmark.go` - Basic performance test with CLI argument for service count
- `advanced_performance_benchmark.go` - Advanced performance test with different DAG patterns
- `benchmark_suite_main.go` - Comprehensive benchmark suite with multiple test configurations
- `README.md` - This documentation file

## Quick Start

### Basic Performance Test

Test with a specific number of services:

```bash
# Test with 1000 services
go run performance_benchmark.go 1000

# Test with 5000 services (will prompt for confirmation)
go run performance_benchmark.go 5000
```

### Advanced Performance Test

Test with different DAG patterns:

```bash
# Test with 1000 services using layered pattern (default)
go run advanced_performance_benchmark.go 1000

# Test with different patterns
go run advanced_performance_benchmark.go 1000 linear
go run advanced_performance_benchmark.go 1000 tree
go run advanced_performance_benchmark.go 1000 star
go run advanced_performance_benchmark.go 1000 mesh
go run advanced_performance_benchmark.go 1000 random
```

### Benchmark Suite

Run comprehensive benchmarks with multiple configurations:

```bash
# Test with 100, 500, 1000 services (default patterns)
go run benchmark_suite_main.go "100,500,1000"

# Test 1000 services with different patterns
go run benchmark_suite_main.go "1000:linear,tree,layered"

# Test multiple counts and patterns with 3 iterations each
go run benchmark_suite_main.go "100,500,1000:linear,tree:3"
```

## DAG Patterns

The benchmarks support different DAG patterns to test various dependency scenarios:

### Linear Pattern
- Each service depends on the previous one
- Creates a chain: Service1 → Service2 → Service3 → ...
- Tests sequential startup performance

### Tree Pattern
- Binary tree structure
- Each service depends on its children
- Tests hierarchical dependency resolution

### Layered Pattern (Default)
- Multiple layers with dependencies between layers
- Services in layer N depend on services in layer N-1
- Most realistic for real-world applications

### Star Pattern
- One central service with many dependencies
- All other services depend on the central hub
- Tests single-point-of-failure scenarios

### Mesh Pattern
- Highly connected mesh
- Each service depends on multiple other services
- Tests complex dependency resolution

### Random Pattern
- Random dependencies (but acyclic)
- Tests worst-case dependency scenarios

## Performance Metrics

The benchmarks measure:

- **DAG Generation Time**: Time to generate the service dependency graph
- **Registration Time**: Time to register all services with the orchestrator
- **Start Time**: Time to start all services (including dependency resolution)
- **Stop Time**: Time to stop all services
- **Total Time**: End-to-end execution time
- **Throughput**: Services started/stopped per second

## Sample Output

```
Go Orchestrator Performance Test
=====================================
Starting performance test with 1000 services...
Generating DAG with 1000 services...
   DAG generated in 2.5ms
Registering 1000 services...
   Services registered in 15.2ms
Starting 1000 services...
   All services started in 1.2s
Running services for 2 seconds...
Stopping 1000 services...
   All services stopped in 800ms

Performance Summary:
   Services: 1000
   DAG Generation: 2.5ms
   Registration: 15.2ms
   Start Time: 1.2s
   Stop Time: 800ms
   Total Time: 2.1s
   Start Throughput: 833.33 services/second
   Stop Throughput: 1250.00 services/second

Performance test completed successfully!
```

## Benchmark Suite Output

The benchmark suite generates JSON results and detailed reports:

```json
[
  {
    "service_count": 1000,
    "pattern": "layered",
    "dag_generation_ms": 2500000,
    "registration_ms": 15200000,
    "start_time_ms": 1200000000,
    "stop_time_ms": 800000000,
    "total_time_ms": 2100000000,
    "start_throughput_per_sec": 833.33,
    "stop_throughput_per_sec": 1250.00,
    "success": true
  }
]
```

## Performance Expectations

Based on testing, here are typical performance expectations:

| Service Count | Start Time | Stop Time | Start Throughput | Stop Throughput |
|---------------|------------|-----------|------------------|-----------------|
| 100           | ~50ms      | ~30ms     | ~2000/s          | ~3300/s         |
| 500           | ~200ms     | ~150ms    | ~2500/s          | ~3300/s         |
| 1000          | ~400ms     | ~300ms    | ~2500/s          | ~3300/s         |
| 5000          | ~2s        | ~1.5s     | ~2500/s          | ~3300/s         |

*Note: Performance may vary based on system resources and DAG complexity.*

## Memory Usage

The benchmarks simulate realistic service workloads:
- Each service has configurable startup/shutdown delays (1-10ms)
- Services perform simulated CPU work during startup/shutdown
- Memory usage scales linearly with service count

## Troubleshooting

### Out of Memory
If you encounter out-of-memory errors with large service counts:
- Reduce the service count
- Increase system memory
- Use simpler DAG patterns (linear, star)

### Timeout Errors
If services fail to start/stop within the timeout:
- Increase the timeout values in the code
- Use fewer services
- Check system performance

### Build Errors
If you encounter build errors:
- Ensure you're in the project root directory
- Run `go mod tidy` to update dependencies
- Check that all imports are correct

## Contributing

To add new benchmark patterns or metrics:

1. Add new pattern generation logic to the `DAGGenerator`
2. Update the pattern constants and parsing logic
3. Add new metrics to the `BenchmarkResult` struct
4. Update the report generation logic

## License

This benchmark suite is part of the Go Orchestrator project and follows the same license terms.
