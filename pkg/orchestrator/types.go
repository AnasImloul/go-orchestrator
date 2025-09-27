package orchestrator

import (
	"context"
	"time"

	"github.com/AnasImloul/go-orchestrator/pkg/di"
	"github.com/AnasImloul/go-orchestrator/pkg/lifecycle"
)

// OrchestratorConfig holds configuration for the orchestrator.
type OrchestratorConfig struct {
	StartupTimeout      time.Duration
	ShutdownTimeout     time.Duration
	HealthCheckInterval time.Duration
	EnableMetrics       bool
	EnableTracing       bool
	FeatureConfig       map[string]interface{}
}

// DefaultOrchestratorConfig returns the default orchestrator configuration.
func DefaultOrchestratorConfig() OrchestratorConfig {
	return OrchestratorConfig{
		StartupTimeout:      30 * time.Second,
		ShutdownTimeout:     15 * time.Second,
		HealthCheckInterval: 30 * time.Second,
		EnableMetrics:       true,
		EnableTracing:       false,
		FeatureConfig:       make(map[string]interface{}),
	}
}

// Orchestrator represents the main orchestrator interface.
type Orchestrator interface {
	RegisterFeature(feature Feature) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	HealthCheck(ctx context.Context) HealthReport
	GetContainer() di.Container
	GetLifecycleManager() lifecycle.LifecycleManager
}

// Feature represents a feature that can be managed by the orchestrator.
type Feature interface {
	GetName() string
	GetDependencies() []string
	GetPriority() int
	RegisterServices(container di.Container) error
	CreateComponent(container di.Container) (lifecycle.Component, error)
	GetRetryConfig() *lifecycle.RetryConfig
	GetMetadata() FeatureMetadata
}

// FeatureMetadata contains metadata about a feature.
type FeatureMetadata struct {
	Name        string
	Description string
	Version     string
	Author      string
	Tags        []string
}

// HealthReport represents the overall health of the application.
type HealthReport struct {
	Status    string
	Timestamp time.Time
	Features  map[string]FeatureHealth
	Summary   HealthSummary
}

// FeatureHealth represents the health of a single feature.
type FeatureHealth struct {
	Status    string
	Message   string
	Timestamp time.Time
	Metadata  FeatureMetadata
}

// HealthSummary provides a summary of the health check.
type HealthSummary struct {
	TotalFeatures     int
	HealthyFeatures   int
	DegradedFeatures  int
	UnhealthyFeatures int
}