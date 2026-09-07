#pragma once
#include "frame.h"

namespace id3v2 {

	class Tag final {
	private:
		Tag() = default;
		Tag(const Tag&) = delete;
		Tag& operator=(const Tag&) = delete; // non-copyable
	public:
		Tag(Tag&&) = default;
		Tag& operator=(Tag&&) = default; // movable

		explicit Tag(wd::StrView filePath) : Tag{wd::file::read(filePath)} { }
		explicit Tag(std::span<const BYTE> src);
	private:
		[[nodiscard]] static size_t parse_header(std::span<const BYTE> src);
		void parse_frames(std::span<const BYTE> src);
		[[nodiscard]] static bool is_mp3_magic(std::span<const BYTE> src);

	public:
		[[nodiscard]] constexpr size_t mp3_offset() const { return _mp3Offset; }
		[[nodiscard]] constexpr size_t padding() const    { return _padding; }
		[[nodiscard]] Tag clone() const;
		[[nodiscard]] std::vector<BYTE> serialize() const;
		void save_to_file(wd::StrView filePath);

		void delete_frames_by_name4(wd::StrView name4);
		void delete_replay_gain();
		[[nodiscard]] constexpr const std::vector<Frame>& frames() const { return _frames; }
		[[nodiscard]] constexpr std::vector<Frame>& frames()             { return _frames; }
		[[nodiscard]] const Frame* frame_by_name4(wd::StrView name4) const;
		[[nodiscard]] Frame* frame_by_name4(wd::StrView name4) { return const_cast<Frame*>(std::as_const(*this).frame_by_name4(name4)); }
		[[nodiscard]] const WCHAR* replay_gain_status() const;

	private:
		size_t _mp3Offset = 0;
		size_t _padding = 0;
		std::vector<Frame> _frames{};
	};

	[[nodiscard]] const Frame* same_frame_across_all_tags(wd::StrView name4, const std::vector<Tag> &tags);
	[[nodiscard]] const BodyPicture* same_apic_frame_across_all_tags(const std::vector<Tag> &tags);

}
