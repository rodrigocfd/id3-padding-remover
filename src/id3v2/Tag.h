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

	// Parses a tag from an MP3 file.
	explicit Tag(std::wstring_view mp3Path);
	
	bool operator==(const Tag&) const = default;
	[[nodiscard]] const Frame* frameByName4(std::wstring_view name4) const;
	[[nodiscard]] Frame* frameByName4(std::wstring_view name4);
	void removeFrameByName4(std::wstring_view name4);
	[[nodiscard]] LPCWSTR replayGainStatus() const;
	void saveToFile();

	[[nodiscard]] static Frame* SameFrameAcrossAllTags(std::wstring_view name4, const std::vector<Tag*>& tagsToCheck);

private:
	struct HeaderInfo final {
		UINT declaredSize = 0;
		UINT mp3Offset = 0;
	};
	struct FramesInfo final {
		std::vector<Frame> frames;
		UINT padding = 0;
	};

	[[nodiscard]] static HeaderInfo _ParseHeader(std::span<BYTE> src);
	[[nodiscard]] static FramesInfo _ParseFrames(std::span<BYTE> src);
	[[nodiscard]] std::optional<size_t> _apicSize() const;
	[[nodiscard]] std::vector<BYTE> _serialize() const;
};

}
