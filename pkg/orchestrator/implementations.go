package orchestrator

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/AnasImloul/go-orchestrator/pkg/di"
	"github.com/AnasImloul/go-orchestrator/pkg/lifecycle"
	"github.com/AnasImloul/go-orchestrator/pkg/logger"
)

// simpleLogger is a basic logger implementation.
type simpleLogger struct{}

func (l *simpleLogger) Debug(msg string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+msg+"\n", args...)
}

func (l *simpleLogger) Info(msg string, args ...interface{}) {
	fmt.Printf("[INFO] "+msg+"\n", args...)
}

func (l *simpleLogger) Warn(msg string, args ...interface{}) {
	fmt.Printf("[WARN] "+msg+"\n", args...)
}

func (l *simpleLogger) Error(msg string, args ...interface{}) {
	fmt.Printf("[ERROR] "+msg+"\n", args...)
}

func (l *simpleLogger) WithComponent(component string) logger.Logger {
	return &simpleLogger{}
}

// basicContainer is a functional DI container implementation.
type basicContainer struct {
	services map[string]interface{}
	mu       sync.RWMutex
}

func (c *basicContainer) Register(serviceType reflect.Type, factory di.Factory, options ...di.Option) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	key := serviceType.String()
	c.services[key] = factory
	return nil
}

func (c *basicContainer) RegisterInstance(serviceType reflect.Type, instance interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	key := serviceType.String()
	c.services[key] = instance
	return nil
}

func (c *basicContainer) RegisterSingleton(serviceType reflect.Type, factory di.Factory, options ...di.Option) error {
	return c.Register(serviceType, factory, options...)
}

func (c *basicContainer) Resolve(serviceType reflect.Type) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	key := serviceType.String()
	if service, exists := c.services[key]; exists {
		if factory, ok := service.(di.Factory); ok {
			return factory(context.Background(), c)
		}
		return service, nil
	}
	
	return nil, fmt.Errorf("service %s not found", key)
}

func (c *basicContainer) CreateScope() di.Scope {
	return &basicScope{container: c}
}

func (c *basicContainer) Validate() error {
	return nil
}

// basicScope is a functional scope implementation.
type basicScope struct {
	container *basicContainer
}

func (s *basicScope) Resolve(serviceType reflect.Type) (interface{}, error) {
	return s.container.Resolve(serviceType)
}

func (s *basicScope) Dispose() error {
	return nil
}

// basicLifecycleManager is a functional lifecycle manager implementation.
type basicLifecycleManager struct {
	components map[string]lifecycle.Component
	phase      lifecycle.Phase
	logger     logger.Logger
	mu         sync.RWMutex
}

func (lm *basicLifecycleManager) RegisterComponent(component lifecycle.Component) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	name := component.Name()
	if _, exists := lm.components[name]; exists {
		return fmt.Errorf("component %s is already registered", name)
	}
	
	lm.components[name] = component
	lm.logger.Info("Component registered", "name", name)
	return nil
}

func (lm *basicLifecycleManager) UnregisterComponent(name string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	delete(lm.components, name)
	lm.logger.Info("Component unregistered", "name", name)
	return nil
}

func (lm *basicLifecycleManager) Start(ctx context.Context) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	lm.phase = lifecycle.PhaseStarting
	lm.logger.Info("Starting lifecycle manager")
	
	// Simple implementation - start all components
	for name, component := range lm.components {
		lm.logger.Info("Starting component", "name", name)
		if err := component.Start(ctx); err != nil {
			lm.logger.Error("Failed to start component", "name", name, "error", err)
			lm.phase = lifecycle.PhaseFailed
			return err
		}
	}
	
	lm.phase = lifecycle.PhaseRunning
	lm.logger.Info("All components started successfully")
	return nil
}

func (lm *basicLifecycleManager) Stop(ctx context.Context) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	
	lm.phase = lifecycle.PhaseStopping
	lm.logger.Info("Stopping lifecycle manager")
	
	// Simple implementation - stop all components
	for name, component := range lm.components {
		lm.logger.Info("Stopping component", "name", name)
		if err := component.Stop(ctx); err != nil {
			lm.logger.Error("Failed to stop component", "name", name, "error", err)
		}
	}
	
	lm.phase = lifecycle.PhaseStopped
	lm.logger.Info("All components stopped")
	return nil
}

func (lm *basicLifecycleManager) GetPhase() lifecycle.Phase {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	return lm.phase
}

func (lm *basicLifecycleManager) GetComponentState(name string) (*lifecycle.ComponentState, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	
	component, exists := lm.components[name]
	if !exists {
		return nil, fmt.Errorf("component %s not found", name)
	}
	
	now := time.Now()
	state := &lifecycle.ComponentState{
		Name:         component.Name(),
		Phase:        lm.phase,
		Dependencies: component.Dependencies(),
		Priority:     component.Priority(),
		Health:       component.Health(context.Background()),
		Metadata:     make(map[string]string),
	}
	
	if lm.phase == lifecycle.PhaseRunning {
		state.StartedAt = &now
	}
	
	return state, nil
}

func (lm *basicLifecycleManager) GetAllComponentStates() map[string]*lifecycle.ComponentState {
	lm.mu.RLock()
	defer lm.mu.RUnlock()
	
	states := make(map[string]*lifecycle.ComponentState)
	for name := range lm.components {
		if state, err := lm.GetComponentState(name); err == nil {
			states[name] = state
		}
	}
	
	return states
}

func (lm *basicLifecycleManager) AddHook(phase lifecycle.Phase, hook lifecycle.Hook) error {
	// Basic implementation - hooks not supported
	return nil
}

func (lm *basicLifecycleManager) RemoveHook(phase lifecycle.Phase, hook lifecycle.Hook) error {
	// Basic implementation - hooks not supported
	return nil
}
