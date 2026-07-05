package main

import (
	"fmt"
	"time"
	"workflow_engine/internal/workflow/compiler"

	"workflow_engine/internal/logger"
	"workflow_engine/internal/workflow"
	"workflow_engine/internal/workflow/retry"
	"workflow_engine/internal/workflow/runtime"
)

func main() {

	logger.Init()

	ctx := runtime.NewWorkflowContext()
	defer ctx.Cancel()

	wf := workflow.NewWorkflow()

	add := func(s *workflow.Step) {
		_ = wf.AddStep(s)
	}

	create := func(name string, deps ...string) *workflow.Step {

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
				fmt.Println(name)
				return nil
			},
		}
	}

	add(create("fetch-ais"))
	add(create("fetch-26as"))

	add(create(
		"merge-tax-data",
		"fetch-ais",
		"fetch-26as",
	))

	add(create(
		"calculate-tax",
		"merge-tax-data",
	))

	add(create(
		"generate-summary",
		"calculate-tax",
	))

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
