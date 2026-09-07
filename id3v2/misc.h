#pragma once
#include "../windlg/lib.hpp"

namespace id3v2 {

	class ParsingError {
	public:
		explicit ParsingError(wd::StrView reason)
			: _reason{reason} { }
		[[nodiscard]] constexpr const std::wstring& reason() const noexcept { return _reason; }
	private:
		std::wstring _reason{};
	};

	enum class PicType : BYTE {
		other,
		file_icon_png_32,
		file_icon_other,
		cover_front,
		cover_back,
		leaflet,
		cd_label_side,
		lead_artist,
		artist,
		conductor,
		band,
		composer,
		lyricist,
		rec_location,
		during_recording,
		during_performance,
		movie_capture,
		bright_coloured_fish,
		illustration,
		band_logo,
		publisher_logo,
	};

	enum class Encoding : BYTE { iso88591=0x00, unicode=0x01 };

	namespace misc {
		enum class NullT { no, yes };

		[[nodiscard]] const WCHAR* fmt_pictype(PicType picType);

		[[nodiscard]] Encoding     parse_enc(std::span<const BYTE> &src);
		[[nodiscard]] std::wstring parse_str(Encoding enc, std::span<const BYTE> &src);
		[[nodiscard]] std::wstring parse_str_iso88591(std::span<const BYTE> &src);
		[[nodiscard]] std::wstring parse_str_unicode(std::span<const BYTE> &src);
		[[nodiscard]] Encoding     serialize_str_enc(std::initializer_list<wd::StrView> ss);
		[[nodiscard]] size_t       serialize_str_len(Encoding enc, wd::StrView s, NullT nullT);
		void                       serialize_str(Encoding enc, std::vector<BYTE> &dest, wd::StrView s, NullT nullT);

		[[nodiscard]] DWORD parse_dword_be(std::span<const BYTE> src);
		void                serialize_dword_be(std::vector<BYTE> &dest, DWORD num);
		[[nodiscard]] DWORD syncsafe_enc(DWORD n);
		[[nodiscard]] DWORD syncsafe_dec(DWORD n);
	}

}
