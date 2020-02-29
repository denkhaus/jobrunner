package jobrunner

import (
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type Delayer struct {
	mutex sync.Mutex
	thMap map[cron.Job]time.Time
}

func NewDelayer() *Delayer {
	t := &Delayer{
		thMap: make(map[cron.Job]time.Time),
	}

	go t.processor()
	return t
}

func (p *Delayer) processor() {
	ch := time.Tick(time.Second)
	for {
		select {
		case tm := <-ch:
			p.mutex.Lock()
			for job, extm := range p.thMap {
				if tm.After(extm) {
					go New(job).Run()
					delete(p.thMap, job)
				}
			}
			p.mutex.Unlock()
		}
	}
}

func (p *Delayer) Process(duration time.Duration, job cron.Job) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.thMap[job] = time.Now().Add(duration)
}
