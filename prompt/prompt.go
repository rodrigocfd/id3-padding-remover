package prompt

import (
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

func Error(parent ui.AnyParent, title string, header win.StrOpt, body string) {
	_Base(parent, title, header, body, co.TDCBF_OK, co.TD_ICON_ERROR)
}

func Info(parent ui.AnyParent, title string, header win.StrOpt, body string) {
	_Base(parent, title, header, body, co.TDCBF_OK, co.TD_ICON_INFORMATION)
}

func OkCancel(parent ui.AnyParent, title string, header win.StrOpt, body string) bool {
	return _Base(parent, title, header, body,
		co.TDCBF_OK|co.TDCBF_CANCEL, co.TD_ICON_WARNING) == co.ID_OK
}

func _Base(parent ui.AnyParent,
	title string, header win.StrOpt, body string,
	btns co.TDCBF, ico co.TD_ICON) co.ID {

	tdc := win.TASKDIALOGCONFIG{
		DwFlags:         co.TDF_ALLOW_DIALOG_CANCELLATION,
		DwCommonButtons: btns,
		PszWindowTitle:  title,
		HMainIcon:       win.TdcIconTdi(ico),
		PszContent:      body,
	}
	if parent != nil {
		tdc.HwndParent = parent.Hwnd()
	}
	if header, ok := header.(win.StrOptVal); ok { // not empty?
		tdc.PszMainInstruction = string(header)
	}

	return win.TaskDialogIndirect(&tdc)
}
