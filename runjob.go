package jobrunner

// A job runner for executing scheduled or ad-hoc tasks asynchronously from HTTP requests.
//
// It adds a couple of features on top of the Robfig cron package:
// 1. Protection against job panics.  (They print to ERROR instead of take down the process)
// 2. (Optional) Limit on the number of jobs that may run simulatenously, to
//    limit resource consumption.
// 3. (Optional) Protection against multiple instances of a single job running
//    concurrently.  If one execution runs into the next, the next will be queued.

import (
	"sync"
	"time"

	"github.com/denkhaus/jobrunner/schedules"
	"github.com/robfig/cron/v3"
)

const (
	InvalidEntryID = -1
)

var (
	//jobList keeps our own list of jobs to maintain detailed state
	jobList = make(map[cron.EntryID]*Job)
	//jobListMu protects jobList
	jobListMu sync.Mutex
)

//addJob adds a Job and triggers a state update
func addJob(job *Job) cron.EntryID {
	jobListMu.Lock()
	defer jobListMu.Unlock()

	job.setState(Idle, false)
	jobList[job.entryID] = job
	triggerOnJobStateChanged(job)

	return job.entryID
}

//removeJob removes a Job and triggers a state update if needed
func removeJob(job *Job, triggerStateUpdate bool) {
	jobListMu.Lock()
	defer jobListMu.Unlock()
	job.setState(Finished, triggerStateUpdate)
	delete(jobList, job.entryID)
}

// cleanCron removes all cron entries with next start time
// equal to zero to avoid job littering
func cleanCron() {
	for _, entry := range mainCron.Entries() {
		if entry.Schedule.Next(Now()).IsZero() {
			mainCron.Remove(entry.ID)
			removeJob(entry.Job.(*Job), true)
		}
	}
}

// Schedule adds a Job to be run on the given schedule.
func Schedule(spec string, job *Job) (cron.EntryID, error) {
	sched, err := cron.ParseStandard(spec)
	if err != nil {
		return InvalidEntryID, err
	}

	job.entryID = mainCron.Schedule(sched, job)
	return addJob(job), nil
}

// Run the given job at a fixed interval.
// The interval provided is the time between the job ending and the job being run again.
// The time that the job takes to run is not included in the interval.
func Every(duration time.Duration, job *Job) cron.EntryID {
	job.entryID = mainCron.Schedule(cron.Every(duration), job)
	return addJob(job)
}

// Run the given job right now.
func OnceNow(job *Job) cron.EntryID {
	job.entryID = mainCron.Schedule(schedules.OnceNow(), job)
	return addJob(job)
}

// Run the given job at a fixed time.
func At(dt time.Time, job *Job) cron.EntryID {
	if dt.Before(Now()) {
		return InvalidEntryID
	}
	job.entryID = mainCron.Schedule(schedules.Absolute(dt), job)
	return addJob(job)
}

// Run the given job N times at a fixed interval.
func NTimesEvery(times int, duration time.Duration, job *Job) cron.EntryID {
	job.entryID = mainCron.Schedule(schedules.NTimesEvery(times, duration), job)
	return addJob(job)
}

// Run the given job debounced. Consecutive calls on a job with the same name
// will defer the execution time by given duration.
func Debounced(dur time.Duration, job *Job) cron.EntryID {
	for _, entry := range mainCron.Entries() {
		if entry.Job.(*Job).name == job.name {
			if entry.Valid() {
				removeJob(entry.Job.(*Job), false)
				mainCron.Remove(entry.ID)
			}
		}
	}

	job.entryID = mainCron.Schedule(
		schedules.Absolute(Now().Add(dur)),
		job,
	)

	return addJob(job)
}

// Stop all active jobs from running at the next scheduled time
func Stop() {
	go mainCron.Stop()
}
