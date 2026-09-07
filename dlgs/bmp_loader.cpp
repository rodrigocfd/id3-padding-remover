#include "bmp_loader.h"
#include <wincodec.h>

#ifdef _MSC_VER
#pragma comment(lib, "windowscodecs.lib")
#endif

void BmpLoader::clear() {
	if (hBmp) {
		DeleteObject(hBmp);
		hBmp = nullptr;
	}
}

void BmpLoader::load(const std::vector<BYTE> &bin) {
	wd::ComPtr<IWICImagingFactory> factory{};
	factory.co_create_instance(CLSID_WICImagingFactory, CLSCTX_INPROC_SERVER);

	wd::ComPtr<IWICStream> stream{};
	factory->CreateStream(stream.pptr());
	HRESULT hr = stream->InitializeFromMemory(const_cast<BYTE*>(bin.data()), static_cast<DWORD>(bin.size()));
	if (FAILED(hr)) [[unlikely]] {
		throw wd::WinErr{hr, L"IWICStream::InitializeFromMemory failed."};
	}

	wd::ComPtr<IWICBitmapDecoder> decoder{};
	hr = factory->CreateDecoderFromStream(stream.ptr(), nullptr, WICDecodeMetadataCacheOnLoad, decoder.pptr());
	if (FAILED(hr)) [[unlikely]] {
		throw wd::WinErr{hr, L"IWICImagingFactory::CreateDecoderFromStream failed."};
	}

	wd::ComPtr<IWICBitmapFrameDecode> frameDec{};
	hr = decoder->GetFrame(0, frameDec.pptr());
	if (FAILED(hr)) [[unlikely]] {
		throw wd::WinErr{hr, L"IWICBitmapDecoder::GetFrame failed."};
	}

	wd::ComPtr<IWICFormatConverter> fmtConv{};
	factory->CreateFormatConverter(fmtConv.pptr());
	fmtConv->Initialize(frameDec.ptr(), GUID_WICPixelFormat32bppBGRA, WICBitmapDitherTypeNone,
		nullptr, 0, WICBitmapPaletteTypeCustom);

	UINT width = 0, height = 0;
	hr = fmtConv->GetSize(&width, &height);
	if (FAILED(hr)) [[unlikely]] {
		throw wd::WinErr{hr, L"IWICFormatConverter::GetSize failed."};
	}
	cx = width; // store
	cy = height;

	const BITMAPINFO bmi{
		.bmiHeader{
			.biSize = sizeof(BITMAPINFOHEADER),
			.biWidth = cx,
			.biHeight = -cy, // top-down
			.biPlanes = 1,
			.biBitCount = 32,
			.biCompression = BI_RGB,
		},
	};

	clear();
	BYTE *pRgbBits = nullptr;
	hBmp = CreateDIBSection(nullptr, &bmi, DIB_RGB_COLORS, reinterpret_cast<void**>(&pRgbBits), nullptr, 0);
	if (!hBmp) [[unlikely]] {
		throw wd::WinErr{GetLastError(), L"CreateDIBSection failed."};
	}

	const UINT stride = cx * 4;
	hr = fmtConv->CopyPixels(nullptr, stride, stride * cy, pRgbBits);
	if (FAILED(hr)) [[unlikely]] {
		throw wd::WinErr{hr, L"IWICFormatConverter::CopyPixels failed."};
	}
}
