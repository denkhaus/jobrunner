package jobrunner

// A job runner for executing scheduled or ad-hoc tasks asynchronously from HTTP requests.
//
// It adds a couple of features on top of the Robfig cron package:
// 1. Protection against job panics.  (They print to ERROR instead of take down the process)
// 2. (Optional) Limit on the number of jobs that may run simulatenously, to
//    limit resource consumption.
// 3. (Optional) Protection against multiple instances of a single job running
//    concurrently.  If one execution runs into the next, the next will be queued.
// 4. Cron expressions may be defined in app.conf and are reusable across jobs.
// 5. Job status reporting. [WIP]

import (
	"sync"
	"time"

	"github.com/denkhaus/jobrunner/schedules"
	"github.com/robfig/cron/v3"
)

var (
	//jobList keeps our own list of jobs to maintain detailed state
	jobList = make(map[cron.EntryID]*Job)
	//jobListMu protects jobList
	jobListMu sync.Mutex
)

//addJob adds a Jon
func addJob(job *Job) {
	jobListMu.Lock()
	jobListMu.Unlock()
	jobList[job.EntryID] = job
}

//removeJob removes a Job
func removeJob(job *Job) {
	jobListMu.Lock()
	jobListMu.Unlock()
	delete(jobList, job.EntryID)
}

// cleanCron removes all cron entries with next start time
// equals zero time to avoid job littering
func cleanCron() {
	now := time.Now().In(mainCron.Location())
	for _, entry := range mainCron.Entries() {
		if entry.Schedule.Next(now).IsZero() {
			mainCron.Remove(entry.ID)
			removeJob(entry.Job.(*Job))
		}
	}
}

func Schedule(spec string, job *Job) (cron.EntryID, error) {
	sched, err := cron.ParseStandard(spec)
	if err != nil {
		return -1, err
	}

	defer func() { go cleanCron() }()
	job.EntryID = mainCron.Schedule(sched, job)
	addJob(job)
	return job.EntryID, nil
}

// Run the given job at a fixed interval.
// The interval provided is the time between the job ending and the job being run again.
// The time that the job takes to run is not included in the interval.
func Every(duration time.Duration, job *Job) cron.EntryID {
	defer func() { go cleanCron() }()
	job.EntryID = mainCron.Schedule(cron.Every(duration), job)
	addJob(job)
	return job.EntryID
}

// Run the given job right now.
func OnceNow(job *Job) cron.EntryID {
	defer func() { go cleanCron() }()
	job.EntryID = mainCron.Schedule(schedules.OnceNow(), job)
	addJob(job)
	return job.EntryID
}

// Run the given job N times at a fixed interval.
func NTimesEvery(times int, duration time.Duration, job *Job) cron.EntryID {
	defer func() { go cleanCron() }()
	job.EntryID = mainCron.Schedule(schedules.NTimesEvery(times, duration), job)
	addJob(job)
	return job.EntryID
}

// Run the given job debounced. Consecutive calls on the same job
// will defer the execution time by given duration.
func Debounced(duration time.Duration, job *Job) cron.EntryID {
	var found *cron.Entry
	for _, entry := range mainCron.Entries() {
		if entry.Job.(*Job).Name == job.Name {
			found = &entry
			break
		}
	}

	if found != nil && found.Valid() {
		removeJob(found.Job.(*Job))
		mainCron.Remove(found.ID)
	}

	defer func() { go cleanCron() }()
	job.EntryID = mainCron.Schedule(schedules.NTimesEvery(1, duration), job)
	addJob(job)
	return job.EntryID
}

// Stop ALL active jobs from running at the next scheduled time
func Stop() {
	go mainCron.Stop()
}

// Remove a specific job from running
// Get EntryID from the list job entries jobrunner.Entries()
// If job is in the middle of running, once the process is finished it will be removed
func Remove(id cron.EntryID) {
	mainCron.Remove(id)
}
