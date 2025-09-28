package main

import (
	"fmt"

	"github.com/AnasImloul/go-orchestrator"
)

// Test interfaces
type DatabaseService interface {
	Connect() error
}

type WrongService interface {
	DoWrong() string
}

// Correct implementation
type databaseService struct{}

func (d *databaseService) Connect() error {
	fmt.Println("Database connected")
	return nil
}

// Wrong implementation
type wrongService struct{}

func (w *wrongService) DoWrong() string {
	return "wrong"
}

func main() {
	registry := orchestrator.New()

	// This should work - correct factory return type
	fmt.Println("✅ Testing correct factory...")
	correctFactory := func() DatabaseService {
		return &databaseService{}
	}
	correctDef := orchestrator.NewServiceFactory[DatabaseService](correctFactory, orchestrator.Singleton)
	registry.Register(correctDef)

	// This should panic at runtime - wrong factory return type
	fmt.Println("❌ Testing wrong factory...")
	wrongFactory := func() WrongService {
		return &wrongService{}
	}

	// This should panic because WrongService doesn't implement DatabaseService
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("✅ Caught expected panic: %v\n", r)
		}
	}()

	fmt.Println("Creating wrong service definition...")
	wrongDef := orchestrator.NewServiceFactory[DatabaseService](wrongFactory, orchestrator.Singleton)
	fmt.Println("Registering wrong service definition...")
	registry.Register(wrongDef)

	fmt.Println("❌ This should not be reached!")
}
