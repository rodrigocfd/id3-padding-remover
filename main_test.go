package main

import (
	"testing"

	"github.com/rodrigocfd/windigo/win"
)

// go test -bench=.

func BenchmarkSyscall(b *testing.B) {
	// m := make(map[int]int)
	for range b.N {
		win.GetCurrentProcessId()
		// win.GetCurrentThreadId()
	}
}
func BenchmarkSyscall2(b *testing.B) {
	for range b.N {
		win.GetCurrentProcessId()
	}
}
