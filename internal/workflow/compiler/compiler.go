package compiler

import (
	"time"
	"workflow_engine/internal/workflow"
	"workflow_engine/internal/workflow/dag"
)

func Compile(wf *workflow.Workflow) (*CompiledWorkflow, error) {
	if err := wf.Validate(); err != nil {
		return nil, err
	}

	nodes := wf.GetNodes()

	graph, err := dag.BuildGraph(nodes)
	if err != nil {
		return nil, err
	}

	if err := dag.Validate(graph); err != nil {
		return nil, err
	}

	stages, err := dag.BuildExecutionStages(graph)
	if err != nil {
		return nil, err
	}

	return &CompiledWorkflow{
		WorkflowId: wf.Id,
		CompiledAt: time.Now(),
		Graph:      graph,
		Stages:     stages,
		Roots:      graph.Roots(),
		Leaves:     graph.Leaves(),
	}, nil
}
