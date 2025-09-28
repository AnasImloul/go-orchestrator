package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// BenchmarkSuite runs multiple performance tests and generates reports
type BenchmarkSuite struct {
	results []BenchmarkResult
}

// BenchmarkResult holds the result of a single benchmark
type BenchmarkResult struct {
	ServiceCount    int           `json:"service_count"`
	Pattern         string        `json:"pattern"`
	DAGGeneration   time.Duration `json:"dag_generation_ms"`
	Registration    time.Duration `json:"registration_ms"`
	StartTime       time.Duration `json:"start_time_ms"`
	StopTime        time.Duration `json:"stop_time_ms"`
	TotalTime       time.Duration `json:"total_time_ms"`
	StartThroughput float64       `json:"start_throughput_per_sec"`
	StopThroughput  float64       `json:"stop_throughput_per_sec"`
	Success         bool          `json:"success"`
	ErrorMessage    string        `json:"error_message,omitempty"`
}

// BenchmarkConfig holds configuration for benchmark runs
type BenchmarkConfig struct {
	ServiceCounts []int
	Patterns      []string
	Iterations    int
	OutputFile    string
}

// MockService implements the Service interface for benchmarking
type MockService struct {
	ID         string
	StartDelay time.Duration
	StopDelay  time.Duration
	Workload   int
}

func (m *MockService) Start(ctx context.Context) error {
	for i := 0; i < m.Workload; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_ = i * i // Simulate work
		}
	}
	time.Sleep(m.StartDelay)
	return nil
}

func (m *MockService) Stop(ctx context.Context) error {
	for i := 0; i < m.Workload/2; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_ = i * i // Simulate work
		}
	}
	time.Sleep(m.StopDelay)
	return nil
}

func (m *MockService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: fmt.Sprintf("Service %s is healthy", m.ID),
	}
}

// ServiceNode represents a node in the DAG
type ServiceNode struct {
	ID           string
	Dependencies []string
	StartTime    time.Duration
	StopTime     time.Duration
	Workload     int
}

// DAGGenerator generates different types of DAGs
type DAGGenerator struct {
	serviceCount int
	pattern      string
}

func NewDAGGenerator(serviceCount int, pattern string) *DAGGenerator {
	return &DAGGenerator{
		serviceCount: serviceCount,
		pattern:      pattern,
	}
}

func (g *DAGGenerator) Generate() []ServiceNode {
	switch g.pattern {
	case "linear":
		return g.generateLinear()
	case "tree":
		return g.generateTree()
	case "layered":
		return g.generateLayered()
	case "star":
		return g.generateStar()
	case "mesh":
		return g.generateMesh()
	case "random":
		return g.generateRandom()
	default:
		return g.generateLayered()
	}
}

func (g *DAGGenerator) generateLinear() []ServiceNode {
	nodes := make([]ServiceNode, g.serviceCount)
	for i := 0; i < g.serviceCount; i++ {
		var dependencies []string
		if i > 0 {
			dependencies = []string{fmt.Sprintf("service_%d", i-1)}
		}

		nodes[i] = ServiceNode{
			ID:           fmt.Sprintf("service_%d", i),
			Dependencies: dependencies,
			StartTime:    time.Duration(1+i%10) * time.Millisecond,
			StopTime:     time.Duration(1+i%5) * time.Millisecond,
			Workload:     100 + i%500,
		}
	}
	return nodes
}

func (g *DAGGenerator) generateTree() []ServiceNode {
	nodes := make([]ServiceNode, g.serviceCount)

	for i := 0; i < g.serviceCount; i++ {
		var dependencies []string

		leftChild := 2*i + 1
		rightChild := 2*i + 2

		if leftChild < g.serviceCount {
			dependencies = append(dependencies, fmt.Sprintf("service_%d", leftChild))
		}
		if rightChild < g.serviceCount {
			dependencies = append(dependencies, fmt.Sprintf("service_%d", rightChild))
		}

		nodes[i] = ServiceNode{
			ID:           fmt.Sprintf("service_%d", i),
			Dependencies: dependencies,
			StartTime:    time.Duration(1+i%10) * time.Millisecond,
			StopTime:     time.Duration(1+i%5) * time.Millisecond,
			Workload:     100 + i%500,
		}
	}
	return nodes
}

func (g *DAGGenerator) generateLayered() []ServiceNode {
	nodes := make([]ServiceNode, g.serviceCount)

	layers := int(float64(g.serviceCount) * 0.3)
	if layers < 2 {
		layers = 2
	}

	servicesPerLayer := g.serviceCount / layers
	remainingServices := g.serviceCount % layers

	serviceIndex := 0

	for layer := 0; layer < layers; layer++ {
		servicesInThisLayer := servicesPerLayer
		if layer < remainingServices {
			servicesInThisLayer++
		}

		for i := 0; i < servicesInThisLayer && serviceIndex < g.serviceCount; i++ {
			var dependencies []string

			if layer > 0 {
				prevLayerStart := (layer - 1) * servicesPerLayer
				if layer-1 < remainingServices {
					prevLayerStart += layer - 1
				} else {
					prevLayerStart += remainingServices
				}

				prevLayerSize := servicesPerLayer
				if layer-1 < remainingServices {
					prevLayerSize++
				}

				dependencyCount := 1 + (i % 3)
				if dependencyCount > prevLayerSize {
					dependencyCount = prevLayerSize
				}

				for j := 0; j < dependencyCount; j++ {
					depIndex := prevLayerStart + (j * prevLayerSize / dependencyCount)
					if depIndex < serviceIndex {
						dependencies = append(dependencies, fmt.Sprintf("service_%d", depIndex))
					}
				}
			}

			nodes[serviceIndex] = ServiceNode{
				ID:           fmt.Sprintf("service_%d", serviceIndex),
				Dependencies: dependencies,
				StartTime:    time.Duration(1+serviceIndex%10) * time.Millisecond,
				StopTime:     time.Duration(1+serviceIndex%5) * time.Millisecond,
				Workload:     100 + serviceIndex%500,
			}
			serviceIndex++
		}
	}

	return nodes
}

func (g *DAGGenerator) generateStar() []ServiceNode {
	nodes := make([]ServiceNode, g.serviceCount)

	nodes[0] = ServiceNode{
		ID:           "service_0",
		Dependencies: []string{},
		StartTime:    time.Duration(1) * time.Millisecond,
		StopTime:     time.Duration(1) * time.Millisecond,
		Workload:     100,
	}

	for i := 1; i < g.serviceCount; i++ {
		nodes[i] = ServiceNode{
			ID:           fmt.Sprintf("service_%d", i),
			Dependencies: []string{"service_0"},
			StartTime:    time.Duration(1+i%10) * time.Millisecond,
			StopTime:     time.Duration(1+i%5) * time.Millisecond,
			Workload:     100 + i%500,
		}
	}

	return nodes
}

func (g *DAGGenerator) generateMesh() []ServiceNode {
	nodes := make([]ServiceNode, g.serviceCount)

	for i := 0; i < g.serviceCount; i++ {
		var dependencies []string

		maxDeps := g.serviceCount / 4
		if maxDeps > 10 {
			maxDeps = 10
		}
		if maxDeps < 2 {
			maxDeps = 2
		}

		depCount := 1 + (i % maxDeps)
		for j := 0; j < depCount && j < i; j++ {
			depIndex := (i - 1 - j) % i
			if depIndex >= 0 {
				dependencies = append(dependencies, fmt.Sprintf("service_%d", depIndex))
			}
		}

		nodes[i] = ServiceNode{
			ID:           fmt.Sprintf("service_%d", i),
			Dependencies: dependencies,
			StartTime:    time.Duration(1+i%10) * time.Millisecond,
			StopTime:     time.Duration(1+i%5) * time.Millisecond,
			Workload:     100 + i%500,
		}
	}

	return nodes
}

func (g *DAGGenerator) generateRandom() []ServiceNode {
	nodes := make([]ServiceNode, g.serviceCount)

	for i := 0; i < g.serviceCount; i++ {
		var dependencies []string

		maxDeps := 1 + (i % 5)
		for j := 0; j < maxDeps && j < i; j++ {
			depIndex := (i*7 + j*13) % i
			if depIndex >= 0 && depIndex < i {
				dependencies = append(dependencies, fmt.Sprintf("service_%d", depIndex))
			}
		}

		nodes[i] = ServiceNode{
			ID:           fmt.Sprintf("service_%d", i),
			Dependencies: dependencies,
			StartTime:    time.Duration(1+i%10) * time.Millisecond,
			StopTime:     time.Duration(1+i%5) * time.Millisecond,
			Workload:     100 + i%500,
		}
	}

	return nodes
}

// createMockServiceFactory creates a factory function for a mock service
func createMockServiceFactory(node ServiceNode) func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
	return func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
		return &MockService{
			ID:         node.ID,
			StartDelay: node.StartTime,
			StopDelay:  node.StopTime,
			Workload:   node.Workload,
		}, nil
	}
}

// runBenchmark runs a single benchmark test
func runBenchmark(serviceCount int, pattern string) *BenchmarkResult {
	result := &BenchmarkResult{
		ServiceCount: serviceCount,
		Pattern:      pattern,
		Success:      false,
	}

	overallStart := time.Now()

	// Generate DAG
	dagStart := time.Now()
	generator := NewDAGGenerator(serviceCount, pattern)
	nodes := generator.Generate()
	result.DAGGeneration = time.Since(dagStart)

	// Create service registry
	registry := orchestrator.New()

	// Register all services
	registrationStart := time.Now()

	for _, node := range nodes {
		serviceDef := &orchestrator.ServiceDefinition{
			Name:         node.ID,
			Dependencies: node.Dependencies,
			Services: []orchestrator.ServiceConfig{
				{
					Name:     node.ID,
					Type:     reflect.TypeOf((*MockService)(nil)).Elem(),
					Factory:  createMockServiceFactory(node),
					Lifetime: orchestrator.Singleton,
				},
			},
		}

		registry.Register(serviceDef)
	}

	result.Registration = time.Since(registrationStart)

	// Start services
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err := registry.Start(ctx)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to start services: %v", err)
		result.TotalTime = time.Since(overallStart)
		return result
	}

	result.StartTime = time.Since(startTime)

	// Wait briefly
	time.Sleep(100 * time.Millisecond)

	// Stop services
	stopTime := time.Now()

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer stopCancel()

	err = registry.Stop(stopCtx)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to stop services: %v", err)
		result.TotalTime = time.Since(overallStart)
		return result
	}

	result.StopTime = time.Since(stopTime)
	result.TotalTime = time.Since(overallStart)

	// Calculate throughput
	result.StartThroughput = float64(serviceCount) / result.StartTime.Seconds()
	result.StopThroughput = float64(serviceCount) / result.StopTime.Seconds()
	result.Success = true

	return result
}

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite() *BenchmarkSuite {
	return &BenchmarkSuite{
		results: make([]BenchmarkResult, 0),
	}
}

// Run runs the benchmark suite with the given configuration
func (bs *BenchmarkSuite) Run(config BenchmarkConfig) {
	fmt.Printf("Go Orchestrator Benchmark Suite\n")
	fmt.Printf("===================================\n")
	fmt.Printf("Service counts: %v\n", config.ServiceCounts)
	fmt.Printf("Patterns: %v\n", config.Patterns)
	fmt.Printf("Iterations: %d\n", config.Iterations)
	fmt.Printf("Output file: %s\n\n", config.OutputFile)

	totalTests := len(config.ServiceCounts) * len(config.Patterns) * config.Iterations
	currentTest := 0

	for _, serviceCount := range config.ServiceCounts {
		for _, pattern := range config.Patterns {
			for iteration := 0; iteration < config.Iterations; iteration++ {
				currentTest++
				fmt.Printf("[%d/%d] Running benchmark: %d services, %s pattern (iteration %d)\n",
					currentTest, totalTests, serviceCount, pattern, iteration+1)

				result := runBenchmark(serviceCount, pattern)
				bs.results = append(bs.results, *result)

				if result.Success {
					fmt.Printf("   Success: Start=%v, Stop=%v, Throughput=%.1f/s\n",
						result.StartTime, result.StopTime, result.StartThroughput)
				} else {
					fmt.Printf("   Failed: %s\n", result.ErrorMessage)
				}
			}
		}
	}
}

// GenerateReport generates a performance report
func (bs *BenchmarkSuite) GenerateReport() {
	fmt.Printf("\nPerformance Report\n")
	fmt.Printf("====================\n")

	// Group results by service count and pattern
	grouped := make(map[string][]BenchmarkResult)

	for _, result := range bs.results {
		if result.Success {
			key := fmt.Sprintf("%d_%s", result.ServiceCount, result.Pattern)
			grouped[key] = append(grouped[key], result)
		}
	}

	// Calculate averages
	for key, results := range grouped {
		if len(results) == 0 {
			continue
		}

		var totalStart, totalStop, totalDAG, totalReg time.Duration
		var totalStartThroughput, totalStopThroughput float64

		for _, result := range results {
			totalStart += result.StartTime
			totalStop += result.StopTime
			totalDAG += result.DAGGeneration
			totalReg += result.Registration
			totalStartThroughput += result.StartThroughput
			totalStopThroughput += result.StopThroughput
		}

		count := len(results)
		avgStart := totalStart / time.Duration(count)
		avgStop := totalStop / time.Duration(count)
		avgDAG := totalDAG / time.Duration(count)
		avgReg := totalReg / time.Duration(count)
		avgStartThroughput := totalStartThroughput / float64(count)
		avgStopThroughput := totalStopThroughput / float64(count)

		parts := strings.Split(key, "_")
		serviceCount := parts[0]
		pattern := parts[1]

		fmt.Printf("\n%s services, %s pattern (%d iterations):\n", serviceCount, pattern, count)
		fmt.Printf("  DAG Generation: %v\n", avgDAG)
		fmt.Printf("  Registration: %v\n", avgReg)
		fmt.Printf("  Start Time: %v\n", avgStart)
		fmt.Printf("  Stop Time: %v\n", avgStop)
		fmt.Printf("  Start Throughput: %.1f services/second\n", avgStartThroughput)
		fmt.Printf("  Stop Throughput: %.1f services/second\n", avgStopThroughput)
	}
}

// SaveResults saves the benchmark results to a JSON file
func (bs *BenchmarkSuite) SaveResults(filename string) error {
	data, err := json.MarshalIndent(bs.results, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run benchmark_suite.go <config>")
		fmt.Println("Config examples:")
		fmt.Println("  '100,500,1000' - Test with 100, 500, 1000 services")
		fmt.Println("  '1000:linear,tree,layered' - Test 1000 services with different patterns")
		fmt.Println("  '100,500,1000:linear,tree:3' - Test multiple counts, patterns, 3 iterations")
		os.Exit(1)
	}

	configStr := os.Args[1]

	// Parse configuration
	config := BenchmarkConfig{
		ServiceCounts: []int{100, 500, 1000},
		Patterns:      []string{"linear", "tree", "layered"},
		Iterations:    1,
		OutputFile:    "benchmark_results.json",
	}

	// Parse the configuration string
	parts := strings.Split(configStr, ":")

	// Parse service counts
	if len(parts) > 0 && parts[0] != "" {
		countStrs := strings.Split(parts[0], ",")
		config.ServiceCounts = make([]int, len(countStrs))
		for i, countStr := range countStrs {
			count, err := strconv.Atoi(strings.TrimSpace(countStr))
			if err != nil {
				log.Fatalf("Invalid service count: %v", err)
			}
			config.ServiceCounts[i] = count
		}
	}

	// Parse patterns
	if len(parts) > 1 && parts[1] != "" {
		config.Patterns = strings.Split(parts[1], ",")
		for i, pattern := range config.Patterns {
			config.Patterns[i] = strings.TrimSpace(pattern)
		}
	}

	// Parse iterations
	if len(parts) > 2 && parts[2] != "" {
		iterations, err := strconv.Atoi(strings.TrimSpace(parts[2]))
		if err != nil {
			log.Fatalf("Invalid iterations: %v", err)
		}
		config.Iterations = iterations
	}

	// Run benchmark suite
	suite := NewBenchmarkSuite()
	suite.Run(config)

	// Generate report
	suite.GenerateReport()

	// Save results
	err := suite.SaveResults(config.OutputFile)
	if err != nil {
		log.Printf("Failed to save results: %v", err)
	} else {
		fmt.Printf("\nResults saved to %s\n", config.OutputFile)
	}

	fmt.Printf("\nBenchmark suite completed!\n")
}
