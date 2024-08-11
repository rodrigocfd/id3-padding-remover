#include "DlgEdit.h"
#include "../res/resource.h"

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
