# go-perfstat

[![Go Reference](https://pkg.go.dev/badge/github.com/go-perfstat/go.svg)](https://pkg.go.dev/github.com/go-perfstat/go)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-perfstat/go)](https://goreportcard.com/report/github.com/go-perfstat/go)
[![Release](https://img.shields.io/github/v/release/go-perfstat/go)](https://github.com/go-perfstat/go/releases)

Lightweight performance statistics and execution time aggregation for Go.

`go-perfstat` records execution time, min/max duration, average duration,
total processing time, and leap count for application components, methods,
background jobs, message flows, and distributed operations.

Statistics may be printed locally, exported to Prometheus/Grafana,
or integrated into custom monitoring pipelines.

---

Create a dedicated stat instance per component, operation, or method.

```go
var perfTypeMethod = perfstat.ForTypeName("package.Type", "Method")
```

Record method leap time

```go
defer perfTypeMethod.Leap(time.Now())
// ... calculations
```

A leap may be recorded using an external timestamp received from another
service or upstream system. This allows measuring end-to-end latency across distributed systems.

```go
perfTypeMethod.Leap(message.Timestamp)
```

Print aggregated statistics

```go
perfstat.Print()
```

Export metrics using Prometheus and build Grafana dashboards.

```go
prometheus.MustRegister(perfstat.NewPerfStatMetricsCollector())
```

See [go-perfstat/prometheus](https://github.com/go-perfstat/go)
