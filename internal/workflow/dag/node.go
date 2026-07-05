package dag

import (
	"workflow_engine/internal/workflow/runtime"
)

type Node struct {
	Name         string
	DependsOn    []string
	Run          func(*runtime.WorkflowContext) error
	Dependencies []*Node
	Dependents   []*Node
}

func NewNode(stepName string, dependsOn []string, run func(*runtime.WorkflowContext) error) *Node {
	return &Node{
		Name:         stepName,
		DependsOn:    dependsOn,
		Run:          run,
		Dependencies: make([]*Node, 0),
		Dependents:   make([]*Node, 0),
	}
}
