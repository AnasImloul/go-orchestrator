package di

import (
	"fmt"
	"sort"

	"github.com/AnasImloul/go-orchestrator/logger"
)

// DefaultContainerBuilder implements the ContainerBuilder interface
type DefaultContainerBuilder struct {
	providers []ServiceProvider
	modules   []Module
	config    ContainerConfig
	logger    logger.Logger
}

// NewContainerBuilder creates a new container builder
func NewContainerBuilder(logger logger.Logger) *DefaultContainerBuilder {
	return &DefaultContainerBuilder{
		providers: make([]ServiceProvider, 0),
		modules:   make([]Module, 0),
		config: ContainerConfig{
			EnableValidation:    true,
			EnableCircularCheck: true,
			EnableInterception:  true,
			DefaultLifetime:     Transient,
			MaxResolutionDepth:  50,
			EnableMetrics:       false,
		},
		logger: logger,
	}
}

// AddServiceProvider adds a service provider
func (b *DefaultContainerBuilder) AddServiceProvider(provider ServiceProvider) ContainerBuilder {
	b.providers = append(b.providers, provider)
	if b.logger != nil {
		b.logger.Debug("Service provider added",
			"provider", provider.GetName(),
		)
	}
	return b
}

// AddModule adds a module
func (b *DefaultContainerBuilder) AddModule(module Module) ContainerBuilder {
	b.modules = append(b.modules, module)
	if b.logger != nil {
		b.logger.Debug("Module added",
			"module", module.GetName(),
		)
	}
	return b
}

// Configure configures the container
func (b *DefaultContainerBuilder) Configure(config ContainerConfig) ContainerBuilder {
	b.config = config
	return b
}

// Build builds the container
func (b *DefaultContainerBuilder) Build() (Container, error) {
	if b.logger != nil {
		b.logger.Info("Building DI container",
			"providers", len(b.providers),
			"modules", len(b.modules),
		)
	}

	// Create container
	container := NewContainer(b.config, b.logger)

	// Sort providers by dependencies
	sortedProviders, err := b.sortProvidersByDependencies()
	if err != nil {
		return nil, fmt.Errorf("failed to sort providers by dependencies: %w", err)
	}

	// Register services from providers
	for _, provider := range sortedProviders {
		if b.logger != nil {
			b.logger.Debug("Registering services from provider",
				"provider", provider.GetName(),
			)
		}

		if err := provider.RegisterServices(container); err != nil {
			return nil, fmt.Errorf("failed to register services from provider %s: %w", provider.GetName(), err)
		}
	}

	// Sort modules by dependencies
	sortedModules, err := b.sortModulesByDependencies()
	if err != nil {
		return nil, fmt.Errorf("failed to sort modules by dependencies: %w", err)
	}

	// Register services from modules
	for _, module := range sortedModules {
		if b.logger != nil {
			b.logger.Debug("Registering services from module",
				"module", module.GetName(),
			)
		}

		if err := module.RegisterServices(container); err != nil {
			return nil, fmt.Errorf("failed to register services from module %s: %w", module.GetName(), err)
		}
	}

	if b.logger != nil {
		b.logger.Info("DI container built successfully",
			"registrations", len(container.GetRegistrations()),
		)
	}

	return container, nil
}

// Private helper methods

// sortProvidersByDependencies sorts providers by their dependencies
func (b *DefaultContainerBuilder) sortProvidersByDependencies() ([]ServiceProvider, error) {
	// Create dependency graph
	providerMap := make(map[string]ServiceProvider)
	dependencies := make(map[string][]string)

	for _, provider := range b.providers {
		name := provider.GetName()
		providerMap[name] = provider
		dependencies[name] = provider.GetDependencies()
	}

	// Topological sort
	sorted, err := b.topologicalSort(dependencies)
	if err != nil {
		return nil, err
	}

	// Convert back to provider slice
	result := make([]ServiceProvider, 0, len(sorted))
	for _, name := range sorted {
		result = append(result, providerMap[name])
	}

	return result, nil
}

// sortModulesByDependencies sorts modules by their dependencies
func (b *DefaultContainerBuilder) sortModulesByDependencies() ([]Module, error) {
	// Create dependency graph
	moduleMap := make(map[string]Module)
	dependencies := make(map[string][]string)

	for _, module := range b.modules {
		name := module.GetName()
		moduleMap[name] = module
		dependencies[name] = module.GetDependencies()
	}

	// Topological sort
	sorted, err := b.topologicalSort(dependencies)
	if err != nil {
		return nil, err
	}

	// Convert back to module slice
	result := make([]Module, 0, len(sorted))
	for _, name := range sorted {
		result = append(result, moduleMap[name])
	}

	return result, nil
}

// topologicalSort performs topological sorting on a dependency graph
func (b *DefaultContainerBuilder) topologicalSort(dependencies map[string][]string) ([]string, error) {
	// Kahn's algorithm for topological sorting
	inDegree := make(map[string]int)

	// Initialize in-degrees
	for name := range dependencies {
		inDegree[name] = 0
	}

	// Calculate in-degrees
	for name, deps := range dependencies {
		for _, dep := range deps {
			if _, exists := dependencies[dep]; !exists {
				return nil, fmt.Errorf("dependency %s not found for %s", dep, name)
			}
			inDegree[name]++
		}
	}

	// Find nodes with no incoming edges
	queue := make([]string, 0)
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	// Sort queue for deterministic order
	sort.Strings(queue)

	var result []string

	for len(queue) > 0 {
		// Remove node from queue
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Update in-degrees of dependent nodes
		for name, deps := range dependencies {
			for _, dep := range deps {
				if dep == current {
					inDegree[name]--
					if inDegree[name] == 0 {
						queue = append(queue, name)
						sort.Strings(queue) // Keep queue sorted
					}
				}
			}
		}
	}

	// Check for circular dependencies
	if len(result) != len(dependencies) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
}

// DefaultServiceProvider provides a base implementation for service providers
type DefaultServiceProvider struct {
	name         string
	dependencies []string
}

// NewServiceProvider creates a new service provider
func NewServiceProvider(name string, dependencies ...string) *DefaultServiceProvider {
	return &DefaultServiceProvider{
		name:         name,
		dependencies: dependencies,
	}
}

// GetName returns the provider name
func (p *DefaultServiceProvider) GetName() string {
	return p.name
}

// GetDependencies returns provider dependencies
func (p *DefaultServiceProvider) GetDependencies() []string {
	return p.dependencies
}

// RegisterServices must be implemented by concrete providers
func (p *DefaultServiceProvider) RegisterServices(container Container) error {
	return fmt.Errorf("RegisterServices not implemented for provider %s", p.name)
}

// DefaultModule provides a base implementation for modules
type DefaultModule struct {
	*DefaultServiceProvider
	config ModuleConfig
}

// NewModule creates a new module
func NewModule(name string, config ModuleConfig, dependencies ...string) *DefaultModule {
	return &DefaultModule{
		DefaultServiceProvider: NewServiceProvider(name, dependencies...),
		config:                 config,
	}
}

// Configure configures the module
func (m *DefaultModule) Configure(config ModuleConfig) error {
	m.config = config
	return nil
}

// GetConfig returns the module configuration
func (m *DefaultModule) GetConfig() ModuleConfig {
	return m.config
}
