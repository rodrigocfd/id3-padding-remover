#pragma once
#include <windlg/lib.h>

class DlgMain final : public lib::DialogMain {
public:
	virtual ~DlgMain() { }

	constexpr DlgMain() = default;
	DlgMain(const DialogMain&) = delete;
	DlgMain(DialogMain&&) = delete;
	DlgMain& operator=(const DlgMain&) = delete;
	DlgMain& operator=(DlgMain&&) = delete;

private:
	INT_PTR dlgProc(UINT uMsg, WPARAM wp, LPARAM lp) override;
	INT_PTR onInitDialog();
};
