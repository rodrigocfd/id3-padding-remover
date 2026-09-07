#include <algorithm>
#include "dlg_edit.h"
#include "dlg_edit_consts.h"

bool DlgEdit::on_init_dialog() {
	_mnuFrames.create_popup()
		.add(L"Move &up", MNU_FRAME_MOVEUP)
		.add(L"Move &down", MNU_FRAME_MOVEDOWN)
		.add_sep()
		.add(L"&Delete", MNU_FRAME_DELETE);
	_mnuPic.create_popup()
		.add(L"Insert &new...", MNU_PIC_INSERTNEW)
		.add_sep()
		.add(L"E&xport...", MNU_PIC_EXPORT)
		.add(L"&Delete", MNU_PIC_DELETE);

	lstFrames.activate_mods(_mnuFrames)
		.set_full_row_sel()
		.set_grid_lines()
		.col_add(L"Frame", wd::dpi::x(56))
		.col_add(L"Value", 1)
		.col(1).set_width_to_fill();

	cmbGenre.item_add(consts::GENRES);
	fill_textboxes_and_pic();
	fill_frames_list();

	wndPic.setup.border = true;
	wndPic.setup.bgColor = COLOR_WINDOW;
	wndPic.create(wd::dpi::pt(410, 40), wd::dpi::sz(220, 220));

	wd::Edit{this, TXT_ARTIST}.focus(); // first field
	return true;
}

void DlgEdit::on_init_menu_popup(HMENU hMenu) {
	if (hMenu == _mnuFrames.hmenu()) { // context menu of frames list view
		bool oneTag = _tags.size() == 1;
		bool hasSel = lstFrames.item_selected_count() > 0;
		bool firstIsSel = lstFrames.item(0).is_selected();
		bool lastIsSel = lstFrames.item(lstFrames.item_count() - 1).is_selected();

		_mnuFrames.enable_cmd(oneTag && hasSel && !firstIsSel, MNU_FRAME_MOVEUP);
		_mnuFrames.enable_cmd(oneTag && hasSel && !lastIsSel, MNU_FRAME_MOVEDOWN);
		_mnuFrames.enable_cmd(oneTag && hasSel, MNU_FRAME_DELETE);

	} else if (hMenu == _mnuPic.hmenu()) { // context menu of picture render window
		bool hasUniqueApic = id3v2::same_frame_across_all_tags(L"APIC", _tags) != nullptr;
		_mnuPic.enable_cmd(hasUniqueApic, MNU_PIC_EXPORT);

		bool atLeastOneApic = std::any_of(_tags.begin(), _tags.end(), [](const id3v2::Tag &tag) -> bool {
			return tag.frame_by_name4(L"APIC") != nullptr;
		});
		_mnuPic.enable_cmd(atLeastOneApic, MNU_PIC_DELETE);
	}
}

bool DlgEdit::on_command(WORD id, WORD code) {
	switch (id) {
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
		case CHK_COMMENT: 
			switch (code) {
				case BN_CLICKED: chk_click(id); return true;
			}
			break;
		case TXT_ARTIST:
		case TXT_TITLE:
		case TXT_SUBTITLE:
		case TXT_ALBUM:
		case TXT_TRACK:
		case TXT_YEAR:
		//case CMB_GENRE:
		case TXT_PERFORMER:
		case TXT_PUBLISHER:
		case TXT_OARTIST:
		case TXT_OALBUM:
		case TXT_OYEAR:
		case TXT_COMPOSER:
		case TXT_LYRICIST:
		case TXT_COMMENT:
			switch (code) {
				case EN_CHANGE: txt_change(id, false); return true;
			}
			break;
		case CMB_GENRE:
			switch (code) {
				case CBN_SELCHANGE:  txt_change(id, true); return true; // use chose genre from dropdown
				case CBN_EDITCHANGE: txt_change(id, false); return true;
			}
			break;
		case MNU_FRAME_MOVEUP:   menu_frame_move(true); return true;
		case MNU_FRAME_MOVEDOWN: menu_frame_move(false); return true;
		case MNU_FRAME_DELETE:   menu_frame_delete(); return true;
		case MNU_PIC_INSERTNEW:  menu_pic_insert_new(); return true;
		case MNU_PIC_EXPORT:     menu_pic_export(); return true;
		case MNU_PIC_DELETE:     menu_pic_delete(); return true;

		case BTN_UNCHECK: btn_uncheck(); return true;
		case IDOK:        btn_ok(); return true;
		case IDCANCEL:    PostMessageW(hwnd(), WM_CLOSE, 0, 0); return true; // close on ESC
	}
	return false;
}

bool DlgEdit::on_notify(NMHDR &nm) {
	switch (nm.idFrom) {
		case LST_FRAMES:
			switch (nm.code) {
				case LVN_KEYDOWN:
					switch (reinterpret_cast<const NMLVKEYDOWN&>(nm).wVKey) {
						case VK_DELETE: menu_frame_delete(); return true;
					}
					break;
			}
			break;
	}
	return false;
}
