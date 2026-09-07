#include <crtdbg.h>
#include "dlg_main.h"
#include "../res/resource.h"

int APIENTRY wWinMain(_In_ HINSTANCE hInst, _In_opt_ HINSTANCE, _In_ LPWSTR, _In_ int cmdShow) {
	int ret = 0;
	{
		auto ole = wd::OleInit{};
		DlgMain dlg{};
		ret = wd::run_main_dialog(hInst, cmdShow, dlg, {
			{wd::Acc::Key::ctrl, 'O', MNU_FILE_OPEN},
			{wd::Acc::Key::none, VK_F1, MNU_FILE_ABOUT},
		});
	}
	if (_CrtDumpMemoryLeaks())
		MessageBoxW(nullptr, L"A memory leak was found.", L"Memory leak", MB_ICONERROR);
	return ret;
}

bool DlgMain::on_init_dialog() {
	_imgList.create16()
		.add_file_ext(L"mp3");

	_mnuFiles.create_popup()
		.add(L"&Open files...\tCtrl+O", MNU_FILE_OPEN)
		.add(L"&Edit...", MNU_FILE_EDIT)
		.add(L"&Remove\tDel", MNU_FILE_REMOVE)
		.add_sep()
		.add(L"Re-&save", MNU_FILE_RESAVE)
		.add(L"Delete &pic", MNU_FILE_DELPIC)
		.add(L"Delete pic and R&G", MNU_FILE_DELPICRG)
		.add_sep()
		.add(L"&About...\tF1", MNU_FILE_ABOUT);

	lstFiles.activate_mods(_mnuFiles)
		.set_full_row_sel()
		.set_image_list16(_imgList)
		.col_add(L"File", 1)
		.col_add(L"Pad", wd::dpi::x(50), wd::ListView::Align::right)
		.col_add(L"Pic", wd::dpi::x(30), wd::ListView::Align::center)
		.col_add(L"RG", wd::dpi::x(30), wd::ListView::Align::center)
		.col_add(L"Artist", wd::dpi::x(90))
		.col_add(L"Year", wd::dpi::x(40), wd::ListView::Align::center)
		.col_add(L"Album", wd::dpi::x(100))
		.col_add(L"T#", wd::dpi::x(30), wd::ListView::Align::right)
		.col_add(L"Title", wd::dpi::x(100))
		.col_add(L"Genre", wd::dpi::x(90))
		.col_add(L"Performer", wd::dpi::x(70))
		.col_add(L"Composer", wd::dpi::x(70))
		.col_add(L"Lyricist", wd::dpi::x(70))
		.col_add(L"Orig. artist", wd::dpi::x(70))
		.col_add(L"Comment", wd::dpi::x(70))
		.col(0).set_width_to_fill();
	sort_list();

	_layout.add(LST_FILES, wd::Lay::resz_resz);

	_lstFilesDropTarget.register_drag_drop(lstFiles)
		.on_drag_enter(std::bind(&DlgMain::on_drag_files, this, std::placeholders::_1))
		.on_drop(std::bind(&DlgMain::on_drop_files, this, std::placeholders::_1));

	return true;
}

void DlgMain::on_size(WORD req, SIZE sz) {
	if (req != SIZE_MINIMIZED) [[likely]] {
		_layout.rearrange(req, sz);
		lstFiles.col(0).set_width_to_fill();
	}
}

void DlgMain::on_init_menu_popup(HMENU hMenu) {
	if (hMenu == _mnuFiles.hmenu()) [[likely]] {
		_mnuFiles.set_default_cmd(MNU_FILE_EDIT);
		_mnuFiles.enable_cmds(lstFiles.item_selected_count() > 0,
			{MNU_FILE_EDIT, MNU_FILE_RESAVE, MNU_FILE_REMOVE, MNU_FILE_DELPIC, MNU_FILE_DELPICRG});
	}
}

bool DlgMain::on_command(WORD id, WORD) {
	switch (id) {
		case MNU_FILE_OPEN:     menu_file_open(); return true;
		case MNU_FILE_EDIT:     menu_file_edit(); return true;
		case MNU_FILE_REMOVE:   lstFiles.item_del_selected(); return true;
		case MNU_FILE_RESAVE:   menu_file_resave(); return true;
		case MNU_FILE_DELPIC:   menu_del_pic(false); return true;
		case MNU_FILE_DELPICRG: menu_del_pic(true); return true;
		case MNU_FILE_ABOUT:    wd::sys_dlg::msg_about(this, ICO_FOULBACHELOR); return true;
	}
	return false;
}

bool DlgMain::on_notify(NMHDR &nm) {
	switch (nm.idFrom) {
		case LST_FILES:
			switch (nm.code) {
				case NM_DBLCLK:       menu_file_edit(); return true;
				case LVN_ITEMCHANGED: update_titlebar_num_files(lstFiles.item_count()); return true;
				case LVN_DELETEITEM:  list_delete_item(nm); return true;
				case LVN_KEYDOWN:
					switch (reinterpret_cast<const NMLVKEYDOWN&>(nm).wVKey) {
						case VK_RETURN: menu_file_edit(); return true;
						case VK_DELETE: lstFiles.item_del_selected(); return true;
					}
					break;
				case HDN_ITEMCLICK: list_header_lick(nm); return true;
			}
			break;
	}
	return false;
}
