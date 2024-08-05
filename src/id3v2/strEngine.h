#pragma once
#include <span>
#include <string>
#include <vector>
#include <Windows.h>

namespace id3::strEngine {

[[nodiscard]] std::vector<std::wstring> parseAny(std::span<BYTE> src);
[[nodiscard]] std::vector<std::wstring> parseIso88591(std::span<BYTE> src);
[[nodiscard]] std::vector<std::wstring> parseUnicode(std::span<BYTE> src);

struct SerializedStrs final {
	// 0x00 = ISO 8859-1; 0x01 = Unicode.
	BYTE enc = 0;
	std::vector<BYTE> data;
};
[[nodiscard]] SerializedStrs serialize(std::vector<std::wstring>& strs);

[[nodiscard]] UINT uintFromBeBytes(std::span<BYTE> src);

namespace syncSafe {
	[[nodiscard]] UINT encode(UINT num);
	[[nodiscard]] UINT decode(UINT num);
}

}
