#pragma once
#include <optional>
#include <span>
#include <vector>
#include <Windows.h>
#include "Frame.h"

namespace id3 {

struct Tag final {
	std::wstring path;
	UINT mp3Offset = 0;
	UINT padding = 0;
	std::vector<Frame> frames;

	Tag() = delete;
	Tag(const Tag&) = delete;
	constexpr Tag(Tag&&) = default;
	Tag& operator=(const Tag&) = delete;
	constexpr Tag& operator=(Tag&&) = default;

	explicit Tag(std::wstring_view mp3);
	
	bool operator==(const Tag&) const = default;
	[[nodiscard]] std::optional<const Frame*> frameByName4(std::wstring_view name4) const;
	[[nodiscard]] std::optional<Frame*> frameByName4(std::wstring_view name4);
	[[nodiscard]] LPCWSTR replayGainStatus() const;
	void saveToFile() const;

	[[nodiscard]] static bool FrameHasSameValueAcrossAllTags(const std::vector<Tag*>& tags, std::wstring_view name4);

private:
	struct HeaderInfo final {
		UINT declaredSize = 0;
		UINT mp3Offset = 0;
	};
	struct FramesInfo final {
		std::vector<Frame> frames;
		UINT padding = 0;
	};

	void _parseBin(std::span<BYTE> src);
	[[nodiscard]] static HeaderInfo _ParseHeader(std::span<BYTE> src);
	[[nodiscard]] static FramesInfo _ParseFrames(std::span<BYTE> src);
	[[nodiscard]] std::optional<size_t> _apicSize() const;
	[[nodiscard]] std::vector<BYTE> _serialize() const;
};

}
