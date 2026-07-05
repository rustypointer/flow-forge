package execution

import "workflow_engine/internal/workflow/runtime"

type Store interface {
	Save(wfCtx *runtime.WorkflowContext, exec *WorkflowExecution) error
	Update(wfCtx *runtime.WorkflowContext, exec *WorkflowExecution) error
	Get(wfCtx *runtime.WorkflowContext, runId string) (*WorkflowExecution, error)
	List(wfCtx *runtime.WorkflowContext, query Query) ([]*WorkflowExecution, error)
	Delete(wfCtx *runtime.WorkflowContext, runId string) error
}
