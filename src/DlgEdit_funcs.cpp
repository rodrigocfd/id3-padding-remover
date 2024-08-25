#include <system_error>
#include "DlgEdit.h"
#include <Shlwapi.h>
#include <olectl.h>
#include "../res/resource.h"
#pragma comment(lib, "Shlwapi.lib")

void DlgEdit::renderTitlebarCounts() const
{
	if (_pTags.size() == 1) {
		setText(lib::str::fmt(L"%s - %d frames", text(), _reorderedFrames.size()));
	} else {
		setText(lib::str::fmt(L"%s - %d files", text(), _pTags.size()));
	}
}

void DlgEdit::renderTextboxes() const
{
	for (auto&& field : _Fields) {
		const id3::Frame* pFrame;
		if (_pTags.size() == 1) {
			pFrame = lib::vec::findIf(_reorderedFrames,
				[&field](const id3::Frame& f) { return lib::str::eqI(field.name4, f.name4); });
		} else {
			pFrame = id3::Tag::SameFrameAcrossAllTags(_pTags, field.name4);
		}

		lib::CheckRadio{this, field.chkId}.checkAndTrigger(pFrame != nullptr);
		lib::NativeControl{this, static_cast<WORD>(field.chkId + 1)}
			.setText(pFrame ? pFrame->asText() : L"");
	}
}

void DlgEdit::loadPicture()
{
	const id3::Frame* pFrame;
	if (_pTags.size() == 1) {
		pFrame = lib::vec::findIf(_reorderedFrames,
			[](const id3::Frame& f) { return lib::str::eqI(f.name4, L"APIC"); });
	} else {
		pFrame = id3::Tag::SameFrameAcrossAllTags(_pTags, L"APIC");
	}

	lib::NativeControl lblPic{this, LBL_PICSIZE};
	if (pFrame) {
		auto pData = pFrame->dataAs<id3::Frame::Picture>();
		lib::ComPtr<IStream> stream = lib::ComPtr{
			SHCreateMemStream(pData->bin.data(), static_cast<UINT>(pData->bin.size())) };
		if (HRESULT hr = OleLoadPicture(
				stream.ptr(), 0, FALSE, IID_IPicture,
				reinterpret_cast<void**>(_pic.pptr())); FAILED(hr)) [[unlikely]] {
			auto err = std::system_category().message(hr);
			dlg.msgBox(L"Picture loading error", {}, lib::str::toWide(err), TDCBF_OK_BUTTON, TD_ERROR_ICON);
			lblPic.setText(L"Image failed to load");
		} else { // image successfully loaded
			OLE_XSIZE_HIMETRIC hmx = 0;
			OLE_YSIZE_HIMETRIC hmy = 0;
			_pic->get_Width(&hmx);
			_pic->get_Height(&hmy);
			auto reso = lib::str::fmt(L"%d x %d pixels",
				lib::dpi::himetricToPixelX(hmx, {}, hWnd()), lib::dpi::himetricToPixelY(hmy, {}, hWnd()));
			lblPic.setText(reso);
		}
	} else { // no APIC frame
		_pic.release();
		lblPic.setText(L"");
	}
}

void DlgEdit::renderFramesList() const
{
	lib::ListView lv{this, LST_FRAMES};
	lv.items.removeAll();
	dlg.enable({LST_FRAMES}, _pTags.size() == 1);

	if (_pTags.size() == 1) {
		for (auto&& frame : _reorderedFrames) {
			auto strFrame = frame.asText();
			lv.items.add(frame.name4, {strFrame});
		}
	} else {
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
			if (auto pFrame = pTag->frameByName4(field.name4); pFrame) { // the frame already exists in this tag
				if (text.empty()) { // empty text will remove the frame
					pTag->removeFrameByName4(field.name4); // note: with TXXX, will remove all TXXX
				} else {
					pFrame->forceText(text);
				}
			} else { // the frame doesn't exist in this tag yet
				if (!text.empty())
					pTag->frames.emplace_back(field.name4, text);
			}
		}
	}
}
