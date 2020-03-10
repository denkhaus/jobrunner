package jobrunner

import "time"

type JobChangedFunc func(*Job)

var (
	onJobStateChanged JobChangedFunc
)

func OnJobStateChanged(fn func(*Job)) {
	onJobStateChanged = fn
}

func triggerOnJobStateChanged(job *Job) {
	if onJobStateChanged != nil && job.changed() {
		onJobStateChanged(job)
	}
}

func triggerStateUpdates(dur time.Duration) {
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
