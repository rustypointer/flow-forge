package execution

import (
	"time"
	"workflow_engine/internal/workflow/state"
)

type Query struct {
	WorkflowId   string
	State        state.WorkflowState
	StartedAfter time.Time
	Limit        int
}
