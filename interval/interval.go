package interval

import (
	"time"
)

// IntervalRunner is used to run a function again only after a certain time
type IntervalRunner struct {
	durationSuccess time.Duration
	durationFailure time.Duration
	lastRun         time.Time
	success         bool
}

func NewIntervalRunner(durationSuccess, durationFailure time.Duration) *IntervalRunner {
	return &IntervalRunner{durationSuccess: durationSuccess, durationFailure: durationFailure}
}

// Run runs the passed in fn function, if the last run is a certain duration ago.
func (ir *IntervalRunner) Run(fn func() bool) {
	var duration time.Duration
	if ir.success {
		duration = ir.durationSuccess
	} else {
		duration = ir.durationFailure
	}

	now := time.Now()
	if ir.lastRun.Add(duration).Before(now) {
		ir.success = fn()
		ir.lastRun = now
	}
}
