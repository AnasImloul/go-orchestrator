package orchestrator

import (
	"context"
	"time"

	"github.com/imloulanas/go-orchestrator/di"
	"github.com/imloulanas/go-orchestrator/lifecycle"
)

// Orchestrator coordinates the application lifecycle and dependency injection
type Orchestrator interface {
	// RegisterFeature registers a feature with the orchestrator
	RegisterFeature(feature Feature) error

	// Start starts the application in the correct order
	Start(ctx context.Context) error

	// Stop stops the application gracefully
	Stop(ctx context.Context) error

	// GetContainer returns the DI container
	GetContainer() di.Container

	// GetLifecycleManager returns the lifecycle manager
	GetLifecycleManager() lifecycle.LifecycleManager

	// GetPhase returns the current application phase
	GetPhase() lifecycle.Phase

	// HealthCheck performs a health check on all components
	HealthCheck(ctx context.Context) HealthReport
}

// Feature represents a feature that can be managed by the orchestrator
type Feature interface {
	// GetName returns the feature name
	GetName() string

	// GetDependencies returns the names of features this feature depends on
	GetDependencies() []string

	// GetPriority returns the feature priority (lower numbers start first)
	GetPriority() int

	// RegisterServices registers services with the DI container
	RegisterServices(container di.Container) error

	// CreateComponent creates the lifecycle component for this feature
	CreateComponent(container di.Container) (lifecycle.Component, error)

	// GetMetadata returns feature metadata
	GetMetadata() FeatureMetadata
}

// FeatureMetadata contains metadata about a feature
type FeatureMetadata struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Version      string                 `json:"version"`
	Author       string                 `json:"author"`
	Dependencies []string               `json:"dependencies"`
	Priority     int                    `json:"priority"`
	Tags         []string               `json:"tags"`
	Config       map[string]interface{} `json:"config"`
}

// HealthReport represents the overall health of the application
type HealthReport struct {
	Status     HealthStatus                         `json:"status"`
	Timestamp  time.Time                            `json:"timestamp"`
	Phase      lifecycle.Phase                      `json:"phase"`
	Features   map[string]FeatureHealth             `json:"features"`
	Components map[string]lifecycle.ComponentHealth `json:"components"`
	Summary    HealthSummary                        `json:"summary"`
}

// FeatureHealth represents the health of a single feature
type FeatureHealth struct {
	Status    HealthStatus    `json:"status"`
	Message   string          `json:"message"`
	Timestamp time.Time       `json:"timestamp"`
	Metadata  FeatureMetadata `json:"metadata"`
}

// HealthStatus represents different health states
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthSummary provides a summary of the health check
type HealthSummary struct {
	TotalFeatures     int `json:"total_features"`
	HealthyFeatures   int `json:"healthy_features"`
	DegradedFeatures  int `json:"degraded_features"`
	UnhealthyFeatures int `json:"unhealthy_features"`
	UnknownFeatures   int `json:"unknown_features"`
}

// OrchestratorConfig holds configuration for the orchestrator
type OrchestratorConfig struct {
	// StartupTimeout is the maximum time to wait for startup
	StartupTimeout time.Duration

	// ShutdownTimeout is the maximum time to wait for shutdown
	ShutdownTimeout time.Duration

	// HealthCheckInterval is the interval for periodic health checks
	HealthCheckInterval time.Duration

	// EnableMetrics enables metrics collection
	EnableMetrics bool

	// EnableTracing enables distributed tracing
	EnableTracing bool

	// FeatureConfig contains feature-specific configuration
	FeatureConfig map[string]interface{}
}

// DefaultOrchestratorConfig returns the default orchestrator configuration
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

// ComponentWrapper wraps a feature as a lifecycle component
type ComponentWrapper struct {
	feature   Feature
	container di.Container
	component lifecycle.Component
}

// Name returns the component name
func (w *ComponentWrapper) Name() string {
	return w.feature.GetName()
}

// Dependencies returns the component dependencies
func (w *ComponentWrapper) Dependencies() []string {
	return w.feature.GetDependencies()
}

// Priority returns the component priority
func (w *ComponentWrapper) Priority() int {
	return w.feature.GetPriority()
}

// Start starts the component
func (w *ComponentWrapper) Start(ctx context.Context) error {
	// Register services first
	if err := w.feature.RegisterServices(w.container); err != nil {
		return err
	}

	// Create and start the actual component
	component, err := w.feature.CreateComponent(w.container)
	if err != nil {
		return err
	}

	w.component = component
	return component.Start(ctx)
}

// Stop stops the component
func (w *ComponentWrapper) Stop(ctx context.Context) error {
	if w.component != nil {
		return w.component.Stop(ctx)
	}
	return nil
}

// Health returns the component health
func (w *ComponentWrapper) Health(ctx context.Context) lifecycle.ComponentHealth {
	if w.component != nil {
		return w.component.Health(ctx)
	}

	return lifecycle.ComponentHealth{
		Status:    lifecycle.HealthStatusUnknown,
		Message:   "Component not initialized",
		Timestamp: time.Now(),
	}
}
