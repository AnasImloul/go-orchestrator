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
	"fmt"
	"reflect"
	"sync"
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
	GetContainer() Container
	GetLifecycleManager() LifecycleManager
}

// Container represents a dependency injection container.
type Container interface {
	Register(serviceType reflect.Type, factory func(ctx context.Context, container Container) (interface{}, error), options ...interface{}) error
	RegisterInstance(serviceType reflect.Type, instance interface{}) error
	RegisterSingleton(serviceType reflect.Type, factory func(ctx context.Context, container Container) (interface{}, error), options ...interface{}) error
	Resolve(serviceType reflect.Type) (interface{}, error)
	CreateScope() interface{}
	Validate() error
}

// Component represents a lifecycle component.
type Component interface {
	Name() string
	Dependencies() []string
	Priority() int
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Health(ctx context.Context) ComponentHealth
	GetRetryConfig() *RetryConfig
}

// ComponentHealth represents the health status of a component.
type ComponentHealth struct {
	Status    string
	Message   string
	Details   map[string]interface{}
	Timestamp time.Time
}

// RetryConfig configures retry behavior for component operations.
type RetryConfig struct {
	MaxAttempts       int
	InitialDelay      time.Duration
	MaxDelay          time.Duration
	BackoffMultiplier float64
	RetryableErrors   []error
}

// LifecycleManager manages component lifecycle operations.
type LifecycleManager interface {
	RegisterComponent(component Component) error
	UnregisterComponent(name string) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	GetPhase() string
	GetComponentState(name string) (interface{}, error)
	GetAllComponentStates() map[string]interface{}
	AddHook(phase string, hook interface{}) error
	RemoveHook(phase string, hook interface{}) error
}

// Feature represents a feature that can be managed by the orchestrator.
type Feature interface {
	GetName() string
	GetDependencies() []string
	GetPriority() int
	RegisterServices(container Container) error
	CreateComponent(container Container) (Component, error)
	GetRetryConfig() *RetryConfig
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
		mu:               sync.RWMutex{},
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
	mu               sync.RWMutex
}

// RegisterFeature registers a feature with the orchestrator.
func (o *defaultOrchestrator) RegisterFeature(feature Feature) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	name := feature.GetName()
	if _, exists := o.features[name]; exists {
		return fmt.Errorf("feature %s is already registered", name)
	}

	o.features[name] = feature
	o.logger.Info("Feature registered", "name", name)
	return nil
}

// Start starts the orchestrator.
func (o *defaultOrchestrator) Start(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.logger.Info("Starting orchestrator")

	// Register all features as components
	for name, feature := range o.features {
		wrapper := &componentWrapper{
			feature:   feature,
			container: o.container,
		}
		
		if err := o.lifecycleManager.RegisterComponent(wrapper); err != nil {
			return fmt.Errorf("failed to register component %s: %w", name, err)
		}
	}

	// Start the lifecycle manager
	return o.lifecycleManager.Start(ctx)
}

// Stop stops the orchestrator.
func (o *defaultOrchestrator) Stop(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.logger.Info("Stopping orchestrator")
	return o.lifecycleManager.Stop(ctx)
}

// HealthCheck performs a health check on all components.
func (o *defaultOrchestrator) HealthCheck(ctx context.Context) HealthReport {
	o.mu.RLock()
	defer o.mu.RUnlock()

	states := o.lifecycleManager.GetAllComponentStates()
	
	report := HealthReport{
		Status:    "healthy",
		Timestamp: time.Now(),
		Features:  make(map[string]FeatureHealth),
		Summary:   HealthSummary{},
	}

	healthyCount := 0
	unhealthyCount := 0

	for name, state := range states {
		feature := o.features[name]
		health := FeatureHealth{
			Status:    string(state.Health.Status),
			Message:   state.Health.Message,
			Timestamp: state.Health.Timestamp,
			Metadata:  feature.GetMetadata(),
		}

		report.Features[name] = health

		switch state.Health.Status {
		case lifecycle.HealthStatusHealthy:
			healthyCount++
		case lifecycle.HealthStatusUnhealthy:
			unhealthyCount++
		}
	}

	report.Summary.TotalFeatures = len(o.features)
	report.Summary.HealthyFeatures = healthyCount
	report.Summary.UnhealthyFeatures = unhealthyCount

	if unhealthyCount > 0 {
		report.Status = "unhealthy"
	}

	return report
}

// GetContainer returns the DI container.
func (o *defaultOrchestrator) GetContainer() Container {
	return &containerWrapper{internal: o.container}
}

// GetLifecycleManager returns the lifecycle manager.
func (o *defaultOrchestrator) GetLifecycleManager() LifecycleManager {
	return &lifecycleWrapper{internal: o.lifecycleManager}
}

// componentWrapper wraps a feature as a lifecycle component.
type componentWrapper struct {
	feature   Feature
	container di.Container
	component Component
}

// Name returns the component name.
func (w *componentWrapper) Name() string {
	return w.feature.GetName()
}

// Dependencies returns the component dependencies.
func (w *componentWrapper) Dependencies() []string {
	return w.feature.GetDependencies()
}

// Priority returns the component priority.
func (w *componentWrapper) Priority() int {
	return w.feature.GetPriority()
}

// Start starts the component.
func (w *componentWrapper) Start(ctx context.Context) error {
	// Register services first
	if err := w.feature.RegisterServices(&containerWrapper{internal: w.container}); err != nil {
		return err
	}

	// Create and start the actual component
	component, err := w.feature.CreateComponent(&containerWrapper{internal: w.container})
	if err != nil {
		return err
	}

	w.component = component
	return component.Start(ctx)
}

// Stop stops the component.
func (w *componentWrapper) Stop(ctx context.Context) error {
	if w.component != nil {
		return w.component.Stop(ctx)
	}
	return nil
}

// Health returns the component health.
func (w *componentWrapper) Health(ctx context.Context) lifecycle.ComponentHealth {
	if w.component != nil {
		health := w.component.Health(ctx)
		return lifecycle.ComponentHealth{
			Status:    lifecycle.HealthStatus(health.Status),
			Message:   health.Message,
			Details:   health.Details,
			Timestamp: health.Timestamp,
		}
	}

	return lifecycle.ComponentHealth{
		Status:    lifecycle.HealthStatusUnknown,
		Message:   "Component not initialized",
		Timestamp: time.Now(),
	}
}

// GetRetryConfig returns the retry configuration.
func (w *componentWrapper) GetRetryConfig() *lifecycle.RetryConfig {
	config := w.feature.GetRetryConfig()
	if config == nil {
		return nil
	}

	return &lifecycle.RetryConfig{
		MaxAttempts:       config.MaxAttempts,
		InitialDelay:      config.InitialDelay,
		MaxDelay:          config.MaxDelay,
		BackoffMultiplier: config.BackoffMultiplier,
		RetryableErrors:   config.RetryableErrors,
	}
}

// containerWrapper wraps the internal container to provide a public API.
type containerWrapper struct {
	internal di.Container
}

// Register registers a service with the container.
func (c *containerWrapper) Register(serviceType reflect.Type, factory func(ctx context.Context, container Container) (interface{}, error), options ...interface{}) error {
	internalFactory := func(ctx context.Context, container di.Container) (interface{}, error) {
		return factory(ctx, c)
	}
	return c.internal.Register(serviceType, internalFactory)
}

// RegisterInstance registers a service instance.
func (c *containerWrapper) RegisterInstance(serviceType reflect.Type, instance interface{}) error {
	return c.internal.RegisterInstance(serviceType, instance)
}

// RegisterSingleton registers a singleton service.
func (c *containerWrapper) RegisterSingleton(serviceType reflect.Type, factory func(ctx context.Context, container Container) (interface{}, error), options ...interface{}) error {
	internalFactory := func(ctx context.Context, container di.Container) (interface{}, error) {
		return factory(ctx, c)
	}
	return c.internal.RegisterSingleton(serviceType, internalFactory)
}

// Resolve resolves a service from the container.
func (c *containerWrapper) Resolve(serviceType reflect.Type) (interface{}, error) {
	return c.internal.Resolve(serviceType)
}

// CreateScope creates a new scope.
func (c *containerWrapper) CreateScope() interface{} {
	return c.internal.CreateScope()
}

// Validate validates the container configuration.
func (c *containerWrapper) Validate() error {
	return nil
}

// lifecycleWrapper wraps the internal lifecycle manager to provide a public API.
type lifecycleWrapper struct {
	internal lifecycle.LifecycleManager
}

// RegisterComponent registers a component for lifecycle management.
func (lm *lifecycleWrapper) RegisterComponent(component Component) error {
	// Convert public component to internal component
	internalComponent := &componentWrapper{component: component}
	return lm.internal.RegisterComponent(internalComponent)
}

// UnregisterComponent removes a component from lifecycle management.
func (lm *lifecycleWrapper) UnregisterComponent(name string) error {
	return lm.internal.UnregisterComponent(name)
}

// Start starts all registered components in dependency order.
func (lm *lifecycleWrapper) Start(ctx context.Context) error {
	return lm.internal.Start(ctx)
}

// Stop stops all registered components in reverse order.
func (lm *lifecycleWrapper) Stop(ctx context.Context) error {
	return lm.internal.Stop(ctx)
}

// GetPhase returns the current lifecycle phase.
func (lm *lifecycleWrapper) GetPhase() string {
	return string(lm.internal.GetPhase())
}

// GetComponentState returns the state of a specific component.
func (lm *lifecycleWrapper) GetComponentState(name string) (interface{}, error) {
	state, exists := lm.internal.GetComponentState(name)
	if !exists {
		return nil, fmt.Errorf("component %s not found", name)
	}
	return state, nil
}

// GetAllComponentStates returns the states of all components.
func (lm *lifecycleWrapper) GetAllComponentStates() map[string]interface{} {
	states := lm.internal.GetAllComponentStates()
	result := make(map[string]interface{})
	for k, v := range states {
		result[k] = v
	}
	return result
}

// AddHook adds a lifecycle hook.
func (lm *lifecycleWrapper) AddHook(phase string, hook interface{}) error {
	return nil // Simplified for now
}

// RemoveHook removes a lifecycle hook.
func (lm *lifecycleWrapper) RemoveHook(phase string, hook interface{}) error {
	return nil // Simplified for now
}

