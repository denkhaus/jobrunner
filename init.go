package jobrunner

import (
	"time"

	"github.com/robfig/cron/v3"
)

const DEFAULT_JOB_POOL_SIZE = 10

var (
	stateUpdateDuration time.Duration
	// Singleton instance of the underlying job scheduler.
	mainCron *cron.Cron

	// This limits the number of jobs allowed to run concurrently.
	workPermits chan struct{}

	// Is a single job allowed to run concurrently with itself?
	selfConcurrent bool
)

var (
	functions = []interface{}{makeWorkPermits, isSelfConcurrent, updateduration}
)

func makeWorkPermits(bufferCapacity int) {
	if bufferCapacity <= 0 {
		workPermits = make(chan struct{}, DEFAULT_JOB_POOL_SIZE)
	} else {
		workPermits = make(chan struct{}, bufferCapacity)
	}
}

func isSelfConcurrent(concurrencyFlag int) {
	if concurrencyFlag <= 0 {
		selfConcurrent = false
	} else {
		selfConcurrent = true
	}
}

func updateduration(dur int) {
	stateUpdateDuration = time.Second * time.Duration(dur)
}

func Start(v ...int) {
	mainCron = cron.New()
	for i, option := range v {
		functions[i].(func(int))(option)
	}

	mainCron.Start()
	go triggerStateUpdates(stateUpdateDuration)
}
