#pragma once
#include <windlg/lib.h>
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

	struct FieldInfo final {
		WORD chkId = 0;
		LPCWSTR name4 = nullptr;
	};
	static FieldInfo _Fields[15];

	static LPCWSTR _Genres[51];
};
