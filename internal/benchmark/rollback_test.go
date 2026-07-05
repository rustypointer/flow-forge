package benchmark

import (
	"errors"
	"fmt"
	"testing"

	"workflow_engine/internal/workflow"
	"workflow_engine/internal/workflow/compiler"
	"workflow_engine/internal/workflow/runtime"
)

func BenchmarkRollback100(b *testing.B) {
	wf := workflow.NewWorkflow()

	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("step-%d", i)
		var deps []string
		if i > 0 {
			deps = []string{
				fmt.Sprintf("step-%d", i-1),
			}
		}

		fail := (i == 80)

		_ = wf.AddStep(&workflow.Step{
			Name:      name,
			DependsOn: deps,
			Execute: func(ctx *runtime.WorkflowContext) error {
				if fail {
					return errors.New("failure")
				}
				return nil
			},
			Rollback: func(ctx *runtime.WorkflowContext) error {
				return nil
			},
		})
	}

	compiled, _ := compiler.Compile(wf)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx := runtime.NewWorkflowContext()
		opts := &workflow.ExecutionOptions{
			MaxWorkers:    10,
			FailurePolicy: workflow.FailFast,
			Store:         nil, // Disable store to isolate rollback scheduling cost
		}
		compiled.Execute(wf, ctx, opts)
		ctx.Cancel()
	}
}
