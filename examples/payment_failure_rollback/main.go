package main

import (
	"errors"
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

	add := func(step *workflow.Step) {
		_ = wf.AddStep(step)
	}

	add(&workflow.Step{
		Name: "create-user",

		Retry: retry.NewRetryPolicy(
			3,
			time.Second,
			5*time.Second,
			0.5,
		),

		Execute: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("create-user")
			return nil
		},

		Rollback: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("delete-user")
			return nil
		},
	})

	add(&workflow.Step{
		Name: "create-wallet",

		DependsOn: []string{
			"create-user",
		},

		Retry: retry.NewRetryPolicy(
			3,
			time.Second,
			5*time.Second,
			0.5,
		),

		Execute: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("create-wallet")
			return nil
		},

		Rollback: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("delete-wallet")
			return nil
		},
	})

	add(&workflow.Step{
		Name: "charge-payment",

		DependsOn: []string{
			"create-wallet",
		},

		Retry: retry.NewRetryPolicy(
			1,
			time.Second,
			5*time.Second,
			0.5,
		),

		Execute: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("payment failed")
			return errors.New("payment declined")
		},

		Rollback: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("refund")
			return nil
		},
	})

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
