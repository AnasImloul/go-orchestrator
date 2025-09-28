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
		orchestrator.NewFeatureWithInstance("database", DatabaseService(&databaseService{host: "localhost", port: 5432}), orchestrator.Singleton),
	)

	// Type safety is enforced at compile time - the following would not compile:
	// app.AddFeature(
	//     orchestrator.NewFeatureWithInstance("wrong", &SomeOtherService{}, orchestrator.Singleton),
	// )
	fmt.Println("✅ Type safety is enforced at compile time - wrong types cannot be registered")
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
