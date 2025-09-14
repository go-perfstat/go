package perfstat_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/go-perfstat/go/perfstat"
	"github.com/stretchr/testify/assert"
)

func init() {
	perfstat.SetDefaultAggregationPeriodMs(1000)
}

func TestPerfStatBasic(t *testing.T) {
	perf := perfstat.ForName("basic")

	perf.Start()
	for i := 0; i < 1000; i++ {
		perf.Start()
		time.Sleep(time.Millisecond)
		perf.Stop()
	}

	stat := perf.GetStat()

	// previous aggregation period
	assert.GreaterOrEqual(t, stat.GetMinTimeSampleMs(), float64(1))
	assert.Greater(t, stat.GetAvgTimeSampleMs(), float64(1))
	assert.Greater(t, stat.GetMaxTimeSampleMs(), float64(1))

	// grand total
	assert.GreaterOrEqual(t, stat.GetMinTimeMs(), float64(1))
	assert.Greater(t, stat.GetAvgTimeMs(), float64(1))
	assert.Greater(t, stat.GetMaxTimeMs(), float64(1))
	assert.Greater(t, stat.GetTotalTimeMs(), float64(1000))
	assert.Equal(t, stat.GetLeapsCount(), int64(1000))
	assert.Equal(t, int64(1), stat.GetFailedLeapsCount())
}

func TestPerfStatConcurrent(t *testing.T) {
	goroutines := 50
	iterations := 200

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			perf := perfstat.ForName("concurrent")
			for j := 0; j < iterations; j++ {
				perf.Start()
				time.Sleep(1 * time.Millisecond)
				perf.Stop()
			}
		}()
	}
	wg.Wait()

	stat := perfstat.GetByName("concurrent")
	assert.Equal(t, int64(goroutines*iterations), stat.GetLeapsCount())
	assert.Equal(t, int64(goroutines), stat.GetPeersCount())
}

func TestPerfStatPerformance(t *testing.T) {
	t.Skip("for manual run")
	iterations := 10_000_000

	benchmark := perfstat.ForName("benchmark")
	perf := perfstat.ForName("benchmark_test")

	benchmark.Start()
	for i := 0; i < iterations; i++ {
		perf.Start()
		perf.Stop()
	}
	totalNs := benchmark.Stop()
	fmt.Printf("Average leap time: %d ns\n", totalNs/int64(iterations)) // ~ 220 ns
}
