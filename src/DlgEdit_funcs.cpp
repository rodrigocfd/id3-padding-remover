#include "DlgEdit.h"
#include "../res/resource.h"

WORD DlgEdit::_Chks[] = {
	CHK_ARTIST,
	CHK_TITLE,
	CHK_SUBTITLE,
	CHK_ALBUM,
	CHK_TRACK,
	CHK_YEAR,
	CHK_GENRE,
	CHK_PERFORMER,
	CHK_PUBLISHER,
	CHK_OARTIST,
	CHK_OALBUM,
	CHK_OYEAR,
	CHK_COMPOSER,
	CHK_LYRICIST,
	CHK_COMMENT,
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
	auto maybeFrame0 = _pTags[0]->frameByName4(L"TALB");

	bool isSame = lib::vec::allIf(_pTags, [&maybeFrame0](const id3::Tag* pTag) -> bool {
		auto maybeFrameN = pTag->frameByName4(L"TALB");
		if (maybeFrame0.has_value() && maybeFrameN.has_value()) {
			auto f0 = maybeFrame0.value();
			auto fN = maybeFrameN.value();
			return *maybeFrame0.value() == *maybeFrameN.value();
		} else {
			return maybeFrame0 == maybeFrameN;
		}
	});


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
