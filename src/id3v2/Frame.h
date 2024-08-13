#pragma once
#include <array>
#include <span>
#include <variant>
#include <vector>
#include <Windows.h>

namespace id3 {

struct Frame final {
	struct Text final {
		std::wstring text;
		constexpr bool operator==(const Text&) const = default;
	};
	struct UserText final {
		std::wstring descr;
		std::wstring text;
		constexpr bool operator==(const UserText&) const = default;
	};
	struct Binary final {
		std::vector<BYTE> bin;
		constexpr bool operator==(const Binary&) const = default;
	};
	struct Comment final {
		WCHAR lang3[4] = {L'\0'};
		std::wstring descr;
		std::wstring text;
		bool operator==(const Comment&) const;
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
		[[nodiscard]] static LPCWSTR TypeToText(Type t);

		std::wstring mime;
		Type type;
		std::wstring descr;
		std::vector<BYTE> bin;
		constexpr bool operator==(const Picture&) const = default;
	};
	using Data = std::variant<Text, UserText, Binary, Comment, Picture>;

	WCHAR name4[5] = {L'\0'};
	UINT declaredSize = 0; // used only at parsing
	std::array<BYTE, 2> flags;
	Data data;

	Frame() = delete;
	Frame(const Frame&) = delete;
	constexpr Frame(Frame&&) = default;
	Frame& operator=(const Frame&) = delete;
	constexpr Frame& operator=(Frame&&) = default;

	explicit Frame(std::span<BYTE> src);
	Frame(std::wstring_view name4, std::wstring_view textContent);

	bool operator==(const Frame&) const;
	[[nodiscard]] std::wstring asText() const;
	void forceText(std::wstring_view text);
	size_t serialize(std::vector<BYTE>& dest) const;

private:
	[[nodiscard]] static Data _ParseData(WCHAR name4[4], std::span<BYTE> src);
	[[nodiscard]] static Comment _ParseComm(std::span<BYTE> src);
	[[nodiscard]] static Picture _ParseApic(std::span<BYTE> src);
	size_t _serializeData(std::vector<BYTE>& dest) const;
};

}
