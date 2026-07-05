package workflow

import (
	"fmt"
	"strings"
	"sync"
	"workflow_engine/internal/workflow/errs"
	"workflow_engine/internal/workflow/retry"
	"workflow_engine/internal/workflow/runtime"
)

type Step struct {
	Name      string
	DependsOn []string
	Retry     *retry.RetryPolicy
	Execute   func(ctx *runtime.WorkflowContext) error
	Rollback  func(ctx *runtime.WorkflowContext) error
	mu        sync.Mutex
}

func (s *Step) Validate() error {
	if strings.TrimSpace(s.Name) == "" {
		return fmt.Errorf("%w: step %q", errs.ErrStepNameEmpty, s.Name)
	}

	if s.Execute == nil {
		return fmt.Errorf("%w: step %q", errs.ErrStepNilExecution, s.Name)
	}

	dependencySet := make(map[string]struct{})

	for _, dep := range s.DependsOn {
		if dep == s.Name {
			return fmt.Errorf("%w: step %q", errs.ErrStepSelfDependency, s.Name)
		}

		if _, exists := dependencySet[dep]; exists {
			return fmt.Errorf("%w: step %q", errs.ErrStepDuplicateDependency, s.Name)
		}

		dependencySet[dep] = struct{}{}
	}

	return nil
}
