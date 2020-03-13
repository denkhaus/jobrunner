package jobrunner

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
)

var (
	// mainCron is a singleton instance of the underlying job scheduler.
	mainCron *cron.Cron
	// options - the startup options
	options = &DefaultOptions
)

func Start(ctx context.Context, opts ...Option) {
	mainCron = cron.New()
	for _, o := range opts {
		o(options)
	}

	mainCron.Start()
	go monitorStateUpdates(ctx, options.StateUpdateDuration)
}

func Now() time.Time {
	return time.Now().In(mainCron.Location())
}
