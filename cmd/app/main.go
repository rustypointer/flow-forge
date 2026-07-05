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

	wfCtx := runtime.NewWorkflowContext()
	defer wfCtx.Cancel()

	wf := workflow.NewWorkflow()

	if err := wf.AddStep(&workflow.Step{
		Name:  "create-user",
		Retry: retry.NewRetryPolicy(3, 2*time.Second, 5*time.Second, 0.5),
		Execute: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("Creating user...")
			ctx.Data.Store("user_id", 123)
			return nil
		},
		Rollback: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("Deleting user...")
			return nil
		},
	}); err != nil {
		panic(err)
	}

	if err := wf.AddStep(&workflow.Step{
		Name:  "create-wallet",
		Retry: retry.NewRetryPolicy(3, 2*time.Second, 5*time.Second, 0.5),
		Execute: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("Creating wallet...")
			return nil
		},
		Rollback: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("Deleting wallet...")
			return nil
		},
	}); err != nil {
		panic(err)
	}

	if err := wf.AddStep(&workflow.Step{
		Name:  "create-order",
		Retry: retry.NewRetryPolicy(3, 2*time.Second, 5*time.Second, 0.5),
		DependsOn: []string{
			"create-user",
			"create-wallet",
		},
		Execute: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("Creating order...")
			return nil
		},
		Rollback: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("Deleting order...")
			return nil
		},
	}); err != nil {
		panic(err)
	}

	compiled, err := compiler.Compile(wf)
	if err != nil {
		panic(err)
	}

	err = compiled.ExportVisualization(
		"workflow.dot",
		nil,
	)
	if err == nil {
		fmt.Println("Visualization completed successfully")
	}

	opts := workflow.DefaultExecutionOptions()

	result := compiled.Execute(wf, wfCtx, opts)

	saved, _ := opts.Store.Get(wfCtx, result.RunId)

	fmt.Printf("Result:\n%v\n", result)
	fmt.Printf("Saved:\n%v\n", saved)
}
