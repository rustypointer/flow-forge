package dag

import (
	"fmt"
	"sort"
	"workflow_engine/internal/workflow/errs"
)

func BuildExecutionStages(graph *Graph) ([][]*Node, error) {
	if len(graph.Nodes) == 0 {
		return nil, fmt.Errorf("%w: cannot validate an empty graph", errs.ErrGraphEmpty)
	}

	var stages [][]*Node
	inDegrees := graph.InDegrees()
	currentStage := graph.Roots()
	processedNodesCount := 0

	if len(currentStage) == 0 {
		return nil, fmt.Errorf("%w: graph has no valid entry points (missing root nodes)", errs.ErrGraphNoEntryPoints)
	}

	for len(currentStage) > 0 {
		// create a dedicated, isolated snapshot slice of this stage
		stageSnapshot := make([]*Node, len(currentStage))
		copy(stageSnapshot, currentStage)

		// append the immutable snapshot instead of the working loop slice
		stages = append(stages, stageSnapshot)
		processedNodesCount += len(stageSnapshot)

		var nextStage []*Node

		for _, currNode := range currentStage {
			for _, dependent := range currNode.Dependents {
				name := dependent.Name
				inDegrees[name]--

				if inDegrees[name] == 0 {
					nextStage = append(nextStage, dependent)
				}
			}
		}

		// sorting for consistent ordering
		sort.Slice(nextStage, func(i, j int) bool {
			return nextStage[i].Name < nextStage[j].Name
		})

		currentStage = nextStage
	}

	if processedNodesCount < len(graph.Nodes) {
		return nil, fmt.Errorf("%w: graph contains an unresolvable cycle or deadlocked dependencies", errs.ErrGraphCyclicDependency)
	}

	return stages, nil
}
