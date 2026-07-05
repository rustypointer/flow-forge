package errs

import "errors"

var (
	ErrStepNameEmpty           = errors.New("empty step name")
	ErrStepNilExecution        = errors.New("step has nil execution")
	ErrStepSelfDependency      = errors.New("step self-dependency")
	ErrStepDuplicateDependency = errors.New("duplicate step dependency")
)

var (
	ErrWorkflowMissingDependency = errors.New("workflow missing dependency")
)

var (
	ErrGraphEmpty            = errors.New("graph is empty")
	ErrGraphNoEntryPoints    = errors.New("graph has no entry points")
	ErrGraphCyclicDependency = errors.New("graph contains cyclic dependency")
)
