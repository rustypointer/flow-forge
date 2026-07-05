package benchmark

import (
	"fmt"
	"testing"
	"time"

	"workflow_engine/internal/workflow"
	"workflow_engine/internal/workflow/compiler"
	"workflow_engine/internal/workflow/runtime"
)

func TestPrintMetrics(t *testing.T) {

	cases := []struct {
		nodes   int
		workers int
	}{
		{100, 10},
		{500, 16},
		{1000, 32},
	}

	fmt.Println()
	fmt.Println("======== Workflow Engine Metrics ========")

	for _, tc := range cases {

		wf := buildParallelWorkflow(tc.nodes)

		compiled, _ := compiler.Compile(wf)

		opts := workflow.DefaultExecutionOptions()
		opts.MaxWorkers = tc.workers

		ctx := runtime.NewWorkflowContext()

		start := time.Now()

		res := compiled.Execute(
			wf,
			ctx,
			opts,
		)

		elapsed := time.Since(start)

		fmt.Printf(`
Nodes:          %d
Workers:        %d
Completed:      %d
State:          %s
Duration:       %v
Throughput:     %.2f steps/sec

`,
			tc.nodes,
			tc.workers,
			len(res.CompletedSteps),
			res.State,
			elapsed,
			float64(tc.nodes)/elapsed.Seconds(),
		)

		ctx.Cancel()
	}

	fmt.Println("========================================")
}
