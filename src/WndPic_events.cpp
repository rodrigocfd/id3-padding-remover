#include "WndPic.h"

LRESULT WndPic::wndProc(UINT uMsg, WPARAM wp, LPARAM lp)
{
	switch (uMsg) {
	case WM_CREATE: return onCreate();
	default:        return DefWindowProcW(hWnd(), uMsg, wp, lp);
	}
}

LRESULT WndPic::onCreate()
{
	OutputDebugStringW(L"CTRL\n");
	return 0;
}
