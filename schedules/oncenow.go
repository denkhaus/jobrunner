package schedules

import (
	"sync"
	"time"
)

// OnceNowSchedule represents a simple immediatly one time execution.
type OnceNowSchedule struct {
	mu          sync.Mutex
	invocations int
}

// OnceNowSchedule returns a crontab Schedule that activates immediatly and runs once.
func OnceNow() *OnceNowSchedule {
	return &OnceNowSchedule{
		invocations: 1,
	}
}

// Next returns the next time this should be run.
func (p *OnceNowSchedule) Next(t time.Time) time.Time {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.invocations > 0 {
		p.invocations--
		return t.Add(-1 * time.Second)
	}

	return time.Time{}
}
