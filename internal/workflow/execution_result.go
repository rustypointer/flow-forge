package workflow

import (
	"sync"
	"workflow_engine/internal/workflow/execution"
	"workflow_engine/internal/workflow/state"
)

type ExecutionResult struct {
	RunId          string
	WorkflowId     string
	Execution      *execution.WorkflowExecution
	State          state.WorkflowState
	CompletedSteps []string
	FailedStep     string
	Error          error
	RollbackErrors []RollbackError
	mu             sync.RWMutex
}

type RollbackError struct {
	StepName string
	Error    error
}

func NewExecutionResult(runId, workflowId string, exec *execution.WorkflowExecution) *ExecutionResult {
	return &ExecutionResult{
		RunId:          runId,
		WorkflowId:     workflowId,
		Execution:      exec,
		State:          state.WorkflowPending,
		CompletedSteps: make([]string, 0),
		RollbackErrors: make([]RollbackError, 0),
	}
}

func (er *ExecutionResult) SetState(state state.WorkflowState) {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.State = state
}

func (er *ExecutionResult) SetSuccess() {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.State = state.WorkflowSucceeded
}

func (er *ExecutionResult) SetFailure(failedStep string, err error) {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.State = state.WorkflowFailed
	er.FailedStep = failedStep
	er.Error = err
}

func (er *ExecutionResult) AddCompletedStep(stepName string) {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.CompletedSteps = append(er.CompletedSteps, stepName)
}

func (er *ExecutionResult) AddRollbackError(stepName string, err error) {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.RollbackErrors = append(er.RollbackErrors, RollbackError{
		StepName: stepName,
		Error:    err,
	})
}
