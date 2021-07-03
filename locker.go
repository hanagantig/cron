package cron

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis"
	"time"
)

const lockTTL = 10 * time.Second

type JobLocker interface {
	Lock(ctx context.Context, key string) error
	Extend(ctx context.Context, key string) error
	Unlock(ctx context.Context, key string) error
}

type redisLocker struct {
	client *redis.Client
	rs     *redsync.Redsync
	jobMutexes map[string] *redsync.Mutex
}

func newRedisLocker(redisClient *redis.Client) JobLocker {
	pool := goredis.NewPool(redisClient)
	rs := redsync.New(pool)
	return &redisLocker{
		client: redisClient,
		rs:     rs,
	}
}

func (rl *redisLocker) getMutex(key string) *redsync.Mutex {
	m, ok := rl.jobMutexes[key]
	if !ok || m == nil {
		m = rl.rs.NewMutex(key, redsync.WithExpiry(lockTTL))
		rl.jobMutexes[key] = m
	}

	return m
}

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

func (rl *redisLocker) Extend(ctx context.Context, key string) error {
	mu := rl.getMutex(key)
	if mu == nil {
		return errors.New("can't extend nil mutex")
	}

	attempts := 0
	for _, err := mu.ExtendContext(ctx); err != nil; attempts++ {
		if attempts == 3 {
			return fmt.Errorf("mutex extend error after %v attempts: %v", attempts, err)
		}
	}

	return nil
}

func (rl *redisLocker) Unlock(ctx context.Context, key string) error {
	mu := rl.getMutex(key)
	delete(rl.jobMutexes, key)

	if mu != nil {
		_, err := mu.UnlockContext(ctx)
		return err
	}

	return nil
}
