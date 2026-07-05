package compiler

import (
	"fmt"
	"sync"
	"time"
	"workflow_engine/internal/logger"
	"workflow_engine/internal/workflow"
	"workflow_engine/internal/workflow/dag"
	"workflow_engine/internal/workflow/events"
	"workflow_engine/internal/workflow/execution"
	"workflow_engine/internal/workflow/runtime"
	"workflow_engine/internal/workflow/state"
)

type CompiledWorkflow struct {
	WorkflowId string
	CompiledAt time.Time
	Graph      *dag.Graph
	Stages     [][]*dag.Node
	Roots      []*dag.Node
	Leaves     []*dag.Node
}

func (c *CompiledWorkflow) Execute(w *workflow.Workflow, wfCtx *runtime.WorkflowContext, opts *workflow.ExecutionOptions) *workflow.ExecutionResult {
	// initialize workflow execution
	workflowExec := execution.NewWorkflowExecution(w.Id)

	if opts.Store != nil {
		_ = opts.Store.Save(wfCtx, workflowExec)
	}

	// initialize step executions
	for stepName := range w.Steps {
		workflowExec.Steps[stepName] = execution.NewStepExecution(stepName)
	}

	result := workflow.NewExecutionResult(workflowExec.RunId, c.WorkflowId, workflowExec)

	workflowExec.SetState(state.WorkflowRunning)
	result.SetState(state.WorkflowRunning)

	publish(opts.EventSink, events.Event{Type: "workflow_started", WorkflowId: c.WorkflowId})

	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error("Recovered from panic", "workflowId", w.Id, "panic", r)

			workflowExec.Finish(state.WorkflowFailed)
			if opts.Store != nil {
				_ = opts.Store.Save(wfCtx, workflowExec)
			}

			publish(opts.EventSink, events.Event{Type: "workflow_failed", WorkflowId: c.WorkflowId})

			result.SetFailure("", fmt.Errorf("panic: %v", r))

			// trigger rollback on whatever completed before the panic hit
			c.rollbackWithContext(w, wfCtx, result, workflowExec)
		}
	}()

	// Hand over execution to the high-concurrency Scheduler
	scheduler := dag.NewScheduler(c.Graph, opts.MaxWorkers, opts.FailurePolicy == workflow.FailFast)
	// get scheduler execution hooks
	hooks := c.getExecutionHooks(w, opts, result, workflowExec)

	logger.Log.Info("Starting concurrent workflow scheduler", "workflowId", w.Id)
	if err := scheduler.Execute(wfCtx, hooks); err != nil {
		workflowExec.Finish(state.WorkflowFailed)

		publish(opts.EventSink, events.Event{Type: "workflow_failed", WorkflowId: c.WorkflowId})

		result.SetFailure("", err)

		// Run compensation logic using graph topological relationships
		c.rollbackWithContext(w, wfCtx, result, workflowExec)
		return result
	}

	workflowExec.Finish(state.WorkflowSucceeded)
	if opts.Store != nil {
		_ = opts.Store.Save(wfCtx, workflowExec)
	}

	publish(opts.EventSink, events.Event{Type: "workflow_finished", WorkflowId: c.WorkflowId})

	result.SetSuccess()
	logger.Log.Info("Workflow completed successfully", "workflowId", w.Id)
	return result
}

func (c *CompiledWorkflow) getExecutionHooks(w *workflow.Workflow, opts *workflow.ExecutionOptions,
	executionResult *workflow.ExecutionResult, workflowExec *execution.WorkflowExecution) *dag.ExecutionHooks {

	return &dag.ExecutionHooks{
		OnStart: func(node *dag.Node) {
			step := w.Steps[node.Name]

			stepExec := workflowExec.Steps[node.Name]
			stepExec.SetState(state.StepRunning)

			publish(opts.EventSink, events.Event{Type: "step_started", WorkflowId: c.WorkflowId, Step: step.Name})

			logger.Log.Info("Executing step", "workflowId", w.Id, "step", step.Name)
		},
		OnFinish: func(taskResult *dag.TaskResult) {
			step := w.Steps[taskResult.Node.Name]

			stepExec := workflowExec.Steps[taskResult.Node.Name]
			stepExec.StartedAt = taskResult.StartedAt
			stepExec.FinishedAt = taskResult.FinishedAt
			stepExec.Duration = taskResult.FinishedAt.Sub(taskResult.StartedAt)

			if taskResult.Err != nil {
				stepExec.SetState(state.StepFailed)
				stepExec.Error = taskResult.Err.Error()
				publish(opts.EventSink, events.Event{Type: "step_failed", WorkflowId: c.WorkflowId, Step: step.Name})
				logger.Log.Error("Step failed", "workflowId", w.Id, "step", step.Name, "err", taskResult.Err)
				return
			}

			stepExec.SetState(state.StepSucceeded)
			publish(opts.EventSink, events.Event{Type: "step_succeeded", WorkflowId: c.WorkflowId, Step: step.Name})
			logger.Log.Info("Step finished", "workflowId", w.Id, "step", step.Name)

			// safely track completed items across multiple concurrent worker routines
			executionResult.AddCompletedStep(step.Name)
		},
	}
}

func (c *CompiledWorkflow) rollbackWithContext(w *workflow.Workflow, wfCtx *runtime.WorkflowContext,
	result *workflow.ExecutionResult, workflowExec *execution.WorkflowExecution) {
	rollbackCtx := wfCtx.NewRollbackContext()
	defer rollbackCtx.Cancel()

	c.concurrentRollback(w, rollbackCtx, result, workflowExec)
}

func (c *CompiledWorkflow) concurrentRollback(w *workflow.Workflow, wfCtx *runtime.WorkflowContext,
	result *workflow.ExecutionResult, workflowExec *execution.WorkflowExecution) {
	logger.Log.Warn("Starting concurrent topological rollback", "workflowId", c.WorkflowId)

	// Fetch plan layers to see the dependency architecture
	stages := c.Stages

	workflowExec.SetState(state.WorkflowRollingBack)

	// Loop through stages BACKWARDS (from the leaves back to the roots)
	for i := len(stages) - 1; i >= 0; i-- {
		stage := stages[i]
		var wg sync.WaitGroup

		for _, node := range stage {
			// Only roll back steps that actually completed successfully
			stepExec := workflowExec.Steps[node.Name]
			if stepExec.GetState() != state.StepSucceeded || w.Steps[node.Name].Rollback == nil {
				continue
			}

			wg.Add(1)
			go func(n *dag.Node, stepExec *execution.StepExecution) {
				defer wg.Done()

				stepExec.RollbackStartedAt = time.Now()

				defer func() {
					if r := recover(); r != nil {

						err := fmt.Errorf("panic recovered during rollback: %v", r)

						stepExec.SetState(state.StepRollbackFailed)
						stepExec.RollbackFinishedAt = time.Now()
						stepExec.RollbackDuration = stepExec.FinishedAt.Sub(stepExec.RollbackStartedAt)
						stepExec.RollbackError = err.Error()

						logger.Log.Error("Rollback panic", "step", n.Name, "panic", r)
						result.AddRollbackError(n.Name, err)
					}
				}()

				stepExec.SetState(state.StepRollingBack)

				logger.Log.Warn("Rolling back step", "step", n.Name)

				if err := w.Steps[n.Name].Rollback(wfCtx); err != nil {
					stepExec.RollbackError = err.Error()
					stepExec.RollbackFinishedAt = time.Now()
					stepExec.RollbackDuration = stepExec.RollbackFinishedAt.Sub(stepExec.RollbackStartedAt)
					stepExec.SetState(state.StepRollbackFailed)

					logger.Log.Error("Rollback failed", "step", n.Name, "error", err)

					// Thread-safe append to result errors
					result.AddRollbackError(n.Name, err)

					return
				}

				stepExec.SetState(state.StepRolledBack)
				stepExec.RollbackFinishedAt = time.Now()
				stepExec.RollbackDuration = stepExec.RollbackFinishedAt.Sub(stepExec.RollbackStartedAt)

				logger.Log.Warn("Step rolled-back successfully", "step", n.Name)
			}(node, stepExec)
		}

		// Wait for this entire stage tier to finish rolling back before moving up to its parents
		wg.Wait()
	}

	workflowExec.SetState(state.WorkflowRolledBack)
}

func publish(sink events.Sink, event events.Event) {
	if sink == nil {
		return
	}

	sink.Publish(event)
}
