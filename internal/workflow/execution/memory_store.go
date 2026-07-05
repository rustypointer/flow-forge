package execution

import (
	"sync"
	"workflow_engine/internal/workflow/runtime"
)

type MemoryStore struct {
	mu   sync.RWMutex
	runs map[string]*WorkflowExecution
}

var _ Store = (*MemoryStore)(nil)

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		runs: make(map[string]*WorkflowExecution),
	}
}

func (s *MemoryStore) Save(wfCtx *runtime.WorkflowContext, exec *WorkflowExecution) error {
	select {
	case <-wfCtx.Done():
		return wfCtx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.runs[exec.RunId] = exec
	return nil
}

func (s *MemoryStore) Update(wfCtx *runtime.WorkflowContext, exec *WorkflowExecution) error {
	select {
	case <-wfCtx.Done():
		return wfCtx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.runs[exec.RunId] = exec
	return nil
}

func (s *MemoryStore) Get(wfCtx *runtime.WorkflowContext, runId string) (*WorkflowExecution, error) {
	select {
	case <-wfCtx.Done():
		return nil, wfCtx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	run := s.runs[runId]

	return run, nil
}

func (s *MemoryStore) List(wfCtx *runtime.WorkflowContext, query Query) ([]*WorkflowExecution, error) {
	select {
	case <-wfCtx.Done():
		return nil, wfCtx.Err()
	default:
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var out []*WorkflowExecution

	for _, run := range s.runs {
		if query.WorkflowId != "" && run.WorkflowId != query.WorkflowId {
			continue
		}

		if query.State != "" && run.State != query.State {
			continue
		}

		out = append(out, run)

		if query.Limit > 0 && len(out) >= query.Limit {
			break
		}
	}

	return out, nil
}

func (s *MemoryStore) Delete(wfCtx *runtime.WorkflowContext, runId string) error {
	select {
	case <-wfCtx.Done():
		return wfCtx.Err()
	default:
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.runs, runId)

	return nil
}
