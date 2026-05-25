# go-perfstat

[![Go Reference](https://pkg.go.dev/badge/github.com/go-perfstat/go.svg)](https://pkg.go.dev/github.com/go-perfstat/go)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-perfstat/go)](https://goreportcard.com/report/github.com/go-perfstat/go)
[![Release](https://img.shields.io/github/v/release/go-perfstat/go)](https://github.com/go-perfstat/go/releases)

Record/calculate performance statistic for aggregation period and grand total.  
Expose metrics using prometheus or deliver statistic report at the end of the flow.

### Create a package instance

```go
var perfTypeMethod = perfstat.ForTypeName("package.Type", "Method")
```

### Use in a method

```go
defer perfTypeMethod.Stop(perfTypeMethod.Start())
...
```

### Print all stats before exit

```go
perfstat.Print()
```

### Register in Prometheus and expose as Grafana dashboard

[github.com/go-perfstat/prometheus](https://github.com/go-perfstat/prometheus)
