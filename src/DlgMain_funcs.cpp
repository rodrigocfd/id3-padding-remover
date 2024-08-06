#include "DlgMain.h"
#include "id3v2/Tag.h"
#include "../res/resource.h"
using std::optional, std::vector, std::wstring, std::wstring_view;

void DlgMain::_addMp3sToList(const vector<wstring>& mp3s)
{
	vector<wstring> invalids;
	for (auto&& mp3 : mp3s) {
		if (!lib::path::hasExtension(mp3, L"mp3") && !lib::path::isDir(mp3))
			invalids.emplace_back(mp3);
	}
	if (!invalids.empty()) {
		auto buf = lib::str::newReserved(22 * invalids.size()); // arbitrary
		buf = L"Non-MP3 file(s):";
		for (const auto& mp3 : invalids) {
			buf.append(L"\n");
			buf.append(mp3);
		}
		dlg.msgBox(L"Non-MP3 file(s)", {}, buf, TDCBF_OK_BUTTON, TD_ERROR_ICON);
		return; // do not continue; no files will be added
	}

	for (auto&& mp3 : mp3s) {
		if (lib::path::isDir(mp3)) {
			for (auto&& f : lib::path::dirList(mp3 + L"\\*.mp3")) // search only 1 level deep
				_addOneMp3ToList(f);
		} else {
			_addOneMp3ToList(mp3);
		}
	}
	_updateNumFiles(lib::ListView{this, LST_FILES}.items.count());
}

void DlgMain::_addOneMp3ToList(wstring_view mp3)
{
	lib::ListView lv{this, LST_FILES};
	int idxItem = -1;
	if (auto curItem = lv.items.find(mp3); curItem.has_value()) { // MP3 already in list
		auto pTagCurrent = curItem.value().data<id3::Tag*>();
		delete pTagCurrent;
		idxItem = curItem.value().index();
	} else { // MP3 not in list yet
		auto newItem = lv.items.add(mp3); // insert new
		idxItem = newItem.index();
	}

	auto item = lv.items[idxItem];
	auto pTag = new id3::Tag{mp3}; // store pointer to Tag in item
	item.setData(pTag);
	item.setText(std::to_wstring(pTag->padding), 1);
	if (auto pic = pTag->frameByName4(L"APIC"); pic.has_value())
		item.setText(L"\u2713", 2); // checkmark
}

void DlgMain::_updateNumFiles(UINT numFiles)
{
	setText(lib::str::fmt(L"ID3 Fit (%d/%d)",
		lib::ListView{this, LST_FILES}.items.countSelected(), numFiles));
}
