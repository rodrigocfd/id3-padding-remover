#include "DlgMain.h"
#include "DlgEdit.h"
#include "id3v2/Tag.h"
#include "../res/resource.h"
using std::array, std::optional, std::vector, std::wstring, std::wstring_view;

int APIENTRY wWinMain(_In_ HINSTANCE hInst, _In_opt_ HINSTANCE, _In_ LPWSTR, _In_ int cmdShow)
{
	lib::ComOle oleLib;
	DlgMain d;
	return lib::runMain(d, hInst, DLG_MAIN, cmdShow, ICO_FOULBACHELOR, ACC_MAIN);
}

INT_PTR DlgMain::dlgProc(UINT uMsg, WPARAM wp, LPARAM lp)
{
	lib::ListView::ProcessMessages(this, LST_FILES, uMsg, wp, lp, MNU_FILE);

	switch (uMsg) {
		case WM_INITDIALOG:    return onInitDialog();
		case WM_SIZE:          return onSize(wp, lp);
		case WM_INITMENUPOPUP: return onInitMenuPopup(wp);
		case WM_COMMAND:
			switch (LOWORD(wp)) {
				case MNU_FILE_OPEN:   return onMenuFileOpen();
				case MNU_FILE_EDIT:   return onMenuFileEdit();
				case MNU_FILE_RESAVE: return onMenuFileReSave();
				case MNU_FILE_REMOVE: return onMenuFileRemove();
				case MNU_FILE_ABOUT:  return onMenuFileAbout();
				default:              return FALSE;
			}
		case WM_NOTIFY:
			switch (reinterpret_cast<NMHDR*>(lp)->idFrom) {
				case LST_FILES:
					switch (reinterpret_cast<NMHDR*>(lp)->code) {
						case NM_DBLCLK:       return onMenuFileEdit();
						case LVN_ITEMCHANGED: return onListItemChanged();
						case LVN_DELETEITEM:  return onListDeleteItem(lp);
						case HDN_ITEMCLICK:   return onListHeaderClick(lp);
						case LVN_KEYDOWN: {
							switch (reinterpret_cast<NMLVKEYDOWN*>(lp)->wVKey) {
								case VK_RETURN: return onMenuFileEdit();
								case VK_DELETE: return onMenuFileRemove();
								default:        return FALSE;
							}
						}
						default: return FALSE;
					}
				default: return FALSE;
			}
		case WM_CLOSE:     DestroyWindow(hWnd()); return TRUE;
		case WM_NCDESTROY: PostQuitMessage(0); return TRUE;
		default:           return FALSE;
	}
}

INT_PTR DlgMain::onInitDialog()
{
	dlg.registerDragDrop()
		.layout(lib::Dialog::Act::Resize, lib::Dialog::Act::Resize, {LST_FILES});

	_imgList.create({16, 16})
		.addShell({L"mp3"});

	lib::ListView lv{this, LST_FILES};
	lv.setImageList(_imgList)
		.setFullRowSelect();
	lv.columns.add(L"File", lib::dpi::x(400)).setSortArrow(HDF_SORTUP);
	lv.columns.add(L"Pad", lib::dpi::x(50)).setJustification(HDF_RIGHT);
	lv.columns.add(L"Pic", lib::dpi::x(30)).setJustification(HDF_CENTER);
	lv.columns.add(L"RG", lib::dpi::x(30)).setJustification(HDF_CENTER);
	lv.columns.add(L"Artist", lib::dpi::x(90));
	lv.columns.add(L"Year", lib::dpi::x(40)).setJustification(HDF_CENTER);
	lv.columns.add(L"Album", lib::dpi::x(100));
	lv.columns.add(L"T#", lib::dpi::x(30)).setJustification(HDF_RIGHT);
	lv.columns.add(L"Title", lib::dpi::x(100));
	lv.columns.add(L"Genre", lib::dpi::x(90));
	lv.columns.add(L"Performer", lib::dpi::x(70));
	lv.columns.add(L"Composer", lib::dpi::x(70));
	lv.columns.add(L"Lyricist", lib::dpi::x(70));
	lv.columns.add(L"Orig. artist", lib::dpi::x(70));
	lv.columns.add(L"Comment", lib::dpi::x(70));
	lv.columns[0].setWidthToFill();

	return TRUE;
}

void DlgMain::onDropTarget(const vector<wstring>& files)
{
	_addMp3sToList(files);
}

INT_PTR DlgMain::onSize(WPARAM wp, LPARAM lp)
{
	if (wp != SIZE_MINIMIZED)
		lib::ListView{this, LST_FILES}.columns[0].setWidthToFill();

	return TRUE;
}

INT_PTR DlgMain::onInitMenuPopup(WPARAM wp)
{
	lib::Menu popupMenu{reinterpret_cast<HMENU>(wp)};
	if (popupMenu.idByPos(0) == MNU_FILE_OPEN) {
		popupMenu.setDefaultItemByCmd(MNU_FILE_EDIT);
		popupMenu.enableItemsByCmd({MNU_FILE_EDIT, MNU_FILE_RESAVE, MNU_FILE_REMOVE},
			lib::ListView{this, LST_FILES}.items.countSelected() > 0);
	}
	return TRUE;
}

INT_PTR DlgMain::onMenuFileOpen()
{
	if (optional<vector<wstring>> files = dlg.showOpenFiles({
		{L"MP3 audio files", L"*.mp3"},
		{L"All files", L"*.*"},
	}); files.has_value()) {
		_addMp3sToList(files.value());
	}
	return TRUE;
}

INT_PTR DlgMain::onMenuFileEdit()
{
	lib::ListView lv{this, LST_FILES};
	auto selItems = lv.items.selected();
	if (selItems.empty())
		return TRUE; // Enter key will hit here even without selected items

	auto pTags = lib::vec::transform(selItems, [](const lib::ListView::Item& item) {
		return item.data<id3::Tag*>();
	});

	DlgEdit dlgEdit{pTags};
	dlgEdit.showModal(this, DLG_EDIT);

	if (dlgEdit.clickedOk()) {
		for (auto&& item : selItems)
			_renderMp3ListItem(item); // tags potentially changed, re-render them in the list
		_saveSelected(); // DlgEdit won't save the files, just change the tags; we save them here
	}
	return TRUE;
}

INT_PTR DlgMain::onMenuFileReSave()
{
	_saveSelected();
	return TRUE;
}

INT_PTR DlgMain::onMenuFileRemove()
{
	lib::ListView{this, LST_FILES}.items.removeSelected();
	return TRUE;
}

INT_PTR DlgMain::onMenuFileAbout()
{
	lib::VersionInfo vi;
	wstring_view productName = vi.strInfo(vi.langsCps()[0], L"ProductName");
	array<WORD, 4> ver = vi.verNum();
	auto body = lib::str::fmt(L"Version %u.%u.%u.\nWritten in C++20.", ver[0], ver[1], ver[2]);

	dlg.msgBox(L"About", {productName}, body, TDCBF_OK_BUTTON, TD_INFORMATION_ICON);
	return TRUE;
}

INT_PTR DlgMain::onListItemChanged()
{
	_updateNumFiles(lib::ListView{this, LST_FILES}.items.count());
	return TRUE;
}

INT_PTR DlgMain::onListDeleteItem(LPARAM lp)
{
	lib::ListView lv{this, LST_FILES};
	auto pNmlv = reinterpret_cast<NMLISTVIEW*>(lp);
	auto pTag = lv.items[pNmlv->iItem].data<id3::Tag*>();
	delete pTag;
	_updateNumFiles(lv.items.count() - 1); // notification is sent before the item is removed
	return TRUE;
}

INT_PTR DlgMain::onListHeaderClick(LPARAM lp)
{
	auto pNmh = reinterpret_cast<NMHEADERW*>(lp);
	lib::ListView lv{this, LST_FILES};
	int arrowFlag = lv.columns[pNmh->iItem].sortArrow();
	bool willSortAsc = !(arrowFlag & HDF_SORTUP);

	lv.columns[pNmh->iItem].setSortArrow(willSortAsc ? HDF_SORTUP : HDF_SORTDOWN); // draw arrow
	_sort = {.col = pNmh->iItem, .asc = willSortAsc}; // update state

	_sortList();
	return TRUE;
}
