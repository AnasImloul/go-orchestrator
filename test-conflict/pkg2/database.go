package pkg2

import (
	"context"
	"fmt"

	"github.com/AnasImloul/go-orchestrator"
)

// DatabaseService interface from package 2 (same name, different package)
type DatabaseService interface {
	orchestrator.Service
	Query(sql string) error
}

type databaseService struct{}

func (d *databaseService) Start(ctx context.Context) error {
	fmt.Println("Package 2 Database service starting")
	return nil
}

func (d *databaseService) Stop(ctx context.Context) error {
	fmt.Println("Package 2 Database service stopping")
	return nil
}

func (d *databaseService) Health(ctx context.Context) orchestrator.HealthStatus {
	return orchestrator.HealthStatus{Status: "healthy", Message: "Package 2 Database is healthy"}
}

func (d *databaseService) Query(sql string) error {
	fmt.Println("Package 2 Database querying:", sql)
	return nil
}

func NewDatabaseService() DatabaseService {
	return &databaseService{}
}
