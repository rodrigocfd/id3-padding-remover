#pragma once
#include <array>
#include <span>
#include <Windows.h>
#include "FrameData.h"

namespace id3 {

struct Frame final {
	WCHAR name4[5] = {L'\0'};
	std::array<BYTE, 2> flags;
	UINT declaredSize = 0;
	FrameData data;

	[[nodiscard]] static Frame Parse(std::span<BYTE> src);
};

}
