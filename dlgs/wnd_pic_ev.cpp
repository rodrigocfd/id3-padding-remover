#include "wnd_pic.h"

LRESULT WndPic::wnd_proc(UINT msg, WPARAM wp, LPARAM lp) {
	switch (msg) {
		case WM_PAINT:       return on_paint();
		case WM_CONTEXTMENU: return on_context_menu(lp);
	}
	return DefWindowProcW(hwnd(), msg, wp, lp);
}

LRESULT WndPic::on_paint() const {
	PAINTSTRUCT ps{};
	const HDC hdc = BeginPaint(hwnd(), &ps);

	if (_bmpLoader.hBmp) {
		const HDC hdcMem = CreateCompatibleDC(hdc);
		const HGDIOBJ hBmpOld = SelectObject(hdcMem, _bmpLoader.hBmp);
		SetStretchBltMode(hdc, HALFTONE);
		SetBrushOrgEx(hdc, 0, 0, nullptr);
		StretchBlt(hdc, 0, 0, ps.rcPaint.right, ps.rcPaint.bottom, hdcMem, 0, 0, _bmpLoader.cx, _bmpLoader.cy, SRCCOPY);
		SelectObject(hdcMem, hBmpOld);
		DeleteDC(hdcMem);
	}

	EndPaint(hwnd(), &ps);
	return 0;
}

LRESULT WndPic::on_context_menu(LPARAM lp) const {
	POINT pt{.x = LOWORD(lp), .y = HIWORD(lp)};
	ScreenToClient(hwnd(), &pt);
	_mnuPic.show_at_point(pt, parent()->hwnd(), hwnd());
	return 0;
}
