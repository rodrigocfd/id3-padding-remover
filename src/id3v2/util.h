#pragma once
#include <optional>
#include <span>
#include <string>
#include <vector>
#include <Windows.h>

namespace id3::util {

[[nodiscard]] std::vector<std::wstring> parseStr(std::span<BYTE> src);
[[nodiscard]] std::vector<std::wstring> parseIso88591(std::span<BYTE> src);
[[nodiscard]] std::vector<std::wstring> parseUnicode(std::span<BYTE> src);

struct SerializedStrs final {
	// 0x00 = ISO 8859-1; 0x01 = Unicode.
	BYTE enc = 0;
	std::vector<BYTE> data;
};
[[nodiscard]] SerializedStrs serializeStrs(std::vector<std::wstring>& strs);

[[nodiscard]] UINT uintFromBeBytes(std::span<BYTE> src);
[[nodiscard]] std::optional<size_t> positionOf2(std::span<BYTE> src, BYTE elem1, BYTE elem2);

namespace syncSafe {
	[[nodiscard]] UINT encode(UINT num);
	[[nodiscard]] UINT decode(UINT num);
}

}
