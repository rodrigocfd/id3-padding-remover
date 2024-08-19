#include <system_error>
#include "DlgEdit.h"
#include <Shlwapi.h>
#include <olectl.h>
#include "../res/resource.h"
#pragma comment(lib, "Shlwapi.lib")

void DlgEdit::renderTitlebarCounts() const
{
	if (_pTags.size() > 1) {
		setText(lib::str::fmt(L"%s - %d files", text(), _pTags.size()));
	} else {
		setText(lib::str::fmt(L"%s - %d frames", text(), _pTags[0]->frames.size()));
	}
}

void DlgEdit::renderTextboxes() const
{
	for (auto&& field : _Fields) {
		if (auto pFrame = id3::Tag::SameFrameAcrossAllTags(_pTags, field.name4); pFrame.has_value()) {
			lib::CheckRadio{this, field.chkId}.checkAndTrigger();
			lib::NativeControl{this, static_cast<WORD>(field.chkId + 1)}.setText(pFrame.value()->asText());
		}
	}
}

void DlgEdit::loadPicture()
{
	if (auto pFrame = id3::Tag::SameFrameAcrossAllTags(_pTags, L"APIC"); pFrame.has_value()) {
		auto pFramePic = pFrame.value()->dataAs<id3::Frame::Picture>();
		lib::ComPtr<IStream> stream = lib::ComPtr{
			SHCreateMemStream(pFramePic->bin.data(), static_cast<UINT>(pFramePic->bin.size())) };
		if (HRESULT hr = OleLoadPicture(
				stream.ptr(), 0, FALSE, IID_IPicture, reinterpret_cast<void**>(_pic.pptr())); FAILED(hr)) [[unlikely]] {
			lib::NativeControl{this, LBL_PICSIZE}.setText(L"");
			auto err = std::system_category().message(hr);
			dlg.msgBox(L"Picture loading error", {}, lib::str::toWide(err), TDCBF_OK_BUTTON, TD_ERROR_ICON);
		} else {
			OLE_XSIZE_HIMETRIC hmx = 0;
			OLE_YSIZE_HIMETRIC hmy = 0;
			_pic->get_Width(&hmx);
			_pic->get_Height(&hmy);
			auto reso = lib::str::fmt(L"%d x %d pixels",
				lib::dpi::himetricToPixelX(hmx, {}, hWnd()), lib::dpi::himetricToPixelY(hmy, {}, hWnd()));
			lib::NativeControl{this, LBL_PICSIZE}.setText(reso);
		}
	} else {
		lib::NativeControl{this, LBL_PICSIZE}.setText(L"");
	}
}

void DlgEdit::renderFramesList() const
{
	lib::ListView lv{this, LST_FRAMES};
	lv.items.removeAll();

	if (_pTags.size() == 1) {
		for (const id3::Frame& frame : _pTags[0]->frames) {
			auto strFrame = frame.asText();
			lv.items.add(frame.name4, {strFrame});
		}
	} else { // multiple files
		auto msg = lib::str::fmt(L"%d files...", _pTags.size());
		lv.items.add(L"", {msg});
	}
}

void DlgEdit::updateTagsWithTexts() const
{
	for (auto&& field : _Fields) {
		if (!lib::CheckRadio{this, field.chkId}.isChecked())
			continue;

		auto text = lib::NativeControl{this, static_cast<WORD>(field.chkId + 1)}.text();
		lib::str::trim(text);

		for (auto&& pTag : _pTags) {
			if (auto pFrame = pTag->frameByName4(field.name4); pFrame.has_value()) { // the frame already exists in this tag
				if (text.empty()) { // empty text will remove the frame
					lib::vec::removeIf(pTag->frames,
						[&field](const id3::Frame& f) { return lib::str::eqI(f.name4, field.name4); }); // will fail with TXXX frames
				} else {
					pFrame.value()->forceText(text);
				}
			} else { // the frame doesn't exist in this tag yet
				if (!text.empty())
					pTag->frames.emplace_back(field.name4, text);
			}
		}
	}
}
