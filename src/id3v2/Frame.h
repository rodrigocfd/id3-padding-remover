#pragma once
#include <array>
#include <span>
#include <variant>
#include <Windows.h>

namespace id3 {

struct Frame final {
	struct Text final {
		std::wstring text;
	};
	struct UserText final {
		std::wstring descr;
		std::wstring text;
	};
	struct Binary final {
		std::vector<BYTE> bin;
	};
	struct Comment final {
		WCHAR lang3[4] = {L'\0'};
		std::wstring descr;
		std::wstring text;
	};
	struct Picture final {
		enum class Type: BYTE {
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
		[[nodiscard]] static LPCWSTR TypeToString(Type t);

		std::wstring mime;
		Type type;
		std::wstring descr;
		std::vector<BYTE> bin;
	};
	using Data = std::variant<Text, UserText, Binary, Comment, Picture>;

	WCHAR name4[5] = {L'\0'};
	std::array<BYTE, 2> flags;
	UINT declaredSize = 0;
	Data data;

	[[nodiscard]] static Frame Parse(std::span<BYTE> src);

private:
	[[nodiscard]] static Data _ParseData(WCHAR name4[4], std::span<BYTE> src);
	[[nodiscard]] static Comment _ParseComm(std::span<BYTE> src);
	[[nodiscard]] static Picture _ParseApic(std::span<BYTE> src);
};

}
