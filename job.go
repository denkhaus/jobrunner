package jobrunner

//go:generate stringer -type=JobState

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"log"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"golang.org/x/sync/semaphore"
)

type JobRunner interface {
	Run() error
}

//JobState is the state a Job can run into
type JobState int

const (
	JobStateNew JobState = iota
	JobStateInitializing
	JobStateIdle
	JobStateRunning
	JobStateFinished
	JobStateExecutionDeferred
)

//JobType defines th jobs running behaviour
type JobType int

const (
	JobTypeOnce JobType = iota + 1
	JobTypeRecurring
)

type Job struct {
	name              string
	runner            JobRunner
	omitRunOnDeferred bool
	currentState      JobState
	lastState         JobState
	typ               JobType
	runStart          time.Time
	runEnd            time.Time
	next              time.Time
	prev              time.Time
	result            error
	entryID           cron.EntryID
	lastHash          []byte
	running           *semaphore.Weighted
	stateMu           sync.Mutex
}

func (j *Job) Type() JobType {
	return j.typ
}

func (j *Job) State() JobState {
	j.stateMu.Lock()
	defer j.stateMu.Unlock()
	return j.currentState
}

func (j *Job) Result() error {
	return j.result
}

func (j *Job) Name() string {
	return j.name
}

func (j *Job) Runner() JobRunner {
	return j.runner
}

func (j *Job) Prev() time.Time {
	return j.prev
}

func (j *Job) Next() time.Time {
	return j.next
}

func (j *Job) RunStart() time.Time {
	return j.runStart
}

func (j *Job) RunEnd() time.Time {
	return j.runEnd
}

// New creates a new Job
func New(name string, omitRunOnDeferred bool, runner JobRunner) *Job {
	return &Job{
		name:              name,
		running:           semaphore.NewWeighted(1),
		omitRunOnDeferred: omitRunOnDeferred,
		entryID:           InvalidEntryID,
		runner:            runner,
	}
}

// setState sets the Jobs state
func (j *Job) setState(state JobState, trigger bool) {
	j.stateMu.Lock()
	j.lastState = j.currentState
	j.currentState = state
	j.stateMu.Unlock()

	if trigger {
		triggerOnJobChanged(j)
	}
}

// Run starts the job
func (j *Job) Run() {
	defer func() {
		j.runEnd = Now()
		if j.typ == JobTypeRecurring {
			j.setState(JobStateIdle, true)
		} else {
			j.setState(JobStateFinished, true)
		}

		cleanCron()

		if err := recover(); err != nil {
			var buf bytes.Buffer
			logger := log.New(&buf, "JobRunner Log: ", log.Lshortfile)
			logger.Panic(err, "\n", string(debug.Stack()))
		}
	}()

	if !options.SelfConcurrent {
		if !j.running.TryAcquire(1) {
			j.setState(JobStateExecutionDeferred, true)
			if j.omitRunOnDeferred {
				return
			}
		}

		defer j.running.Release(1)
	}

	if options.WorkPermits != nil {
		if len(options.WorkPermits) == cap(options.WorkPermits) {
			j.setState(JobStateExecutionDeferred, true)
			if j.omitRunOnDeferred {
				return
			}
		}

		options.WorkPermits <- struct{}{}
		defer func() { <-options.WorkPermits }()
	}

	j.runStart = Now()
	j.setState(JobStateRunning, true)
	j.result = j.runner.Run()
	j.prev = j.runStart
}

// String  is the Jobs string representation
func (j *Job) String() string {
	return fmt.Sprintf("%s-[%s->%s]", j.Name(), j.lastState, j.State())
}

// hash computes the Jobs hash value. Used to determine changed state.
func (j *Job) hash() []byte {
	h := make([]byte, 16)
	h = xor16(h, hashMd5([]byte(j.name)))
	h = xor16(h, hashMd5([]byte(j.runStart.Format(time.RFC3339))))
	h = xor16(h, hashMd5([]byte(j.runEnd.Format(time.RFC3339))))
	h = xor16(h, hashMd5([]byte(j.prev.Format(time.RFC3339))))
	h = xor16(h, hashMd5([]byte(j.next.Format(time.RFC3339))))
	h = xor16(h, hashMd5([]byte(strconv.Itoa(int(j.lastState)))))
	h = xor16(h, hashMd5([]byte(strconv.Itoa(int(j.currentState)))))
	h = xor16(h, hashMd5([]byte(strconv.Itoa(int(j.entryID)))))
	if j.result != nil {
		h = xor16(h, hashMd5([]byte(j.result.Error())))
	}

	return h
}

// changed compares the last Job hash with the current
// to find out if any properties have changed
func (j *Job) changed() bool {
	j.stateMu.Lock()
	defer j.stateMu.Unlock()

	h := j.hash()
	c := bytes.Compare(h, j.lastHash)
	j.lastHash = h
	return c != 0
}

// xor16 is a hash helper func
func xor16(v1, v2 []byte) []byte {
	r := make([]byte, 16)
	for i := 0; i < len(v1); i++ {
		r[i] = v1[i] ^ v2[i]
	}
	return r
}

// hashMd5 is a hash helper func
func hashMd5(t []byte) []byte {
	h := md5.New()
	h.Write(t)
	return h.Sum(nil)
}
