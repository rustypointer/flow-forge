package dag_test

import (
	"errors"
	"testing"
	"workflow_engine/internal/workflow/dag"
	"workflow_engine/internal/workflow/errs"
)

func TestBuildExecutionStages(t *testing.T) {
	tests := []struct {
		name          string
		setupGraph    func() *dag.Graph
		expectErr     error
		expectedCount int // Number of expected layers/stages
		verifyStages  func(t *testing.T, stages [][]*dag.Node)
	}{
		{
			name: "Error: Empty Graph",
			setupGraph: func() *dag.Graph {
				return &dag.Graph{Nodes: make(map[string]*dag.Node)}
			},
			expectErr: errs.ErrGraphEmpty,
		},
		{
			name: "Error: No Roots (Pure Cyclic Loop)",
			setupGraph: func() *dag.Graph {
				// A -> B -> A (Every node has dependencies, no entry points)
				nA := &dag.Node{Name: "A"}
				nB := &dag.Node{Name: "B"}

				nA.Dependencies = []*dag.Node{nB}
				nA.Dependents = []*dag.Node{nB}
				nB.Dependencies = []*dag.Node{nA}
				nB.Dependents = []*dag.Node{nA}

				return &dag.Graph{
					Nodes: map[string]*dag.Node{"A": nA, "B": nB},
				}
			},
			expectErr: errs.ErrGraphNoEntryPoints,
		},
		{
			name: "Error: Partial Cycle (Has Root, But Detached Loop Exists)",
			setupGraph: func() *dag.Graph {
				// Valid path: Root -> Child
				nRoot := &dag.Node{Name: "Root"}
				nChild := &dag.Node{Name: "Child"}
				nChild.Dependencies = []*dag.Node{nRoot}
				nRoot.Dependents = []*dag.Node{nChild}

				// Cyclic Island: C -> D -> C
				nC := &dag.Node{Name: "C"}
				nD := &dag.Node{Name: "D"}
				nC.Dependencies = []*dag.Node{nD}
				nC.Dependents = []*dag.Node{nD}
				nD.Dependencies = []*dag.Node{nC}
				nD.Dependents = []*dag.Node{nC}

				return &dag.Graph{
					Nodes: map[string]*dag.Node{
						"Root": nRoot, "Child": nChild,
						"C": nC, "D": nD,
					},
				}
			},
			expectErr: errs.ErrGraphCyclicDependency,
		},
		{
			name: "Success: Linear Topology (A -> B -> C)",
			setupGraph: func() *dag.Graph {
				nA := &dag.Node{Name: "A"}
				nB := &dag.Node{Name: "B"}
				nC := &dag.Node{Name: "C"}

				nA.Dependents = []*dag.Node{nB}
				nB.Dependencies = []*dag.Node{nA}
				nB.Dependents = []*dag.Node{nC}
				nC.Dependencies = []*dag.Node{nB}

				return &dag.Graph{
					Nodes: map[string]*dag.Node{"A": nA, "B": nB, "C": nC},
				}
			},
			expectedCount: 3,
			verifyStages: func(t *testing.T, stages [][]*dag.Node) {
				if stages[0][0].Name != "A" {
					t.Errorf("Stage 0 should be A, got %s", stages[0][0].Name)
				}
				if stages[1][0].Name != "B" {
					t.Errorf("Stage 1 should be B, got %s", stages[1][0].Name)
				}
				if stages[2][0].Name != "C" {
					t.Errorf("Stage 2 should be C, got %s", stages[2][0].Name)
				}
			},
		},
		{
			name: "Success: Fan-Out Topology (A -> [B, C, D])",
			setupGraph: func() *dag.Graph {
				nA := &dag.Node{Name: "A"}
				nB := &dag.Node{Name: "B"}
				nC := &dag.Node{Name: "C"}
				nD := &dag.Node{Name: "D"}

				nA.Dependents = []*dag.Node{nB, nC, nD}
				nB.Dependencies = []*dag.Node{nA}
				nC.Dependencies = []*dag.Node{nA}
				nD.Dependencies = []*dag.Node{nA}

				return &dag.Graph{
					Nodes: map[string]*dag.Node{"A": nA, "B": nB, "C": nC, "D": nD},
				}
			},
			expectedCount: 2,
			verifyStages: func(t *testing.T, stages [][]*dag.Node) {
				if len(stages[1]) != 3 {
					t.Errorf("Stage 1 should have 3 parallel nodes, got %d", len(stages[1]))
				}
				names := map[string]bool{}
				for _, node := range stages[1] {
					names[node.Name] = true
				}
				if !names["B"] || !names["C"] || !names["D"] {
					t.Errorf("Stage 1 elements missing or misallocated: %v", names)
				}
			},
		},
		{
			name: "Success: Fan-In Topology ([B, C] -> D)",
			setupGraph: func() *dag.Graph {
				nB := &dag.Node{Name: "B"}
				nC := &dag.Node{Name: "C"}
				nD := &dag.Node{Name: "D"}

				nB.Dependents = []*dag.Node{nD}
				nC.Dependents = []*dag.Node{nD}
				nD.Dependencies = []*dag.Node{nB, nC}

				return &dag.Graph{
					Nodes: map[string]*dag.Node{"B": nB, "C": nC, "D": nD},
				}
			},
			expectedCount: 2,
			verifyStages: func(t *testing.T, stages [][]*dag.Node) {
				if len(stages[0]) != 2 {
					t.Errorf("Stage 0 should have 2 root nodes, got %d", len(stages[0]))
				}
				if stages[1][0].Name != "D" {
					t.Errorf("Stage 1 should be D, got %s", stages[1][0].Name)
				}
			},
		},
		{
			name: "Success: Disconnected Components ([A -> B] and [X -> Y])",
			setupGraph: func() *dag.Graph {
				nA := &dag.Node{Name: "A"}
				nB := &dag.Node{Name: "B"}
				nX := &dag.Node{Name: "X"}
				nY := &dag.Node{Name: "Y"}

				nA.Dependents = []*dag.Node{nB}
				nB.Dependencies = []*dag.Node{nA}

				nX.Dependents = []*dag.Node{nY}
				nY.Dependencies = []*dag.Node{nX}

				return &dag.Graph{
					Nodes: map[string]*dag.Node{"A": nA, "B": nB, "X": nX, "Y": nY},
				}
			},
			expectedCount: 2,
			verifyStages: func(t *testing.T, stages [][]*dag.Node) {
				if len(stages[0]) != 2 {
					t.Errorf("Stage 0 should combine roots from both islands, got %d", len(stages[0]))
				}
				if len(stages[1]) != 2 {
					t.Errorf("Stage 1 should combine dependents from both islands, got %d", len(stages[1]))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			graph := tt.setupGraph()

			stages, err := dag.BuildExecutionStages(graph)

			if tt.expectErr != nil {
				if err == nil {
					t.Fatalf("Expected error %v, got nil", tt.expectErr)
				}
				if !errors.Is(err, tt.expectErr) {
					t.Fatalf("Expected error type %v, got %v", tt.expectErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(stages) != tt.expectedCount {
				t.Errorf("Expected %d execution stages, got %d", tt.expectedCount, len(stages))
			}

			if tt.verifyStages != nil {
				tt.verifyStages(t, stages)
			}
		})
	}
}
