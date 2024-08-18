#include <system_error>
#include "WndPic.h"
#include <Shlwapi.h>
#include <olectl.h>
#pragma comment(lib, "Shlwapi.lib")
using std::span;

void WndPic::_loadPic(span<BYTE> src)
{
	auto stream = lib::ComPtr{SHCreateMemStream(src.data(), static_cast<UINT>(src.size()))};
	if (HRESULT hr = OleLoadPicture(
		stream.ptr(), 0, FALSE, IID_IPicture, reinterpret_cast<void**>(_pic.pptr())); FAILED(hr)) [[unlikely]] {
		auto err = std::system_category().message(hr);
		MessageBoxA(GetParent(hWnd()), err.c_str(), "Picture loading error", MB_ICONERROR);
	}
}

void WndPic::_renderPic(const PAINTSTRUCT& ps) const
{
	if (_pic.ptr()) {
		OLE_XSIZE_HIMETRIC hmx = 0;
		OLE_YSIZE_HIMETRIC hmy = 0;
		_pic->get_Width(&hmx);
		_pic->get_Height(&hmy);

		RECT dummy{};
		_pic->Render(ps.hdc, 0, 0, ps.rcPaint.right, ps.rcPaint.bottom, 0, hmy, hmx, -hmy, &dummy);
	}
}
