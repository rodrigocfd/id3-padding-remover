#pragma once
#include <span>
#include <vector>
#include <Windows.h>
#include "Frame.h"

namespace id3 {

struct Tag final {
	UINT mp3Offset = 0;
	UINT padding = 0;
	std::vector<Frame> frames;

	[[nodiscard]] static Tag Parse(std::span<BYTE> src);

private:
	struct HeaderInfo final {
		UINT declaredSize = 0;
		UINT mp3Offset = 0;
	};
	struct FramesInfo final {
		std::vector<Frame> frames;
		UINT padding = 0;
	};

	[[nodiscard]] static HeaderInfo _ParseHeader(std::span<BYTE> src);
	[[nodiscard]] static FramesInfo _ParseFrames(std::span<BYTE> src);
};

}
