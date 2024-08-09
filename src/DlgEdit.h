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
	INT_PTR onBtnUncheck();
	INT_PTR onBtnOk();

	void _writeTitlebarCounts() const;
	void _writeFields() const;
	void _renderFrames() const;

	const std::vector<id3::Tag*> _pTags;
};
