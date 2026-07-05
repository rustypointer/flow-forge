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

	add := func(step *workflow.Step) {
		if err := wf.AddStep(step); err != nil {
			panic(err)
		}
	}

	add(&workflow.Step{
		Name:  "create-user",
		Retry: retry.NewRetryPolicy(3, time.Second, 5*time.Second, .5),
		Execute: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("User created")
			return nil
		},
		Rollback: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("Delete user")
			return nil
		},
	})

	add(&workflow.Step{
		Name:  "create-wallet",
		Retry: retry.NewRetryPolicy(3, time.Second, 5*time.Second, .5),
		Execute: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("Wallet created")
			return nil
		},
		Rollback: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("Delete wallet")
			return nil
		},
	})

	add(&workflow.Step{
		Name: "create-order",
		DependsOn: []string{
			"create-user",
			"create-wallet",
		},
		Retry: retry.NewRetryPolicy(3, time.Second, 5*time.Second, .5),

		Execute: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("Order created")
			return nil
		},

		Rollback: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("Delete order")
			return nil
		},
	})

	add(&workflow.Step{
		Name: "process-payment",
		DependsOn: []string{
			"create-order",
		},

		Retry: retry.NewRetryPolicy(3, time.Second, 5*time.Second, .5),

		Execute: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("Payment processed")
			return nil
		},

		Rollback: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("Refund payment")
			return nil
		},
	})

	add(&workflow.Step{
		Name: "send-email",

		DependsOn: []string{
			"process-payment",
		},

		Retry: retry.NewRetryPolicy(1, time.Second, time.Second, .5),

		Execute: func(ctx *runtime.WorkflowContext) error {
			fmt.Println("Email sent")
			return nil
		},
	})

	run(wf, ctx)
}

func run(
	wf *workflow.Workflow,
	ctx *runtime.WorkflowContext,
) {

	compiled, _ := compiler.Compile(wf)

	_ = compiled.ExportVisualization(
		"workflow.dot",
		nil,
	)

	result :=
		compiled.Execute(
			wf,
			ctx,
			workflow.DefaultExecutionOptions(),
		)

	fmt.Println(result.State)
}
