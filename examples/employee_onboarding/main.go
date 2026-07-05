package main

import (
	"fmt"
	"time"

	"workflow_engine/internal/logger"
	"workflow_engine/internal/workflow"
	"workflow_engine/internal/workflow/compiler"
	"workflow_engine/internal/workflow/retry"
	"workflow_engine/internal/workflow/runtime"
)

func main() {

	logger.Init()

	ctx := runtime.NewWorkflowContext()
	defer ctx.Cancel()

	wf := workflow.NewWorkflow()

	add := func(s *workflow.Step) {
		if err := wf.AddStep(s); err != nil {
			panic(err)
		}
	}

	step := func(name string, deps ...string) *workflow.Step {
		return &workflow.Step{
			Name:      name,
			DependsOn: deps,
			Retry: retry.NewRetryPolicy(
				3,
				time.Second,
				5*time.Second,
				0.5,
			),

			Execute: func(ctx *runtime.WorkflowContext) error {
				fmt.Println("Executing:", name)
				return nil
			},

			Rollback: func(ctx *runtime.WorkflowContext) error {
				fmt.Println("Rollback:", name)
				return nil
			},
		}
	}

	add(step("create-employee"))
	add(step("create-email", "create-employee"))
	add(step("grant-github", "create-email"))
	add(step("grant-vpn", "create-email"))
	add(step("notify-manager", "grant-github", "grant-vpn"))

	run(wf, ctx)
}

func run(wf *workflow.Workflow, ctx *runtime.WorkflowContext) {
	compiled, _ := compiler.Compile(wf)

	_ = compiled.ExportVisualization("workflow.dot", nil)

	res := compiled.Execute(
		wf,
		ctx,
		workflow.DefaultExecutionOptions(),
	)

	fmt.Println(res.State)
}
