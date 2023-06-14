package dlgrun

import (
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win/com/com"
	"github.com/rodrigocfd/windigo/win/com/com/comco"
	"github.com/rodrigocfd/windigo/win/com/shell"
	"github.com/rodrigocfd/windigo/win/com/shell/shellco"
)

// Displays the marquee progress bar while running a job in background.
type DlgRun struct {
	wnd     ui.WindowModal
	proRun  ui.ProgressBar
	taskbar shell.ITaskbarList4
	job     func() []error // Job to run in a parallel thread.
	errors  []error        // Errors returned by the job.
}

func NewDlgRun() *DlgRun {
	wnd := ui.NewWindowModalDlg(DLG_RUN)

	me := &DlgRun{
		wnd:    wnd,
		proRun: ui.NewProgressBarDlg(wnd, PRO_RUN, ui.HORZ_NONE, ui.VERT_NONE),
	}

	me.events()
	return me
}

func (me *DlgRun) Show(parent ui.AnyParent, job func() []error) bool {
	me.taskbar = shell.NewITaskbarList4(
		com.CoCreateInstance(
			shellco.CLSID_TaskbarList, nil,
			comco.CLSCTX_INPROC_SERVER,
			shellco.IID_ITaskbarList4),
	)
	defer me.taskbar.Release()

	me.job = job
	defer func() { me.job = nil }()

	me.wnd.ShowModal(parent)
	return len(me.errors) == 0 // no errors? all good
}
