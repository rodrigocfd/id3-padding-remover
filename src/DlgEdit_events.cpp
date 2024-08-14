#include "DlgEdit.h"
#include "../res/resource.h"
using std::span;

INT_PTR DlgEdit::dlgProc(UINT uMsg, WPARAM wp, LPARAM lp)
{
	switch (uMsg) {
		case WM_INITDIALOG: return onInitDialog();
		case WM_COMMAND:
			if (lib::vec::anyIf(span{_Fields}, [wp](auto&& f) { return f.chkId == LOWORD(wp); })) { // one of the checkboxes
				return onChk(wp);
			} else {
				switch LOWORD(wp) {
					case BTN_UNCHECK: return onBtnUncheck();
					case IDOK:        return onBtnOk();
					case IDCANCEL:    PostMessageW(hWnd(), WM_CLOSE, 0, 0); return TRUE;
					default:          return FALSE;
				}
			}
		case WM_CLOSE: EndDialog(hWnd(), 0); return TRUE;
		default:       return FALSE;
	}
}

INT_PTR DlgEdit::onInitDialog()
{
	lib::ComboBox{this, CMB_GENRE}.add(_Genres);

	_wndPic.create(this, {
		.x = lib::dpi::x(414),
		.y = lib::dpi::y(42),
		.cx = lib::dpi::cx(210),
		.cy = lib::dpi::cy(210),
		.style = WS_CHILD | WS_GROUP | WS_VISIBLE | WS_CLIPCHILDREN | WS_CLIPSIBLINGS,
	});

	lib::ListView lv{this, LST_FRAMES};
	lv.setFullRowSelect()
		.setGridLines()
		.columns.add({{L"Frame", 56}, {L"Value", 100}});

	_renderTitlebarCounts();
	_renderTextboxes();
	_renderFramesList();
	lv.columns[1].setWidthToFill();

	return TRUE;
}

INT_PTR DlgEdit::onChk(WPARAM wp)
{
	WORD chkId = LOWORD(wp);
	WORD txtId = chkId + 1;
	if (lib::CheckRadio{this, chkId}.isChecked()) { // when checked, enable textbox and focus it
		dlg.enable({txtId}, TRUE);
		lib::NativeControl{this, txtId}.focus();
	} else {
		dlg.enable({txtId}, FALSE);
	}
	return TRUE;
}

INT_PTR DlgEdit::onBtnUncheck()
{
	for (auto&& field : _Fields)
		lib::CheckRadio{this, field.chkId}.checkAndTrigger(false);

	return TRUE;
}

INT_PTR DlgEdit::onBtnOk()
{
	_updateTagsWithTexts();

	size_t numSaved = 0;
	auto t0 = lib::TimeCount::Immediately();

	for (auto&& pTag : _pTags) {
		try {
			pTag->saveToFile();
			++numSaved;
		} catch (const std::exception& e) {
			dlg.msgBox(L"Saving error", {},
				lib::str::fmt(L"Tag saving failed:\n%s\n\n%s", pTag->path, lib::str::toWide(e.what())),
				TDCBF_OK_BUTTON, TD_ERROR_ICON);
		}
	}

	auto t1 = t0.now();

	for (auto&& field : _Fields) { // disable all controls just for visual feedback purposes
		EnableWindow(GetDlgItem(hWnd(), field.chkId), FALSE);
		EnableWindow(GetDlgItem(hWnd(), field.chkId + 1), FALSE);
	}
	dlg.enable({BTN_UNCHECK, IDOK, IDCANCEL, LST_FRAMES}, false);

	dlg.msgBox(L"Tag(s) saved", {},
		lib::str::fmt(L"%d tag(s) saved in %02d.%03d.", numSaved, t1.sec, t1.ms),
		TDCBF_OK_BUTTON, TD_INFORMATION_ICON);

	EndDialog(hWnd(), 0);
	return TRUE;
}
