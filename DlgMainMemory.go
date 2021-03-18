package main

import (
	"fmt"
	"runtime"

	"github.com/rodrigocfd/windigo/win"
)

var memStats runtime.MemStats

func (me *DlgMain) updateMemoryStatus() {
	me.wnd.Hwnd().SetTimer(1, 1000, func(msElapsed uint32) {
		runtime.ReadMemStats(&memStats)
		me.statusBar.Parts().SetAllTexts(
			fmt.Sprintf("Alloc: %s", win.Str.FmtBytes(memStats.Alloc)),
			fmt.Sprintf("Accum alloc: %s", win.Str.FmtBytes(memStats.TotalAlloc)),
			fmt.Sprintf("Obtained: %s", win.Str.FmtBytes(memStats.Sys)),
			fmt.Sprintf("GC cycles: %d", memStats.NumGC),
		)
	})
}
