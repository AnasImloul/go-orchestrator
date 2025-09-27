package main

import (
	"fmt"

	"github.com/AnasImloul/go-orchestrator"
)

// DatabaseService interface
type DatabaseService interface {
	Connect() error
	Disconnect() error
}

// SomeOtherService that doesn't implement DatabaseService
type SomeOtherService struct{}

func (s *SomeOtherService) DoSomething() error {
	return nil
}

func main() {
	app := orchestrator.New()
	
	// This should work - correct interface implementation
	fmt.Println("Testing correct interface implementation...")
	app.AddFeature(
		orchestrator.WithService[DatabaseService](&databaseService{host: "localhost", port: 5432})(
			orchestrator.NewFeature("database"),
		).
			WithLifetime(orchestrator.Singleton),
	)
	
	// This should panic - wrong interface implementation
	fmt.Println("Testing incorrect interface implementation (should panic)...")
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("✅ Correctly caught type safety error: %v\n", r)
		}
	}()
	
	app.AddFeature(
		orchestrator.WithService[DatabaseService](&SomeOtherService{})(
			orchestrator.NewFeature("wrong"),
		).
			WithLifetime(orchestrator.Singleton), // This doesn't implement DatabaseService
	)
	
	fmt.Println("❌ Type safety check failed - this should not be reached!")
}

// databaseService implementation
type databaseService struct {
	host string
	port int
}

func (d *databaseService) Connect() error {
	return nil
}

func (d *databaseService) Disconnect() error {
	return nil
}
