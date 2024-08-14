#include "WndPic.h"
#include <Shlwapi.h>
#include <olectl.h>
#pragma comment(lib, "Shlwapi.lib")

LRESULT WndPic::wndProc(UINT uMsg, WPARAM wp, LPARAM lp)
{
	switch (uMsg) {
		case WM_CREATE:  return onCreate();
		case WM_DESTROY: return onDestroy();
		default:         return DefWindowProcW(hWnd(), uMsg, wp, lp);
	}
}

LRESULT WndPic::onCreate()
{
	if (auto pFrame = id3::Tag::SameFrameAcrossAllTags(_pTags, L"APIC"); pFrame.has_value()) {
		auto pFramePic = pFrame.value()->dataAs<id3::Frame::Picture>();
		auto stream = lib::ComPtr{SHCreateMemStream(pFramePic->bin.data(), static_cast<UINT>(pFramePic->bin.size()))};
		lib::checkHr(
			OleLoadPicture(stream.ptr(), 0, FALSE, IID_IPicture, reinterpret_cast<void**>(_pic.pptr())),
			"OleLoadPicture");
	}
	return 0;
}

LRESULT WndPic::onDestroy()
{
	_pic.release();
	return 0;
}
