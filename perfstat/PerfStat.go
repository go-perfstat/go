package perfstat

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/go-perfstat/go/concurrent"
)

var defaultAggregationPeriodMs int64 = 15000
var stats = concurrent.NewHashMap[string, *concurrent.HashMap[string, *Stat]]()

// Record and aggregate performance statistics such as execution time,
// min/max duration, average duration, and leap count.
//
// Create a dedicated Stat instance per component, subsystem, or operation.
//
//	perf := perfstat.ForName("domain")
//
// Measure local execution time:
//
//	start := perf.Start()
//	// ... calculations
//	fmt.Printf("%.2f ms\n", perfstat.Round(perf.Stop(start)))
//
// Or use defer:
//
//	defer perf.Leap(time.Now())
//
// Leap can also be recorded from external/distributed timestamps:
//
//	perf.Leap(message.CreatedAt)
//
// Periodically export aggregated samples to monitoring systems
// such as Grafana/Prometheus.
type PerfStat struct {
	stat                *Stat
	aggregationPeriodMs int64
}

func ForName(name string) *PerfStat {
	return ForTypeNamePeriod("", name, defaultAggregationPeriodMs)
}

func ForTypeName(typ, name string) *PerfStat {
	return ForTypeNamePeriod(typ, name, defaultAggregationPeriodMs)
}

func ForTypeNamePeriod(typ, name string, period int64) *PerfStat {
	innerMap := stats.Get(typ)
	if innerMap == nil {
		innerMap = concurrent.NewHashMap[string, *Stat]()
		innerMap = stats.PutIfAbsent(typ, innerMap)
	}
	stat := innerMap.Get(name)
	if stat == nil {
		stat = newStat(typ, name)
		stat = innerMap.PutIfAbsent(name, stat)
	}
	stat.peersCount.Add(1)
	perfStat := &PerfStat{
		stat:                stat,
		aggregationPeriodMs: period,
	}
	runtime.SetFinalizer(perfStat, func(p *PerfStat) {
		p.stat.peersCount.Add(-1)
	})
	return perfStat
}

func (this *PerfStat) Leap(start time.Time) int64 {
	now := time.Now()
	timeNs := now.Sub(start).Nanoseconds()
	timeMs := Round(timeNs)

	this.stat.mu.Lock()
	defer this.stat.mu.Unlock()

	this.stat.leapTimeMs = timeMs
	this.stat.totalTimeNs += timeNs
	this.stat.totalTimeThresholdNs += timeNs

	if timeMs > this.stat.maxTimeMs {
		this.stat.maxTimeMs = timeMs
	}
	if timeMs > this.stat.maxTimeThresholdMs {
		this.stat.maxTimeThresholdMs = timeMs
	}
	if timeMs < this.stat.minTimeMs || this.stat.minTimeMs == 0 {
		this.stat.minTimeMs = timeMs
	}
	if timeMs < this.stat.minTimeThresholdMs || this.stat.minTimeThresholdMs == 0 {
		this.stat.minTimeThresholdMs = timeMs
	}
	this.stat.leapsCount++
	this.stat.leapsCountThreshold++

	nowUnixMilli := now.UnixMilli()
	if nowUnixMilli > this.stat.lastAggregationMs+this.aggregationPeriodMs {
		this.stat.lastAggregationMs = nowUnixMilli
		this.stat.avgTimeSampleMs = Round(this.stat.totalTimeThresholdNs / this.stat.leapsCountThreshold)
		this.stat.leapsCountSample = this.stat.leapsCountThreshold
		this.stat.leapsCountThreshold = 0
		this.stat.totalTimeThresholdNs = 0
		this.stat.maxTimeSampleMs = this.stat.maxTimeThresholdMs
		this.stat.maxTimeThresholdMs = 0
		this.stat.minTimeSampleMs = this.stat.minTimeThresholdMs
		this.stat.minTimeThresholdMs = 0
	}

	return timeNs
}

func (this *PerfStat) Reset() {
	this.stat.Reset()
}

func (this *PerfStat) GetStat() *Stat {
	return this.stat
}

func (this *PerfStat) GetType() string {
	return this.stat.GetType()
}

func (this *PerfStat) GetName() string {
	return this.stat.GetName()
}

func (this *PerfStat) GetFullName() string {
	return this.stat.GetFullName()
}

func (this *PerfStat) String() string {
	return this.stat.String()
}

func GetDefaultAggregationPeriodMs() int64 {
	return defaultAggregationPeriodMs
}

func SetDefaultAggregationPeriodMs(ms int64) {
	defaultAggregationPeriodMs = ms
}

func GetByName(name string) *Stat {
	return GetByTypeName("", name)
}

func GetByTypeName(typ, name string) *Stat {
	innerMap := stats.Get(typ)
	if innerMap == nil {
		return nil
	}
	return innerMap.Get(name)
}

func GetAll() map[string]map[string]*Stat {
	result := make(map[string]map[string]*Stat)
	for _, typ := range stats.Keys() {
		innerMap := stats.Get(typ)
		copyMap := make(map[string]*Stat)
		for _, name := range innerMap.Keys() {
			copyMap[name] = innerMap.Get(name)
		}
		result[typ] = copyMap
	}
	return result
}

func Merge(rhs map[string]map[string]*Stat) {
	for typ, rhsStats := range rhs {
		innerMap := stats.Get(typ)
		if innerMap == nil {
			innerMap = concurrent.NewHashMap[string, *Stat]()
			stats.Put(typ, innerMap)
		}

		for name, rhsStat := range rhsStats {
			existing := innerMap.Get(name)
			if existing == nil {
				innerMap.Put(name, rhsStat)
				continue
			}

			existing.mu.Lock()
			existing.leapsCount += rhsStat.leapsCount
			existing.totalTimeNs += rhsStat.totalTimeNs
			if rhsStat.minTimeMs < existing.minTimeMs || existing.minTimeMs == 0 {
				existing.minTimeMs = rhsStat.minTimeMs
			}
			if rhsStat.maxTimeMs > existing.maxTimeMs {
				existing.maxTimeMs = rhsStat.maxTimeMs
			}
			existing.mu.Unlock()
		}
	}
}

func RemoveByType(typ string) {
	stats.Remove(typ)
}

func RemoveByName(name string) {
	RemoveByTypeName("", name)
}

func RemoveByTypeName(typ, name string) {
	innerMap := stats.Get(typ)
	if innerMap != nil {
		innerMap.Remove(name)
	}
}

func ResetAll() {
	for _, typ := range stats.Keys() {
		innerMap := stats.Get(typ)
		for _, name := range innerMap.Keys() {
			innerMap.Get(name).Reset()
		}
	}
}

// returns #.### ms
func Round(nanos int64) float64 {
	scale := 1000.0
	nsInMs := 1_000_000.0
	return math.Round(float64(nanos)*scale/nsInMs) / scale
}

func Print() {
	fmt.Printf("%-70s %10s %10s %10s %15s %10s %9s\n", "Type/Name", "Min(ms)", "Avg(ms)", "Max(ms)", "Total", "Leaps", "Peers")
	fmt.Println(strings.Repeat("-", 140))
	forEachOrdered(GetAll(), func(typ string, innerMap map[string]*Stat) {
		forEachOrdered(innerMap, func(name string, st *Stat) {
			fmt.Printf("%-70s %10.3f %10.3f %10.3f %15s %10d %9d\n",
				strings.Join([]string{typ, name}, "."), st.GetMinTimeMs(), st.GetAvgTimeMs(), st.GetMaxTimeMs(),
				time.Duration(st.GetTotalTimeMs())*time.Millisecond, st.GetLeapsCount(), st.GetPeersCount())
		})
	})
}

func forEachOrdered[V any](m map[string]V, fn func(key string, value V)) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fn(k, m[k])
	}
}
