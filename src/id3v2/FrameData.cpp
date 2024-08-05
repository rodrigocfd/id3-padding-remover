#include <stdexcept>
#include <windlg/lib.h>
#include "FrameData.h"
#include "strEngine.h"
using std::span, std::vector, std::wstring;
using namespace lib;
using namespace id3;

LPCWSTR PicType::toString(Type t)
{
	switch (t) {
		case Type::Other: return L"Other";
		case Type::FileIconPng32: return L"32x32 pixels 'file icon' (PNG only)";
		case Type::FileIconOther: return L"Other file icon";
		case Type::CoverFront: return L"Cover (front)";
		case Type::CoverBack: return L"Cover (back)";
		case Type::Leaflet: return L"Leaflet page";
		case Type::CdLabelSide: return L"Media (e.g. label side of CD)";
		case Type::LeadArtist: return L"Lead artist/lead performer/soloist";
		case Type::Artist: return L"Artist/performer";
		case Type::Conductor: return L"Conductor";
		case Type::Band: return L"Band/Orchestra";
		case Type::Composer: return L"Composer";
		case Type::Lyricist: return L"Lyricist/text writer";
		case Type::RecLocation: return L"Recording Location";
		case Type::DuringRecording: return L"During recording";
		case Type::DuringPerformance: return L"During performance";
		case Type::MovieCapture: return L"Movie/video screen capture";
		case Type::BrightColouredFish: return L"A bright coloured fish";
		case Type::Illustration: return L"Illustration";
		case Type::BandLogo: return L"Band/artist logotype";
		case Type::PublisherLogo: return L"Publisher/Studio logotype";
		default: throw new std::invalid_argument("Unknown picture type");
	}
}


FrameData FrameData::Parse(WCHAR name4[4], span<BYTE> src)
{
	FrameData frameData{};
	if (str::eq(name4, L"COMM")) {
		frameData.val = _ParseComm(src);
	} else if (str::eq(name4, L"APIC")) {
		frameData.val = _ParseApic(src);
	} else if (name4[0] == L'T') {
		vector<wstring> texts = strEngine::parseAny(src);
		switch (texts.size()) {
		[[unlikely]] case 0:
			throw std::runtime_error( str::toAnsi(str::fmt(L"Frame %s contains no texts", name4)) );
		case 1:
			frameData.val = {FrameText{.text = std::move(texts[0])}};
			break;
		case 2:
			frameData.val = {FrameUserText{.descr = std::move(texts[0]), .text = std::move(texts[1])}};
			break;
		[[unlikely]] default:
			throw std::runtime_error( str::toAnsi(str::fmt(L"Frame %s contains %d texts", name4, texts.size())) );
		}
	} else { // anything else is treated as raw binary
		frameData.val = {FrameBinary{.data = {src.begin(), src.end()}}};
	}
	return frameData;
}

FrameComment FrameData::_ParseComm(span<BYTE> src)
{
	BYTE encBy = src[0];
	if (encBy != 0x00 && encBy != 0x01) [[unlikely]] {
		throw std::runtime_error( str::toAnsi(str::fmt(L"Unknown COMM encoding: %d", encBy)) );
	}
	src = src.subspan(1); // skip encoding byte

	FrameComment comm{};
	for (size_t i = 0; i < 3; ++i) comm.lang3[i] = src[i];
	src = src.subspan(3); // skip lang chars

	vector<wstring> texts = strEngine::parseAny(src);
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

FramePicture FrameData::_ParseApic(span<BYTE> src)
{
	FramePicture picture{};


	return picture;
}
