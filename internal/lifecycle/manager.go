package lifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/AnasImloul/go-orchestrator/internal/logger"
)

// DefaultLifecycleManager implements the LifecycleManager interface
type DefaultLifecycleManager struct {
	dag    *DAG
	hooks  map[Phase][]Hook
	states map[string]*ComponentState
	phase  Phase
	logger logger.Logger
	mu     sync.RWMutex
}

// NewLifecycleManager creates a new lifecycle manager
func NewLifecycleManager(logger logger.Logger) *DefaultLifecycleManager {
	return &DefaultLifecycleManager{
		dag:    NewDAG(),
		hooks:  make(map[Phase][]Hook),
		states: make(map[string]*ComponentState),
		phase:  PhaseStopped,
		logger: logger,
	}
}

// RegisterComponent registers a component for lifecycle management
func (lm *DefaultLifecycleManager) RegisterComponent(component Component) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	name := component.Name()

	// Check if component is already registered
	if _, exists := lm.states[name]; exists {
		return fmt.Errorf("component %s is already registered", name)
	}

	// Add to DAG
	if err := lm.dag.AddNode(component); err != nil {
		return fmt.Errorf("failed to add component %s to DAG: %w", name, err)
	}

	// Initialize component state
	lm.states[name] = &ComponentState{
		Name:         name,
		Phase:        PhaseStopped,
		Dependencies: component.Dependencies(),
		Health: ComponentHealth{
			Status:    HealthStatusUnknown,
			Timestamp: time.Now(),
		},
	}

	if lm.logger != nil {
		lm.logger.Info("Component registered",
			"component", name,
			"dependencies", component.Dependencies(),
		)
	}

	return nil
}

// UnregisterComponent removes a component from lifecycle management
func (lm *DefaultLifecycleManager) UnregisterComponent(name string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	// Check if component exists
	if _, exists := lm.states[name]; !exists {
		return fmt.Errorf("component %s is not registered", name)
	}

	// Remove from DAG
	if err := lm.dag.RemoveNode(name); err != nil {
		return fmt.Errorf("failed to remove component %s from DAG: %w", name, err)
	}

	// Remove state
	delete(lm.states, name)

	if lm.logger != nil {
		lm.logger.Info("Component unregistered",
			"component", name,
		)
	}

	return nil
}

// Start starts all components in dependency order
func (lm *DefaultLifecycleManager) Start(ctx context.Context) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.phase != PhaseStopped {
		return fmt.Errorf("lifecycle manager is not in stopped phase (current: %s)", lm.phase)
	}

	lm.phase = PhaseStartup
	if lm.logger != nil {
		lm.logger.Info("Starting lifecycle manager")
	}

	// Fire startup hooks
	if err := lm.fireHooks(ctx, PhaseStartup, "lifecycle", nil); err != nil {
		lm.phase = PhaseStopped
		return fmt.Errorf("startup hooks failed: %w", err)
	}

	// Get startup levels for parallel execution
	startupLevels, err := lm.dag.GetStartupLevels()
	if err != nil {
		lm.phase = PhaseStopped
		return fmt.Errorf("failed to determine startup levels: %w", err)
	}

	// Start components level by level, with parallel execution within each level
	for levelIndex, level := range startupLevels {
		if lm.logger != nil {
			lm.logger.Info("Starting components at level",
				"level", levelIndex,
				"components", len(level),
			)
		}

		if err := lm.startComponentsInParallel(ctx, level); err != nil {
			if lm.logger != nil {
				lm.logger.Error("Failed to start components at level, initiating rollback",
					"level", levelIndex,
					"error", err.Error(),
				)
			}

			// Rollback: stop all started components
			lm.rollbackStartup(ctx, "")
			lm.phase = PhaseStopped
			return fmt.Errorf("failed to start components at level %d: %w", levelIndex, err)
		}
	}

	lm.phase = PhaseRunning
	if lm.logger != nil {
		lm.logger.Info("All components started successfully")
	}

	return nil
}

// Stop stops all components in reverse dependency order
func (lm *DefaultLifecycleManager) Stop(ctx context.Context) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	if lm.phase != PhaseRunning {
		if lm.logger != nil {
			lm.logger.Warn("Attempting to stop lifecycle manager not in running phase",
				"current_phase", lm.phase,
			)
		}
	}

	lm.phase = PhaseShutdown
	if lm.logger != nil {
		lm.logger.Info("Stopping lifecycle manager")
	}

	// Fire shutdown hooks
	if err := lm.fireHooks(ctx, PhaseShutdown, "lifecycle", nil); err != nil {
		if lm.logger != nil {
			lm.logger.Error("Shutdown hooks failed",
				"error", err.Error(),
			)
		}
	}

	// Get shutdown order
	shutdownOrder, err := lm.dag.GetShutdownOrder()
	if err != nil {
		if lm.logger != nil {
			lm.logger.Error("Failed to determine shutdown order",
				"error", err.Error(),
			)
		}
		// Continue with best effort shutdown
		shutdownOrder = lm.getAllNodesInReverseOrder()
	}

	// Stop components in order
	var lastError error
	for _, node := range shutdownOrder {
		if err := lm.stopComponent(ctx, node); err != nil {
			if lm.logger != nil {
				lm.logger.Error("Failed to stop component",
					"component", node.Name,
					"error", err.Error(),
				)
			}
			lastError = err
		}
	}

	lm.phase = PhaseStopped
	if lm.logger != nil {
		lm.logger.Info("Lifecycle manager stopped")
	}

	return lastError
}

// AddHook adds a lifecycle hook for a specific phase
func (lm *DefaultLifecycleManager) AddHook(phase Phase, hook Hook) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.hooks[phase] = append(lm.hooks[phase], hook)

	if lm.logger != nil {
		lm.logger.Debug("Lifecycle hook added",
			"phase", phase,
		)
	}

	return nil
}

// RemoveHook removes a lifecycle hook
func (lm *DefaultLifecycleManager) RemoveHook(phase Phase, hook Hook) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	hooks := lm.hooks[phase]
	for i, h := range hooks {
		// Compare function pointers (this is a limitation of Go)
		if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", hook) {
			lm.hooks[phase] = append(hooks[:i], hooks[i+1:]...)
			if lm.logger != nil {
				lm.logger.Debug("Lifecycle hook removed",
					"phase", phase,
				)
			}
			return nil
		}
	}

	return fmt.Errorf("hook not found for phase %s", phase)
}

// GetComponentState returns the state of a specific component
func (lm *DefaultLifecycleManager) GetComponentState(name string) (ComponentState, bool) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	if state, exists := lm.states[name]; exists {
		// Return a copy to prevent external modification
		return *state, true
	}

	return ComponentState{}, false
}

// GetAllComponentStates returns the state of all components
func (lm *DefaultLifecycleManager) GetAllComponentStates() map[string]ComponentState {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	// Return copies to prevent external modification
	states := make(map[string]ComponentState)
	for name, state := range lm.states {
		states[name] = *state
	}

	return states
}

// GetPhase returns the current lifecycle phase
func (lm *DefaultLifecycleManager) GetPhase() Phase {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	return lm.phase
}

// HealthCheck performs a health check on all components
func (lm *DefaultLifecycleManager) HealthCheck(ctx context.Context) map[string]ComponentHealth {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	health := make(map[string]ComponentHealth)

	for name, state := range lm.states {
		if node, exists := lm.dag.GetNode(name); exists {
			componentHealth := node.Component.Health(ctx)
			health[name] = componentHealth

			// Update the stored state
			state.Health = componentHealth
		} else {
			health[name] = ComponentHealth{
				Status:    HealthStatusUnknown,
				Message:   "Component not found in DAG",
				Timestamp: time.Now(),
			}
		}
	}

	return health
}

// Private helper methods

// startComponentsInParallel starts multiple components in parallel
func (lm *DefaultLifecycleManager) startComponentsInParallel(ctx context.Context, nodes []*Node) error {
	if len(nodes) == 0 {
		return nil
	}

	if len(nodes) == 1 {
		// Single component, no need for goroutines
		return lm.startComponent(ctx, nodes[0])
	}

	// Use goroutines for parallel execution
	type result struct {
		node *Node
		err  error
	}

	results := make(chan result, len(nodes))
	
	// Start all components in parallel
	for _, node := range nodes {
		go func(n *Node) {
			err := lm.startComponent(ctx, n)
			results <- result{node: n, err: err}
		}(node)
	}

	// Collect results
	var errors []error
	for i := 0; i < len(nodes); i++ {
		res := <-results
		if res.err != nil {
			errors = append(errors, fmt.Errorf("component %s: %w", res.node.Name, res.err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to start components: %v", errors)
	}

	return nil
}

// startComponent starts a single component
func (lm *DefaultLifecycleManager) startComponent(ctx context.Context, node *Node) (err error) {
	// Add panic recovery
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during component start %s: %v", node.Name, r)
		}
	}()

	name := node.Name
	state := lm.states[name]

	if lm.logger != nil {
		lm.logger.Info("Starting component",
			"component", name,
		)
	}

	// Update state
	state.Phase = PhaseStartup
	now := time.Now()
	state.StartedAt = &now

	// Start the component with retry logic if configured
	var startErr error
	if retryConfig := node.Component.GetRetryConfig(); retryConfig != nil {
		startErr = RetryWithBackoff(ctx, *retryConfig, func() error {
			return node.Component.Start(ctx)
		})
	} else {
		startErr = node.Component.Start(ctx)
	}

	if startErr != nil {
		state.Phase = PhaseStopped
		state.Error = startErr
		state.StartedAt = nil
		return startErr
	}

	// Update state
	state.Phase = PhaseRunning
	state.Error = nil

	// Fire component-specific startup hooks
	if err := lm.fireHooks(ctx, PhaseStartup, name, map[string]interface{}{
		"component": name,
	}); err != nil {
		lm.logger.Warn("Failed to fire startup hooks", "component", name, "error", err.Error())
	}

	if lm.logger != nil {
		lm.logger.Info("Component started successfully",
			"component", name,
		)
	}

	return nil
}

// stopComponent stops a single component
func (lm *DefaultLifecycleManager) stopComponent(ctx context.Context, node *Node) error {
	name := node.Name
	state := lm.states[name]

	if state.Phase != PhaseRunning {
		if lm.logger != nil {
			lm.logger.Debug("Skipping component stop (not running)",
				"component", name,
				"phase", state.Phase,
			)
		}
		return nil
	}

	if lm.logger != nil {
		lm.logger.Info("Stopping component",
			"component", name,
		)
	}

	// Update state
	state.Phase = PhaseShutdown

	// Stop the component with retry logic if configured
	var stopErr error
	if retryConfig := node.Component.GetRetryConfig(); retryConfig != nil {
		stopErr = RetryWithBackoff(ctx, *retryConfig, func() error {
			return node.Component.Stop(ctx)
		})
	} else {
		stopErr = node.Component.Stop(ctx)
	}

	if stopErr != nil {
		state.Error = stopErr
		if lm.logger != nil {
			lm.logger.Error("Component stop failed",
				"component", name,
				"error", stopErr.Error(),
			)
		}
		// Continue with shutdown despite error
	}

	// Update state
	state.Phase = PhaseStopped
	now := time.Now()
	state.StoppedAt = &now

	// Fire component-specific shutdown hooks
	if err := lm.fireHooks(ctx, PhaseShutdown, name, map[string]interface{}{
		"component": name,
	}); err != nil {
		lm.logger.Warn("Failed to fire shutdown hooks", "component", name, "error", err.Error())
	}

	if lm.logger != nil {
		lm.logger.Info("Component stopped",
			"component", name,
		)
	}

	return state.Error
}

// rollbackStartup stops all components that were started before a failure
func (lm *DefaultLifecycleManager) rollbackStartup(ctx context.Context, failedComponent string) {
	if lm.logger != nil {
		lm.logger.Warn("Rolling back startup",
			"failed_component", failedComponent,
		)
	}

	// Get all nodes and stop those that are running
	nodes := lm.dag.GetAllNodes()
	for name, node := range nodes {
		if name == failedComponent {
			break // Don't try to stop the failed component
		}

		if state := lm.states[name]; state.Phase == PhaseRunning {
			if err := lm.stopComponent(ctx, node); err != nil {
				lm.logger.Warn("Failed to stop component during cleanup", "component", name, "error", err.Error())
			}
		}
	}
}

// getAllNodesInReverseOrder returns all nodes in reverse registration order as fallback
func (lm *DefaultLifecycleManager) getAllNodesInReverseOrder() []*Node {
	nodes := lm.dag.GetAllNodes()
	var result []*Node

	for _, node := range nodes {
		result = append(result, node)
	}

	// Reverse the order
	for i := 0; i < len(result)/2; i++ {
		j := len(result) - 1 - i
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// fireHooks executes all hooks for a given phase
func (lm *DefaultLifecycleManager) fireHooks(ctx context.Context, phase Phase, component string, data map[string]interface{}) error {
	hooks := lm.hooks[phase]
	if len(hooks) == 0 {
		return nil
	}

	event := Event{
		Phase:     phase,
		Component: component,
		Timestamp: time.Now(),
		Data:      data,
	}

	for _, hook := range hooks {
		if err := hook(ctx, event); err != nil {
			if lm.logger != nil {
				lm.logger.Error("Lifecycle hook failed",
					"phase", phase,
					"component", component,
					"error", err.Error(),
				)
			}
			return err
		}
	}

	return nil
}
