package schedules

import (
	"sync"
	"time"
)

// AbsoluteSchedule represents a absolute time execution.
type AbsoluteSchedule struct {
	mu sync.Mutex
	dt time.Time
}

// AbsoluteSchedule returns a crontab Schedule that runs  at an absolute time.
func Absolute(dt time.Time) *AbsoluteSchedule {
	return &AbsoluteSchedule{
		dt: dt,
	}
}

func (p *AbsoluteSchedule) Reset(dt time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.dt.IsZero() {
		p.dt = dt
	}
}

// Next returns the next time this should be run.
func (p *AbsoluteSchedule) Next(t time.Time) time.Time {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.dt.IsZero() && p.dt.Before(t) {
		dt := p.dt
		p.dt = time.Time{}
		return dt
	}

	return p.dt
}
