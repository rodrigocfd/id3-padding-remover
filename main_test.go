package main

import (
	"testing"

	"github.com/rodrigocfd/windigo/win"
)

// go test -bench=.

func BenchmarkThreadId(b *testing.B) {
	// m := make(map[int]int)
	for range b.N {
		win.GetCurrentThreadId()
	}
}
func BenchmarkProcId(b *testing.B) {
	for range b.N {
		win.GetCurrentProcessId()
	}
}
