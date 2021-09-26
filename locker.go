package cron

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redsync/redsync/v4"
	redsyncredis "github.com/go-redsync/redsync/v4/redis"
	"time"
)

const (
	lockTTL           = 10 * time.Second
	extendMaxAttempts = 3
)

// JobLocker is the interface used in this package for distributed locking a job, so that any backend
// can be plugged in.
type JobLocker interface {
	Lock(ctx context.Context, key string) error
	Extend(ctx context.Context, key string) error
	Unlock(ctx context.Context, key string) error
}

type redisLocker struct {
	rs         *redsync.Redsync
	jobMutexes map[string]*redsync.Mutex
}

func newRedisLocker(pool redsyncredis.Pool) JobLocker {
	rs := redsync.New(pool)
	return &redisLocker{
		rs:         rs,
		jobMutexes: make(map[string]*redsync.Mutex),
	}
}

func (rl *redisLocker) getMutex(key string) *redsync.Mutex {
	m, ok := rl.jobMutexes[key]
	if !ok || m == nil {
		m = rl.rs.NewMutex(key+"_cronLock", redsync.WithExpiry(lockTTL))
		rl.jobMutexes[key] = m
	}

	return m
}

// Lock is JobLocker interface method, that implements locks for redsync distributed lock by a job unique key.
func (rl *redisLocker) Lock(ctx context.Context, key string) error {
	mutex := rl.getMutex(key)

	if mutex == nil {
		return errors.New("mutex is nil")
	}

	if err := mutex.LockContext(ctx); err != nil {
		return err
	}

	return nil
}

// Extend is JobLocker interface method, that extends lock for redsync distributed lock by a job unique key.
func (rl *redisLocker) Extend(ctx context.Context, key string) error {
	mu := rl.getMutex(key)
	if mu == nil {
		return errors.New("can't extend nil mutex")
	}

	attempts := 0
	for _, err := mu.ExtendContext(ctx); err != nil; attempts++ {
		if attempts == extendMaxAttempts {
			return fmt.Errorf("mutex extend error after %v attempts: %v", attempts, err)
		}
	}

	return nil
}

// Unlock is JobLocker interface method, that removes redsync distributed lock for a job by the key.
func (rl *redisLocker) Unlock(ctx context.Context, key string) error {
	mu := rl.getMutex(key)
	delete(rl.jobMutexes, key)

	if mu != nil {
		_, err := mu.UnlockContext(ctx)
		return err
	}

	return nil
}
