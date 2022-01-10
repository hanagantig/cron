[![GoDoc](http://godoc.org/github.com/hanagantig/cron?status.png)](http://godoc.org/github.com/hanagantig/cron)
[![Build Status](https://travis-ci.org/hanagantig/cron.svg?branch=main)](https://travis-ci.com/hanagantig/cron)

# cron

*Essentially it's a fork of [robfig cron](https://github.com/robfig/cron) with **distributed lock** feature.*

**This cron package is helpful when you scale your cron instances and want to run same jobs once at a time**.

There is built-in implementation with redsync distributed lock
([Redis-based distributed mutual exclusion lock](https://github.com/go-redsync/redsync)).
> :warning: *Be careful and make sure to understand how the redlock algorithm works. Check [this discussion](http://antirez.com/news/101).*

Moreover, you are capable to provide your own locker implementation.
Check the [feature example](#job-distributed-lock). 

## Getting started
To download the package, run:
```bash
go get github.com/hanagantig/cron
```

Import it in your program as:
```go
import "github.com/hanagantig/cron"
```

A simple usage:
```go
myCron := cron.New(cron.WithLogger(logger), cron.WithRedsyncLocker(redisPool))
myCron.AddFunc("*/15 * * * *", "myTaskKey", myTaskFunc)
myCron.Start()
```

Refer to the documentation here: http://godoc.org/github.com/hanagantig/cron

## Features
### Spec parsers
Standard cron spec parsing by default (first field is "minute"), with an easy way to opt into the seconds field (quartz-compatible). Although, note that the year field (optional in Quartz) is not supported.
```go
// Seconds field, required
cron.New(cron.WithSeconds())

// Seconds field, optional
cron.New(cron.WithParser(cron.NewParser(
	cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)))
```
### Logging
Extensible, key/value logging via an interface that complies with the https://github.com/go-logr/logr project.
Configure a logger:
```go
cron.New(
    cron.WithLogger(cron.VerbosePrintfLogger(logger))
)
```

### Interceptors
Chain & JobWrapper types allow you to install "interceptors" to add cross-cutting behavior like the following:
  - Recover any panics from jobs
  - Delay a job's execution if the previous run hasn't completed yet
  - Skip a job's execution if the previous run hasn't completed yet
  - Log each job's invocations
  - Notification when jobs are completed
  - Panic recovery and configure the panic logger
```go
cron.New(cron.WithChain(
  cron.Recover(logger),  // or use cron.DefaultLogger
))
```
### Job distributed lock
You can use built-in redsync locker to lock your jobs or provide a custom locker.
```go
// with built-in redsync locker
cron.New(
    cron.WithRedsyncLocker(pool)
)

// with your own locker
cron.New(
    cron.WithLocks(locker)
)
```

Check the [documentation](https://pkg.go.dev/github.com/hanagantig/cron#WithRedsyncLocker) for `WithRedsyncLocker` option.
## Background - Cron spec format

There are two cron spec formats in common usage:

- The "standard" cron format, described on [the Cron wikipedia page] and used by
  the cron Linux system utility.

- The cron format used by [the Quartz Scheduler], commonly used for scheduled
  jobs in Java software

[the Cron wikipedia page]: https://en.wikipedia.org/wiki/Cron
[the Quartz Scheduler]: http://www.quartz-scheduler.org/documentation/quartz-2.3.0/tutorials/tutorial-lesson-06.html

The original version of this package included an optional "seconds" field, which
made it incompatible with both of these formats. Now, the "standard" format is
the default format accepted, and the Quartz format is opt-in.