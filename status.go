package jobrunner

import (
	"time"

	"github.com/robfig/cron/v3"
)

type StatusData struct {
	Id        cron.EntryID
	JobRunner *Job
	Next      time.Time
	Prev      time.Time
}

// Return detailed list of currently running recurring jobs
// to remove an entry, first retrieve the ID of entry
func Entries() []cron.Entry {
	return mainCron.Entries()
}

func Status() []StatusData {
	ents := mainCron.Entries()

	Statuses := make([]StatusData, len(ents))
	for k, v := range ents {
		Statuses[k].Id = v.ID
		Statuses[k].JobRunner = AddJob(v.Job)
		Statuses[k].Next = v.Next
		Statuses[k].Prev = v.Prev

	}

	return Statuses
}

func StatusJson() map[string]interface{} {
	return map[string]interface{}{
		"jobrunner": Status(),
	}
}

func AddJob(job cron.Job) *Job {
	return job.(*Job)
}
