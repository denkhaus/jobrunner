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
)

type JobRunner interface {
	Run() error
}

//TaskState is the state a Job can run into
type JobState int

const (
	Initializing JobState = iota + 1
	Idle
	Running
	Finished
)

type Job struct {
	name     string
	runner   JobRunner
	state    JobState
	runStart time.Time
	runEnd   time.Time
	next     time.Time
	prev     time.Time
	result   error
	entryID  cron.EntryID
	lastHash []byte
	running  sync.Mutex
	stateMu  sync.Mutex
}

func (j *Job) State() JobState {
	j.stateMu.Lock()
	defer j.stateMu.Unlock()
	return j.state
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
func New(name string, runner JobRunner) *Job {
	return &Job{
		name:    name,
		entryID: -1,
		runner:  runner,
	}
}

// setState sets the Jobs state
func (j *Job) setState(state JobState, trigger bool) {
	j.stateMu.Lock()
	j.state = state
	j.stateMu.Unlock()

	if trigger {
		triggerOnJobStateChanged(j)
	}
}

// Run starts the job
func (j *Job) Run() {
	defer func() {
		j.runEnd = Now()
		j.setState(Idle, true)
		cleanCron()

		if err := recover(); err != nil {
			var buf bytes.Buffer
			logger := log.New(&buf, "JobRunner Log: ", log.Lshortfile)
			logger.Panic(err, "\n", string(debug.Stack()))
		}
	}()

	if !options.SelfConcurrent {
		j.running.Lock()
		defer j.running.Unlock()
	}

	if options.WorkPermits != nil {
		options.WorkPermits <- struct{}{}
		defer func() { <-options.WorkPermits }()
	}

	j.setState(Running, true)
	j.runStart = Now()
	j.result = j.runner.Run()
}

// String  is the Jobs string representation
func (j *Job) String() string {
	return fmt.Sprintf("%s-%s", j.Name(), j.state)
}

// hash computes the Jobs hash value. Used to determne changed state.
func (j *Job) hash() []byte {
	h := make([]byte, 16)
	h = xor16(h, hashMd5([]byte(j.name)))
	h = xor16(h, hashMd5([]byte(j.runStart.Format(time.RFC3339))))
	h = xor16(h, hashMd5([]byte(j.runEnd.Format(time.RFC3339))))
	h = xor16(h, hashMd5([]byte(j.prev.Format(time.RFC3339))))
	h = xor16(h, hashMd5([]byte(j.next.Format(time.RFC3339))))
	h = xor16(h, hashMd5([]byte(strconv.Itoa(int(j.state)))))
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
