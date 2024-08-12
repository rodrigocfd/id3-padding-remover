#include "DlgEdit.h"
#include "../res/resource.h"

DlgEdit::FieldInfo DlgEdit::_Fields[] = {
	{CHK_ARTIST, L"TPE1"},
	{CHK_TITLE, L"TIT2"},
	{CHK_SUBTITLE, L"TIT3"},
	{CHK_ALBUM, L"TALB"},
	{CHK_TRACK, L"TRCK"},
	{CHK_YEAR, L"TYER"},
	{CHK_GENRE, L"TCON"},
	{CHK_PERFORMER, L"TPE3"},
	{CHK_PUBLISHER, L"TPUB"},
	{CHK_OARTIST, L"TOPE"},
	{CHK_OALBUM, L"TOAL"},
	{CHK_OYEAR, L"TORY"},
	{CHK_COMPOSER, L"TCOM"},
	{CHK_LYRICIST, L"TEXT"},
	{CHK_COMMENT, L"COMM"},
};

void DlgEdit::_renderTitlebarCounts() const
{
	if (_pTags.size() > 1) {
		setText(lib::str::fmt(L"%s - %d files", text(), _pTags.size()));
	} else {
		setText(lib::str::fmt(L"%s - %d frames", text(), _pTags[0]->frames.size()));
	}
}

void DlgEdit::_renderTextboxes() const
{
	for (auto&& field : _Fields) {
		auto maybeFrame0 = _pTags[0]->frameByName4(field.name4); // assumes at least 1 tag was passed to DlgEdit

		bool isSameValue = lib::vec::allIf(_pTags, [&field, &maybeFrame0](const id3::Tag* pTag) -> bool {
			auto maybeFrameN = pTag->frameByName4(field.name4);
			if (maybeFrame0.has_value() && maybeFrameN.has_value()) {
				return *maybeFrame0.value() == *maybeFrameN.value();
			} else {
				return maybeFrame0 == maybeFrameN;
			}
		});

		if (isSameValue) {
			for (auto&& pTag : _pTags) {
				if (pTag->frameByName4(field.name4).has_value()) { // 1st tag which has this frame
					lib::CheckRadio{this, field.chkId}.checkAndTrigger();
					lib::NativeControl{this, static_cast<WORD>(field.chkId + 1)}.setText(
						pTag->frameByName4(field.name4).value()->toText());
				}
			}
		}
	}
}

void DlgEdit::_renderFrames() const
{
	lib::ListView lv{this, LST_FRAMES};

	if (_pTags.size() == 1) {
		for (const id3::Frame& frame : _pTags[0]->frames) {
			auto strFrame = frame.toText();
			lv.items.add(frame.name4, {strFrame});
		}
	} else { // multiple files
		auto msg = lib::str::fmt(L"%d files...", _pTags.size());
		lv.items.add(L"", {msg});
	}
}
