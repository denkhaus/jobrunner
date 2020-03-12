package jobrunner

//go:generate stringer -type=TaskState

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
	uuid "github.com/satori/go.uuid"
)

type JobFunc func() error

//TaskState is the stats a Job can run into
type TaskState int

const (
	Idle TaskState = iota + 1
	Running
	Finished
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
	UUID     uuid.UUID
	EntryID  cron.EntryID
	lastHash []byte
	running  sync.Mutex
	stateMu  sync.Mutex
}

// New creates a new Job
func New(name string, fn JobFunc) *Job {
	return &Job{
		Name:    name,
		UUID:    uuid.NewV4(),
		EntryID: -1,
		jobFunc: fn,
	}
}

// setState sets the Jobs state
func (j *Job) setState(state TaskState, trigger bool) {
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
		j.RunEnd = time.Now().UTC()
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
	j.RunStart = time.Now().UTC()
	j.Result = j.jobFunc()
}

// String  is the Jobs string representation
func (j *Job) String() string {
	return fmt.Sprintf("%s-%s", j.Name, j.state)
}

// hash computes the Jobs hash value. Used to determne changed state.
func (j *Job) hash() []byte {
	h := make([]byte, 16)
	h = xor16(h, hashMd5([]byte(j.Name)))
	h = xor16(h, hashMd5([]byte(j.RunStart.Format(time.RFC3339))))
	h = xor16(h, hashMd5([]byte(j.RunEnd.Format(time.RFC3339))))
	h = xor16(h, hashMd5([]byte(j.Prev.Format(time.RFC3339))))
	h = xor16(h, hashMd5([]byte(j.Next.Format(time.RFC3339))))
	h = xor16(h, hashMd5([]byte(strconv.Itoa(int(j.state)))))
	h = xor16(h, hashMd5([]byte(strconv.Itoa(int(j.EntryID)))))
	if j.Result != nil {
		h = xor16(h, hashMd5([]byte(j.Result.Error())))
	}

	return h
}

// changed compares the last Job hash with the current
// so we find out if any properties have changed
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
