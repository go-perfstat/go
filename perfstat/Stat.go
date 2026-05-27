package perfstat

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Stat struct {
	mu                   sync.Mutex
	typ                  string
	name                 string
	leapsCount           int64
	leapsCountSample     int64
	leapsCountThreshold  int64
	totalTimeNs          int64
	totalTimeThresholdNs int64
	avgTimeSampleMs      float64
	leapTimeMs           float64
	minTimeMs            float64
	minTimeSampleMs      float64
	minTimeThresholdMs   float64
	maxTimeMs            float64
	maxTimeSampleMs      float64
	maxTimeThresholdMs   float64
	peersCount           atomic.Int64
	lastAggregationMs    int64
}

func newStat(typ, name string) *Stat {
	return &Stat{typ: typ, name: name}
}

func (this *Stat) Reset() {
	this.mu.Lock()
	defer this.mu.Unlock()

	this.totalTimeNs = 0
	this.totalTimeThresholdNs = 0
	this.maxTimeMs = 0
	this.maxTimeSampleMs = 0
	this.maxTimeThresholdMs = 0
	this.minTimeMs = 0
	this.minTimeSampleMs = 0
	this.minTimeThresholdMs = 0
	this.leapsCount = 0
	this.leapsCountSample = 0
	this.leapsCountThreshold = 0
	this.lastAggregationMs = 0
	this.avgTimeSampleMs = 0
	this.leapTimeMs = 0
}

func (this *Stat) GetType() string {
	this.mu.Lock()
	defer this.mu.Unlock()

	return this.typ
}

func (this *Stat) GetName() string {
	this.mu.Lock()
	defer this.mu.Unlock()

	return this.name
}

func (this *Stat) GetFullName() string {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.typ == "" {
		return this.name
	}
	return this.typ + "/" + this.name
}

func (this *Stat) GetLeapsCount() int64 {
	this.mu.Lock()
	defer this.mu.Unlock()

	return this.leapsCount
}

func (this *Stat) GetLeapsCountSample() int64 {
	this.mu.Lock()
	defer this.mu.Unlock()

	return this.leapsCountSample
}

func (this *Stat) GetTotalTimeMs() float64 {
	this.mu.Lock()
	defer this.mu.Unlock()

	return Round(this.totalTimeNs)
}

func (this *Stat) GetLeapTimeMs() float64 {
	this.mu.Lock()
	defer this.mu.Unlock()

	return this.leapTimeMs
}

func (this *Stat) GetMinTimeMs() float64 {
	this.mu.Lock()
	defer this.mu.Unlock()

	return this.minTimeMs
}

func (this *Stat) GetMinTimeSampleMs() float64 {
	this.mu.Lock()
	defer this.mu.Unlock()

	return this.minTimeSampleMs
}

func (this *Stat) GetMaxTimeMs() float64 {
	this.mu.Lock()
	defer this.mu.Unlock()

	return this.maxTimeMs
}

func (this *Stat) GetMaxTimeSampleMs() float64 {
	this.mu.Lock()
	defer this.mu.Unlock()

	return this.maxTimeSampleMs
}

func (this *Stat) GetAvgTimeMs() float64 {
	this.mu.Lock()
	defer this.mu.Unlock()

	if this.leapsCount == 0 {
		return 0
	}
	return Round(this.totalTimeNs / this.leapsCount)
}

func (this *Stat) GetAvgTimeSampleMs() float64 {
	this.mu.Lock()
	defer this.mu.Unlock()

	return this.avgTimeSampleMs
}

func (this *Stat) GetPeersCount() int64 {
	return this.peersCount.Load()
}

func (this *Stat) String() string {
	this.mu.Lock()
	defer this.mu.Unlock()

	avgTimeMs := float64(0)
	if this.leapsCount != 0 {
		avgTimeMs = Round(this.totalTimeNs / this.leapsCount)
	}

	return fmt.Sprintf("%.2f %.2f %.2f", this.minTimeMs, avgTimeMs, this.maxTimeMs)
}
