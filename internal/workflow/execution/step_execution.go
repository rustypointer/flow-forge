package execution

import (
	"sync"
	"time"
	"workflow_engine/internal/workflow/state"
)

type StepExecution struct {
	StepName           string
	State              state.StepState
	StartedAt          time.Time
	FinishedAt         time.Time
	Duration           time.Duration
	Attempts           int
	Error              string
	RollbackStartedAt  time.Time
	RollbackFinishedAt time.Time
	RollbackDuration   time.Duration
	RollbackError      string
	mu                 sync.RWMutex
}

func NewStepExecution(stepName string) *StepExecution {
	return &StepExecution{
		StepName: stepName,
		State:    state.StepPending,
	}
}

func (se *StepExecution) SetState(state state.StepState) {
	se.mu.Lock()
	defer se.mu.Unlock()
	se.State = state
}

func (se *StepExecution) GetState() state.StepState {
	se.mu.RLock()
	defer se.mu.RUnlock()
	return se.State
}
