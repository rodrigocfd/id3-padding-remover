#pragma once
#include <windlg/lib.h>

class DlgMain final : public lib::DialogMain {
public:
	virtual ~DlgMain() { }

	constexpr DlgMain() = default;
	DlgMain(const DlgMain&) = delete;
	DlgMain(DlgMain&&) = delete;
	DlgMain& operator=(const DlgMain&) = delete;
	DlgMain& operator=(DlgMain&&) = delete;

private:
	INT_PTR dlgProc(UINT uMsg, WPARAM wp, LPARAM lp) override;
	INT_PTR onInitDialog();
	void    onDropTarget(const std::vector<std::wstring>& files) override;
	INT_PTR onSize(WPARAM wp, LPARAM lp);
	INT_PTR onInitMenuPopup(WPARAM wp);
	INT_PTR onMenuFileOpen();
	INT_PTR onMenuFileEdit();
	INT_PTR onMenuFileReSave();
	INT_PTR onMenuFileRemove();
	INT_PTR onMenuFileAbout();
	INT_PTR onListItemChanged();
	INT_PTR onListDeleteItem(LPARAM lp);
	INT_PTR onListHeaderClick(LPARAM lp);

	void _addMp3sToList(const std::vector<std::wstring>& mp3s) const;
	void _addOneMp3ToList(std::wstring_view mp3) const;
	void _renderMp3ListItem(lib::ListView::Item item) const;
	void _updateNumFiles(UINT numFiles) const;
	void _sortList() const;
	void _saveSelected() const;

	lib::ImgList _imgList;
	struct { int col; bool asc; } _sort = {.col = 0, .asc = true};
};
