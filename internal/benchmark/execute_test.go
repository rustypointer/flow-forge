package benchmark

import (
	"testing"
	"time"

	"workflow_engine/internal/workflow"
	"workflow_engine/internal/workflow/compiler"
	"workflow_engine/internal/workflow/execution"
	"workflow_engine/internal/workflow/runtime"
)

// pureSchedulingBench executes the workflow with persistence disabled (Store = nil)
// to measure raw channel communication and worker coordination overhead.
func pureSchedulingBench(b *testing.B, nodes int, workers int, delay time.Duration, isDiamond bool) {
	var wf *workflow.Workflow
	if isDiamond {
		wf = buildDiamondWorkflow(nodes, delay)
	} else {
		wf = buildParallelWorkflowWithDelay(nodes, delay)
	}

	compiled, _ := compiler.Compile(wf)
	opts := &workflow.ExecutionOptions{
		MaxWorkers:    workers,
		FailurePolicy: workflow.FailFast,
		Store:         nil, // Disable persistence to isolate scheduler overhead
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx := runtime.NewWorkflowContext()
		compiled.Execute(wf, ctx, opts)
		ctx.Cancel()
	}
}

// persistenceBench executes the workflow with a fresh memory store per run
// to avoid memory leaks while measuring persistence overhead.
func persistenceBench(b *testing.B, nodes int, workers int) {
	wf := buildParallelWorkflow(nodes)
	compiled, _ := compiler.Compile(wf)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx := runtime.NewWorkflowContext()
		opts := &workflow.ExecutionOptions{
			MaxWorkers:    workers,
			FailurePolicy: workflow.FailFast,
			Store:         execution.NewMemoryStore(), // Fresh store per run
		}
		compiled.Execute(wf, ctx, opts)
		ctx.Cancel()
	}
}

// ---- Pure Scheduler Overhead (No Delay, No Store) ----

func BenchmarkExecuteParallel_100Nodes_10Workers(b *testing.B) {
	pureSchedulingBench(b, 100, 10, 0, false)
}

func BenchmarkExecuteParallel_1000Nodes_32Workers(b *testing.B) {
	pureSchedulingBench(b, 1000, 32, 0, false)
}

func BenchmarkExecuteDiamond_100Nodes_10Workers(b *testing.B) {
	pureSchedulingBench(b, 100, 10, 0, true)
}

// ---- Persistence Overhead Benchmarks ----

func BenchmarkExecuteWithStore_100Nodes_10Workers(b *testing.B) {
	persistenceBench(b, 100, 10)
}

// ---- Concurrency & Worker Pool Scaling (with 100µs Step Delay) ----

func BenchmarkWorkerScaling_100Nodes_1Worker(b *testing.B) {
	pureSchedulingBench(b, 100, 1, 100*time.Microsecond, false)
}

func BenchmarkWorkerScaling_100Nodes_2Workers(b *testing.B) {
	pureSchedulingBench(b, 100, 2, 100*time.Microsecond, false)
}

func BenchmarkWorkerScaling_100Nodes_4Workers(b *testing.B) {
	pureSchedulingBench(b, 100, 4, 100*time.Microsecond, false)
}

func BenchmarkWorkerScaling_100Nodes_8Workers(b *testing.B) {
	pureSchedulingBench(b, 100, 8, 100*time.Microsecond, false)
}

func BenchmarkWorkerScaling_100Nodes_16Workers(b *testing.B) {
	pureSchedulingBench(b, 100, 16, 100*time.Microsecond, false)
}

func BenchmarkWorkerScaling_100Nodes_32Workers(b *testing.B) {
	pureSchedulingBench(b, 100, 32, 100*time.Microsecond, false)
}
