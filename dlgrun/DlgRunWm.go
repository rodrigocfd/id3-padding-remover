package dlgrun

import (
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/ui/wm"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
	"github.com/rodrigocfd/windigo/win/com/shell/shellco"
)

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
					ui.TaskDlg.Error(me.wnd, "Error", win.StrOptSome("Errors found"), text)
				}
				me.taskbar.SetProgressState(hRootOwner, shellco.TBPF_NOPROGRESS)
				me.wnd.Hwnd().SendMessage(co.WM_CLOSE, 0, 0)
			})
		}()

		return true
	})
}
