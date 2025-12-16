package parallel

import (
	"fmt"
	"sync"
)

// DependencyGraph manages task dependencies and determines execution order
type DependencyGraph struct {
	nodes map[TaskID][]TaskID // node -> dependencies
	mu    sync.RWMutex
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		nodes: make(map[TaskID][]TaskID),
	}
}

// AddNode adds a node with its dependencies
func (g *DependencyGraph) AddNode(taskID TaskID, dependencies []TaskID) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.nodes[taskID] = dependencies
}

// GetDependencies returns dependencies for a task
func (g *DependencyGraph) GetDependencies(taskID TaskID) []TaskID {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.nodes[taskID]
}

// TopologicalSort returns tasks in execution order using Kahn's algorithm
// Returns error if cycle detected
func (g *DependencyGraph) TopologicalSort() ([]TaskID, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Build in-degree map and adjacency list
	inDegree := make(map[TaskID]int)
	adjList := make(map[TaskID][]TaskID)

	// Initialize in-degree for all nodes
	for node := range g.nodes {
		if _, ok := inDegree[node]; !ok {
			inDegree[node] = 0
		}
	}

	// Build adjacency list (reverse of dependencies)
	for node, deps := range g.nodes {
		for _, dep := range deps {
			// dep -> node (node depends on dep)
			adjList[dep] = append(adjList[dep], node)
			inDegree[node]++
		}
	}

	// Find nodes with no dependencies (in-degree = 0)
	queue := []TaskID{}
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	// Process nodes in topological order
	result := []TaskID{}
	for len(queue) > 0 {
		// Dequeue
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		// Process neighbors
		for _, neighbor := range adjList[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	// Check for cycle
	if len(result) != len(g.nodes) {
		return nil, fmt.Errorf("cycle detected in dependency graph")
	}

	return result, nil
}

// GroupByWave groups tasks into waves where tasks in same wave can run in parallel
// Wave N contains tasks whose dependencies are all satisfied by waves 0..N-1
func (g *DependencyGraph) GroupByWave() ([][]TaskID, error) {
	order, err := g.TopologicalSort()
	if err != nil {
		return nil, err
	}

	// Map task to wave number
	waveMap := make(map[TaskID]int)

	// Assign wave numbers
	for _, taskID := range order {
		deps := g.GetDependencies(taskID)

		// Find maximum wave of dependencies
		maxDepWave := -1
		for _, dep := range deps {
			if depWave, ok := waveMap[dep]; ok {
				if depWave > maxDepWave {
					maxDepWave = depWave
				}
			}
		}

		// This task goes in wave after max dependency wave
		waveMap[taskID] = maxDepWave + 1
	}

	// Group tasks by wave
	maxWave := 0
	for _, wave := range waveMap {
		if wave > maxWave {
			maxWave = wave
		}
	}

	waves := make([][]TaskID, maxWave+1)
	for taskID, wave := range waveMap {
		waves[wave] = append(waves[wave], taskID)
	}

	return waves, nil
}

// HasCycle checks if graph has a cycle
func (g *DependencyGraph) HasCycle() bool {
	_, err := g.TopologicalSort()
	return err != nil
}

// Validate checks if all dependencies exist as nodes
func (g *DependencyGraph) Validate() error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	for node, deps := range g.nodes {
		for _, dep := range deps {
			if _, exists := g.nodes[dep]; !exists {
				return fmt.Errorf("task %s depends on non-existent task %s", node, dep)
			}
		}
	}

	return nil
}

// Size returns number of nodes in graph
func (g *DependencyGraph) Size() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return len(g.nodes)
}
