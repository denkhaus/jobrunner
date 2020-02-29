package schedules

import (
	"time"
)

// OnceIn represents a simple one time execution, e.g. "In 5 minutes".
type NTimesEverySchedule struct {
	Delay       time.Duration
	Invocations int
}

// OnceIn returns a crontab Schedule that activates once after a given duration.
// Delays of less than a second are not supported (will round up to 1 second).
// Any fields less than a Second are truncated.
func NTimesEvery(times int, duration time.Duration) *NTimesEverySchedule {
	if duration < time.Second {
		duration = time.Second
	}

	return &NTimesEverySchedule{
		Invocations: times + 1,
		Delay:       duration - time.Duration(duration.Nanoseconds())%time.Second,
	}
}

// Next returns the next time this should be run.
// This rounds so that the next activation time will be on the second.
func (schedule *NTimesEverySchedule) Next(t time.Time) time.Time {
	if schedule.Invocations > 0 {
		schedule.Invocations--
		return t.Add(schedule.Delay - time.Duration(t.Nanosecond())*time.Nanosecond)
	}

	return time.Time{}
}
