#include <stdexcept>
#include <windlg/lib.h>
#include "Tag.h"
#include "util.h"
using std::optional, std::span, std::vector, std::wstring_view;
using namespace lib;
using namespace id3;

Tag::Tag(wstring_view mp3)
{
	lib::FileMapped f{mp3, lib::FileMapped::Access::ExistingReadOnly};
	_parseBin(f.asSpan());
}

optional<const Frame*> Tag::frameByName4(std::wstring_view name4) const
{
	for (auto&& frame : frames) {
		if (lib::str::eqI(frame.name4, name4))
			return &frame;
	}
	return std::nullopt;
}

void Tag::_parseBin(span<BYTE> src)
{
	HeaderInfo headerNfo = _ParseHeader(src);
	if (!headerNfo.declaredSize && !headerNfo.mp3Offset)
		return; // MP3 has no ID3v2 tag

	FramesInfo framesNfo = _ParseFrames(src.subspan(10, headerNfo.mp3Offset - 10));
	mp3Offset = headerNfo.mp3Offset;
	padding = framesNfo.padding;
	frames = std::move(framesNfo.frames);
}

Tag::HeaderInfo Tag::_ParseHeader(span<BYTE> src)
{
	HeaderInfo nfo{};

	// Retrieve MP3 offset.
	auto maybeMp3Offset = util::positionOf2(src, 0xff, 0xfb); // https://stackoverflow.com/a/7302482/6923555
	if (!maybeMp3Offset.has_value()) [[unlikely]] {
		throw std::runtime_error("No MP3 signature found");
	}
	nfo.mp3Offset = static_cast<UINT>(maybeMp3Offset.value());

	// Check ID3 magic bytes.
	if ( !(src[0] == 'I' && src[1] == 'D' && src[2] == '3') ) {
		return nfo; // MP3 has no ID3v2 tag
	}

	// Validate tag version.
	if ( !(src[3] == 3 && src[4] == 0) ) { // the "2" of "2.3.0" is not serialized
		throw std::runtime_error( str::toAnsi(str::fmt(L"Tag v2.%d.%d not supported, only v2.3.0", src[3], src[4])) );
	}

	// Validate unsupported flags.
	if (src[5] & 0b1000'0000) {
		throw std::runtime_error("Unsynchronised tag not supported");
	} else if (src[5] & 0b0100'0000) {
		throw std::runtime_error("Tag extended header not supported");
	}

	nfo.declaredSize = util::syncSafe::decode(util::uintFromBeBytes(src.subspan(6, 4)));
	if (nfo.declaredSize > nfo.mp3Offset) {
		auto msg = str::fmt(L"--- Declared size: %d > offset: %d\n", nfo.declaredSize, nfo.mp3Offset);
		OutputDebugStringW(msg.c_str());
	}

	return nfo;
}

Tag::FramesInfo Tag::_ParseFrames(span<BYTE> src)
{
	FramesInfo nfo{};
	for (;;) {
		if (src.empty()) { // end of tag, no padding found
			break;
		} else if (vec::all(src, 0x00)) { // we entered a padding region after all frames
			nfo.padding = static_cast<UINT>(src.size());
			break;
		}

		Frame frame{src};
		if (frame.declaredSize > src.size()) { // means the size was serialized with error
			throw std::runtime_error(
				str::toAnsi(str::fmt(L"Declared frame size greater than available size: %d vs %d",
					frame.declaredSize, src.size())) );
		}

		src = src.subspan(frame.declaredSize);
		nfo.frames.emplace_back(std::move(frame));
	}
	return nfo;
}
