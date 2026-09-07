#include <algorithm>
#include "dlg_main.h"
#include "dlg_edit.h"
#include "../id3v2/tag.h"

bool DlgMain::on_drag_files(const std::vector<std::wstring> &files) const {
	for (auto &&file : files) {
		if (wd::file::is_dir(file)) {
			for (auto &&subFile : wd::DirList{file}) { // go only 1 level deep
				if (wd::str::ends_with_i(subFile, L".mp3"))
					return true; // at least 1 MP3 file, proceed to drop
			}
		} else {
			if (wd::str::ends_with_i(file, L".mp3"))
				return true; // at least 1 MP3 file, proceed to drop
		}
	}
	return false; // no MP3 found amongst all files, don't proceed to drop
}

void DlgMain::on_drop_files(const std::vector<std::wstring> &files) const {
	add_mp3s_to_list(files);
}

void DlgMain::list_delete_item(const NMHDR &nm) const {
	auto nmlv = reinterpret_cast<const NMLISTVIEW&>(nm);
	auto pTag = reinterpret_cast<id3v2::Tag*>(nmlv.lParam);
	delete pTag;
}

void DlgMain::list_header_lick(const NMHDR &nm) {
	auto nmh = reinterpret_cast<const NMHEADERW&>(nm);
	if (_curSortCol == nmh.iItem) { // same column
		_curSortDir = (_curSortDir == wd::ListView::Sort::asc_i) ?
			wd::ListView::Sort::desc_i : wd::ListView::Sort::asc_i; // invert sort direction
	} else {
		_curSortCol = nmh.iItem; // new col
		_curSortDir = wd::ListView::Sort::asc_i;
	}
	sort_list();
}

void DlgMain::menu_file_open() const {
	const std::vector<std::wstring> files = wd::sys_dlg::file_open_multi(this, {
		{L"MP3 audio files", L"*.mp3"},
		{L"All files", L"*.*"},
	});
	if (!files.empty())
		add_mp3s_to_list(files);
}

void DlgMain::menu_file_edit() {
	if (!lstFiles.item_selected_count()) [[unlikely]] {
		return; // necessary check because ENTER key will hit here, whether you have selected items or not
	}

	const std::vector<wd::ListView::Item> selItems = lstFiles.item_selected();
	std::vector<id3v2::Tag> clonedTags{};
	clonedTags.reserve(selItems.size());
	for (auto &&selItem : selItems)
		clonedTags.emplace_back( selItem.data<id3v2::Tag*>()->clone() );

	DlgEdit dlgEdit{clonedTags};
	wd::run_modal_dialog(this, dlgEdit);
	if (dlgEdit.ok()) {
		for (size_t i = 0; i < selItems.size(); ++i) {
			const std::wstring mp3Path = selItems[i].text(0);
			clonedTags[i].save_to_file(mp3Path); // saving to disk will set padding to zero
			*selItems[i].data<id3v2::Tag*>() = std::move(clonedTags[i]); // update modified tags in memory
			render_mp3_in_list(selItems[i]);
		}
	}
}

void DlgMain::menu_file_resave() {
	const std::wstring msg = wd::str::fmt(L"Do you want to rewrite the tags of %u file(s)?", lstFiles.item_selected_count());
	if (wd::sys_dlg::msg_ask(this, L"Save files", msg, L"&Rewrite")) [[likely]] {
		for (auto &&selItem : lstFiles.item_selected()) {
			const std::wstring mp3Path = selItem.text(0);
			auto pTag = selItem.data<id3v2::Tag*>();
			pTag->save_to_file(mp3Path); // will set padding to zero
			render_mp3_in_list(selItem);
		}
	}
}

void DlgMain::menu_del_pic(bool delRg) {
	const wd::StrView prompt = delRg
		? L"Do you want to remove picture and ReplayGain from %u file(s)?"
		: L"Do you want to remove picture from %u file(s)?";
	const std::wstring msg = wd::str::fmt(prompt, lstFiles.item_selected_count());

	if (wd::sys_dlg::msg_ask(this, L"Remove frames", msg, L"&Remove")) [[likely]] {
		for (auto &&selItem : lstFiles.item_selected()) {
			auto pTag = selItem.data<id3v2::Tag*>();
			pTag->delete_frames_by_name4(L"APIC");
			if (delRg)
				pTag->delete_replay_gain();

			const std::wstring mp3Path = selItem.text(0);
			pTag->save_to_file(mp3Path);
			render_mp3_in_list(selItem);
		}
	}
}

void DlgMain::update_titlebar_num_files(size_t numFiles) const {
	set_text(wd::str::fmt(L"ID3 Fit (%u/%u)",
		lstFiles.item_selected_count(), numFiles));
}

void DlgMain::sort_list() const {
	if (_curSortCol == 0) {
		lstFiles.sort({{0, _curSortDir}});
	} else { // on subsequent cols, the second sort criteria is the file path
		lstFiles.sort({{_curSortCol, _curSortDir}, {0, _curSortDir}});
	}
}

void DlgMain::add_mp3s_to_list(const std::vector<std::wstring> &files) const {
	std::vector<std::wstring> mp3s{};
	mp3s.reserve(files.size()); // arbitrary

	for (auto &&file : files) {
		if (wd::file::is_dir(file)) {
			for (auto &&subFile : wd::DirList{file}) { // go only 1 level deep
				if (wd::str::ends_with_i(subFile, L".mp3"))
					mp3s.emplace_back(subFile);
			}
		} else {
			if (wd::str::ends_with_i(file, L".mp3"))
				mp3s.emplace_back(file);
		}
	}

	if (mp3s.empty()) [[unlikely]] {
		return;
	}

	for (auto &&mp3 : mp3s)
		add_one_mp3_to_list(mp3);

	update_titlebar_num_files(lstFiles.item_count());
	lstFiles.col(0).set_width_to_fill();
	sort_list();
}

void DlgMain::add_one_mp3_to_list(wd::StrView filePath) const {
	if (wd::ListView::Item itemExisting = lstFiles.item_find_i(filePath, 0); itemExisting.valid())
		itemExisting.del(); // will fire LVN_DELETEITEM and free the stored tag

	const wd::ListView::Item newItem = lstFiles.item_add({filePath}, 0);
	auto pTag = new id3v2::Tag{filePath}; // load tag from MP3
	newItem.set_data(pTag); // store tag; will be released in LVN_DELETEITEM
	render_mp3_in_list(newItem);
}

void DlgMain::render_mp3_in_list(wd::ListView::Item item) const {
	auto pTag = item.data<const id3v2::Tag*>(); // retrieve stored tag
	auto file = item.text(0);
	auto pic = pTag->frame_by_name4(L"APIC") ? L"\u2713" : L""; // checkmark symbol

	auto frameTxt = [&pTag](wd::StrView name4) -> std::wstring {
		auto pFrame = pTag->frame_by_name4(name4);
		return pFrame ? pFrame->as_simple_text() : L"";
	};

	item.set_texts({
		file,
		wd::str::fmt(L"%u", pTag->padding()),
		pic,
		pTag->replay_gain_status(),
		frameTxt(L"TPE1"),
		frameTxt(L"TYER"),
		frameTxt(L"TALB"),
		frameTxt(L"TRCK"),
		frameTxt(L"TIT2"),
		frameTxt(L"TCON"),
		frameTxt(L"TPE3"),
		frameTxt(L"TCOM"),
		frameTxt(L"TEXT"),
		frameTxt(L"TOPE"),
		frameTxt(L"TEXT"),
		frameTxt(L"COMM"),
	});
}
