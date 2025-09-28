package main

import (
	"fmt"

	"github.com/AnasImloul/go-orchestrator"
	"test-conflict/pkg1"
	"test-conflict/pkg2"
)

func main() {
	fmt.Println("Testing interface name conflicts across packages...")

	registry := orchestrator.New()

	// Register service from package 1
	registry.Register(
		orchestrator.NewServiceSingleton[pkg1.DatabaseService](
			pkg1.NewDatabaseService(),
		),
	)

	// This will cause a conflict because both services will have the same inferred name "database"
	// The library strips the package prefix and only uses the interface name
	registry.Register(
		orchestrator.NewServiceSingleton[pkg2.DatabaseService](
			pkg2.NewDatabaseService(),
		),
	)

	fmt.Println("Both services registered successfully!")
}
