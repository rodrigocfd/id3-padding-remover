package dlgrun

import (
	"id3fit/prompt"

	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
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
	job     func() []error
	errors  []error
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

func (me *DlgRun) events() {
	me.wnd.On().WmInitDialog(func(_ wm.InitDialog) bool {
		if me.taskbar.Ptr() == nil {
			panic("DlgRun modal cannot be reused, create another one.")
		}

		hRootOwner := me.wnd.Hwnd().GetWindow(co.GW_OWNER)
		me.proRun.SetMarquee(true)
		me.taskbar.SetProgressState(hRootOwner, shellco.TBPF_INDETERMINATE)

		go func() { // launch another thread for the job
			me.errors = me.job()
			me.wnd.RunUiThread(func() { // return to UI thread after job is finished
				if len(me.errors) > 0 {
					text := ""
					for _, err := range me.errors { // show errors of all files
						text += err.Error() + "\n\n"
					}
					text = text[:len(text)-2]
					prompt.Error(me.wnd, "Error", win.StrOptSome("Errors found"), text)
				}
				me.taskbar.SetProgressState(hRootOwner, shellco.TBPF_NOPROGRESS)
				me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0)
			})
		}()
		return true
	})
}
