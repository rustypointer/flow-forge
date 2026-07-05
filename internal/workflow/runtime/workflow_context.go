package runtime

import (
	"context"
	"sync"
)

type WorkflowContext struct {
	ctx    context.Context
	cancel context.CancelFunc
	Data   *sync.Map
}

func NewWorkflowContext() *WorkflowContext {
	return NewWorkflowContextFrom(context.Background())
}

func NewWorkflowContextFrom(parent context.Context) *WorkflowContext {
	ctx, cancel := context.WithCancel(parent)

	return &WorkflowContext{
		ctx:    ctx,
		cancel: cancel,
		Data:   &sync.Map{},
	}
}

func (w *WorkflowContext) NewRollbackContext() *WorkflowContext {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkflowContext{
		ctx:    ctx,
		cancel: cancel,
		Data:   w.Data,
	}
}

func (w *WorkflowContext) Context() context.Context {
	return w.ctx
}

func (w *WorkflowContext) Cancel() {
	if w.cancel != nil {
		w.cancel()
	}
}

func (w *WorkflowContext) Done() <-chan struct{} {
	return w.ctx.Done()
}

func (w *WorkflowContext) Err() error {
	return w.ctx.Err()
}
