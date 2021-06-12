package cron

import (
	"github.com/go-redis/redis"
	"github.com/go-redsync/redsync/v3"
	redsyncredis "github.com/go-redsync/redsync/v3/redis"
	"github.com/go-redsync/redsync/v3/redis/goredis"
	"time"
)

const lockTTL = 10 * time.Second

type JobLocker interface {
	Lock(key string) error
	Unlock(key string)
	//IsLocked(entry Entry) bool
}

type redisLocker struct {
	client *redis.Client
	rs     *redsync.Redsync
}

func newRedisLocker(redisClient *redis.Client) JobLocker {
	pool := goredis.NewGoredisPool(redisClient)
	rs := redsync.New([]redsyncredis.Pool{pool})
	return &redisLocker{
		client: redisClient,
		rs:     rs,
	}
}

func (rl *redisLocker) Lock(key string) error {
	mutex := rl.rs.NewMutex(key)

	if err := mutex.Lock(); err != nil {
		return err
	}

	if mutex != nil {
		defer func() {
			_, _ = mutex.Unlock()
		}()
	}

	exists := rl.client.Exists(key)
	if exists.Err() != nil {
		return exists.Err()
	}

	res := rl.client.Set(key, "locked", lockTTL)
	if res.Err() != nil {
		return res.Err()
	}

	return nil
}

func (rl *redisLocker) Unlock(key string) {
	rl.client.Del(key)
}
