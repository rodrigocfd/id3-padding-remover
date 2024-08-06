#include <stdexcept>
#include <windlg/lib.h>
#include "Frame.h"
#include "util.h"
using std::span, std::vector, std::wstring;
using namespace lib;
using namespace id3;

LPCWSTR Frame::Picture::TypeToString(Type t)
{
	using enum Type;
	switch (t) {
		case Other: return L"Other";
		case FileIconPng32: return L"32x32 pixels 'file icon' (PNG only)";
		case FileIconOther: return L"Other file icon";
		case CoverFront: return L"Cover (front)";
		case CoverBack: return L"Cover (back)";
		case Leaflet: return L"Leaflet page";
		case CdLabelSide: return L"Media (e.g. label side of CD)";
		case LeadArtist: return L"Lead artist/lead performer/soloist";
		case Artist: return L"Artist/performer";
		case Conductor: return L"Conductor";
		case Band: return L"Band/Orchestra";
		case Composer: return L"Composer";
		case Lyricist: return L"Lyricist/text writer";
		case RecLocation: return L"Recording Location";
		case DuringRecording: return L"During recording";
		case DuringPerformance: return L"During performance";
		case MovieCapture: return L"Movie/video screen capture";
		case BrightColouredFish: return L"A bright coloured fish";
		case Illustration: return L"Illustration";
		case BandLogo: return L"Band/artist logotype";
		case PublisherLogo: return L"Publisher/Studio logotype";
		default: throw new std::invalid_argument("Unknown picture type");
	}
}


Frame::Frame(span<BYTE> src)
{
	// Parse the 10-byte frame header.
	for (size_t i = 0; i < 4; ++i) name4[i] = src[i];
	
	declaredSize = util::uintFromBeBytes(src.subspan(4, 4)) + 10; // also count 10-byte frame header
	if (declaredSize > src.size()) // if serialized with error, be complacent
		declaredSize = static_cast<UINT>(src.size());

	flags = {src[8], src[9]};

	// Skip frame header, truncate to declared frame size.
	src = src.subspan(10, declaredSize - 10);

	// Parse the frame contents.
	data = _ParseData(name4, src);
}

Frame::Data Frame::_ParseData(WCHAR name4[4], span<BYTE> src)
{
	if (str::eq(name4, L"COMM")) {
		return _ParseComm(src);
	} else if (str::eq(name4, L"APIC")) {
		return _ParseApic(src);
	} else if (name4[0] == L'T') {
		vector<wstring> texts = util::parseStr(src);
		switch (texts.size()) {
		[[unlikely]] case 0:
			throw std::runtime_error( str::toAnsi(str::fmt(L"Frame %s contains no texts", name4)) );
		case 1:
			return {Text{.text = std::move(texts[0])}};
		case 2:
			return {UserText{.descr = std::move(texts[0]), .text = std::move(texts[1])}};
		[[unlikely]] default:
			throw std::runtime_error( str::toAnsi(str::fmt(L"Frame %s contains %d texts", name4, texts.size())) );
		}
	} else { // anything else is treated as raw binary
		return {Binary{.bin = {src.begin(), src.end()}}};
	}
}

Frame::Comment Frame::_ParseComm(span<BYTE> src)
{
	BYTE encBy = src[0];
	if (encBy != 0x00 && encBy != 0x01) [[unlikely]] {
		throw std::runtime_error( str::toAnsi(str::fmt(L"Unknown COMM encoding: %d", encBy)) );
	}
	src = src.subspan(1); // skip encoding byte

	Comment comm{};
	for (size_t i = 0; i < 3; ++i) comm.lang3[i] = src[i];
	src = src.subspan(3); // skip lang chars

	vector<wstring> texts = util::parseStr(src);
	switch (texts.size()) {
	[[unlikely]] case 0:
		throw std::runtime_error("COMM frame has no texts");
	case 1:
		comm.text = std::move(texts[0]); // in case of 1 text, be lenient and assume empty description
		break;
	case 2:
		comm.descr = std::move(texts[0]);
		comm.text = std::move(texts[1]);
		break;
	[[unlikely]] default:
		throw std::runtime_error( str::toAnsi(str::fmt(L"COMM frame has %d texts", texts.size())) );
	}
	return comm;
}

Frame::Picture Frame::_ParseApic(span<BYTE> src)
{
	BYTE encBy = src[0];
	if (encBy != 0x00 && encBy != 0x01) [[unlikely]] {
		throw std::runtime_error( str::toAnsi(str::fmt(L"Unknown APIC encoding: %d", encBy)) );
	}
	src = src.subspan(1); // skip encoding byte

	Picture picture{};
	vector<span<BYTE>> mimeParts = vec::split(src, 0x00, {2});

	picture.mime = str::newReserved(mimeParts[0].size());
	for (size_t i = 0; i < mimeParts[0].size(); ++i)
		picture.mime += static_cast<WCHAR>(mimeParts[0][i]);

	src = mimeParts[1];
	picture.type = static_cast<Picture::Type>(src[0]);
	src = src.subspan(1); // skip picture type

	if (encBy == 0x00) { // ISO 8859-1
		vector<span<BYTE>> descrParts = vec::split(src, 0x00, {2});
		vector<wstring> texts = util::parseIso88591(descrParts[0]);
		if (!texts.empty()) // description may be absent
			picture.descr = std::move(texts[0]);
		src = descrParts[1];
	} else { // Unicode
		auto idxZero = util::positionOf2(src, 0x00, 0x01).value();
		vector<wstring> texts = util::parseUnicode(src.subspan(0, idxZero));
		if (!texts.empty()) // description may be absent
			picture.descr = std::move(texts[0]);
		src = src.subspan(idxZero + 1);
	}

	picture.bin = {src.begin(), src.end()};
	return picture;
}
