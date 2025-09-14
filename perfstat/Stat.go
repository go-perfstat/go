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
}

func newStat(typ, name string) *Stat {
	return &Stat{typ: typ, name: name}
}

func (s *Stat) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalTimeNs = 0
	s.totalTimeThresholdNs = 0
	s.maxTimeMs = 0
	s.maxTimeSampleMs = 0
	s.maxTimeThresholdMs = 0
	s.minTimeMs = 0
	s.minTimeSampleMs = 0
	s.minTimeThresholdMs = 0
	s.leapsCount = 0
	s.leapsCountSample = 0
	s.leapsCountThreshold = 0
}

func (s *Stat) GetType() string {
	return s.typ
}

func (s *Stat) GetName() string {
	return s.name
}

func (s *Stat) GetFullName() string {
	if s.typ == "" {
		return s.name
	}
	return s.typ + "/" + s.name
}

func (s *Stat) GetLeapsCount() int64 {
	return s.leapsCount
}

func (s *Stat) GetLeapsCountSample() int64 {
	return s.leapsCountSample
}

func (s *Stat) GetTotalTimeMs() float64 {
	return Round(s.totalTimeNs)
}

func (s *Stat) GetLeapTimeMs() float64 {
	return s.leapTimeMs
}

func (s *Stat) GetMinTimeMs() float64 {
	return s.minTimeMs
}

func (s *Stat) GetMinTimeSampleMs() float64 {
	return s.minTimeSampleMs
}

func (s *Stat) GetMaxTimeMs() float64 {
	return s.maxTimeMs
}

func (s *Stat) GetMaxTimeSampleMs() float64 {
	return s.maxTimeSampleMs
}

func (s *Stat) GetAvgTimeMs() float64 {
	if s.leapsCount == 0 {
		return 0
	}
	return Round(s.totalTimeNs / s.leapsCount)
}

func (s *Stat) GetAvgTimeSampleMs() float64 {
	return s.avgTimeSampleMs
}

func (s *Stat) GetPeersCount() int64 {
	return s.peersCount.Load()
}

func (s *Stat) String() string {
	return fmt.Sprintf("%.2f %.2f %.2f", s.minTimeMs, s.GetAvgTimeMs(), s.maxTimeMs)
}
