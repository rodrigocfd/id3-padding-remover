#include "WndPic.h"

LRESULT WndPic::wndProc(UINT uMsg, WPARAM wp, LPARAM lp)
{
	switch (uMsg) {
		case WM_CREATE:  return onCreate();
		case WM_PAINT:   return onPaint();
		case WM_DESTROY: return onDestroy();
		default:         return DefWindowProcW(hWnd(), uMsg, wp, lp);
	}
}

LRESULT WndPic::onCreate()
{
	if (auto pFrame = id3::Tag::SameFrameAcrossAllTags(_pTags, L"APIC"); pFrame.has_value()) {
		auto pFramePic = pFrame.value()->dataAs<id3::Frame::Picture>();
		_loadPic(pFramePic->bin);
	}
	return 0;
}

LRESULT WndPic::onPaint()
{
	PAINTSTRUCT ps{};
	HDC hdc = BeginPaint(hWnd(), &ps);
	_renderPic(ps);
	EndPaint(hWnd(), &ps);
	return 0;
}

LRESULT WndPic::onDestroy()
{
	_pic.release();
	return 0;
}
