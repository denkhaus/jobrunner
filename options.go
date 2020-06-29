package jobrunner

import "time"

type Options struct {
	//StateUpdateDuration is the time between state updates
	StateUpdateDuration time.Duration
}

var DefaultOptions = Options{
	StateUpdateDuration: 15 * time.Second,
}

type Option func(*Options)

func WithStateUpdateDuration(dur time.Duration) Option {
	return func(opt *Options) {
		opt.StateUpdateDuration = dur
	}
}
