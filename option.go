package cron

import (
	redsyncredis "github.com/go-redsync/redsync/v4/redis"
	"time"
)

// Option represents a modification to the default behavior of a Cron.
type Option func(*Cron)

// WithLocation overrides the timezone of the cron instance.
func WithLocation(loc *time.Location) Option {
	return func(c *Cron) {
		c.location = loc
	}
}

// WithSeconds overrides the parser used for interpreting job schedules to
// include a seconds field as the first one.
func WithSeconds() Option {
	return WithParser(NewParser(
		Second | Minute | Hour | Dom | Month | Dow | Descriptor,
	))
}

// WithParser overrides the parser used for interpreting job schedules.
func WithParser(p ScheduleParser) Option {
	return func(c *Cron) {
		c.parser = p
	}
}

// WithChain specifies Job wrappers to apply to all jobs added to this cron.
// Refer to the Chain* functions in this package for provided wrappers.
func WithChain(wrappers ...JobWrapper) Option {
	return func(c *Cron) {
		c.chain = NewChain(wrappers...)
	}
}

// WithLogger uses the provided logger.
func WithLogger(logger Logger) Option {
	return func(c *Cron) {
		c.logger = logger
	}
}

// WithDistributedLock specifies JobLocker realization for all jobs added to this cron.
func WithDistributedLock(locker JobLocker) Option {
	return func(c *Cron) {
		c.jobLocker = locker
	}
}

// WithRedsyncLocker specifies a built-in implementation of JobLocker with redsync distributed lock.
// Redis-based distributed mutual exclusion lock https://github.com/go-redsync/redsync
func WithRedsyncLocker(pool redsyncredis.Pool) Option  {
	return WithDistributedLock(newRedisLocker(pool))
}
