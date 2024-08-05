#include "Frame.h"
#include "util.h"
using std::span;
using namespace id3;

Frame Frame::Parse(span<BYTE> src)
{
	Frame frame{};

	// Parse the 10-byte frame header.
	for (size_t i = 0; i < 4; ++i) frame.name4[i] = src[i];
	frame.declaredSize = util::uintFromBeBytes(src.subspan(4, 4)) + 10; // also count 10-byte frame header
	frame.flags = {src[8], src[9]};

	// Skip frame header, truncate to declared frame size.
	src = src.subspan(10, frame.declaredSize - 10);

	// Parse the frame contents.
	frame.data = FrameData::Parse(frame.name4, src);
	return frame;
}
