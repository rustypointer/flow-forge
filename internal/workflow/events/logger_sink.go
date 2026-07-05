package events

import "workflow_engine/internal/logger"

type LoggerSink struct{}

var _ Sink = (*LoggerSink)(nil)

func NewLoggerSink() *LoggerSink {
	return &LoggerSink{}
}

func (s *LoggerSink) Publish(event Event) {
	logger.Log.Info(
		"workflow-event",
		"type", event.Type,
		"workflow_id", event.WorkflowId,
		"step", event.Step,
		"time", event.Time,
		"error", event.Error,
	)
}
