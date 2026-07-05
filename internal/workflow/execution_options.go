package workflow

import (
	"workflow_engine/internal/workflow/events"
	"workflow_engine/internal/workflow/execution"
)

type ExecutionOptions struct {
	MaxWorkers    int
	FailurePolicy FailurePolicy
	EventSink     events.Sink
	Store         execution.Store
}

func NewExecutionOptions(maxWorkers int) *ExecutionOptions {
	eo := &ExecutionOptions{
		MaxWorkers: maxWorkers,
	}

	eo.normalize()
	return eo
}

func DefaultExecutionOptions() *ExecutionOptions {
	eo := &ExecutionOptions{
		MaxWorkers:    10,
		FailurePolicy: FailFast,
		EventSink:     events.NewLoggerSink(),
		Store:         execution.NewMemoryStore(),
	}

	eo.normalize()
	return eo
}

func (eo *ExecutionOptions) normalize() {
	if eo.MaxWorkers < 1 {
		eo.MaxWorkers = 10
	}
}
