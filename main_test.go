package main

import (
	"testing"

	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/heap"
)

// go test -bench=.

func BenchmarkSyscall(b *testing.B) {
	// m := make(map[int]int)
	for range b.N {
		// win.GetCurrentProcessId()
		win.GetCurrentThreadId()
	}
}

//	func BenchmarkAllocOS(b *testing.B) {
//		for range b.N {
//			a := heap.NewWideStr[heap.Stack20]()
//			defer a.Free()
//			a.Set("123456", heap.ALLOW_EMPTY)
//		}
//	}
func BenchmarkAllocGC(b *testing.B) {
	for range b.N {
		a := heap.NewWideStr[heap.Stack20]()
		a.Set("123456", heap.ALLOW_EMPTY)
	}
}
