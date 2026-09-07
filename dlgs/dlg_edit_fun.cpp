#include <algorithm>
#include <ranges>
#include "dlg_edit.h"
#include "dlg_edit_consts.h"

void DlgEdit::chk_click(WORD chkId) const {
	const wd::Edit txt{this, static_cast<WORD>(chkId + 1)};
	if (wd::CheckBox{this, chkId}.is_checked()) {
		txt.enable(true).focus();
	} else {
		txt.enable(false);
	}
}

void DlgEdit::txt_change(WORD txtId, bool isDropdownChange) {
	if (!lstFrames.item_count())
		return; // prevents running before WM_INITDIALOG, because EN_CHANGE will hit for every field

	const std::wstring text = isDropdownChange
		? cmbGenre.item_selected().text()
		: wd::Edit{this, txtId}.text();

	const wd::StrView name4 = consts::FIELDS_NAME4[(txtId - TXT_ARTIST) / 2];
	for (auto &&tag : _tags) {
		if (id3v2::Frame *f = tag.frame_by_name4(name4); f) { // such frame already exists
			f->set_simple_text(text);
		} else { // non-existing frame, create new
			auto newF = id3v2::Frame{name4, text};
			tag.frames().emplace_back(std::move(newF));
		}
	}
	fill_frames_list(); // update the text in the frames listview
}

void DlgEdit::menu_frame_move(bool isUp) {
	const wd::ListView::Item focusedItem = lstFrames.item_focused();
	const std::vector<wd::ListView::Item> selItems = lstFrames.item_selected();
	std::vector<id3v2::Frame> &frames = _tags[0].frames();

	if (isUp) {
		for (auto &&selItem : selItems)
			std::swap(frames[selItem.index()], frames[selItem.index() - 1]); // move frames within the tag
	} else {
		for (auto &&selItem : std::views::reverse(selItems))
			std::swap(frames[selItem.index()], frames[selItem.index() + 1]);
	}
	fill_frames_list();

	for (auto &&selItem : selItems)
		lstFrames.item(selItem.index() + (isUp ? -1 : 1)).select(true); // adjust the listview items selection
	if (focusedItem.valid())
		lstFrames.item(focusedItem.index() + (isUp ? -1 : 1)).focus();
}

void DlgEdit::menu_frame_delete() {
	const std::wstring msg = wd::str::fmt(L"Do you want to remove %u frame(s)?", lstFrames.item_selected_count());
	if (!wd::sys_dlg::msg_ask(this, L"Remove frames", msg, L"&Remove")) [[unlikely]] {
		return; // user cancelled
	}

	std::vector<id3v2::Frame> &frames = _tags[0].frames(); // we're here solely with 1 tag being edited
	for (auto &&selItem : std::views::reverse(lstFrames.item_selected()))
		frames.erase(frames.begin() + selItem.index());

	lstFrames.item_del_all(); // necessary because fill_textboxes_and_pic() will trigger EN_CHANGE in all textboxes
	fill_textboxes_and_pic();
	fill_frames_list();
}

void DlgEdit::menu_pic_insert_new() {
	const std::wstring imagePath = wd::sys_dlg::file_open(this, {
		{L"JPG images", L"*.jpg;*.jpeg"},
		{L"PNG images", L"*.png"},
		{L"All images", L"*.jpg;*.jpeg;*.png"},
	});
	if (imagePath.empty())
		return; // user cancelled

	std::vector<BYTE> imageBytes{};
	try {
		imageBytes = wd::file::read(imagePath);
	} catch (const wd::WinErr &e) {
		wd::sys_dlg::msg_err(this, L"Loading image", wd::str::fmt(L"Reading image file failed:\n%s", imagePath), e);
		return;
	}

	std::wstring mimeType{};
	wd::StrView vwImagePath{imagePath};
	if (vwImagePath.ends_with_i(L".jpg") || vwImagePath.ends_with_i(L".jpeg")) {
		mimeType = L"image/jpeg";
	} else if (vwImagePath.ends_with_i(L".png")) {
		mimeType = L"image/png";
	} else {
		wd::sys_dlg::msg_err(this, L"Bad image type", wd::str::fmt(L"Images must be JPG or PNG only.\n\nInvalid:%s", imagePath));
		return;
	}

	for (auto &&tag : _tags) {
		if (id3v2::Frame *pFrame = tag.frame_by_name4(L"APIC"); pFrame) // already has APIC
			tag.delete_frames_by_name4(L"APIC"); // simply discard it

		tag.frames().emplace_back(imageBytes, mimeType);
	}

	lstFrames.item_del_all(); // necessary because fill_textboxes_and_pic() will trigger EN_CHANGE in all textboxes
	fill_textboxes_and_pic();
	fill_frames_list();
}

void DlgEdit::menu_pic_export() const {
	if (const id3v2::BodyPicture *pApic = id3v2::same_apic_frame_across_all_tags(_tags); pApic) {
		std::wstring savePath{};
		if (pApic->mime == L"image/jpeg") {
			savePath = wd::sys_dlg::file_save(this, {
				{L"JPG images", L"*.jpg"},
				{L"All files", L"*.*"},
				});
		} else if (pApic->mime == L"image/png") {
			savePath = wd::sys_dlg::file_save(this, {
				{L"PNG images", L"*.png"},
				{L"All files", L"*.*"},
				});
		} else {
			savePath = wd::sys_dlg::file_save(this, {{L"All files", L"*.*"}});
		}
		if (!savePath.empty()) {
			try {
				wd::file::write(savePath, pApic->bin);
			} catch (const wd::WinErr &e) {
				wd::sys_dlg::msg_err(this, L"Exporting image", wd::str::fmt(L"Exporting image failed:\n%s", savePath), e);
			}			
		}
	}
}

void DlgEdit::menu_pic_delete() {
	size_t numMp3s = std::count_if(_tags.begin(), _tags.end(), [](const id3v2::Tag &tag) -> bool {
		return tag.frame_by_name4(L"APIC") != nullptr;
	});
	if (!wd::sys_dlg::msg_ask(this, L"Delete picture", wd::str::fmt(L"Delete picture from %u file(s)?", numMp3s), L"&Delete"))
		return;

	for (auto &&tag : _tags)
		tag.delete_frames_by_name4(L"APIC");
	
	lstFrames.item_del_all(); // necessary because fill_textboxes_and_pic() will trigger EN_CHANGE in all textboxes
	fill_textboxes_and_pic();
	fill_frames_list();
}

void DlgEdit::btn_uncheck() const {
	for (WORD i = 0; i < consts::FIELDS_NAME4.size(); ++i) {
		const wd::CheckBox chk{this, static_cast<WORD>(CHK_ARTIST + i * 2)};
		chk.set_check_and_trigger(false);
	}
}

void DlgEdit::btn_ok() {
	for (WORD i = 0; i < consts::FIELDS_NAME4.size(); ++i) {
		const wd::CheckBox chk{this, static_cast<WORD>(CHK_ARTIST + i * 2)};
		if (chk.is_checked()) {
			const wd::Edit txt{this, static_cast<WORD>(TXT_ARTIST + i * 2)};
			std::wstring newText{txt.text()};
			wd::str::trim_spaces(newText); // trim every textbox

			for (auto &&tag : _tags) {
				if (newText.empty()) { // checked, empty text means frame removal
					tag.delete_frames_by_name4(consts::FIELDS_NAME4[i]); // note: if TXXX, all will be removed
				} else {
					auto pFrame = tag.frame_by_name4(consts::FIELDS_NAME4[i]); // it's checked, assume frame exists
					pFrame->set_simple_text(newText); // set trimmed text
				}
			}
		}
	}

	PostMessageW(hwnd(), WM_CLOSE, 0, 0);
	_ok = true; // to be returned from the modal
}

void DlgEdit::fill_textboxes_and_pic() {
	for (WORD i = 0; i < consts::FIELDS_NAME4.size(); ++i) { // APIC not listed here, just text fields
		const id3v2::Frame *pFrame = id3v2::same_frame_across_all_tags(consts::FIELDS_NAME4[i], _tags);
		const wd::CheckBox chk{this, static_cast<WORD>(CHK_ARTIST + i * 2)};
		const wd::Edit txt{this, static_cast<WORD>(TXT_ARTIST + i * 2)};

		if (!pFrame) { // if we don't have the same value across all files, show nothing
			chk.set_check_and_trigger(false);
			txt.set_text(L"");
		} else {
			chk.set_check_and_trigger(true); // will enable the textbox
			const std::wstring ss = pFrame->as_simple_text();
			txt.set_text(ss);
		}
	}

	if (const id3v2::BodyPicture *pApic = id3v2::same_apic_frame_across_all_tags(_tags); pApic) { // do we have single APIC?
		try {
			_bmpLoader.load(pApic->bin);
		} catch (const wd::WinErr &e) {
			wd::sys_dlg::msg_err(this, L"Image loading", L"Loading image failed.", e);
		}
		lblPicSize.set_text(wd::str::fmt(L"%u x %u pixels", _bmpLoader.cx, _bmpLoader.cy));
	} else {
		_bmpLoader.clear();
		lblPicSize.set_text(_tags.size() == 1 ? L"(no picture)" : L"(multiple pictures)");
	}
	wndPic.redraw();
}

void DlgEdit::fill_frames_list() const {
	lstFrames.item_del_all();
	if (_tags.size() == 1) { // editing 1 tag
		lstFrames.enable(true);
		for (auto &&f : _tags[0].frames()) {
			lstFrames.item_add({f.name4(), f.as_simple_text()});
		}
	} else { // editing multiple tags
		lstFrames.item_add({L"", wd::str::fmt(L"%u files...", _tags.size())});
		lstFrames.enable(false);
	}
	lstFrames.col(1).set_width_to_fill();
}
