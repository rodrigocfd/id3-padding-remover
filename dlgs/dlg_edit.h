#pragma once
#include "../windlg/lib.hpp"
#include "bmp_loader.h"
#include "wnd_pic.h"
#include "../id3v2/tag.h"
#include "../res/resource.h"

class DlgEdit final : public wd::BaseDialog {
public:
	explicit DlgEdit(std::vector<id3v2::Tag> &tags) :
		BaseDialog{DLG_EDIT},
		wndPic{this, _mnuPic, _bmpLoader}, _tags{tags}, _ok{false} { }

	[[nodiscard]] constexpr bool ok() const { return _ok; }

private:
	bool on_init_dialog() override;
	void on_init_menu_popup(HMENU hMenu) override;
	bool on_command(WORD id, WORD code) override;
	bool on_notify(NMHDR &nm) override;

	void chk_click(WORD chkId) const;
	void txt_change(WORD txtId, bool isDropdownChange);
	void menu_frame_move(bool isUp);
	void menu_frame_delete();
	void menu_pic_insert_new();
	void menu_pic_export() const;
	void menu_pic_delete();
	void btn_uncheck() const;
	void btn_ok();
	void fill_textboxes_and_pic();
	void fill_frames_list() const;

	wd::ComboBox cmbGenre{this, CMB_GENRE};
	wd::Button btnUncheck{this, BTN_UNCHECK};
	wd::ListView lstFrames{this, LST_FRAMES};
	wd::Static lblPicSize{this, LBL_PICSIZE};
	WndPic wndPic;

	std::vector<id3v2::Tag> &_tags;
	wd::MenuResource _mnuFrames{}, _mnuPic{};
	BmpLoader _bmpLoader{};
	bool _ok; // to be returned from the modal
};
