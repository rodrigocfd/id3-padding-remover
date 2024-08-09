#pragma once
#include <optional>
#include <span>
#include <string>
#include <vector>
#include <Windows.h>

namespace id3::util {

// Overload for std::variant's std::visit().
template<class... Ts>
struct Overload : Ts... {
	using Ts::operator()...;
};

[[nodiscard]] std::vector<std::wstring> parseStr(std::span<BYTE> src);
[[nodiscard]] std::vector<std::wstring> parseIso88591(std::span<BYTE> src);
[[nodiscard]] std::vector<std::wstring> parseUnicode(std::span<BYTE> src);

struct SerializedStrs final {
	// 0x00 = ISO 8859-1; 0x01 = Unicode.
	BYTE enc = 0;
	std::vector<BYTE> data;
};
[[nodiscard]] SerializedStrs serializeStrs(std::initializer_list<std::wstring_view> strs);

[[nodiscard]] UINT uintFromBeBytes(std::span<BYTE> src);
void serializeInPlaceUintBe(UINT n, std::vector<BYTE>::iterator dest);
[[nodiscard]] std::optional<size_t> positionOf2(std::span<BYTE> src, BYTE elem1, BYTE elem2);

namespace syncSafe {
	[[nodiscard]] UINT encode(UINT num);
	[[nodiscard]] UINT decode(UINT num);
}

}
