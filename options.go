package jobrunner

import "time"

type Options struct {
	//SelfConcurrent defines if a single job is allowed to run concurrently with itself?
	SelfConcurrent bool
	//StateUpdateDuration is the time between state updates
	StateUpdateDuration time.Duration
	// WorkPermits limits the number of jobs allowed to run concurrently.
	WorkPermits chan struct{}
}

var DefaultOptions = Options{
	StateUpdateDuration: 15,
	SelfConcurrent:      false,
	WorkPermits:         make(chan struct{}, 10),
}

type Option func(*Options)

func WithPoolSize(poolSize int) Option {
	return func(opt *Options) {
		opt.WorkPermits = make(chan struct{}, poolSize)
	}
}

func WithSelfConcurrent(sc bool) Option {
	return func(opt *Options) {
		opt.SelfConcurrent = sc
	}
}

func WithStateUpdateDuration(dur time.Duration) Option {
	return func(opt *Options) {
		opt.StateUpdateDuration = dur
	}
}
