# stats [![CircleCI](https://circleci.com/gh/segmentio/stats.svg?style=shield)](https://circleci.com/gh/segmentio/stats) [![Go Report Card](https://goreportcard.com/badge/github.com/segmentio/stats)](https://goreportcard.com/report/github.com/segmentio/stats) [![GoDoc](https://godoc.org/github.com/segmentio/stats?status.svg)](https://godoc.org/github.com/segmentio/stats)

A Go package for abstracting stats collection.

Installation
------------

```
go get github.com/segmentio/stats
```

Migration to v4
---------------

Version 4 of the stats package introduced a new way of producing metrics based
on defining struct types with tags on certain fields that define how to interpret
the values. This approach allows for much more efficient metric production as it
allows the program to do quick assignments and increments of the struct fields to
set the values to be reported, and submit them all with one call to the stats
engine, resulting in orders of magnitude faster metrics production. Here's an
example:
```go
type funcMetrics struct {
    calls struct {
        count int           `metric:"count" type:"counter"`
        time  time.Duration `metric:"time"  type:"histogram"`
    } `metric:"func.calls"`
}
```
```go
t := time.Now()
f()
callTime := time.Now().Sub(t)

m := &funcMetrics{}
m.calls.count = 1
m.calls.time = callTime

// Equivalent to:
//
//   stats.Incr("func.calls.count")
//   stats.Observe("func.calls.time", callTime)
//
stats.Report(m)
```
To avoid greatly increasing the complexity of the codebase some old APIs were
removed in favor of this new approach, other were transformed to provide more
flexibility and leverage new features.

The stats package used to only support float values, metrics can now be of
various numeric types (see stats.MakeMeasure for a detailed description),
therefore functions like `stats.Add` now accept an `interface{}` value instead
of `float64`. `stats.ObserveDuration` was also removed since this new approach
makes it obsolete (durations can be passed to `stats.Observe` directly).

The `stats.Engine` type used to be configured through a configuration object
passed to its constructor function, and a few methods (like `Register`) were
exposed to mutate engine instances. This required synchronization in order to
be safe to modify an engine from multiple goroutines. We haven't had a use case
for modifying an engine after creating it so the constraint on being thread-safe
were lifted and the fields exposed on the `stats.Engine` struct type directly to
communicate that they are unsafe to modify concurrently. The helper methods
remain tho to make migration of existing code smoother.

Histogram buckets (mostly used for the prometheus client) are now defined by
default on the `stats.Buckets` global variable instead of within the engine.
This decoupling was made to avoid paying the cost of doing histogram bucket
lookups when producing metrics to backends that don't use them (like datadog
or influxdb for example).

The data model also changed a little. Handlers for metrics produced by an engine
now accept a list of measures instead of single metrics, each measure being made
of a name, a set of fields, and tags to apply to each of those fields. This
allows a more generic and more efficient approach to metric production, better
fits the influxdb data model, while still being compatible with other clients
(datadog, prometheus, ...). A single timeseries is usually identified by the
combination of the measure name, a field name and value, and the set of tags set
on that measure. Refer to each client for a details about how measures are
translated to individual metrics.

Note that no changes were made to the end metrics being produced by each
sub-package (httpstats, procstats, ...). This was important as we must keep
the behavior backward compatible since making changes here would implicitly
break dashboards or monitors set on the various metric collection systems that
this package supports, potentially causing production issues.

_If you find a bug or an API is not available anymore but deserves to be ported feel free to open an issue._

Quick Start
-----------

### Engine

A core concept of the `stats` package is the `Engine`. Every program importing
the package gets a default engine where all metrics produced are aggregated.
The program then has to instantiate clients that will consume from the engine
at regular time intervals and report the state of the engine to metrics
collection platforms.

```go
package main

import (
    "github.com/segmentio/stats"
    "github.com/segmentio/stats/datadog"
)

func main() {
    // Creates a new datadog client publishing metrics to localhost:8125
    dd := datadog.NewClient("localhost:8125")

    // Register the client so it receives metrics from the default engine.
    stats.Register(dd)

    // Flush the default stats engine on return to ensure all buffered
    // metrics are sent to the dogstatsd server.
    defer stats.Flush()

    // That's it! Metrics produced by the application will now be reported!
    // ...
}
```

### Metrics

- [Gauges](https://godoc.org/github.com/segmentio/stats#Gauge)
- [Counters](https://godoc.org/github.com/segmentio/stats#Counter)
- [Histograms](https://godoc.org/github.com/segmentio/stats#Histogram)
- [Timers](https://godoc.org/github.com/segmentio/stats#Timer)

```go
package main

import (
    "github.com/segmentio/stats"
    "github.com/segmentio/stats/datadog"
)

func main() {
    stats.Register(datadog.NewClient("localhost:8125"))
    defer stats.Flush()

    // Increment counters.
    stats.Incr("user.login")
    defer stats.Incr("user.logout")

    // Set a tag on a counter increment.
    stats.Incr("user.login", stats.Tag{"user", "luke"})

    // ...
}
```

### Flushing Metrics

Metrics are stored in a buffer, which will be flushed when it reaches its
capacity. _For most use-cases, you do not need to explicitly send out metrics._

If you're producing metrics only very infrequently, you may have metrics that
stay in the buffer and never get sent out. In that case, you can manually
trigger stats flushes like so:

```go
func main() {
    stats.Register(datadog.NewClient("localhost:8125"))
    defer stats.Flush()

    // Force a metrics flush every second
    go func() {
      for range time.Tick(time.Second) {
        stats.Flush()
      }
    }()

    // ...
}
```

Monitoring
----------

### Processes

The
[github.com/segmentio/stats/procstats](https://godoc.org/github.com/segmentio/stats/procstats)
package exposes an API for creating a statistics collector on local processes.
Statistics are collected for the current process and metrics including Goroutine
count and memory usage are reported.

Here's an example of how to use the collector:
```go
package main

import (
    "github.com/segmentio/stats/datadog"
    "github.com/segmentio/stats/procstats"
)


func main() {
     stats.Register(datadog.NewClient("localhost:8125"))
     defer stats.Flush()

    // Start a new collector for the current process, reporting Go metrics.
    c := procstats.StartCollector(procstats.NewGoMetrics())

    // Gracefully stops stats collection.
    defer c.Close()

    // ...
}
```

One can also collect additional statistics on resource delays, such as
CPU delays, block I/O delays, and paging/swapping delays.  This capability
is currently only available on Linux, and can be optionally enabled as follows:

```
func main() {
    // As above...

    // Start a new collector for the current process, reporting Go metrics.
    c := procstats.StartCollector(procstats.NewDelayMetrics())
    defer c.Close()
}
```

### HTTP Servers

The [github.com/segmentio/stats/httpstats](https://godoc.org/github.com/segmentio/stats/httpstats)
package exposes a decorator of `http.Handler` that automatically adds metric
collection to a HTTP handler, reporting things like request processing time,
error counters, header and body sizes...

Here's an example of how to use the decorator:
```go
package main

import (
    "net/http"

    "github.com/segmentio/stats/datadog"
    "github.com/segmentio/stats/httpstats"
)

func main() {
     stats.Register(datadog.NewClient("localhost:8125"))
     defer stats.Flush()

    // ...

    http.ListenAndServe(":8080", httpstats.NewHandler(
        http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
            // This HTTP handler is automatically reporting metrics for all
            // requests it handles.
            // ...
        }),
    ))
}
```

### HTTP Clients

The [github.com/segmentio/stats/httpstats](https://godoc.org/github.com/segmentio/stats/httpstats)
package exposes a decorator of `http.RoundTripper` which collects and reports
metrics for client requests the same way it's done on the server side.

Here's an example of how to use the decorator:
```go
package main

import (
    "net/http"

    "github.com/segmentio/stats/datadog"
    "github.com/segmentio/stats/httpstats"
)

func main() {
     stats.Register(datadog.NewClient("localhost:8125"))
     defer stats.Flush()

    // Make a new HTTP client with a transport that will report HTTP metrics,
    // set the engine to nil to use the default.
    httpc := &http.Client{
        Transport: httpstats.NewTransport(
            &http.Transport{},
        ),
    }

    // ...
}
```

You can also modify the default HTTP client to automatically get metrics for all
packages using it, this is very convinient to get insights into dependencies.
```go
package main

import (
    "net/http"

    "github.com/segmentio/stats/datadog"
    "github.com/segmentio/stats/httpstats"
)

func main() {
     stats.Register(datadog.NewClient("localhost:8125"))
     defer stats.Flush()

    // Wraps the default HTTP client's transport.
    http.DefaultClient.Transport = httpstats.NewTransport(http.DefaultClient.Transport)

    // ...
}
```

### Redis

The [github.com/segmentio/stats/redisstats](https://godoc.org/github.com/segmentio/stats/redisstats)
package exposes:

* a decorator of
  [`redis.RoundTripper`](https://godoc.org/github.com/segmentio/redis-go#RoundTripper)
  which collects metrics for client requests, and
* a decorator or
  [`redis.ServeRedis`](https://godoc.org/github.com/segmentio/redis-go#HandlerFunc.ServeRedis)
  which collects metrics for server requests.

Here's an example of how to use the decorator on the client side:
```go
package main

import (
    "github.com/segmentio/redis-go"
    "github.com/segmentio/stats/redisstats"
)

func main() {
    stats.Register(datadog.NewClient("localhost:8125"))
    defer stats.Flush()

    client := redis.Client{
        Addr:      "127.0.0.1:6379",
        Transport: redisstats.NewTransport(&redis.Transport{}),
    }

    // ...
}
```

And on the server side:

```go
package main

import (
    "github.com/segmentio/redis-go"
    "github.com/segmentio/stats/redisstats"
)

func main() {
    stats.Register(datadog.NewClient("localhost:8125"))
    defer stats.Flush()

    handler := redis.HandlerFunc(func(res redis.ResponseWriter, req *redis.Request) {
      // Implement handler function here
    })

    server := redis.Server{
        Handler: redisstats.NewHandler(&handler),
    }

    server.ListenAndServe()

    // ...
}
```
