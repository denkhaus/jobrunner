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

type BuildWrapperFunc func(*Job) cron.Job

func BuildWrapperDefault(job *Job) cron.Job {
	return job
}

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

	job.setState(JobStateIdle, false)
	jobList[job.entryID] = job
	triggerOnJobChanged(job)

	return job.entryID
}

//removeJob removes a Job and triggers a state update if needed
func removeJob(id cron.EntryID, triggerStateUpdate bool) {
	jobListMu.Lock()
	defer jobListMu.Unlock()
	if job, ok := jobList[id]; ok {
		job.setState(JobStateFinished, triggerStateUpdate)
	}

	delete(jobList, id)
}

// cleanCron removes all cron entries with next start time
// equal to zero to avoid job littering
func cleanCron() {
	for _, entry := range mainCron.Entries() {
		if entry.Schedule.Next(Now()).IsZero() {
			mainCron.Remove(entry.ID)
			removeJob(entry.ID, true)
		}
	}
}

// Schedule adds a Job to be run on the given schedule.
func Schedule(spec string, job *Job, buildWrapper BuildWrapperFunc) (cron.EntryID, error) {
	sched, err := cron.ParseStandard(spec)
	if err != nil {
		return InvalidEntryID, err
	}

	job.entryID = mainCron.Schedule(sched, buildWrapper(job))
	job.typ = JobTypeOnce
	return addJob(job), nil
}

// Run the given job at a fixed interval.
// The interval provided is the time between the job ending and the job being run again.
// The time that the job takes to run is not included in the interval.
func Every(duration time.Duration, job *Job, buildWrapper BuildWrapperFunc) cron.EntryID {
	job.entryID = mainCron.Schedule(cron.Every(duration), buildWrapper(job))
	job.typ = JobTypeRecurring
	return addJob(job)
}

// Run the given job right now.
func OnceNow(job *Job, buildWrapper BuildWrapperFunc) cron.EntryID {
	job.entryID = mainCron.Schedule(schedules.OnceNow(), buildWrapper(job))
	job.typ = JobTypeOnce
	return addJob(job)
}

// Run the given job at a fixed time.
func At(dt time.Time, job *Job, buildWrapper BuildWrapperFunc) cron.EntryID {
	if dt.Before(Now()) {
		return InvalidEntryID
	}

	job.entryID = mainCron.Schedule(schedules.Absolute(dt), buildWrapper(job))
	job.typ = JobTypeOnce
	return addJob(job)
}

// Run the given job N times at a fixed interval.
func NTimesEvery(times int, duration time.Duration, job *Job, buildWrapper BuildWrapperFunc) cron.EntryID {
	job.entryID = mainCron.Schedule(schedules.NTimesEvery(times, duration), buildWrapper(job))
	job.typ = JobTypeRecurring
	return addJob(job)
}

// Run the given job debounced. Consecutive calls on a job with the same name
// will defer the execution time by given duration.
func Debounced(dur time.Duration, job *Job, buildWrapper BuildWrapperFunc) cron.EntryID {
	for _, entry := range mainCron.Entries() {
		if entry.Job.(*Job).name == job.name {
			if entry.Valid() {
				removeJob(entry.ID, false)
				mainCron.Remove(entry.ID)
			}
		}
	}

	job.entryID = mainCron.Schedule(
		schedules.Absolute(Now().Add(dur)),
		buildWrapper(job),
	)

	return addJob(job)
}

// Stop all active jobs from running at the next scheduled time
func Stop() {
	go mainCron.Stop()
}
