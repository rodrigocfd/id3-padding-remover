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
	if (id3::Tag::FrameHasSameValueAcrossAllTags(_pTags, L"APIC")) {
		for (auto&& pTag : _pTags) {
			if (auto frame = pTag->frameByName4(L"APIC"); frame.has_value()) { // 1st tag which has this frame
				auto pFramePic = frame.value()->dataAs<id3::Frame::Picture>();
				
			}
		}
	}
	
	return 0;
}
