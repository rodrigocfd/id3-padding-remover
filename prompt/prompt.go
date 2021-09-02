package prompt

import (
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

func Error(parent ui.AnyParent, title, header, body string) {
	_Base(parent, title, header, body, co.TDCBF_OK, co.TD_ICON_ERROR)
}

func Info(parent ui.AnyParent, title, header, body string) {
	_Base(parent, title, header, body, co.TDCBF_OK, co.TD_ICON_INFORMATION)
}

func OkCancel(parent ui.AnyParent, title, header, body string) co.ID {
	return _Base(parent, title, header, body,
		co.TDCBF_OK|co.TDCBF_CANCEL, co.TD_ICON_WARNING)
}

func _Base(parent ui.AnyParent,
	title, header, body string,
	btns co.TDCBF, ico co.TD_ICON) co.ID {

	var tdc win.TASKDIALOGCONFIG
	tdc.SetCbSize()
	tdc.SetHwndParent(parent.Hwnd())
	tdc.SetDwFlags(co.TDF_ALLOW_DIALOG_CANCELLATION)
	tdc.SetDwCommonButtons(btns)
	tdc.SetHMainIcon(ico)
	tdc.SetPszWindowTitle(title)
	if header != "" {
		tdc.SetPszMainInstruction(header)
	}
	tdc.SetPszContent(body)

	return win.TaskDialogIndirect(&tdc)
}
