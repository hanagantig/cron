package cron

import (
	"context"
	"github.com/go-redsync/redsync/v4/redis"
	"testing"
	"time"
)

type connMock struct {}
type testPool struct {}

func(c *connMock) Get(name string) (string, error) {return name, nil}
func(c *connMock) Set(name string, value string) (bool, error) {return true, nil}
func(c *connMock) SetNX(name string, value string, expiry time.Duration) (bool, error) {return true, nil}
func(c *connMock) Eval(script *redis.Script, keysAndArgs ...interface{}) (interface{}, error) {return keysAndArgs, nil}
func(c *connMock) PTTL(name string) (time.Duration, error) {return time.Second, nil}
func(c *connMock) Close() error {return nil}


func (p *testPool) Get(ctx context.Context) (redis.Conn, error) {
	return &connMock{}, nil
}

func TestRedisLockerRace(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	l := newRedisLocker(&testPool{}, &testPool{})

	job1, job2 := "test_job1", "test_job2"

	err := l.Lock(ctx, job1)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				err := l.Extend(ctx, job1)
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}()

	time.Sleep(time.Second)
	err = l.Lock(ctx, job2)
	if err != nil {
		t.Fatal(err)
	}
	cancel()

	err = l.Unlock(ctx, job1)
	if err != nil {
		t.Fatal(err)
	}

	err = l.Unlock(ctx, job2)
	if err != nil {
		t.Fatal(err)
	}
}
