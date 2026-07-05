package dag

import "time"

type TaskResult struct {
	Node       *Node
	Err        error
	StartedAt  time.Time
	FinishedAt time.Time
}
