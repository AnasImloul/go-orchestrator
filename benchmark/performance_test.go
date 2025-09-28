package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/AnasImloul/go-orchestrator"
)

// ServiceNode represents a node in the DAG
type ServiceNode struct {
	ID           string
	Dependencies []string
	StartTime    time.Duration
	StopTime     time.Duration
}

// MockService implements the Service interface for performance testing
type MockService struct {
	ID        string
	StartDelay time.Duration
	StopDelay  time.Duration
}

func (m *MockService) Start(ctx context.Context) error {
	time.Sleep(m.StartDelay)
	return nil
}

func (m *MockService) Stop(ctx context.Context) error {
	time.Sleep(m.StopDelay)
	return nil
}

func (m *MockService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{
		Status:  orchestrator.HealthStatusHealthy,
		Message: fmt.Sprintf("Service %s is healthy", m.ID),
	}
}

// generateDAG creates a DAG with the specified number of services
// Uses a layered approach where each layer depends on the previous layer
func generateDAG(serviceCount int) []ServiceNode {
	nodes := make([]ServiceNode, serviceCount)
	
	// Calculate number of layers (roughly sqrt of service count for balanced DAG)
	layers := int(float64(serviceCount) * 0.3) // 30% of services as layers
	if layers < 2 {
		layers = 2
	}
	
	servicesPerLayer := serviceCount / layers
	remainingServices := serviceCount % layers
	
	serviceIndex := 0
	
	for layer := 0; layer < layers; layer++ {
		// Calculate services in this layer
		servicesInThisLayer := servicesPerLayer
		if layer < remainingServices {
			servicesInThisLayer++
		}
		
		// Generate services for this layer
		for i := 0; i < servicesInThisLayer && serviceIndex < serviceCount; i++ {
			nodeID := fmt.Sprintf("service_%d", serviceIndex)
			var dependencies []string
			
			// Services in first layer have no dependencies
			// Services in subsequent layers depend on some services from previous layer
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
				
				// Each service depends on 1-3 services from previous layer
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
				ID:           nodeID,
				Dependencies: dependencies,
				StartTime:    time.Duration(1+serviceIndex%10) * time.Millisecond, // 1-10ms start time
				StopTime:     time.Duration(1+serviceIndex%5) * time.Millisecond,  // 1-5ms stop time
			}
			serviceIndex++
		}
	}
	
	return nodes
}

// createMockServiceFactory creates a factory function for a mock service
func createMockServiceFactory(node ServiceNode) func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
	return func(ctx context.Context, container *orchestrator.Container) (interface{}, error) {
		return &MockService{
			ID:        node.ID,
			StartDelay: node.StartTime,
			StopDelay:  node.StopTime,
		}, nil
	}
}

// runPerformanceTest runs the performance test with the specified number of services
func runPerformanceTest(serviceCount int) error {
	fmt.Printf("Starting performance test with %d services...\n", serviceCount)
	
	// Generate DAG
	fmt.Printf("Generating DAG with %d services...\n", serviceCount)
	start := time.Now()
	nodes := generateDAG(serviceCount)
	dagGenerationTime := time.Since(start)
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
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	
	err := registry.Start(ctx)
	if err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	}
	
	startDuration := time.Since(startTime)
	fmt.Printf("   All services started in %v\n", startDuration)
	
	// Wait a bit to simulate work
	fmt.Printf("Running services for 2 seconds...\n")
	time.Sleep(2 * time.Second)
	
	// Stop services
	fmt.Printf("Stopping %d services...\n", serviceCount)
	stopTime := time.Now()
	
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer stopCancel()
	
	err = registry.Stop(stopCtx)
	if err != nil {
		return fmt.Errorf("failed to stop services: %w", err)
	}
	
	stopDuration := time.Since(stopTime)
	fmt.Printf("   All services stopped in %v\n", stopDuration)
	
	// Print performance summary
	fmt.Printf("\nPerformance Summary:\n")
	fmt.Printf("   Services: %d\n", serviceCount)
	fmt.Printf("   DAG Generation: %v\n", dagGenerationTime)
	fmt.Printf("   Registration: %v\n", registrationTime)
	fmt.Printf("   Start Time: %v\n", startDuration)
	fmt.Printf("   Stop Time: %v\n", stopDuration)
	fmt.Printf("   Total Time: %v\n", time.Since(start))
	
	// Calculate throughput
	startThroughput := float64(serviceCount) / startDuration.Seconds()
	stopThroughput := float64(serviceCount) / stopDuration.Seconds()
	
	fmt.Printf("   Start Throughput: %.2f services/second\n", startThroughput)
	fmt.Printf("   Stop Throughput: %.2f services/second\n", stopThroughput)
	
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run performance_test.go <service_count>")
		fmt.Println("Example: go run performance_test.go 1000")
		os.Exit(1)
	}
	
	serviceCount, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatalf("Invalid service count: %v", err)
	}
	
	if serviceCount <= 0 {
		log.Fatalf("Service count must be positive, got: %d", serviceCount)
	}
	
	if serviceCount > 10000 {
		fmt.Printf("Warning: Testing with %d services may take a long time and use significant memory\n", serviceCount)
		fmt.Print("Continue? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Test cancelled")
			os.Exit(0)
		}
	}
	
	fmt.Printf("Go Orchestrator Performance Test\n")
	fmt.Printf("=====================================\n")
	
	err = runPerformanceTest(serviceCount)
	if err != nil {
		log.Fatalf("Performance test failed: %v", err)
	}
	
	fmt.Printf("\nPerformance test completed successfully!\n")
}
