package execution

import (
	"github.com/google/uuid"
	"sync"
	"time"
	"workflow_engine/internal/workflow/state"
)

type WorkflowExecution struct {
	RunId      string
	WorkflowId string
	State      state.WorkflowState
	StartedAt  time.Time
	FinishedAt time.Time
	Duration   time.Duration
	Steps      map[string]*StepExecution
	mu         sync.RWMutex
}

func NewWorkflowExecution(workflowId string) *WorkflowExecution {
	return &WorkflowExecution{
		RunId:      uuid.NewString(),
		WorkflowId: workflowId,
		State:      state.WorkflowPending,
		StartedAt:  time.Now(),
		Steps:      make(map[string]*StepExecution),
	}
}

func (we *WorkflowExecution) SetState(state state.WorkflowState) {
	we.mu.Lock()
	defer we.mu.Unlock()
	we.State = state
}

func (we *WorkflowExecution) GetState() state.WorkflowState {
	we.mu.RLock()
	defer we.mu.RUnlock()

	return we.State
}

func (we *WorkflowExecution) Finish(state state.WorkflowState) {
	we.mu.Lock()
	defer we.mu.Unlock()
	we.State = state
	we.FinishedAt = time.Now()
	we.Duration = we.FinishedAt.Sub(we.StartedAt)
}
