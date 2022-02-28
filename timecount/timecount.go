package timecount

import (
	"github.com/rodrigocfd/windigo/win"
)

var frequency float64 = 0 // global, retrieved only once

// High-resolution elapsed time counter.
type TimeCount struct {
	t0 float64
}

// Creates a new high-performance elapsed time counter.
func New() TimeCount {
	if frequency == 0 { // not cached yet?
		frequency = float64(win.QueryPerformanceFrequency())
	}

	return TimeCount{
		t0: float64(win.QueryPerformanceCounter()),
	}
}

// Returns how many milliseconds ellapsed since the TimeCount creation.
func (me *TimeCount) ElapsedMs() float64 {
	tFin := float64(win.QueryPerformanceCounter())
	return ((tFin - me.t0) / frequency) * 1000
}
