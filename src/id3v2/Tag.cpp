#include <algorithm>
#include <stdexcept>
#include <windlg/lib.h>
#include "Tag.h"
#include "util.h"
using std::optional, std::span, std::vector, std::wstring_view;
using namespace lib;
using namespace id3;

Tag::Tag(wstring_view mp3)
	: path{mp3}
{
	lib::FileMapped f{mp3, lib::FileMapped::Access::ExistingReadOnly};
	_parseBin(f.asSpan());
}

optional<const Frame*> Tag::frameByName4(wstring_view name4) const
{
	for (const Frame& frame : frames) {
		if (lib::str::eqI(frame.name4, name4))
			return &frame;
	}
	return std::nullopt;
}

optional<Frame*> Tag::frameByName4(wstring_view name4)
{
	auto pFrame = const_cast<const Tag*>(this)->frameByName4(name4);
	return pFrame.has_value() ? optional{const_cast<Frame*>(pFrame.value())} : std::nullopt;
}

LPCWSTR Tag::replayGainStatus() const
{
	bool hasTrack = false;
	bool hasAlbum = false;

	for (const Frame& frame : frames) {
		if (hasTrack && hasAlbum) break;

		if (lib::str::eqI(frame.name4, L"TXXX")) {
			if (auto pData = std::get_if<Frame::UserText>(&frame.data); pData) {
				if (lib::str::startsWithI(pData->descr, L"replaygain_track_"))
					hasTrack = true;
				else if (lib::str::startsWithI(pData->descr, L"replaygain_album_"))
					hasAlbum = true;
			}
		}
	}
	
	if (hasTrack && hasAlbum) return L"TA";
	else if (hasTrack) return L"T";
	else if (hasAlbum) return L"A";
	else return L"";
}

void Tag::saveToFile() const
{
	if (path.empty())
		throw std::runtime_error("Tag has no path");

	lib::File fout{path, lib::File::Access::ExistingRW};
	vector<BYTE> currentContents = fout.readAll();
	HeaderInfo headerNfo = _ParseHeader(currentContents);

	fout.setSize(0);
	if (!frames.empty()) {
		vector<BYTE> tagBlob = _serialize();
		fout.write(tagBlob);
	}
	fout.write({currentContents.begin() + headerNfo.mp3Offset, currentContents.end()}); // MP3 data
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
	optional<size_t> maybeMp3Offset = util::positionOf2(src, 0xff, 0xfb); // https://stackoverflow.com/a/7302482/6923555
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
#ifdef _DEBUG
	if (nfo.declaredSize > nfo.mp3Offset) {
		auto msg = str::fmt(L"--- Declared size: %d > offset: %d\n", nfo.declaredSize, nfo.mp3Offset);
		OutputDebugStringW(msg.c_str());
	}
#endif

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

optional<size_t> Tag::_apicSize() const
{
	if (optional<const Frame*> frame = frameByName4(L"APIC"); frame.has_value()) {
		auto pic = std::get_if<Frame::Picture>(&frame.value()->data);
		return pic->bin.size();
	}
	return std::nullopt;
}

vector<BYTE> Tag::_serialize() const
{
	size_t apicSize = _apicSize().value_or(0);

	auto buf = vec::newReserved<BYTE>(10 + 10 * frames.size() + apicSize); // arbitrary
	util::serializeChars(L"ID3", buf); // magic bytes
	vec::append(buf, 0x03, 0x00); // tag version
	vec::append(buf, 0x00); // flags

	size_t offsetSz = buf.size();
	buf.insert(buf.end(), 4, 0x00); // data size placeholder

	size_t szFrames = 0; // won't count 10-byte tag header
	for (const Frame& frame : frames)
		szFrames += frame.serialize(buf);

	util::serializeInPlaceUintBe(util::syncSafe::encode(static_cast<UINT>(szFrames)), buf.begin() + offsetSz);
	return buf;
}
