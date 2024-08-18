#include "WndPic.h"

LRESULT WndPic::wndProc(UINT uMsg, WPARAM wp, LPARAM lp)
{
	switch (uMsg) {
		case WM_PAINT: return onPaint();
		default:       return DefWindowProcW(hWnd(), uMsg, wp, lp);
	}
}

LRESULT WndPic::onPaint()
{
	PAINTSTRUCT ps{};
	HDC hdc = BeginPaint(hWnd(), &ps);

	if (_pic.ptr()) {
		OLE_XSIZE_HIMETRIC hmx = 0;
		OLE_YSIZE_HIMETRIC hmy = 0;
		_pic->get_Width(&hmx);
		_pic->get_Height(&hmy);

		RECT dummy{};
		_pic->Render(ps.hdc, 0, 0, ps.rcPaint.right, ps.rcPaint.bottom, 0, hmy, hmx, -hmy, &dummy);
	}

	EndPaint(hWnd(), &ps);
	return 0;
}
