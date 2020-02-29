package schedules

import (
	"time"
)

// OnceNowSchedule represents a simple immediatly one time execution.
type OnceNowSchedule struct {
	Invocations int
}

// OnceNowSchedule returns a crontab Schedule that activates immediatly and runs once.
func OnceNow() *OnceNowSchedule {
	return &OnceNowSchedule{
		Invocations: 2,
	}
}

// Next returns the next time this should be run.
func (schedule *OnceNowSchedule) Next(t time.Time) time.Time {
	if schedule.Invocations > 0 {
		schedule.Invocations--
		return t.Add(-1 * time.Second)
	}

	return time.Time{}
}
