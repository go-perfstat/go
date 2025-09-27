package perfstat

import (
	"math"
	"runtime"
	"time"

	"github.com/go-perfstat/go/concurrent"
)

var defaultAggregationPeriodMs int64 = 5000
var stats = concurrent.NewHashMap[string, *concurrent.HashMap[string, *Stat]]()

// Record/calculate performance statistic for aggregation period and grand total.
//
// Create an instance per component
//
//	perf := perfstat.ForName("domain")
//
// Use in method
//
//	t := perf.Start()
//	... calculations
//	fmt.Printf("%f ms\n", perfstat.Round(perf.Stop(t)))
//
// Once per aggregation time period flush samples somethere for Graphana to pick-up
type PerfStat struct {
	stat                *Stat
	lastAggregationMs   int64
	aggregationPeriodMs int64
}

func ForName(name string) *PerfStat {
	return ForTypeNamePeriod("", name, defaultAggregationPeriodMs)
}

func ForTypeName(typ, name string) *PerfStat {
	return ForTypeNamePeriod(typ, name, defaultAggregationPeriodMs)
}

func ForTypeNamePeriod(typ, name string, period int64) *PerfStat {
	innerMap, ok := stats.Get(typ)
	if !ok {
		innerMap = concurrent.NewHashMap[string, *Stat]()
		innerMap = stats.PutIfAbsent(typ, innerMap)
	}
	stat, ok := innerMap.Get(name)
	if !ok || stat == nil {
		stat = newStat(typ, name)
		stat = innerMap.PutIfAbsent(name, stat)
	}
	stat.peersCount.Add(1)
	perfStat := &PerfStat{
		stat:                stat,
		lastAggregationMs:   time.Now().UnixMilli(),
		aggregationPeriodMs: period,
	}
	runtime.SetFinalizer(perfStat, func(p *PerfStat) {
		p.stat.peersCount.Add(-1)
	})
	return perfStat
}

func (p *PerfStat) Start() time.Time {
	now := time.Now().UnixMilli()
	if p.stat.leapsCountThreshold > 0 && now > p.lastAggregationMs+p.aggregationPeriodMs {
		p.stat.mu.Lock()
		if p.stat.leapsCountThreshold > 0 && now > p.lastAggregationMs+p.aggregationPeriodMs {
			p.lastAggregationMs = now
			p.stat.avgTimeSampleMs = Round(p.stat.totalTimeThresholdNs / p.stat.leapsCountThreshold)
			p.stat.leapsCountSample = p.stat.leapsCountThreshold
			p.stat.leapsCountThreshold = 0
			p.stat.totalTimeThresholdNs = 0
			p.stat.maxTimeSampleMs = p.stat.maxTimeThresholdMs
			p.stat.maxTimeThresholdMs = 0
			p.stat.minTimeSampleMs = p.stat.minTimeThresholdMs
			p.stat.minTimeThresholdMs = 0
		}
		p.stat.mu.Unlock()
	}
	return time.Now()
}

func (p *PerfStat) Stop(start time.Time) int64 {
	timeNs := time.Since(start).Nanoseconds()
	timeMs := Round(timeNs)

	p.stat.mu.Lock()
	defer p.stat.mu.Unlock()

	p.stat.leapTimeMs = timeMs
	p.stat.totalTimeNs += timeNs
	p.stat.totalTimeThresholdNs += timeNs

	if timeMs > p.stat.maxTimeMs {
		p.stat.maxTimeMs = timeMs
	}
	if timeMs > p.stat.maxTimeThresholdMs {
		p.stat.maxTimeThresholdMs = timeMs
	}
	if timeMs < p.stat.minTimeMs || p.stat.minTimeMs == 0 {
		p.stat.minTimeMs = timeMs
	}
	if timeMs < p.stat.minTimeThresholdMs || p.stat.minTimeThresholdMs == 0 {
		p.stat.minTimeThresholdMs = timeMs
	}
	p.stat.leapsCount++
	p.stat.leapsCountThreshold++

	return timeNs
}

func (p *PerfStat) Reset() {
	p.stat.Reset()
}

func (p *PerfStat) GetStat() *Stat {
	return p.stat
}

func (p *PerfStat) GetType() string {
	return p.stat.GetType()
}

func (p *PerfStat) GetName() string {
	return p.stat.GetName()
}

func (p *PerfStat) GetFullName() string {
	return p.stat.GetFullName()
}

func (p *PerfStat) String() string {
	return p.stat.String()
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
	innerMap, ok := stats.Get(typ)
	if !ok {
		return nil
	}
	st, ok := innerMap.Get(name)
	if !ok {
		return nil
	}
	return st
}

func GetAll() map[string]map[string]*Stat {
	result := make(map[string]map[string]*Stat)
	for _, typ := range stats.Keys() {
		innerMap, _ := stats.Get(typ)
		copyMap := make(map[string]*Stat)
		for _, name := range innerMap.Keys() {
			st, _ := innerMap.Get(name)
			copyMap[name] = st
		}
		result[typ] = copyMap
	}
	return result
}

func Merge(rhs map[string]map[string]*Stat) {
	for typ, rhsStats := range rhs {
		innerMap, ok := stats.Get(typ)
		if !ok {
			innerMap = concurrent.NewHashMap[string, *Stat]()
			stats.Put(typ, innerMap)
		}

		for name, rhsStat := range rhsStats {
			existing, ok := innerMap.Get(name)
			if !ok || existing == nil {
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

func ResetAll() {
	for _, typ := range stats.Keys() {
		innerMap, _ := stats.Get(typ)
		for _, name := range innerMap.Keys() {
			st, _ := innerMap.Get(name)
			st.Reset()
		}
	}
}

// returns #.### ms
func Round(nanos int64) float64 {
	scale := 1000.0
	nsInMs := 1_000_000.0
	return math.Round(float64(nanos)*scale/nsInMs) / scale
}
