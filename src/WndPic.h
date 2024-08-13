#pragma once
#include <windlg/lib.h>

class WndPic : public lib::CustomControl {
public:
	virtual ~WndPic() { }

	constexpr WndPic() = default;
	WndPic(const WndPic&) = delete;
	WndPic(WndPic&&) = delete;
	WndPic& operator=(const WndPic&) = delete;
	WndPic& operator=(WndPic&&) = delete;

private:
	LRESULT wndProc(UINT uMsg, WPARAM wp, LPARAM lp) override;
	LRESULT onCreate();
};
