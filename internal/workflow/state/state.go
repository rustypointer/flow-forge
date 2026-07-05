package state

type StepState string

const (
	StepPending        StepState = "PENDING"
	StepRunning        StepState = "RUNNING"
	StepSucceeded      StepState = "SUCCEEDED"
	StepFailed         StepState = "FAILED"
	StepRollingBack    StepState = "ROLLING_BACK"
	StepRolledBack     StepState = "ROLLED_BACK"
	StepRollbackFailed StepState = "ROLLBACK_FAILED"
)

type WorkflowState string

const (
	WorkflowPending     WorkflowState = "PENDING"
	WorkflowRunning     WorkflowState = "RUNNING"
	WorkflowSucceeded   WorkflowState = "SUCCEEDED"
	WorkflowFailed      WorkflowState = "FAILED"
	WorkflowRollingBack WorkflowState = "ROLLING_BACK"
	WorkflowRolledBack  WorkflowState = "ROLLED_BACK"
)
