#include <algorithm>
#include "DlgEdit.h"
#include "../res/resource.h"
using std::span;

INT_PTR DlgEdit::dlgProc(UINT uMsg, WPARAM wp, LPARAM lp)
{
	lib::ListView::ProcessMessages(this, LST_FRAMES, uMsg, wp, lp, MNU_FRAME);

	switch (uMsg) {
		case WM_INITDIALOG:    return onInitDialog();
		case WM_INITMENUPOPUP: return onInitMenuPopup(wp);
		case WM_COMMAND:
			switch LOWORD(wp) {
				case CHK_ARTIST:
				case CHK_TITLE:
				case CHK_SUBTITLE:
				case CHK_ALBUM:
				case CHK_TRACK:
				case CHK_YEAR:
				case CHK_GENRE:
				case CHK_PERFORMER:
				case CHK_PUBLISHER:
				case CHK_OARTIST:
				case CHK_OALBUM:
				case CHK_OYEAR:
				case CHK_COMPOSER:
				case CHK_LYRICIST:
				case CHK_COMMENT:        return onChk(wp);
				case MNU_FRAME_MOVEUP:   return onMnuFrameMove(true);
				case MNU_FRAME_MOVEDOWN: return onMnuFrameMove(false);
				case BTN_UNCHECK:        return onBtnUncheck();
				case IDOK:               return onBtnOk();
				case IDCANCEL:           PostMessageW(hWnd(), WM_CLOSE, 0, 0); return TRUE;
				default:                 return FALSE;
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

INT_PTR DlgEdit::onInitMenuPopup(WPARAM wp)
{
	lib::Menu popupMenu{reinterpret_cast<HMENU>(wp)};
	if (popupMenu.idByPos(0) == MNU_FRAME_MOVEUP) {
		lib::ListView lv{this, LST_FRAMES};
		UINT numSel = lv.items.countSelected();
		popupMenu.enableItemsByCmd({MNU_FRAME_MOVEUP}, numSel > 0 && !lv.items[0].isSelected());
		popupMenu.enableItemsByCmd({MNU_FRAME_MOVEDOWN}, numSel > 0 && !lv.items[lv.items.count() - 1].isSelected());
	}
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

INT_PTR DlgEdit::onMnuFrameMove(bool isUp)
{
	lib::ListView lv{this, LST_FRAMES};
	auto selItems = lv.items.selected();
	auto focused = lv.items.focused();
	int adjust = isUp ? -1 : 1;

	for (auto&& item : selItems) {
		std::iter_swap(_pTags[0]->frames.begin() + item.index(), // frames are shown only with 1 tag 
			_pTags[0]->frames.begin() + item.index() + adjust);
	}

	_renderFramesList();
	for (auto&& item : selItems)
		lv.items[item.index() + adjust].select(); // re-select the moved items
	if (focused.has_value())
		lv.items[focused.value().index() + adjust].focus();

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
	_updateTagsWithTexts(); // the file saving itself is made by DlgMain
	_clickedOk = true;
	EndDialog(hWnd(), 0);
	return TRUE;
}
