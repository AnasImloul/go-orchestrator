package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/AnasImloul/go-orchestrator/pkg/di"
	"github.com/AnasImloul/go-orchestrator/pkg/lifecycle"
	"github.com/AnasImloul/go-orchestrator/pkg/logger"
)

// NewOrchestrator creates a new orchestrator with the default configuration.
func NewOrchestrator() (Orchestrator, error) {
	return NewOrchestratorWithConfig(DefaultOrchestratorConfig())
}

// NewOrchestratorWithConfig creates a new orchestrator with the specified configuration.
func NewOrchestratorWithConfig(config OrchestratorConfig) (Orchestrator, error) {
	// Create a simple logger (external users can provide their own)
	logger := &simpleLogger{}

	// Create a simple DI container
	container := &simpleContainer{
		services: make(map[string]interface{}),
	}

	// Create a simple lifecycle manager
	lifecycleManager := &simpleLifecycleManager{
		components: make(map[string]lifecycle.Component),
		phase:      lifecycle.PhaseStopped,
		logger:     logger,
	}

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
	config           OrchestratorConfig
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
			Status:    state.Health.Status.String(),
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
func (o *defaultOrchestrator) GetContainer() di.Container {
	return o.container
}

// GetLifecycleManager returns the lifecycle manager.
func (o *defaultOrchestrator) GetLifecycleManager() lifecycle.LifecycleManager {
	return o.lifecycleManager
}

// componentWrapper wraps a feature as a lifecycle component.
type componentWrapper struct {
	feature   Feature
	container di.Container
	component lifecycle.Component
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
		return w.component.Health(ctx)
	}

	return lifecycle.ComponentHealth{
		Status:    lifecycle.HealthStatusUnknown,
		Message:   "Component not initialized",
		Timestamp: time.Now(),
	}
}

// GetRetryConfig returns the retry configuration.
func (w *componentWrapper) GetRetryConfig() *lifecycle.RetryConfig {
	return w.feature.GetRetryConfig()
}