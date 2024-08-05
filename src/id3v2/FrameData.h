#pragma once
#include <span>
#include <string>
#include <variant>
#include <vector>
#include <Windows.h>

namespace id3 {

enum class PicType: BYTE {
	Other = 0x00,
	FileIconPng32 = 0x01,
	FileIconOther = 0x02,
	CoverFront = 0x03,
	CoverBack = 0x04,
	Leaflet = 0x05,
	CdLabelSide = 0x06,
	LeadArtist = 0x07,
	Artist = 0x08,
	Conductor = 0x09,
	Band = 0x0a,
	Composer = 0x0b,
	Lyricist = 0x0c,
	RecLocation = 0x0d,
	DuringRecording = 0x0e,
	DuringPerformance = 0x0f,
	MovieCapture = 0x10,
	BrightColouredFish = 0x11,
	Illustration = 0x12,
	BandLogo = 0x13,
	PublisherLogo = 0x14,
};
[[nodiscard]] static LPCWSTR picTypeToString(PicType t);


struct FrameText final {
	std::wstring text;
};
struct FrameUserText final {
	std::wstring descr;
	std::wstring text;
};
struct FrameBinary final {
	std::vector<BYTE> data;
};
struct FrameComment final {
	WCHAR lang3[4] = {L'\0'};
	std::wstring descr;
	std::wstring text;
};
struct FramePicture final {
	std::wstring mime;
	PicType type;
	std::wstring descr;
	std::vector<BYTE> data;
};


struct FrameData final {
	std::variant<FrameText, FrameUserText, FrameBinary, FrameComment, FramePicture> val;

	[[nodiscard]] static FrameData Parse(WCHAR name4[4], std::span<BYTE> src);

private:
	[[nodiscard]] static FrameComment _ParseComm(std::span<BYTE> src);
	[[nodiscard]] static FramePicture _ParseApic(std::span<BYTE> src);
};

}
