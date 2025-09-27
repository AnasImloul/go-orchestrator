package di

import (
	"context"
	"fmt"
	"reflect"
	"sync"

	"github.com/imloulanas/go-orchestrator/logger"
)

// DefaultScope implements the Scope interface
type DefaultScope struct {
	container       *DefaultContainer
	scopedInstances map[reflect.Type]interface{}
	logger          logger.Logger
	mu              sync.RWMutex
	disposed        bool
}

// NewScope creates a new DI scope
func NewScope(container *DefaultContainer, logger logger.Logger) *DefaultScope {
	return &DefaultScope{
		container:       container,
		scopedInstances: make(map[reflect.Type]interface{}),
		logger:          logger,
	}
}

// Resolve resolves a service within this scope
func (s *DefaultScope) Resolve(serviceType reflect.Type) (interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.disposed {
		return nil, fmt.Errorf("scope is disposed")
	}

	// Check if we have a scoped instance
	if instance, exists := s.scopedInstances[serviceType]; exists {
		return instance, nil
	}

	// Get registration from container
	registration, exists := s.container.registrations[serviceType]
	if !exists {
		return nil, fmt.Errorf("service of type %s is not registered", serviceType.String())
	}

	// Handle different lifetimes
	switch registration.Lifetime {
	case Singleton:
		// Singletons are resolved from the container
		return s.container.resolve(context.Background(), serviceType, 0)

	case Scoped:
		// Create scoped instance
		instance, err := s.createScopedInstance(registration)
		if err != nil {
			return nil, err
		}
		s.scopedInstances[serviceType] = instance
		return instance, nil

	case Transient:
		// Transients are always created new
		return s.createTransientInstance(registration)

	default:
		return nil, fmt.Errorf("unsupported service lifetime: %v", registration.Lifetime)
	}
}

// ResolveByName resolves a service by name within this scope
func (s *DefaultScope) ResolveByName(name string) (interface{}, error) {
	s.mu.RLock()
	registration, exists := s.container.namedServices[name]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("service with name '%s' not found", name)
	}

	return s.Resolve(registration.ServiceType)
}

// Dispose disposes the scope and all scoped instances
func (s *DefaultScope) Dispose() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.disposed {
		return nil
	}

	s.disposed = true

	// Dispose all scoped instances that implement disposable interface
	for serviceType, instance := range s.scopedInstances {
		if disposable, ok := instance.(Disposable); ok {
			if err := disposable.Dispose(); err != nil {
				if s.logger != nil {
					s.logger.Error("Failed to dispose scoped instance",
						"type", serviceType.String(),
						"error", err.Error(),
					)
				}
			}
		}
	}

	// Clear scoped instances
	s.scopedInstances = nil

	if s.logger != nil {
		s.logger.Debug("Scope disposed")
	}

	return nil
}

// Private helper methods

// createScopedInstance creates a scoped service instance
func (s *DefaultScope) createScopedInstance(registration *ServiceRegistration) (interface{}, error) {
	if registration.Factory == nil {
		return nil, fmt.Errorf("no factory provided for scoped service %s", registration.ServiceType.String())
	}

	// Create instance using factory with scope context
	ctx := context.Background()
	instance, err := registration.Factory(ctx, s.container)
	if err != nil {
		return nil, err
	}

	if s.logger != nil {
		s.logger.Debug("Scoped instance created",
			"type", registration.ServiceType.String(),
		)
	}

	return instance, nil
}

// createTransientInstance creates a transient service instance
func (s *DefaultScope) createTransientInstance(registration *ServiceRegistration) (interface{}, error) {
	if registration.Factory == nil {
		return nil, fmt.Errorf("no factory provided for transient service %s", registration.ServiceType.String())
	}

	// Create instance using factory
	ctx := context.Background()
	instance, err := registration.Factory(ctx, s.container)
	if err != nil {
		return nil, err
	}

	if s.logger != nil {
		s.logger.Debug("Transient instance created",
			"type", registration.ServiceType.String(),
		)
	}

	return instance, nil
}
