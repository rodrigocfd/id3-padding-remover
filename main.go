//go:build windows

package main

import (
	"id3fit/dlg/dlgmain"
	"runtime"

	"github.com/rodrigocfd/windigo/win/ole"
)

func main() {
	runtime.LockOSThread()

	// go dbgMem()

	ole.OleInitialize()
	defer ole.OleUninitialize()

	d := dlgmain.New()
	d.Run()
}

// func dbgMem() {
// 	lastFreed := uint64(0)
// 	var stats runtime.MemStats
// 	for {
// 		runtime.ReadMemStats(&stats)
// 		fmt.Printf("GC cycles: %d; Alloc: %s, Next GC: %s; Frees: %d (+%d)\n",
// 			stats.NumGC, win.Str.FmtBytes(stats.HeapAlloc),
// 			win.Str.FmtBytes(stats.NextGC), stats.Frees, stats.Frees-lastFreed)
// 		lastFreed = stats.Frees
// 		time.Sleep(time.Second * 2)
// 	}
// }
