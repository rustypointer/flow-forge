package dag

import (
	"fmt"
	"workflow_engine/internal/workflow/errs"
)

type color uint8

const (
	white color = iota // unvisited
	gray               // visiting
	black              // visited
)

func Validate(graph *Graph) error {
	if len(graph.Nodes) == 0 {
		return fmt.Errorf("%w: cannot validate an empty graph", errs.ErrGraphEmpty)
	}

	if len(graph.Roots()) == 0 {
		return fmt.Errorf("%w: graph has no valid entry points (missing root nodes)", errs.ErrGraphNoEntryPoints)
	}

	if err := checkCycles(graph); err != nil {
		return err
	}

	return nil
}

func checkCycles(graph *Graph) error {
	colors := make(map[string]color)

	var dfs func(*Node) error
	dfs = func(node *Node) error {
		stepName := node.Name

		// looped back to currently visiting node; cycle detected
		if colors[stepName] == gray {
			return fmt.Errorf("%w: cyclic dependency detected involving step %q", errs.ErrGraphCyclicDependency, stepName)
		}

		// already visited node; return
		if colors[stepName] == black {
			return nil
		}

		// new path; mark currently visiting
		colors[stepName] = gray

		// walk forward to all dependents
		for _, dependentNode := range node.Dependents {
			if err := dfs(dependentNode); err != nil {
				return err
			}
		}

		// fully explored; mark visited
		colors[stepName] = black
		return nil
	}

	// check for all nodes (covers disconnected components)
	for stepName, node := range graph.Nodes {
		if colors[stepName] == white {
			if err := dfs(node); err != nil {
				return err
			}
		}
	}

	return nil
}
