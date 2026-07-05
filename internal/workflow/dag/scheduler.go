package dag

import (
	"sync"
	"time"
	"workflow_engine/internal/workflow/runtime"
)

type Scheduler struct {
	maxWorkers   int
	graph        *Graph
	failFastMode bool
}

func NewScheduler(graph *Graph, maxWorkers int, failFastMode bool) *Scheduler {
	if maxWorkers < 1 {
		maxWorkers = 10
	}

	return &Scheduler{
		maxWorkers:   maxWorkers,
		graph:        graph,
		failFastMode: failFastMode,
	}
}

func (s *Scheduler) Execute(wfCtx *runtime.WorkflowContext, hooks *ExecutionHooks) error {
	totalTasks := len(s.graph.Nodes)

	inDegrees := s.graph.InDegrees()

	readyQueue := make(chan *Node, totalTasks)
	resultChan := make(chan *TaskResult, totalTasks)

	var wg sync.WaitGroup

	s.startWorkers(wfCtx, readyQueue, resultChan, &wg, hooks)

	for _, node := range s.graph.Roots() {
		readyQueue <- node
	}

	completedTasks := 0

	failed := make(map[string]struct{})

	for completedTasks < totalTasks {
		select {
		case <-wfCtx.Done():
			close(readyQueue)
			wg.Wait()
			return wfCtx.Err()
		case result := <-resultChan:
			if hooks != nil && hooks.OnFinish != nil {
				hooks.OnFinish(result)
			}

			if result.Err != nil {

				failed[result.Node.Name] = struct{}{}

				if s.failFastMode {
					wfCtx.Cancel()
					close(readyQueue)
					wg.Wait()
					return result.Err
				}

				completedTasks++
				continue
			}

			completedTasks++

			for _, dependent := range result.Node.Dependents {
				name := dependent.Name
				inDegrees[name]--

				if inDegrees[name] == 0 {
					shouldSkip := false

					for _, dep := range dependent.Dependencies {
						if _, exists := failed[dep.Name]; exists {
							shouldSkip = true
							break
						}
					}

					if shouldSkip {
						completedTasks++
						continue
					}

					select {
					case readyQueue <- dependent:
					case <-wfCtx.Done():
						return wfCtx.Err()
					}
				}
			}
		}
	}

	close(readyQueue)
	wg.Wait()
	return nil
}

func (s *Scheduler) startWorkers(wfCtx *runtime.WorkflowContext,
	readyQueue <-chan *Node, resultChan chan<- *TaskResult, wg *sync.WaitGroup, hooks *ExecutionHooks) {

	for i := 0; i < s.maxWorkers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for {
				select {
				case <-wfCtx.Done():
					return
				case node, ok := <-readyQueue:
					if !ok {
						return
					}

					if hooks != nil && hooks.OnStart != nil {
						hooks.OnStart(node)
					}

					startedAt := time.Now()

					err := node.Run(wfCtx)

					result := &TaskResult{
						Node:       node,
						Err:        err,
						StartedAt:  startedAt,
						FinishedAt: time.Now(),
					}

					select {
					case resultChan <- result:
					case <-wfCtx.Done():
						return
					}
				}
			}
		}()
	}
}
