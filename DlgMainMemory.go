package main

import (
	"fmt"
	"runtime"

	"github.com/rodrigocfd/windigo/win"
)

var memStats runtime.MemStats
var prevAlloc int64 = 0

func (me *DlgMain) updateMemoryStatus() {
	me.wnd.Hwnd().SetTimer(1, 1000, func(msElapsed uint32) {
		runtime.ReadMemStats(&memStats)

		dummy := make([]byte, 10*1024*1024)
		println("dummy", len(dummy))

		allocDiff := int64(memStats.Alloc) - prevAlloc
		if allocDiff < 0 {
			allocDiff = 0
		}

		me.statusBar.Parts().SetAllTexts(
			fmt.Sprintf("Alloc: %s (+%s)",
				win.Str.FmtBytes(memStats.Alloc),
				win.Str.FmtBytes(uint64(allocDiff))),
			fmt.Sprintf("GC cycles: %d", memStats.NumGC),
			fmt.Sprintf("Next GC: %s", win.Str.FmtBytes(memStats.NextGC)),
			fmt.Sprintf("Heap sys: %s", win.Str.FmtBytes(memStats.HeapSys)),
		)

		prevAlloc = int64(memStats.Alloc)
	})
}
