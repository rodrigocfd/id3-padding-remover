package dlgrun

import (
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
	"github.com/rodrigocfd/windigo/win/com/shell"
	"github.com/rodrigocfd/windigo/win/com/shell/shellco"
)

type DlgRun struct {
	wnd     ui.WindowModal
	proRun  ui.ProgressBar
	taskbar shell.ITaskbarList4
	job     func()
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

func (me *DlgRun) Show(parent ui.AnyParent, job func()) {
	me.taskbar = shell.NewITaskbarList4(
		win.CoCreateInstance(
			shellco.CLSID_TaskbarList, nil,
			co.CLSCTX_INPROC_SERVER,
			shellco.IID_ITaskbarList4),
	)
	defer me.taskbar.Release()

	me.job = job
	defer func() { me.job = nil }()

	me.wnd.ShowModal(parent)
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
			me.job()
			me.wnd.RunUiThread(func() { // return to UI thread after job is finished
				me.taskbar.SetProgressState(hRootOwner, shellco.TBPF_NOPROGRESS)
				me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0)
			})
		}()
		return true
	})
}
