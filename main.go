package main

import (
	"fmt"
	"id3fit/dlgmain"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
)

func main() {
	debug.SetGCPercent(50)
	runtime.LockOSThread()

	defer func() {
		if r := recover(); r != nil {
			ui.TaskDlg.Error(nil, "Panic", win.StrOptNone(),
				fmt.Sprintf("PANIC @ %v\n\n%v\n\n%s",
					time.Now(), r, string(debug.Stack())))
		}
	}()

	dlgmain.NewDlgMain().Run()
}
