#include <algorithm>
#include <stdexcept>
#include <windlg/lib.h>
#include "util.h"
using std::optional, std::span, std::vector, std::wstring, std::wstring_view;
using namespace lib;
using namespace id3;

constexpr WORD BOM_LE = 0xfeff;
constexpr WORD BOM_BE = 0xfffe;

vector<wstring> util::parseStr(span<BYTE> src)
{
	switch (src[0]) {
		case 0x00: return parseIso88591(src.subspan(1));
		case 0x01: return parseUnicode(src.subspan(1));
		default:   throw std::invalid_argument(str::toAnsi( str::fmt(L"Unrecognized encoding: %d", src[0]) ));
	}
}

vector<wstring> util::parseIso88591(span<BYTE> src)
{
	auto idxLastNonZero = vec::positionRevIf(src, [](const BYTE& by) -> bool { return by != 0x00; });
	if (idxLastNonZero.has_value())
		src = src.subspan(0, idxLastNonZero.value() + 1); // right-trim zeros to avoid an extra empty string
	if (src.empty())
		return {}; // no strings

	vector<span<BYTE>> blocks = vec::split(src, 0x00);
	vector<wstring> texts = vec::newReserved<wstring>(blocks.size());

	for (auto&& block : blocks) {
		if (block.empty()) {
			texts.emplace_back(); // empty strings are also added
		} else {
			auto buf = str::newReserved(block.size());
			for (auto&& by : block)
				buf += static_cast<WCHAR>(by);
			texts.emplace_back(std::move(buf));
		}
	}
	return texts;
}

vector<wstring> util::parseUnicode(span<BYTE> src)
{
	if (src.size() % 2) [[unlikely]] {
		// Length is not even, something is not quite right.
		// Discard last byte and hope for the best.
		src = src.subspan(0, src.size() - 1);
	}

	span<WORD> wsrc{reinterpret_cast<WORD*>(src.data()), src.size() / 2};

	auto idxLastNonZero = vec::positionRevIf(wsrc, [](const WORD& ch) -> bool { return ch != 0x0000; });
	if (idxLastNonZero.has_value())
		wsrc = wsrc.subspan(0, idxLastNonZero.value() + 1); // right-trim zeros to avoid an extra empty string
	if (wsrc.empty())
		return {}; // no strings

	vector<span<WORD>> blocks = vec::split(wsrc, 0x0000);
	vector<wstring> texts = vec::newReserved<wstring>(blocks.size());

	for (auto&& block : blocks) {
		bool isLE = true; // little-endian by default
		if (block[0] == BOM_LE || block[0] == BOM_BE) { // we have a BOM
			if (block[0] == BOM_LE)
				isLE = false;
			block = block.subspan(1); // skip BOM
		}

		if (block.empty()) {
			texts.emplace_back(); // empty strings are also added
		} else {
			auto buf = str::newReserved(block.size());
			for (auto&& ch : block)
				buf += (isLE ? MAKEWORD(HIWORD(ch), LOWORD(ch)) : ch);
			texts.emplace_back(std::move(buf));
		}
	}
	return texts;
}

util::SerializedStrs util::serializeStrs(std::vector<std::wstring>& strs)
{
	bool isUnicode = false;
	size_t estimatedLenBytes = 0;

	for (const auto& str : strs) {
		estimatedLenBytes += str.length() + 1; // all strings will be null-terminated
		if (!isUnicode) { // we still don't know if it will be Unicode
			bool hasUnicodeCh = std::any_of(str.begin(), str.end(), [](const WCHAR& ch) -> bool { return ch > 0xff; });
			if (hasUnicodeCh) // at least 1 string is Unicode
				isUnicode = true;
		}
	}

	if (isUnicode) { // chars will be serialized as WORD
		estimatedLenBytes *= 2;
		estimatedLenBytes += 2 * strs.size(); // one BOM to each string
	}

	auto buf = vec::newReserved<BYTE>(estimatedLenBytes);
	for (const auto& str : strs) {
		if (isUnicode) {
			// Insert BOM bytes for each string.
			// Strings will be encoded as little-endian.
			vec::append(buf, LOBYTE(BOM_LE), HIBYTE(BOM_BE));
		}

		for (WCHAR ch : str) { // write each char of the string
			if (isUnicode) {
				vec::append(buf, LOBYTE(ch), HIBYTE(ch));	
			} else {
				buf.push_back(static_cast<BYTE>(ch));
			}
		}

		if (isUnicode) {
			vec::append(buf, 0x00, 0x00); // append terminating null
		} else {
			buf.push_back(0x00);
		}
	}

	return {
		.enc = static_cast<BYTE>(isUnicode ? 0x01 : 0x00),
		.data = std::move(buf),
	};
}

UINT util::uintFromBeBytes(span<BYTE> src)
{
	if (src.size() != 4) [[unlikely]] {
		throw std::invalid_argument("UINT must be converted from 4 bytes");
	}
	return MAKELONG(MAKEWORD(src[3], src[2]), MAKEWORD(src[1], src[0]));
}

optional<size_t> util::positionOf2(span<BYTE> src, BYTE elem1, BYTE elem2)
{
	for (size_t i = 0; i < src.size() - 1; ++i) {
		if (src[i] == elem1 && src[i + 1] == elem2)
			return {i};
	}
	return std::nullopt;
}



UINT util::syncSafe::encode(UINT num)
{
	int out, mask = 0x7f;
	while (mask ^ 0x7fff'ffff) {
		out = num & ~mask;
		out <<= 1;
		out |= num & mask;
		mask = ((mask + 1) << 8) - 1;
		num = out;
	}
	return out;
}

UINT util::syncSafe::decode(UINT num)
{
	int out = 0, mask = 0x7f00'0000;
	while (mask) {
		out >>= 1;
		out |= num & mask;
		mask >>= 8;
	}
	return out;
}
