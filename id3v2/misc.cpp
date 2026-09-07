#include <algorithm>
#include "misc.h"
using namespace id3v2;

const WORD BOM_BE = 0xfeff;
const WORD BOM_LE = 0xfffe;

const WCHAR* id3v2::misc::fmt_pictype(PicType picType) {
	switch (picType) {
	using enum PicType;
		case other:                return L"Other";
		case file_icon_png_32:     return L"32x32 pixels 'file icon' (PNG only)";
		case file_icon_other:      return L"Other file icon";
		case cover_front:          return L"Cover (front)";
		case cover_back:           return L"Cover (back)";
		case leaflet:              return L"Leaflet page";
		case cd_label_side:        return L"Media (e.g. label side of CD)";
		case lead_artist:          return L"Lead artist/lead performer/soloist";
		case artist:               return L"Artist/performer";
		case conductor:            return L"Conductor";
		case band:                 return L"Band/Orchestra";
		case composer:             return L"Composer";
		case lyricist:             return L"Lyricist/text writer";
		case rec_location:         return L"Recording Location";
		case during_recording:     return L"During recording";
		case during_performance:   return L"During performance";
		case movie_capture:        return L"Movie/video screen capture";
		case bright_coloured_fish: return L"A bright coloured fish";
		case illustration:         return L"Illustration";
		case band_logo:            return L"Band/artist logotype";
		case publisher_logo:       return L"Publisher/Studio logotype";
		default:                   throw ParsingError{wd::str::fmt(L"Invalid picture type: %u.", static_cast<BYTE>(picType))}; // should never happen
	}
}

id3v2::Encoding id3v2::misc::parse_enc(std::span<const BYTE> &src) {
	Encoding enc = static_cast<Encoding>(src[0]);
	if (enc != Encoding::iso88591 && enc != Encoding::unicode) [[unlikely]] {
		throw ParsingError{wd::str::fmt(L"Unknown encoding: %u.", enc)};
	}
	src = src.subspan(1);
	return enc;
}

std::wstring id3v2::misc::parse_str(Encoding enc, std::span<const BYTE> &src) {
	switch (enc) {
	using enum Encoding;
		case iso88591: return parse_str_iso88591(src);
		case unicode:  return parse_str_unicode(src);
		default:       throw ParsingError{wd::str::fmt(L"Unknown parsing encoding: %u.", enc)}; // should never happen
	}
}

std::wstring id3v2::misc::parse_str_iso88591(std::span<const BYTE> &src) {
	if (src.empty())
		return {};

	std::span<const BYTE>::iterator itZero = std::find_if(src.begin(), src.end(), [](BYTE b) {
		return b == 0x00;
	});
	size_t szNull = (itZero != src.end()) ? 1 : 0; // in bytes

	size_t nChars = std::distance(src.begin(), itZero);
	std::wstring buf{};
	buf.reserve(nChars); // alloc receiving buffer

	for (auto it = src.begin(); it != itZero; ++it)
		buf.push_back(*it); // brute force conversion

	src = src.subspan(nChars + szNull); // advance span
	return buf;
}

std::wstring id3v2::misc::parse_str_unicode(std::span<const BYTE> &src) {
	std::span<const WORD> wsrc{reinterpret_cast<const WORD*>(src.data()), src.size() / sizeof(WORD)}; // will discard an odd byte
	if (wsrc.empty())
		return {};

	bool isLE = true; // default
	size_t szBom = 0; // in bytes
	if (wsrc[0] == BOM_BE || wsrc[0] == BOM_LE) { // we have a BOM
		szBom = 1 * sizeof(WORD);
		isLE = wsrc[0] == BOM_LE;
		wsrc = wsrc.subspan(1); // skip BOM
	}

	if (wsrc.empty())
		return {};

	std::span<const WORD>::iterator itZero = std::find_if(wsrc.begin(), wsrc.end(), [](WORD w) {
		return w == 0x0000;
	});
	size_t szNull = (itZero != wsrc.end()) ? (1 * sizeof(WORD)) : 0; // in bytes

	size_t nChars = std::distance(wsrc.begin(), itZero);
	std::wstring buf{};
	buf.reserve(nChars); // alloc receiving buffer

	for (auto it = wsrc.begin(); it != itZero; ++it) {
		WORD ch = isLE ? MAKEWORD(HIBYTE(*it), LOBYTE(*it)) : *it;
		buf.push_back(ch);
	}

	src = src.subspan(szBom + (nChars * sizeof(WORD)) + szNull); // advance span
	return buf;
}

Encoding id3v2::misc::serialize_str_enc(std::initializer_list<wd::StrView> ss) {
	for (auto &&s : ss) {
		for (auto &&ch : s) {
			if (ch > 0xff)
				return Encoding::unicode;
		}
	}
	return Encoding::iso88591;
}

size_t id3v2::misc::serialize_str_len(Encoding enc, wd::StrView s, NullT nullT) {
	size_t szNull = (nullT == NullT::no) ? 0 : 1; // terminating null, if requested
	switch (enc) {
	using enum Encoding;
		case iso88591: return s.length() + szNull;
		case unicode:  return (1 + s.length() + szNull) * sizeof(WORD); // plus BOM
		default:       throw ParsingError{wd::str::fmt(L"Unknown serializing encoding: %u.", enc)}; // should never happen
	}
}

void id3v2::misc::serialize_str(Encoding enc, std::vector<BYTE> &dest, wd::StrView s, NullT nullT) {
	if (enc == Encoding::unicode) {
		dest.push_back(LOBYTE(BOM_LE)); // BOM bytes; we serialize as little-endian
		dest.push_back(HIBYTE(BOM_LE));
	}

	for (auto &&ch : s) { // each char, no terminating null
		switch (enc) {
		case Encoding::iso88591:
			dest.push_back(static_cast<BYTE>(ch));
			break;
		case Encoding::unicode:
			dest.push_back(LOBYTE(ch)); // little-endian
			dest.push_back(HIBYTE(ch));
		}
	}

	if (nullT == NullT::yes) { // terminating null, if requested
		dest.push_back(0x00);
		if (enc == Encoding::unicode)
			dest.push_back(0x00);
	}
}

DWORD id3v2::misc::parse_dword_be(std::span<const BYTE> src) {
	if (src.size() < sizeof(DWORD)) [[unlikely]] {
		throw ParsingError{wd::str::fmt(L"DWORD BR cannot be read from %u bytes.", src.size())};
	}
	return MAKELONG(MAKEWORD(src[3], src[2]), MAKEWORD(src[1], src[0]));
}

void id3v2::misc::serialize_dword_be(std::vector<BYTE> &dest, DWORD num) {
	dest.push_back(HIBYTE(HIWORD(num)));
	dest.push_back(LOBYTE(HIWORD(num)));
	dest.push_back(HIBYTE(LOWORD(num)));
	dest.push_back(LOBYTE(LOWORD(num)));
}


DWORD id3v2::misc::syncsafe_enc(DWORD num) {
	DWORD out = 0, mask = 0x7f;
	while (mask ^ 0x7fff'ffff) {
		out = num & ~mask;
		out <<= 1;
		out |= num & mask;
		mask = ((mask + 1) << 8) - 1;
		num = out;
	}
	return out;
}

DWORD id3v2::misc::syncsafe_dec(DWORD num) {
	DWORD out = 0, mask = 0x7f00'0000;
	while (mask) {
		out >>= 1;
		out |= num & mask;
		mask >>= 8;
	}
	return out;
}
