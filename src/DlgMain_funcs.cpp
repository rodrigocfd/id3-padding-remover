#include "DlgMain.h"
#include "id3v2/Tag.h"
#include "../res/resource.h"
using std::vector, std::wstring;

void DlgMain::_addMp3sToList(const vector<wstring>& mp3s)
{
	vector<wstring> invalids;
	for (auto&& mp3 : mp3s) {
		if (!lib::path::hasExtension(mp3, L"mp3"))
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
		return;
	}

	lib::ListView lv{this, LST_FILES};
	for (auto&& mp3 : mp3s) {
		if (auto curItem = lv.items.find(mp3); curItem.has_value()) { // MP3 already in list
			auto pTagCurrent = curItem.value().data<id3::Tag*>();
			delete pTagCurrent;
			auto pReloadedTag = new id3::Tag{mp3}; // reload the tag from file
			curItem.value().setData(pReloadedTag);
		} else { // MP3 not in list yet
			auto newItem = lv.items.add(mp3);
			auto pTag = new id3::Tag{mp3};
			newItem.setData(pTag);
		}
	}
	_updateNumFiles(lv.items.count());
}

void DlgMain::_updateNumFiles(UINT numFiles)
{
	setText(lib::str::fmt(L"ID3 Fit (%d/%d)",
		lib::ListView{this, LST_FILES}.items.countSelected(), numFiles));
}
