#pragma once
#include <windlg/lib.h>

class DlgEdit final : public lib::DialogModal {
public:
	virtual ~DlgEdit() { }

	constexpr DlgEdit() = default;
	DlgEdit(const DlgEdit&) = delete;
	DlgEdit(DlgEdit&&) = delete;
	DlgEdit& operator=(const DlgEdit&) = delete;
	DlgEdit& operator=(DlgEdit&&) = delete;

private:
	INT_PTR dlgProc(UINT uMsg, WPARAM wp, LPARAM lp) override;
	INT_PTR onInitDialog();
};
