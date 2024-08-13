#pragma once
#include <windlg/lib.h>
#include "WndPic.h"
#include "id3v2/Tag.h"

class DlgEdit final : public lib::DialogModal {
public:
	virtual ~DlgEdit() { }

	constexpr explicit DlgEdit(const std::vector<id3::Tag*> pTags) : _pTags{pTags} { }
	DlgEdit(const DlgEdit&) = delete;
	DlgEdit(DlgEdit&&) = delete;
	DlgEdit& operator=(const DlgEdit&) = delete;
	DlgEdit& operator=(DlgEdit&&) = delete;

private:
	INT_PTR dlgProc(UINT uMsg, WPARAM wp, LPARAM lp) override;
	INT_PTR onInitDialog();
	INT_PTR onChk(WPARAM wp);
	INT_PTR onBtnUncheck();
	INT_PTR onBtnOk();

	void _renderTitlebarCounts() const;
	void _renderTextboxes() const;
	void _renderFramesList() const;

	const std::vector<id3::Tag*> _pTags;
	WndPic _wndPic;

	struct FieldInfo final {
		WORD chkId = 0;
		std::wstring_view name4;
	};
	static std::initializer_list<FieldInfo> _Fields;
	static std::initializer_list<std::wstring_view> _Genres;
};
