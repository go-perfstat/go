# go-perfstat

[![Go Reference](https://pkg.go.dev/badge/github.com/go-perfstat/go.svg)](https://pkg.go.dev/github.com/go-perfstat/go)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-perfstat/go)](https://goreportcard.com/report/github.com/go-perfstat/go)
[![Release](https://img.shields.io/github/v/release/go-perfstat/go)](https://github.com/go-perfstat/go/releases)

Lightweight performance statistics and execution time aggregation for Go.

go-perfstat records execution time, min/max duration, average duration,
total processing time, and leap count for application components, methods,
background jobs, message flows, and distributed operations.

A stat instance is safe to share and call concurrently. It may represent a
global operation, a component, a method, or an object instance. For example,
a stat created per database connection may track query execution time, while
the peers count gives a rough indication of the connection pool size.

Peers count is reduced when unused stat peers are garbage collected, so it is
not intended to be an exact real-time object counter. It is useful as an
operational signal: for example, to understand approximate pool size or detect
unexpected growth in the number of tracked objects.

Statistics may be printed locally, exported to Prometheus/Grafana,
or integrated into custom monitoring pipelines.

---

Create a dedicated stat instance per component, subsystem, service, shared object,
or globally for a specific operation type.

A stat instance is intended to be reused and called concurrently from multiple
goroutines rather than created per method invocation or operation.

```go
var perfTypeMethod = perfstat.ForTypeName("package.Type", "Method")
```

Record leap time:

```go
defer perfTypeMethod.Leap(time.Now())
// ... calculations
```

A leap may also be recorded using an external timestamp received from another
service or upstream system. This allows measuring end-to-end latency across
distributed systems.

```go
perfTypeMethod.Leap(message.Timestamp)
```

Print aggregated statistics:

```go
perfstat.Print()
```

Example output:

```bash
Type/Name                                                                 Min(ms)    Avg(ms)    Max(ms)           Total      Leaps     Peers
--------------------------------------------------------------------------------------------------------------------------------------------
ApplicationRunner1.Run                                                      1.170      1.170      1.170             1ms          1         1
Service1.Start                                                              0.000      0.000      0.000              0s          0         1
Service2.Start                                                           5005.196   5005.196   5005.196          5.005s          1         1
```

Export metrics using Prometheus and build Grafana dashboards.

```go
prometheus.MustRegister(perfstat.NewPerfStatMetricsCollector())
```

See [go-perfstat/prometheus](https://github.com/go-perfstat/prometheus).
