package timecount

import (
	"github.com/rodrigocfd/windigo/win"
)

// High-resolution elapsed time counter.
type TimeCount interface {
	// Returns the number of milliseconds elapsed since the timer started.
	ElapsedMs() float64
}

//------------------------------------------------------------------------------

var frequency float64 = 0

type _TimeCount struct {
	t0 float64
}

// Creates a new high-performance elapsed time counter.
func New() TimeCount {
	if frequency == 0 { // not cached yet?
		frequency = float64(win.QueryPerformanceFrequency())
	}

	return &_TimeCount{
		t0: float64(win.QueryPerformanceCounter()),
	}
}

func (me *_TimeCount) ElapsedMs() float64 {
	tFin := float64(win.QueryPerformanceCounter())
	return ((tFin - me.t0) / frequency) * 1000
}
