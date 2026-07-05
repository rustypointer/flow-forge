package dag

import (
	"fmt"
	"workflow_engine/internal/workflow/errs"
)

func BuildGraph(nodes []*Node) (*Graph, error) {
	graph := NewGraph(nodes)

	// connect edges
	for stepName, currNode := range graph.Nodes {
		for _, dependencyNodeName := range currNode.DependsOn {
			dependencyNode, exists := graph.Nodes[dependencyNodeName]
			if !exists {
				return nil, fmt.Errorf("%w: step %q depends on unknown step %q", errs.ErrWorkflowMissingDependency, stepName, dependencyNodeName)
			}

			// inward edge on current node from dependency node; c <- d
			currNode.Dependencies = append(currNode.Dependencies, dependencyNode)
			// outward edge from dependency node to current node; d -> c
			dependencyNode.Dependents = append(dependencyNode.Dependents, currNode)
		}
	}

	return graph, nil
}
