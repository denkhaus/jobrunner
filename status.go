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
	if onJobStateChanged != nil {
		onJobStateChanged(job)
	}
}

func triggerStateUpdates(dur time.Duration) {
	updateState := func() {
		jobListMu.Lock()
		jobListMu.Unlock()

		for _, entry := range mainCron.Entries() {
			changed := false
			jobList[entry.ID].stateMu.Lock()
			if jobList[entry.ID].Next != entry.Next {
				jobList[entry.ID].Next = entry.Next
				changed = true
			}
			if jobList[entry.ID].Prev != entry.Prev {
				jobList[entry.ID].Prev = entry.Prev
				changed = true
			}
			jobList[entry.ID].stateMu.Unlock()
			if changed {
				triggerOnJobStateChanged(jobList[entry.ID])
			}
		}
	}

	for _ = range time.Tick(dur) {
		updateState()
	}
}
