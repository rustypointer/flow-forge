package workflow

type FailurePolicy uint8

const (
	FailFast FailurePolicy = iota
	ContinueOnFailure
)