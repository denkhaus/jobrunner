package jobrunner

import (
	"context"
	"time"
)

type JobChangedFunc func(*Job)

var (
	onJobStateChanged JobChangedFunc
)

// With OnJobStateChanged you ddefine the state update callback
func OnJobStateChanged(fn JobChangedFunc) {
	onJobStateChanged = fn
}

//triggerOnJobStateChanged triggers state updates if a job has changed
func triggerOnJobStateChanged(job *Job) {
	if onJobStateChanged != nil && job.changed() {
		onJobStateChanged(job)
	}
}

//monitorStateUpdates triggers state updates if a job has changed
//this func is blocking and intended to run as goroutine
func monitorStateUpdates(ctx context.Context, dur time.Duration) {
	updateState := func() {
		jobListMu.Lock()
		defer jobListMu.Unlock()

		for _, entry := range mainCron.Entries() {
			jobList[entry.ID].stateMu.Lock()
			jobList[entry.ID].next = entry.Next
			jobList[entry.ID].prev = entry.Prev
			jobList[entry.ID].stateMu.Unlock()
			triggerOnJobStateChanged(jobList[entry.ID])
		}
	}

	ch := time.Tick(dur)

	for {
		select {
		case <-ch:
			updateState()
		case <-ctx.Done():
			return
		}
	}
}
