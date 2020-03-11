package jobrunner

import "time"

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
func monitorStateUpdates(dur time.Duration) {
	updateState := func() {
		jobListMu.Lock()
		jobListMu.Unlock()

		for _, entry := range mainCron.Entries() {
			jobList[entry.ID].stateMu.Lock()
			jobList[entry.ID].Next = entry.Next
			jobList[entry.ID].Prev = entry.Prev
			jobList[entry.ID].stateMu.Unlock()
			triggerOnJobStateChanged(jobList[entry.ID])
		}
	}

	for _ = range time.Tick(dur) {
		updateState()
	}
}
