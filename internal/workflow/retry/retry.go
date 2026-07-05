package retry

import (
	"math/rand"
	"time"
	"workflow_engine/internal/logger"
	"workflow_engine/internal/workflow/runtime"
)

type RetryPolicy struct {
	maxAttempts     int           // max allowed attempts
	baseBackoff     time.Duration // starting delay duration
	maxBackoff      time.Duration // absolute max cap for a single delay
	minPercentFloor float64       // min percentage of baseBackoff duration that is strictly guaranteed
}

func NewRetryPolicy(maxAttempts int, baseBackoff, maxBackOff time.Duration, minPercentFloor float64) *RetryPolicy {
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	if minPercentFloor < 0.0 || minPercentFloor > 1.0 {
		minPercentFloor = 0.50
	}

	if maxBackOff <= 0 {
		maxBackOff = baseBackoff * 2
	}

	return &RetryPolicy{
		maxAttempts:     maxAttempts,
		baseBackoff:     baseBackoff,
		maxBackoff:      maxBackOff,
		minPercentFloor: minPercentFloor,
	}
}

func (r *RetryPolicy) ExecuteWithRetry(ctx *runtime.WorkflowContext, stepName string, runnable func(ctx *runtime.WorkflowContext) error) error {
	var err error

	for attempt := 1; attempt <= r.maxAttempts; attempt++ {
		err = runnable(ctx)
		if err == nil {
			return nil
		}

		logger.Log.Warn("Step attempt failed", "step", stepName, "attempt", attempt, "error", err)

		// exponential backoff with random jitter - Tackling "Thundering herd problem"
		if attempt < r.maxAttempts {
			r.backoffRetry(attempt)
		}
	}
	return err
}

func (r *RetryPolicy) backoffRetry(attempt int) {
	// calculate raw exponential backoff: base * 2^(attempt - 1)
	backoffFactor := float64(int(1) << uint(attempt-1))
	currentBackoff := float64(r.baseBackoff) * backoffFactor

	// clamp currentBackoff
	if currentBackoff > float64(r.maxBackoff) {
		currentBackoff = float64(r.maxBackoff)
	}

	// fallback protection for floor percent
	floorPercent := r.minPercentFloor
	if floorPercent < 0.0 || floorPercent > 1.0 {
		floorPercent = 0.50
	}

	floorDuration := currentBackoff * floorPercent
	jitterRange := currentBackoff * (1.0 - floorPercent)

	randomJitter := rand.Float64() * jitterRange
	finalBackoff := time.Duration(floorDuration + randomJitter)

	time.Sleep(finalBackoff)
}
