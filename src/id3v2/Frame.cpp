#include <stdexcept>
#include <windlg/lib.h>
#include "Frame.h"
#include "util.h"
using std::span, std::vector, std::wstring, std::wstring_view;
using namespace lib;
using namespace id3;

bool Frame::Comment::operator==(const Comment& other) const
{
	return str::eqI(lang3, other.lang3)
		&& descr == other.descr
		&& text == other.text;
}

LPCWSTR Frame::Picture::TypeToText(Type t)
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

Frame::Frame(wstring_view name4, wstring_view textContent)
{
	this->name4 = name4;

	if (str::eqI(name4, L"COMM")) { // comment frame
		data = Comment{
			.lang3 = L"eng",
			.text = textContent.data(),
		};
	} else {
		data = Text{ // assume simple text frame
			.text = textContent.data(),
		};
	}
}

bool Frame::operator==(const Frame& other) const
{
	return str::eqI(name4, other.name4) // note: declaredSize is not compared
		&& flags == other.flags
		&& data == other.data;
}

wstring Frame::asText() const
{
	return std::visit(util::Overload{
		[](const Text& t) {
			return t.text;
		},
		[](const UserText& ut) {
			return ut.descr + L" " + ut.text;
		},
		[](const Binary& b) {
			return str::fmtBytes(b.bin.size());
		},
		[](const Comment& c) {
			return c.descr.empty() ? c.text : (c.descr + L" " + c.text);
		},
		[](const Picture& p) {
			return str::fmt(L"%s %s %s",
				Picture::TypeToText(p.type), p.mime, str::fmtBytes(p.bin.size()));
		},
	}, data);
}

void Frame::forceText(wstring_view text)
{
	std::visit(util::Overload{
		[&text](Text& t) {
			t.text = text;
		},
		[&text](UserText& ut) {
			ut.descr = L"";
			ut.text = text;
		},
		[](Binary& b) {
			throw std::invalid_argument("Can't assign text to binary frame");
		},
		[&text](Comment& c) {
			c.lang3 = L"eng";
			c.descr = L"";
			c.text = text;
		},
		[](Picture& p) {
			throw std::invalid_argument("Can't assign text to picture frame");
		},
	}, data);
}

size_t Frame::serialize(vector<BYTE>& dest) const
{
	dest.reserve(dest.size() + 10);
	util::serializeChars(name4, dest); // no terminating null

	size_t offsetSz = dest.size();
	dest.insert(dest.end(), 4, 0x00); // data size placeholder

	vec::append(dest, flags);

	size_t sz = _serializeData(dest); // won't count 10-byte frame header
	util::serializeInPlaceUintBe(static_cast<UINT>(sz), dest.begin() + offsetSz);
	return sz + 10; // count 10-byte frame header
}

Frame::Data Frame::_ParseData(wstring_view name4, span<BYTE> src)
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
	for (BYTE by : mimeParts[0])
		picture.mime += static_cast<WCHAR>(by);

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
		size_t idxZero = util::positionOf2(src, 0x00, 0x01).value();
		vector<wstring> texts = util::parseUnicode(src.subspan(0, idxZero));
		if (!texts.empty()) // description may be absent
			picture.descr = std::move(texts[0]);
		src = src.subspan(idxZero + 1);
	}

	picture.bin = {src.begin(), src.end()};
	return picture;
}

size_t Frame::_serializeData(vector<BYTE>& dest) const
{
	return std::visit(util::Overload{
		[&dest](const Text& t) {
			auto serializedStrs = util::serializeStrs({t.text});
			dest.push_back(serializedStrs.enc);
			vec::append(dest, serializedStrs.data);
			return 1 + serializedStrs.data.size();
		},
		[&dest](const UserText& ut) {
			auto serializedStrs = util::serializeStrs({ut.descr, ut.text});
			dest.push_back(serializedStrs.enc);
			vec::append(dest, serializedStrs.data);
			return 1 + serializedStrs.data.size();
		},
		[&dest](const Binary& b) {
			vec::append(dest, b.bin);
			return b.bin.size();
		},
		[&dest](const Comment& c) {
			auto serializedStrs = util::serializeStrs({c.descr, c.text});
			dest.push_back(serializedStrs.enc);
			util::serializeChars(c.lang3, dest);
			vec::append(dest, serializedStrs.data);
			return 1 + 3 + serializedStrs.data.size();
		},
		[&dest](const Picture& p) {
			auto serializedStrs = util::serializeStrs({p.descr});
			dest.push_back(serializedStrs.enc);
			util::serializeChars(p.mime, dest);
			dest.push_back(0x00);
			dest.push_back(static_cast<BYTE>(p.type));
			vec::append(dest, serializedStrs.data);
			vec::append(dest, p.bin);
			return 1 + p.mime.length() + 1 + 1 + serializedStrs.data.size() + p.bin.size();
		},
	}, data);
}
