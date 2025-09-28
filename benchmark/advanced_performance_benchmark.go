package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// DAGPattern represents different DAG generation patterns
type DAGPattern string

const (
	LinearPattern    DAGPattern = "linear"    // Each service depends on the previous one
	TreePattern      DAGPattern = "tree"      // Binary tree structure
	LayeredPattern   DAGPattern = "layered"   // Multiple layers with dependencies
	StarPattern      DAGPattern = "star"      // One central service with many dependencies
	MeshPattern      DAGPattern = "mesh"      // Highly connected mesh
	RandomPattern    DAGPattern = "random"    // Random dependencies
)

// PerformanceMetrics holds performance measurement data
type PerformanceMetrics struct {
	ServiceCount        int
	Pattern             DAGPattern
	DAGGenerationTime   time.Duration
	RegistrationTime    time.Duration
	StartTime           time.Duration
	StopTime            time.Duration
	TotalTime           time.Duration
	StartThroughput     float64
	StopThroughput      float64
	MemoryUsage         uint64 // In bytes (if available)
}

// MockService implements the Service interface for performance testing
type MockService struct {
	ID         string
	StartDelay time.Duration
	StopDelay  time.Duration
	Workload   int // Simulated work units
}

func (m *MockService) Start(ctx context.Context) error {
	// Simulate some work during startup
	for i := 0; i < m.Workload; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Simulate CPU work
			_ = i * i
		}
	}
	time.Sleep(m.StartDelay)
	return nil
}

func (m *MockService) Stop(ctx context.Context) error {
	// Simulate cleanup work
	for i := 0; i < m.Workload/2; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Simulate CPU work
			_ = i * i
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

// DAGGenerator generates different types of DAGs
type DAGGenerator struct {
	serviceCount int
	pattern      DAGPattern
}

func NewDAGGenerator(serviceCount int, pattern DAGPattern) *DAGGenerator {
	return &DAGGenerator{
		serviceCount: serviceCount,
		pattern:      pattern,
	}
}

func (g *DAGGenerator) Generate() []ServiceNode {
	switch g.pattern {
	case LinearPattern:
		return g.generateLinear()
	case TreePattern:
		return g.generateTree()
	case LayeredPattern:
		return g.generateLayered()
	case StarPattern:
		return g.generateStar()
	case MeshPattern:
		return g.generateMesh()
	case RandomPattern:
		return g.generateRandom()
	default:
		return g.generateLayered() // Default to layered
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
	
	// Build binary tree structure
	for i := 0; i < g.serviceCount; i++ {
		var dependencies []string
		
		// Left child: 2*i + 1, Right child: 2*i + 2
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
	
	// Calculate number of layers
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
				// Depend on some services from previous layer
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
	
	// First service is the central hub
	nodes[0] = ServiceNode{
		ID:           "service_0",
		Dependencies: []string{},
		StartTime:    time.Duration(1) * time.Millisecond,
		StopTime:     time.Duration(1) * time.Millisecond,
		Workload:     100,
	}
	
	// All other services depend on the central hub
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
		
		// Each service depends on multiple other services (high connectivity)
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
		
		// Random dependencies (but avoid cycles)
		maxDeps := 1 + (i % 5)
		for j := 0; j < maxDeps && j < i; j++ {
			depIndex := (i * 7 + j * 13) % i // Pseudo-random but deterministic
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

// ServiceNode represents a node in the DAG
type ServiceNode struct {
	ID           string
	Dependencies []string
	StartTime    time.Duration
	StopTime     time.Duration
	Workload     int
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

// runAdvancedPerformanceTest runs the performance test with different patterns
func runAdvancedPerformanceTest(serviceCount int, pattern DAGPattern) (*PerformanceMetrics, error) {
	fmt.Printf("Starting advanced performance test...\n")
	fmt.Printf("   Services: %d\n", serviceCount)
	fmt.Printf("   Pattern: %s\n", pattern)
	
	overallStart := time.Now()
	
	// Generate DAG
	fmt.Printf("Generating %s DAG with %d services...\n", pattern, serviceCount)
	dagStart := time.Now()
	generator := NewDAGGenerator(serviceCount, pattern)
	nodes := generator.Generate()
	dagGenerationTime := time.Since(dagStart)
	fmt.Printf("   DAG generated in %v\n", dagGenerationTime)
	
	// Create service registry
	registry := orchestrator.New()
	
	// Register all services
	fmt.Printf("Registering %d services...\n", serviceCount)
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
	
	registrationTime := time.Since(registrationStart)
	fmt.Printf("   Services registered in %v\n", registrationTime)
	
	// Start services
	fmt.Printf("Starting %d services...\n", serviceCount)
	startTime := time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	
	err := registry.Start(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start services: %w", err)
	}
	
	startDuration := time.Since(startTime)
	fmt.Printf("   All services started in %v\n", startDuration)
	
	// Wait a bit to simulate work
	fmt.Printf("Running services for 1 second...\n")
	time.Sleep(1 * time.Second)
	
	// Stop services
	fmt.Printf("Stopping %d services...\n", serviceCount)
	stopTime := time.Now()
	
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer stopCancel()
	
	err = registry.Stop(stopCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to stop services: %w", err)
	}
	
	stopDuration := time.Since(stopTime)
	fmt.Printf("   All services stopped in %v\n", stopDuration)
	
	totalTime := time.Since(overallStart)
	
	// Calculate throughput
	startThroughput := float64(serviceCount) / startDuration.Seconds()
	stopThroughput := float64(serviceCount) / stopDuration.Seconds()
	
	metrics := &PerformanceMetrics{
		ServiceCount:        serviceCount,
		Pattern:             pattern,
		DAGGenerationTime:   dagGenerationTime,
		RegistrationTime:    registrationTime,
		StartTime:           startDuration,
		StopTime:            stopDuration,
		TotalTime:           totalTime,
		StartThroughput:     startThroughput,
		StopThroughput:      stopThroughput,
	}
	
	return metrics, nil
}

func printMetrics(metrics *PerformanceMetrics) {
	fmt.Printf("\nPerformance Metrics:\n")
	fmt.Printf("   Services: %d\n", metrics.ServiceCount)
	fmt.Printf("   Pattern: %s\n", metrics.Pattern)
	fmt.Printf("   DAG Generation: %v\n", metrics.DAGGenerationTime)
	fmt.Printf("   Registration: %v\n", metrics.RegistrationTime)
	fmt.Printf("   Start Time: %v\n", metrics.StartTime)
	fmt.Printf("   Stop Time: %v\n", metrics.StopTime)
	fmt.Printf("   Total Time: %v\n", metrics.TotalTime)
	fmt.Printf("   Start Throughput: %.2f services/second\n", metrics.StartThroughput)
	fmt.Printf("   Stop Throughput: %.2f services/second\n", metrics.StopThroughput)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run advanced_performance_test.go <service_count> [pattern]")
		fmt.Println("Patterns: linear, tree, layered, star, mesh, random")
		fmt.Println("Example: go run advanced_performance_test.go 1000 layered")
		os.Exit(1)
	}
	
	serviceCount, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid service count: %v", err)
	}
	
	if serviceCount <= 0 {
		log.Fatalf("Service count must be positive, got: %d", serviceCount)
	}
	
	// Parse pattern
	pattern := LayeredPattern // Default
	if len(os.Args) > 2 {
		patternStr := strings.ToLower(os.Args[2])
		switch patternStr {
		case "linear":
			pattern = LinearPattern
		case "tree":
			pattern = TreePattern
		case "layered":
			pattern = LayeredPattern
		case "star":
			pattern = StarPattern
		case "mesh":
			pattern = MeshPattern
		case "random":
			pattern = RandomPattern
		default:
			fmt.Printf("Unknown pattern '%s', using default 'layered'\n", patternStr)
		}
	}
	
	if serviceCount > 5000 {
		fmt.Printf("Warning: Testing with %d services may take a long time and use significant memory\n", serviceCount)
		fmt.Print("Continue? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Test cancelled")
			os.Exit(0)
		}
	}
	
	fmt.Printf("Go Orchestrator Advanced Performance Test\n")
	fmt.Printf("============================================\n")
	
	metrics, err := runAdvancedPerformanceTest(serviceCount, pattern)
	if err != nil {
		log.Fatalf("Performance test failed: %v", err)
	}
	
	printMetrics(metrics)
	
	fmt.Printf("\nAdvanced performance test completed successfully!\n")
}
