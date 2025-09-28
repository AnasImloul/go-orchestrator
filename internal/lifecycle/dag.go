package lifecycle

import (
	"fmt"
)

// DAG represents a Directed Acyclic Graph for dependency resolution
type DAG struct {
	nodes map[string]*Node
	edges map[string][]string
}

// Node represents a node in the DAG
type Node struct {
	Name         string
	Component    Component
	Dependencies []string
	visited      bool
	visiting     bool
}

// NewDAG creates a new DAG
func NewDAG() *DAG {
	return &DAG{
		nodes: make(map[string]*Node),
		edges: make(map[string][]string),
	}
}

// AddNode adds a component to the DAG
func (d *DAG) AddNode(component Component) error {
	name := component.Name()

	if _, exists := d.nodes[name]; exists {
		return fmt.Errorf("component %s already exists in DAG", name)
	}

	node := &Node{
		Name:         name,
		Component:    component,
		Dependencies: component.Dependencies(),
	}

	d.nodes[name] = node
	d.edges[name] = component.Dependencies()

	return nil
}

// RemoveNode removes a component from the DAG
func (d *DAG) RemoveNode(name string) error {
	if _, exists := d.nodes[name]; !exists {
		return fmt.Errorf("component %s does not exist in DAG", name)
	}

	// Check if any other nodes depend on this node
	for nodeName, deps := range d.edges {
		if nodeName == name {
			continue
		}
		for _, dep := range deps {
			if dep == name {
				return fmt.Errorf("cannot remove component %s: component %s depends on it", name, nodeName)
			}
		}
	}

	delete(d.nodes, name)
	delete(d.edges, name)

	return nil
}

// ValidateDependencies validates that all dependencies exist and there are no cycles
func (d *DAG) ValidateDependencies() error {
	// Check that all dependencies exist
	for name, deps := range d.edges {
		for _, dep := range deps {
			if _, exists := d.nodes[dep]; !exists {
				return fmt.Errorf("component %s has missing dependency: %s", name, dep)
			}
		}
	}

	// Check for cycles using DFS
	for name := range d.nodes {
		d.resetVisited()
		if err := d.hasCycle(name); err != nil {
			return err
		}
	}

	return nil
}

// GetStartupLevels returns components grouped by dependency level for parallel execution
func (d *DAG) GetStartupLevels() ([][]*Node, error) {
	if err := d.ValidateDependencies(); err != nil {
		return nil, err
	}

	// Calculate dependency levels for each node
	levels := make(map[string]int)
	d.calculateLevels(levels)

	// Group nodes by level
	levelGroups := make(map[int][]*Node)
	for name, node := range d.nodes {
		level := levels[name]
		levelGroups[level] = append(levelGroups[level], node)
	}

	// Convert to ordered slice of levels
	var result [][]*Node
	maxLevel := 0
	for level := range levelGroups {
		if level > maxLevel {
			maxLevel = level
		}
	}

	for level := 0; level <= maxLevel; level++ {
		if group, exists := levelGroups[level]; exists {
			result = append(result, group)
		}
	}

	return result, nil
}

// GetStartupOrder returns components in the order they should be started (for backward compatibility)
func (d *DAG) GetStartupOrder() ([]*Node, error) {
	levels, err := d.GetStartupLevels()
	if err != nil {
		return nil, err
	}

	var result []*Node
	for _, level := range levels {
		result = append(result, level...)
	}

	return result, nil
}

// GetShutdownOrder returns components in the order they should be stopped (reverse of startup)
func (d *DAG) GetShutdownOrder() ([]*Node, error) {
	startupOrder, err := d.GetStartupOrder()
	if err != nil {
		return nil, err
	}

	// Reverse the startup order for shutdown
	shutdownOrder := make([]*Node, len(startupOrder))
	for i, node := range startupOrder {
		shutdownOrder[len(startupOrder)-1-i] = node
	}

	return shutdownOrder, nil
}

// GetNode returns a node by name
func (d *DAG) GetNode(name string) (*Node, bool) {
	node, exists := d.nodes[name]
	return node, exists
}

// GetAllNodes returns all nodes
func (d *DAG) GetAllNodes() map[string]*Node {
	// Return a copy to prevent external modification
	nodes := make(map[string]*Node)
	for name, node := range d.nodes {
		nodes[name] = node
	}
	return nodes
}

// GetDependents returns all components that depend on the given component
func (d *DAG) GetDependents(name string) []string {
	var dependents []string

	for nodeName, deps := range d.edges {
		if nodeName == name {
			continue
		}
		for _, dep := range deps {
			if dep == name {
				dependents = append(dependents, nodeName)
				break
			}
		}
	}

	return dependents
}

// GetDependencies returns the direct dependencies of a component
func (d *DAG) GetDependencies(name string) []string {
	if deps, exists := d.edges[name]; exists {
		// Return a copy to prevent external modification
		result := make([]string, len(deps))
		copy(result, deps)
		return result
	}
	return nil
}

// resetVisited resets the visited flags for all nodes
func (d *DAG) resetVisited() {
	for _, node := range d.nodes {
		node.visited = false
		node.visiting = false
	}
}

// hasCycle performs DFS to detect cycles
func (d *DAG) hasCycle(name string) error {
	node := d.nodes[name]

	if node.visiting {
		return fmt.Errorf("circular dependency detected involving component: %s", name)
	}

	if node.visited {
		return nil
	}

	node.visiting = true

	for _, dep := range node.Dependencies {
		if err := d.hasCycle(dep); err != nil {
			return err
		}
	}

	node.visiting = false
	node.visited = true

	return nil
}

// topologicalSort performs topological sorting using DFS
func (d *DAG) topologicalSort(name string, result *[]*Node) error {
	node := d.nodes[name]

	if node.visiting {
		return fmt.Errorf("circular dependency detected involving component: %s", name)
	}

	if node.visited {
		return nil
	}

	node.visiting = true

	// Visit all dependencies first
	for _, dep := range node.Dependencies {
		if err := d.topologicalSort(dep, result); err != nil {
			return err
		}
	}

	node.visiting = false
	node.visited = true

	// Add to result after all dependencies
	*result = append(*result, node)

	return nil
}

// calculateLevels calculates the dependency level for each node
func (d *DAG) calculateLevels(levels map[string]int) {
	d.resetVisited()

	for name := range d.nodes {
		if !d.nodes[name].visited {
			d.calculateLevel(name, levels)
		}
	}
}

// calculateLevel calculates the dependency level for a specific node
func (d *DAG) calculateLevel(name string, levels map[string]int) int {
	node := d.nodes[name]

	if node.visited {
		return levels[name]
	}

	node.visited = true

	maxDepLevel := -1
	for _, dep := range node.Dependencies {
		depLevel := d.calculateLevel(dep, levels)
		if depLevel > maxDepLevel {
			maxDepLevel = depLevel
		}
	}

	level := maxDepLevel + 1
	levels[name] = level

	return level
}

