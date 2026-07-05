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

	add := func(name string, deps ...string) {

		_ = wf.AddStep(&workflow.Step{

			Name: name,

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
		})
	}

	add("upload")

	add("extract-audio", "upload")
	add("extract-thumbnail", "upload")
	add("transcode-video", "upload")

	add(
		"publish",
		"extract-audio",
		"extract-thumbnail",
		"transcode-video",
	)

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
