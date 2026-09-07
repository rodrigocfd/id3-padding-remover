#pragma once
#include "../windlg/lib.hpp"
#include "../res/resource.h"

class DlgMain final : public wd::BaseDialog {
public:
	DlgMain() : BaseDialog{DLG_MAIN, ICO_FOULBACHELOR} { }

private:
	bool on_init_dialog() override;
	void on_size(WORD req, SIZE sz) override;
	void on_init_menu_popup(HMENU hMenu) override;
	bool on_command(WORD id, WORD code) override;
	bool on_notify(NMHDR &nm) override;

	bool on_drag_files(const std::vector<std::wstring> &files) const;
	void on_drop_files(const std::vector<std::wstring> &files) const;
	void list_delete_item(const NMHDR &nm) const;
	void list_header_lick(const NMHDR &nm);
	void menu_file_open() const;
	void menu_file_edit();
	void menu_file_resave();
	void menu_del_pic(bool delRg);

	void update_titlebar_num_files(size_t numFiles) const;
	void sort_list() const;
	void add_mp3s_to_list(const std::vector<std::wstring> &files) const;
	void add_one_mp3_to_list(wd::StrView filePath) const;
	void render_mp3_in_list(wd::ListView::Item item) const;

	wd::ListView lstFiles{this, LST_FILES};
	wd::Layout _layout{this, 1};
	wd::DropTarget _lstFilesDropTarget{};
	wd::MenuResource _mnuFiles{};
	wd::ImageList _imgList{};
	int _curSortCol = 0;
	wd::ListView::Sort _curSortDir = wd::ListView::Sort::asc_i;
};
