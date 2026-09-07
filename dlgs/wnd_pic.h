#pragma once
#include "../windlg/lib.hpp"
#include "bmp_loader.h"

class WndPic final : public wd::BaseControl {
public:
	WndPic(const wd::BaseDialog *pParent, wd::Menu &mnuPic, const BmpLoader &bmpLoader)
		: BaseControl{pParent}, _mnuPic{mnuPic}, _bmpLoader{bmpLoader} { }

	void redraw() const { InvalidateRect(hwnd(), nullptr, TRUE); }

private:
	LRESULT wnd_proc(UINT msg, WPARAM wp, LPARAM lp) override;
	LRESULT on_paint() const;
	LRESULT on_context_menu(LPARAM lp) const;

	const BmpLoader &_bmpLoader; // loaded by DlgEdit
	wd::Menu &_mnuPic; // belongs to DlgEdit, we'll use as context menu
};
