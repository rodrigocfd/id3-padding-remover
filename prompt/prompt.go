package prompt

import (
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

func Error(parent ui.AnyParent, title, body string) {
	base(parent, title, body, co.TDCBF_OK, co.TD_ICON_ERROR)
}

func Info(parent ui.AnyParent, title, body string) {
	base(parent, title, body, co.TDCBF_OK, co.TD_ICON_INFORMATION)
}

func OkCancel(parent ui.AnyParent, title, body string) co.ID {
	return base(parent, title, body,
		co.TDCBF_OK|co.TDCBF_CANCEL, co.TD_ICON_WARNING)
}

func base(parent ui.AnyParent,
	title, body string, btns co.TDCBF, ico co.TD_ICON) co.ID {

	tdc := win.TASKDIALOGCONFIG{}
	tdc.SetCbSize()
	*tdc.HwndParent() = parent.Hwnd()
	*tdc.DwFlags() = co.TDF_ALLOW_DIALOG_CANCELLATION
	*tdc.DwCommonButtons() = btns
	*tdc.HMainIcon() = ico
	*tdc.PszWindowTitle() = win.Str.ToUint16Ptr(title)
	*tdc.PszContent() = win.Str.ToUint16Ptr(body)

	return win.TaskDialogIndirect(&tdc)
}
