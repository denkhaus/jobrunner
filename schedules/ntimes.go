package schedules

import (
	"sync"
	"time"
)

// NTimesEverySchedule represents a simple one time execution, e.g. "In 5 minutes".
type NTimesEverySchedule struct {
	mu          sync.Mutex
	delay       time.Duration
	invocations int
}

// NTimesEvery returns a crontab Schedule that activates once after a given duration.
// Delays of less than a second are not supported (will round up to 1 second).
// Any fields less than a Second are truncated.
func NTimesEvery(times int, duration time.Duration) *NTimesEverySchedule {
	if duration < time.Second {
		duration = time.Second
	}

	return &NTimesEverySchedule{
		invocations: times + 1,
		delay:       duration - time.Duration(duration.Nanoseconds())%time.Second,
	}
}

// Next returns the next time this should be run.
// This rounds so that the next activation time will be on the second.
func (p *NTimesEverySchedule) Next(t time.Time) time.Time {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.invocations > 0 {
		p.invocations--
		return t.Add(p.delay - time.Duration(t.Nanosecond())*time.Nanosecond)
	}

	return time.Time{}
}
