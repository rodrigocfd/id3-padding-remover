#include "DlgMain.h"
#include "id3v2/Tag.h"
#include "../res/resource.h"
using std::optional, std::vector, std::wstring, std::wstring_view;

void DlgMain::_addMp3sToList(const vector<wstring>& mp3s) const
{
	lib::ListView lv{this, LST_FILES};

	vector<wstring> nonMp3s; // keep track of files that aren't MP3
	for (const wstring& mp3 : mp3s) {
		if (!lib::path::hasExtension(mp3, L"mp3") && !lib::path::isDir(mp3))
			nonMp3s.emplace_back(mp3);
	}
	if (!nonMp3s.empty()) {
		wstring buf = lib::str::newReserved(22 * nonMp3s.size()); // arbitrary
		buf = L"Non-MP3 file(s):";
		for (const wstring& mp3 : nonMp3s) {
			buf.append(L"\n");
			buf.append(mp3);
		}
		dlg.msgBox(L"Non-MP3 file(s)", {}, buf, TDCBF_OK_BUTTON, TD_ERROR_ICON);
		return; // do not continue; no files will be added
	}

	for (const wstring& mp3 : mp3s) { // all files are MP3, let's add them
		if (lib::path::isDir(mp3)) {
			for (const wstring& f : lib::path::dirList(mp3 + L"\\*.mp3")) // search only 1 level deep
				_addOneMp3ToList(f);
		} else {
			_addOneMp3ToList(mp3);
		}
	}
	_updateNumFiles(lv.items.count());
	lv.columns[0].setWidthToFill();
	_sortList();
}

void DlgMain::_addOneMp3ToList(wstring_view mp3) const
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
	id3::Tag* pTag = nullptr;
	try {
		pTag = new id3::Tag{mp3}; // store pointer to Tag in item
	} catch (const std::exception& e) {
		item.remove();
		auto msg = lib::str::fmt(L"%s\n\n%s", mp3, lib::str::toWide(e.what()));
		dlg.msgBox(L"Tag parsing error", {}, msg, TDCBF_OK_BUTTON, TD_ERROR_ICON);
		return;
	}
	item.setData(pTag);
	_renderMp3ListItem(item);
}

void DlgMain::_renderMp3ListItem(lib::ListView::Item item) const
{
	auto pTag = item.data<const id3::Tag*>();

	auto renderSimple = [item, pTag](wstring_view name4, UINT col) {
		if (optional<const id3::Frame*> frame = pTag->frameByName4(name4); frame.has_value()) {
			item.setText(std::get_if<id3::Frame::Text>(&frame.value()->data)->text, col);
		} else {
			item.setText(L"", col); // removed frames need to have their text erased
		}
	};

	item.setText(std::to_wstring(pTag->padding), 1);
	if (auto pic = pTag->frameByName4(L"APIC"); pic.has_value()) {
		item.setText(L"\u2713", 2); // checkmark
	} else {
		item.setText(L"", 2);
	}
	item.setText(pTag->replayGainStatus(), 3);
	renderSimple(L"TPE1", 4);
	renderSimple(L"TYER", 5);
	renderSimple(L"TALB", 6);
	renderSimple(L"TRCK", 7);
	renderSimple(L"TIT2", 8);
	renderSimple(L"TCON", 9);
	renderSimple(L"TPE3", 10);
	renderSimple(L"TCOM", 11);
	renderSimple(L"TEXT", 12);
	renderSimple(L"TOPE", 13);
	if (auto comm = pTag->frameByName4(L"COMM"); comm.has_value()) {
		item.setText(std::get_if<id3::Frame::Comment>(&comm.value()->data)->text, 14);
	} else {
		item.setText(L"", 14);
	}
}

void DlgMain::_updateNumFiles(UINT numFiles) const
{
	setText(lib::str::fmt(L"ID3 Fit (%d/%d)",
		lib::ListView{this, LST_FILES}.items.countSelected(), numFiles));
}

void DlgMain::_sortList() const
{
	using lib::ListView;
	ListView{this, LST_FILES}.items.sort([this](ListView::Item a, ListView::Item b) -> int {
		int cmp = 0;
		if (_sort.col == 0 || _sort.col == 1) {
			auto pTagA = a.data<const id3::Tag*>();
			auto pTagB = b.data<const id3::Tag*>();
			if (_sort.col == 0) { // by path
				cmp = lib::str::cmpI(pTagA->path, pTagB->path);
			} else { // by padding size
				cmp = pTagA->padding - pTagB->padding;
			}
		} else { // by column text
			cmp = lib::str::cmpI(a.text(_sort.col), b.text(_sort.col));
		}
		return _sort.asc ? cmp : -cmp;
	});
}
