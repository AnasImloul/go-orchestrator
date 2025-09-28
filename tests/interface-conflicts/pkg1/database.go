package pkg1

import (
	"context"
	"fmt"

	"github.com/AnasImloul/go-orchestrator"
)

// DatabaseService interface from package 1
type DatabaseService interface {
	orchestrator.Service
	Connect() error
}

type databaseService struct{}

func (d *databaseService) Start(ctx context.Context) error {
	fmt.Println("Package 1 Database service starting")
	return nil
}

func (d *databaseService) Stop(ctx context.Context) error {
	fmt.Println("Package 1 Database service stopping")
	return nil
}

func (d *databaseService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{Status: "healthy", Message: "Package 1 Database is healthy"}
}

func (d *databaseService) Connect() error {
	fmt.Println("Package 1 Database connecting")
	return nil
}

func NewDatabaseService() DatabaseService {
	return &databaseService{}
}
