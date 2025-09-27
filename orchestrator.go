// Package orchestrator provides a simple, unified API for application orchestration.
//
// This is the main entry point for the Go Orchestrator library. Most users
// should import this package directly:
//
//	import "github.com/AnasImloul/go-orchestrator"
//
// For advanced usage, you can import specific sub-packages:
//
//	import "github.com/AnasImloul/go-orchestrator/di"
//	import "github.com/AnasImloul/go-orchestrator/lifecycle"
//	import "github.com/AnasImloul/go-orchestrator/logger"
package orchestrator

import (
	"context"
	"time"

	"github.com/AnasImloul/go-orchestrator/internal/di"
	"github.com/AnasImloul/go-orchestrator/internal/lifecycle"
	"github.com/AnasImloul/go-orchestrator/internal/logger"
)

// Config holds configuration for the orchestrator.
type Config struct {
	StartupTimeout      time.Duration
	ShutdownTimeout     time.Duration
	HealthCheckInterval time.Duration
	EnableMetrics       bool
	EnableTracing       bool
	FeatureConfig       map[string]interface{}
}

// DefaultConfig returns the default orchestrator configuration.
func DefaultConfig() Config {
	return Config{
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

// New creates a new orchestrator with the default configuration.
func New() (Orchestrator, error) {
	return NewWithConfig(DefaultConfig())
}

// NewWithConfig creates a new orchestrator with the specified configuration.
func NewWithConfig(config Config) (Orchestrator, error) {
	// Create logger
	logger := logger.NewSlogAdapter(nil) // Will use default slog logger

	// Create DI container
	diConfig := di.ContainerConfig{
		EnableValidation:    true,
		EnableCircularCheck: true,
		EnableInterception:  config.EnableTracing,
		DefaultLifetime:     di.Singleton,
		MaxResolutionDepth:  50,
		EnableMetrics:       config.EnableMetrics,
	}

	container := di.NewContainer(diConfig, logger)

	// Create lifecycle manager
	lifecycleManager := lifecycle.NewLifecycleManager(logger)

	// Create orchestrator
	orch := &defaultOrchestrator{
		container:        container,
		lifecycleManager: lifecycleManager,
		features:         make(map[string]Feature),
		config:           config,
		logger:           logger,
	}

	return orch, nil
}

// defaultOrchestrator implements the Orchestrator interface.
type defaultOrchestrator struct {
	container        di.Container
	lifecycleManager lifecycle.LifecycleManager
	features         map[string]Feature
	config           Config
	logger           logger.Logger
}

// RegisterFeature registers a feature with the orchestrator.
func (o *defaultOrchestrator) RegisterFeature(feature Feature) error {
	// Implementation here
	return nil
}

// Start starts the orchestrator.
func (o *defaultOrchestrator) Start(ctx context.Context) error {
	// Implementation here
	return nil
}

// Stop stops the orchestrator.
func (o *defaultOrchestrator) Stop(ctx context.Context) error {
	// Implementation here
	return nil
}

// HealthCheck performs a health check on all components.
func (o *defaultOrchestrator) HealthCheck(ctx context.Context) HealthReport {
	// Implementation here
	return HealthReport{}
}
