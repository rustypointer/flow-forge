package benchmark

import (
	"fmt"
	"time"

	"workflow_engine/internal/workflow"
	"workflow_engine/internal/workflow/runtime"
)

func buildLinearWorkflow(nodes int) *workflow.Workflow {
	wf := workflow.NewWorkflow()

	for i := 0; i < nodes; i++ {
		name := fmt.Sprintf("step-%d", i)

		var deps []string
		if i > 0 {
			deps = []string{
				fmt.Sprintf("step-%d", i-1),
			}
		}

		_ = wf.AddStep(&workflow.Step{
			Name:      name,
			DependsOn: deps,
			Execute: func(ctx *runtime.WorkflowContext) error {
				return nil
			},
			Rollback: func(ctx *runtime.WorkflowContext) error {
				return nil
			},
		})
	}

	return wf
}

func buildParallelWorkflow(nodes int) *workflow.Workflow {
	return buildParallelWorkflowWithDelay(nodes, 0)
}

func buildParallelWorkflowWithDelay(nodes int, delay time.Duration) *workflow.Workflow {
	wf := workflow.NewWorkflow()

	for i := 0; i < nodes; i++ {
		_ = wf.AddStep(&workflow.Step{
			Name: fmt.Sprintf("parallel-%d", i),
			Execute: func(ctx *runtime.WorkflowContext) error {
				if delay > 0 {
					time.Sleep(delay)
				}
				return nil
			},
			Rollback: func(ctx *runtime.WorkflowContext) error {
				return nil
			},
		})
	}

	return wf
}

func buildDiamondWorkflow(nodes int, delay time.Duration) *workflow.Workflow {
	wf := workflow.NewWorkflow()

	// Root step
	_ = wf.AddStep(&workflow.Step{
		Name: "root",
		Execute: func(ctx *runtime.WorkflowContext) error {
			if delay > 0 {
				time.Sleep(delay)
			}
			return nil
		},
		Rollback: func(ctx *runtime.WorkflowContext) error {
			return nil
		},
	})

	var parallelNames []string
	for i := 0; i < nodes-2; i++ {
		name := fmt.Sprintf("parallel-%d", i)
		parallelNames = append(parallelNames, name)
		_ = wf.AddStep(&workflow.Step{
			Name:      name,
			DependsOn: []string{"root"},
			Execute: func(ctx *runtime.WorkflowContext) error {
				if delay > 0 {
					time.Sleep(delay)
				}
				return nil
			},
			Rollback: func(ctx *runtime.WorkflowContext) error {
				return nil
			},
		})
	}

	// Leaf step
	_ = wf.AddStep(&workflow.Step{
		Name:      "leaf",
		DependsOn: parallelNames,
		Execute: func(ctx *runtime.WorkflowContext) error {
			if delay > 0 {
				time.Sleep(delay)
			}
			return nil
		},
		Rollback: func(ctx *runtime.WorkflowContext) error {
			return nil
		},
	})

	return wf
}
