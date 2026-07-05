package workflow

import (
	"fmt"
	"github.com/google/uuid"
	"sync"
	"workflow_engine/internal/workflow/dag"
	"workflow_engine/internal/workflow/errs"
	"workflow_engine/internal/workflow/runtime"
)

type Workflow struct {
	Id    string
	Steps map[string]*Step
	mu    sync.RWMutex
}

func NewWorkflow() *Workflow {
	return &Workflow{
		Id:    uuid.NewString(),
		Steps: make(map[string]*Step),
	}
}

func (w *Workflow) Validate() error {
	for stepName, step := range w.Steps {
		for _, dep := range step.DependsOn {
			if _, exists := w.Steps[dep]; !exists {
				return fmt.Errorf("%w: step %q depends on unknown step %q", errs.ErrWorkflowMissingDependency, stepName, dep)
			}
		}
	}

	return nil
}

func (w *Workflow) AddStep(step *Step) error {
	if err := step.Validate(); err != nil {
		return err
	}

	w.Steps[step.Name] = step

	return nil
}

func (w *Workflow) GetNodes() []*dag.Node {
	nodes := make([]*dag.Node, 0)

	for _, s := range w.Steps {
		step := s

		nodes = append(nodes, dag.NewNode(
			step.Name,
			step.DependsOn,
			func(wfCtx *runtime.WorkflowContext) error {
				if step.Retry == nil {
					return step.Execute(wfCtx)
				}
				return step.Retry.ExecuteWithRetry(wfCtx, step.Name, step.Execute)
			},
		))
	}

	return nodes
}
