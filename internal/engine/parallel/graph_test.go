package parallel

import (
	"testing"
)

// TestTopologicalSort tests basic topological sorting
func TestTopologicalSort(t *testing.T) {
	graph := NewDependencyGraph()

	// Create dependency chain: A -> B -> C
	graph.AddNode("A", []TaskID{})
	graph.AddNode("B", []TaskID{"A"})
	graph.AddNode("C", []TaskID{"B"})

	sorted, err := graph.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort failed: %v", err)
	}

	if len(sorted) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(sorted))
	}

	// Verify order: A must come before B, B must come before C
	posA, posB, posC := -1, -1, -1
	for i, taskID := range sorted {
		switch taskID {
		case "A":
			posA = i
		case "B":
			posB = i
		case "C":
			posC = i
		}
	}

	if posA >= posB || posB >= posC {
		t.Fatalf("Invalid sort order: A=%d, B=%d, C=%d", posA, posB, posC)
	}
}

// TestTopologicalSortParallel tests parallel tasks
func TestTopologicalSortParallel(t *testing.T) {
	graph := NewDependencyGraph()

	// Create parallel tasks: A, B, C (no dependencies)
	graph.AddNode("A", []TaskID{})
	graph.AddNode("B", []TaskID{})
	graph.AddNode("C", []TaskID{})

	sorted, err := graph.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort failed: %v", err)
	}

	if len(sorted) != 3 {
		t.Fatalf("Expected 3 tasks, got %d", len(sorted))
	}

	// All tasks can be in any order (no dependencies)
}

// TestTopologicalSortDiamond tests diamond dependency
func TestTopologicalSortDiamond(t *testing.T) {
	graph := NewDependencyGraph()

	// Diamond: A -> B, A -> C, B -> D, C -> D
	graph.AddNode("A", []TaskID{})
	graph.AddNode("B", []TaskID{"A"})
	graph.AddNode("C", []TaskID{"A"})
	graph.AddNode("D", []TaskID{"B", "C"})

	sorted, err := graph.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort failed: %v", err)
	}

	if len(sorted) != 4 {
		t.Fatalf("Expected 4 tasks, got %d", len(sorted))
	}

	// Find positions
	pos := make(map[TaskID]int)
	for i, taskID := range sorted {
		pos[taskID] = i
	}

	// Verify constraints
	if pos["A"] >= pos["B"] {
		t.Fatal("A must come before B")
	}
	if pos["A"] >= pos["C"] {
		t.Fatal("A must come before C")
	}
	if pos["B"] >= pos["D"] {
		t.Fatal("B must come before D")
	}
	if pos["C"] >= pos["D"] {
		t.Fatal("C must come before D")
	}
}

// TestCycleDetection tests cycle detection
func TestCycleDetection(t *testing.T) {
	graph := NewDependencyGraph()

	// Create cycle: A -> B -> C -> A
	graph.AddNode("A", []TaskID{"C"})
	graph.AddNode("B", []TaskID{"A"})
	graph.AddNode("C", []TaskID{"B"})

	_, err := graph.TopologicalSort()
	if err == nil {
		t.Fatal("Expected cycle detection error, got nil")
	}

	if !graph.HasCycle() {
		t.Fatal("Expected cycle detection to return true")
	}
}

// TestMissingDependency tests missing dependency detection
func TestMissingDependency(t *testing.T) {
	graph := NewDependencyGraph()

	// Task B depends on non-existent task A
	graph.AddNode("B", []TaskID{"A"})

	_, err := graph.TopologicalSort()
	if err == nil {
		t.Fatal("Expected missing dependency error, got nil")
	}
}

// TestGroupByWave tests wave grouping
func TestGroupByWave(t *testing.T) {
	graph := NewDependencyGraph()

	// Wave 0: A, B (no deps)
	// Wave 1: C (depends on A), D (depends on B)
	// Wave 2: E (depends on C and D)
	graph.AddNode("A", []TaskID{})
	graph.AddNode("B", []TaskID{})
	graph.AddNode("C", []TaskID{"A"})
	graph.AddNode("D", []TaskID{"B"})
	graph.AddNode("E", []TaskID{"C", "D"})

	waves, err := graph.GroupByWave()
	if err != nil {
		t.Fatalf("GroupByWave failed: %v", err)
	}

	if len(waves) != 3 {
		t.Fatalf("Expected 3 waves, got %d", len(waves))
	}

	// Wave 0: A, B
	if len(waves[0]) != 2 {
		t.Fatalf("Expected 2 tasks in wave 0, got %d", len(waves[0]))
	}

	// Wave 1: C, D
	if len(waves[1]) != 2 {
		t.Fatalf("Expected 2 tasks in wave 1, got %d", len(waves[1]))
	}

	// Wave 2: E
	if len(waves[2]) != 1 {
		t.Fatalf("Expected 1 task in wave 2, got %d", len(waves[2]))
	}
}
