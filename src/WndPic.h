#pragma once
#include <windlg/lib.h>
#include <ocidl.h>
#include "id3v2/Tag.h"

class WndPic : public lib::CustomControl {
public:
	virtual ~WndPic() { }

	constexpr explicit WndPic(const std::vector<id3::Tag*>& pTags, const lib::ComPtr<IPicture>& pic)
		: _pTags{pTags}, _pic{pic} { }
	WndPic(const WndPic&) = delete;
	WndPic(WndPic&&) = delete;
	WndPic& operator=(const WndPic&) = delete;
	WndPic& operator=(WndPic&&) = delete;

private:
	LRESULT wndProc(UINT uMsg, WPARAM wp, LPARAM lp) override;
	LRESULT onPaint();

	const std::vector<id3::Tag*>& _pTags;
	const lib::ComPtr<IPicture>& _pic;
};
