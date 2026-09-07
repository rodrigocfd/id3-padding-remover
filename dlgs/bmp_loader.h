#pragma once
#include "../windlg/lib.hpp"

struct BmpLoader final {
	HBITMAP hBmp = nullptr;
	int cx = 0, cy = 0;

	~BmpLoader() { clear(); }
	void clear();
	void load(const std::vector<BYTE> &bin);
};
