package prompt

import (
	"github.com/rodrigocfd/windigo/ui"
	"github.com/rodrigocfd/windigo/win"
	"github.com/rodrigocfd/windigo/win/co"
)

func Error(parent ui.AnyParent, title string, header win.StrOrNil, body string) {
	_Base(parent, title, header, body, co.TDCBF_OK, co.TD_ICON_ERROR)
}

func Info(parent ui.AnyParent, title string, header win.StrOrNil, body string) {
	_Base(parent, title, header, body, co.TDCBF_OK, co.TD_ICON_INFORMATION)
}

func OkCancel(parent ui.AnyParent, title string, header win.StrOrNil, body string) bool {
	return _Base(parent, title, header, body,
		co.TDCBF_OK|co.TDCBF_CANCEL, co.TD_ICON_WARNING) == co.ID_OK
}

func _Base(parent ui.AnyParent,
	title string, header win.StrOrNil, body string,
	btns co.TDCBF, ico co.TD_ICON) co.ID {

	tdc := win.TASKDIALOGCONFIG{}
	tdc.SetCbSize()
	if parent != nil {
		tdc.SetHwndParent(parent.Hwnd())
	}
	tdc.SetDwFlags(co.TDF_ALLOW_DIALOG_CANCELLATION)
	tdc.SetDwCommonButtons(btns)
	tdc.SetHMainIcon(win.TdcIconTdi(ico))
	tdc.SetPszWindowTitle(title)
	if header, ok := header.(win.StrVal); ok { // not nil?
		tdc.SetPszMainInstruction(string(header))
	}
	tdc.SetPszContent(body)

	return win.TaskDialogIndirect(&tdc)
}
