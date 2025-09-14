package perfstat

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
)

type Stat struct {
	mu                 sync.Mutex
	typ                string
	name               string
	leapsCount         int64
	leapsCountSample   int64
	totalTimeNs        int64
	totalTimeSampleNs  int64
	avgTimeSampleMs    float64
	leapTimeMs         float64
	minTimeMs          float64
	minTimeSampleMs    float64
	minTimeThresholdMs float64
	maxTimeMs          float64
	maxTimeSampleMs    float64
	maxTimeThresholdMs float64
	peersCount         atomic.Int64
	// fields for approximate percentiles
	binCounts map[int64]int64 // ms -> count
	binMinMs  int64
	binMaxMs  int64
}

func newStat(typ, name string) *Stat {
	return &Stat{
		typ:       typ,
		name:      name,
		binCounts: make(map[int64]int64),
	}
}

func (s *Stat) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.totalTimeNs = 0
	s.totalTimeSampleNs = 0
	s.maxTimeMs = 0
	s.maxTimeSampleMs = 0
	s.maxTimeThresholdMs = 0
	s.minTimeMs = 0
	s.minTimeSampleMs = 0
	s.minTimeThresholdMs = 0
	s.leapsCount = 0
	s.leapsCountSample = 0
	s.binCounts = make(map[int64]int64)
	s.binMinMs = 0
	s.binMaxMs = 0
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

// Approximate percentiles
func (s *Stat) Percentiles() (p50, p90, p99 float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	total := int64(0)
	for _, c := range s.binCounts {
		total += c
	}
	if total == 0 {
		return 0, 0, 0
	}

	percentile := func(q float64) float64 {
		target := int64(math.Ceil(q * float64(total)))
		count := int64(0)
		for i := s.binMinMs; i <= s.binMaxMs; i++ {
			count += s.binCounts[i]
			if count >= target {
				return float64(i)
			}
		}
		return float64(s.binMaxMs)
	}

	return percentile(0.50), percentile(0.90), percentile(0.99)
}

func (s *Stat) String() string {
	return fmt.Sprintf("%.2f %.2f %.2f", s.minTimeMs, s.GetAvgTimeMs(), s.maxTimeMs)
}

func (s *Stat) addSampleMs(timeMs float64) {
	bin := int64(math.Floor(timeMs))
	s.binCounts[bin]++
	if len(s.binCounts) == 1 {
		s.binMinMs = bin
		s.binMaxMs = bin
	} else {
		if bin < s.binMinMs {
			s.binMinMs = bin
		}
		if bin > s.binMaxMs {
			s.binMaxMs = bin
		}
	}
}
