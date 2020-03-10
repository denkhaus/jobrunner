package jobrunner

import (
	"bytes"
	"log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type TaskState uint32
type JobFunc func() error

const (
	TaskStateIdle TaskState = iota + 1
	TaskStateRunning
	TaskStateFinished
)

type Job struct {
	Name     string
	jobFunc  JobFunc
	state    TaskState
	RunStart time.Time
	RunEnd   time.Time
	Next     time.Time
	Prev     time.Time
	Result   error
	EntryID  cron.EntryID
	running  sync.Mutex
	stateMu  sync.Mutex
}

func New(name string, fn JobFunc) *Job {
	return &Job{
		Name:    name,
		EntryID: -1,
		jobFunc: fn,
	}
}

func (j *Job) setState(state TaskState) {
	j.stateMu.Lock()
	defer j.stateMu.Unlock()
	j.state = state
}

func (j *Job) Run() {
	defer func() {
		if err := recover(); err != nil {
			var buf bytes.Buffer
			logger := log.New(&buf, "JobRunner Log: ", log.Lshortfile)
			logger.Panic(err, "\n", string(debug.Stack()))
		}
	}()

	if !selfConcurrent {
		j.running.Lock()
		defer j.running.Unlock()
	}

	if workPermits != nil {
		workPermits <- struct{}{}
		defer func() { <-workPermits }()
	}

	j.setState(TaskStateRunning)
	j.RunStart = time.Now().UTC()

	defer func() {
		j.setState(TaskStateIdle)
		j.RunEnd = time.Now().UTC()
	}()

	j.Result = j.jobFunc()
}
