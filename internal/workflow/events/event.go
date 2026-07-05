package events

import "time"

type Event struct {
	Type       string
	WorkflowId string
	Step       string
	Time       time.Time
	Error      string
}
