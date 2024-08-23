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

	if (_pTags.size() == 1)
		_reorderedFrames = _pTags[0]->frames; // copy all frames into the back-buffer

	renderTitlebarCounts();
	renderTextboxes();
	loadPicture();
	renderFramesList();
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
		dlg.enable({txtId});
		lib::NativeControl{this, txtId}.focus();
	} else {
		dlg.enable({txtId}, false);
	}
	return TRUE;
}

INT_PTR DlgEdit::onMnuFrameMove(bool isUp)
{
	lib::ListView lv{this, LST_FRAMES};
	auto selItems = lv.items.selected();
	auto focused = lv.items.focused();
	
	if (isUp) {
		for (auto&& item : selItems) {
			std::iter_swap(_reorderedFrames.begin() + item.index() - 1,
				_reorderedFrames.begin() + item.index());
		}
	} else {
		for (auto it = selItems.rbegin(); it != selItems.rend(); ++it) {
			std::iter_swap(_reorderedFrames.begin() + it->index(),
				_reorderedFrames.begin() + it->index() + 1);
		}
	}

	renderFramesList();
	for (auto&& item : selItems)
		lv.items[item.index() + (isUp ? -1 : 1)].select(); // re-select the moved items
	if (focused.has_value())
		lv.items[focused.value().index() + (isUp ? -1 : 1)].focus();

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
	if (_pTags.size() == 1)
		_pTags[0]->frames = std::move(_reorderedFrames);

	updateTagsWithTexts(); // the file saving itself is made by DlgMain
	
	for (auto&& pTag : _pTags)
		pTag->mp3Offset = 0; // when saved, padding is zeroed

	_clickedOk = true;
	EndDialog(hWnd(), 0);
	return TRUE;
}
