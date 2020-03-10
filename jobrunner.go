package jobrunner

import (
	"bytes"
	"crypto/md5"
	"log"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

type TaskState int
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
	lastHash []byte
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

/////////////////////////////////////////////////////////////////////////////////////
func (j *Job) hash() []byte {
	h := make([]byte, 16)
	h = xor16(h, hashMd5([]byte(j.Name)))
	h = xor16(h, hashMd5([]byte(j.RunStart.Format(time.RFC3339))))
	h = xor16(h, hashMd5([]byte(j.RunEnd.Format(time.RFC3339))))
	h = xor16(h, hashMd5([]byte(j.Prev.Format(time.RFC3339))))
	h = xor16(h, hashMd5([]byte(j.Next.Format(time.RFC3339))))
	h = xor16(h, hashMd5([]byte(j.Result.Error())))
	h = xor16(h, hashMd5([]byte(strconv.Itoa(int(j.state)))))
	h = xor16(h, hashMd5([]byte(strconv.Itoa(int(j.EntryID)))))

	return h
}

/////////////////////////////////////////////////////////////////////////////////////
func (j *Job) changed() bool {
	h := j.hash()
	c := bytes.Compare(h, j.lastHash)
	j.lastHash = h
	return c != 0
}

func xor16(v1, v2 []byte) []byte {
	r := make([]byte, 16)
	for i := 0; i < len(v1); i++ {
		r[i] = v1[i] ^ v2[i]
	}
	return r
}

func hashMd5(t []byte) []byte {
	h := md5.New()
	h.Write(t)
	return h.Sum(nil)
}
