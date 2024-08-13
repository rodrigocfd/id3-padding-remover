#pragma once
#include <windlg/lib.h>
#include "id3v2/Tag.h"

class WndPic : public lib::CustomControl {
public:
	virtual ~WndPic() { }

	constexpr explicit WndPic(const std::vector<id3::Tag*>& pTags) : _pTags{pTags} { }
	WndPic(const WndPic&) = delete;
	WndPic(WndPic&&) = delete;
	WndPic& operator=(const WndPic&) = delete;
	WndPic& operator=(WndPic&&) = delete;

private:
	LRESULT wndProc(UINT uMsg, WPARAM wp, LPARAM lp) override;
	LRESULT onCreate();

	const std::vector<id3::Tag*>& _pTags;
};
