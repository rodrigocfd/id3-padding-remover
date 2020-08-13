package main

import (
	"fmt"
	"runtime"
)

// Prints runtime memory usage.
func mem() {
	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc %.2f, TotAlloc %.2f, Sys %.2f, GcCycles %d\n",
		float32(m.Alloc)/1024/1024, float32(m.TotalAlloc)/1024/1024,
		float32(m.Sys)/1024/1024, m.NumGC)
}
