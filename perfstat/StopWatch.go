package perfstat

import "time"

type StopWatch struct {
	start    time.Time
	Started  bool
	Duration time.Duration
}

func newStopWatch() *StopWatch {
	return &StopWatch{}
}

func (sw *StopWatch) Start() {
	sw.start = time.Now()
	sw.Started = true
}

func (sw *StopWatch) Stop() {
	if sw.Started {
		sw.Duration = time.Since(sw.start)
		sw.Started = false
	}
}

func (sw *StopWatch) Reset() {
	sw.Duration = 0
	sw.Started = false
}

func (sw *StopWatch) ElapsedNs() int64 {
	if sw.Started {
		return time.Since(sw.start).Nanoseconds()
	}
	return sw.Duration.Nanoseconds()
}
