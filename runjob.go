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
package jobrunner

import (
	"time"

	"github.com/denkhaus/jobrunner/schedules"

	"github.com/robfig/cron/v3"
)

var (
	delayer = NewDelayer()
)

// Callers can use jobs.Func to wrap a raw func.
// (Copying the type to this package makes it more visible)
//
// For example:
//    jobrunner.Schedule("cron.frequent", jobs.Func(myFunc))
type Func func()

func (r Func) Run() { r() }

// cleanCron removes all cron entries with next start time equals zero time
func cleanCron() {
	now := time.Now().In(MainCron.Location())
	for _, entry := range MainCron.Entries() {
		if entry.Schedule.Next(now).IsZero() {
			MainCron.Remove(entry.ID)
		}
	}
}

func Schedule(spec string, job cron.Job) error {
	sched, err := cron.ParseStandard(spec)
	if err != nil {
		return err
	}

	MainCron.Schedule(sched, New(job))
	return nil
}

// Run the given job at a fixed interval.
// The interval provided is the time between the job ending and the job being run again.
// The time that the job takes to run is not included in the interval.
func Every(duration time.Duration, job cron.Job) {
	MainCron.Schedule(cron.Every(duration), New(job))
	go cleanCron()
}

// Run the given job right now.
func OnceNow(job cron.Job) {
	MainCron.Schedule(schedules.OnceNow(), New(job))
	go cleanCron()
}

// Run the given job N times at a fixed interval.
func NTimesEvery(times int, duration time.Duration, job cron.Job) {
	MainCron.Schedule(schedules.NTimesEvery(times, duration), New(job))
	go cleanCron()
}

// Run the given job debounced. Consecutive calls on the same job
// will defer the execution time by given duration.
func Debounced(duration time.Duration, job cron.Job) {
	newJob := New(job)
	var found cron.Entry
	for _, entry := range MainCron.Entries() {
		if entry.Job.(*Job).Name == newJob.Name {
			found = entry
			break
		}
	}

	if found.Valid() {
		MainCron.Remove(found.ID)
	}

	MainCron.Schedule(schedules.NTimesEvery(1, duration), newJob)
	go cleanCron()
}
