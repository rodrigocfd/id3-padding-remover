#include "DlgEdit.h"
#include "../res/resource.h"

INT_PTR DlgEdit::dlgProc(UINT uMsg, WPARAM wp, LPARAM lp)
{
	switch (uMsg) {
		case WM_INITDIALOG: return onInitDialog();
		case WM_COMMAND:
			switch LOWORD(wp) {
				case BTN_UNCHECK: return onBtnUncheck();
				case IDOK:        return onBtnOk();
				case IDCANCEL:    PostMessageW(hWnd(), WM_CLOSE, 0, 0); return TRUE;
				default:          return FALSE;
			}
		case WM_CLOSE: EndDialog(hWnd(), 0); return TRUE;
		default:       return FALSE;
	}
}

INT_PTR DlgEdit::onInitDialog()
{
	lib::ListView lv{this, LST_FRAMES};
	lv.setFullRowSelect()
		.setGridLines()
		.columns.add({{L"Frame", 56}, {L"Value", 100}});

	_writeTitlebarCounts();
	_writeFields();
	_renderFrames();
	lv.columns[1].setWidthToFill();

	return TRUE;
}

INT_PTR DlgEdit::onBtnUncheck()
{

	return TRUE;
}

INT_PTR DlgEdit::onBtnOk()
{
	for (auto&& pTag : _pTags) {
		pTag->sendApicToLast();

		//try {
		//	pTag->saveToFile();
		//} catch (const std::exception& e) {
		//	dlg.msgBox(L"Saving error", {},
		//		lib::str::fmt(L"Tag saving failed:\n%s\n\n%s", pTag->path, lib::str::toWide(e.what())),
		//		TDCBF_OK_BUTTON, TD_ERROR_ICON);
		//}
	}

	return TRUE;
}
