#include <algorithm>
#include "tag.h"
using namespace id3v2;

Tag::Tag(std::span<const BYTE> src) {
	if (src.empty()) [[unlikely]] { // this is a new, empty file
		return;
	}

	size_t declaredSize = parse_header(src);
	if (!declaredSize) [[unlikely]] {
		return; // MP3 has no ID3v2 tag
	}

	parse_frames(src.subspan(10)); // skip 10-byte tag header
}

size_t Tag::parse_header(std::span<const BYTE> src) {
	// Check ID3 magic bytes.
	const char MAGIC[] = "ID3";
	if (std::memcmp(MAGIC, src.data(), 3) != 0) [[unlikely]] {
		return 0; // MP3 has no tag
	}

	// Validate tag version 2.3.0.
	const BYTE TAG_VER[] = {3, 0};
	if (std::memcmp(TAG_VER, src.subspan(3).data(), 2) != 0) [[unlikely]] {
		throw ParsingError(
			wd::str::fmt(L"Tag version 2.%u.%u is not supported, only 2.3.0.", src[3], src[4]));
	}

	// Validate unsupported flags.
	if (src[5] & 0b1000'0000) [[unlikely]] {
		throw ParsingError(L"Unsynchronised tag not supported.");
	} else if (src[5] & 0b0100'0000) [[unlikely]] {
		throw ParsingError(L"Tag extended header not supported.");
	}

	DWORD declaredSize = misc::syncsafe_dec(misc::parse_dword_be(src.subspan(6)));
	return declaredSize + 10; // also count 10-byte tag header
}

void Tag::parse_frames(std::span<const BYTE> src) {
	_frames.reserve(10); // arbitrary
	_mp3Offset = 10; // start at 10 because src already skipped 10-byte header

	while (!src.empty()) { // will be empty if we're parsing a tag without a whole MP3 file
		if (is_mp3_magic(src)) {
			// We found the beginning of the MP3 file, no padding.
			return;
		}

		if (src[0] == 0x00) {
			// We entered a padding region after all frames.
			// Skip the 1st byte, which is 0x00; don't count last, we're checking 2.
			for (size_t i = 1; i < src.size(); ++i) {
				if (is_mp3_magic(src.subspan(i))) {
					_mp3Offset += i;
					_padding = i;
					return;
				}
			}
			throw ParsingError{L"MP3 offset not found."};
		}

		auto f = Frame{src};
		if (f.declared_size() > src.size()) [[unlikely]] { // means the size was serialized with error
			throw ParsingError{wd::str::fmt(
				L"Declared frame size greater than available size: %u vs %u.", f.declared_size(), src.size())};
		}

		_mp3Offset += f.declared_size();
		src = src.subspan(f.declared_size());
		_frames.emplace_back(std::move(f));
	}
}

bool Tag::is_mp3_magic(std::span<const BYTE> src) {
	// Known magic byte sequences that identify the beginning of a MP3.
	// https://stackoverflow.com/a/7302482/6923555
	// https://en.wikipedia.org/wiki/List_of_file_signatures
	// https://github.com/sindresorhus/file-type/issues/75#issuecomment-320650344

	const BYTE MP3_MAGIC[][2] {
		{0xff, 0xfb}, {0xff, 0xfb}, {0xff, 0xf2}, {0xff, 0xfa}, {0xff, 0xf3},
	};
	for (auto &&magic : MP3_MAGIC) {
		if (std::memcmp(magic, src.data(), 2) == 0)
			return true;
	}
	return false;
}

Tag Tag::clone() const {
	Tag t{};
	t._mp3Offset = _mp3Offset;
	t._padding = _padding;

	t._frames.reserve(_frames.size());
	for (auto &&f : _frames)
		t._frames.emplace_back(f.clone());

	return t;
}

std::vector<BYTE> Tag::serialize() const {
	size_t szTag = 10; // start with 10-byte tag header
	for (auto &&f : _frames)
		szTag += f.serialize_len();

	std::vector<BYTE> buf{};
	buf.reserve(szTag);
	misc::serialize_str(Encoding::iso88591, buf, L"ID3", misc::NullT::no);
	buf.push_back(0x03); // tag version
	buf.push_back(0x00);
	buf.push_back(0x00); // flags
	misc::serialize_dword_be(buf, misc::syncsafe_enc(static_cast<DWORD>(szTag) - 10)); // don't count 10-byte header size

	for (auto &&f : _frames)
		f.serialize(buf);

	return buf;
}

void Tag::save_to_file(wd::StrView filePath) {
	wd::File fout{};
	fout.open_rw(filePath);
	std::vector<BYTE> currentContents = fout.read();
	auto oldTag = Tag{currentContents};

	if (!_frames.empty()) [[likely]] {
		fout.truncate().write(serialize()); // write new tag
	}

	if (!currentContents.empty()) [[likely]] {
		fout.write(currentContents.begin() + oldTag.mp3_offset(), currentContents.end()); // write rest of the MP3
	}
	_padding = 0; // we write no padding
}

void Tag::delete_frames_by_name4(wd::StrView name4) {
	std::erase_if(_frames, [name4](const Frame &f) -> bool {
		return f.name4() == name4;
	});
}

void Tag::delete_replay_gain() {
	std::erase_if(_frames, [](const Frame &f) -> bool {
		if (f.name4() == L"TXXX") {
			if (auto pUserText = dynamic_cast<const BodyUserText*>(f.body()); pUserText) [[likely]] { // should always be
				return wd::str::starts_with_i(pUserText->descr, L"replaygain_track_")
					|| wd::str::starts_with_i(pUserText->descr, L"replaygain_album_");
			}
		}
		return false;
	});
}

const Frame* Tag::frame_by_name4(wd::StrView name4) const {
	for (auto &&f : _frames) {
		if (f.name4() == name4)
			return &f;
	}
	return nullptr; // not found
}

const WCHAR* Tag::replay_gain_status() const {
	bool hasTrack = false, hasAlbum = false;
	for (auto &&f : _frames) {
		if (hasTrack && hasAlbum)
			break;

		if (f.name4() == L"TXXX") {
			if (auto b = dynamic_cast<const BodyUserText*>(f.body()); b) {
				if (wd::str::starts_with_i(b->descr, L"replaygain_track_")) {
					hasTrack = true;
				} else if (wd::str::starts_with_i(b->descr, L"replaygain_album_")) {
					hasAlbum = true;
				}
			}
		}
	}

	if (hasTrack && hasAlbum) {
		return L"TA";
	} else if (hasTrack) {
		return L"T";
	} else if (hasAlbum) {
		return L"A";
	} else {
		return L"";
	}
}


const Frame* id3v2::same_frame_across_all_tags(wd::StrView name4, const std::vector<Tag> &tags) {
	if (tags.empty()) {
		return nullptr; // no tags, no frame
	} else if (tags.size() == 1) {
		return tags[0].frame_by_name4(name4); // only 1 tag, direct query
	}

	const Frame *pFrame0 = tags[0].frame_by_name4(name4); // query the first tag right away
	if (!pFrame0)
		return nullptr; // first tag doesn't have the frame, stop right now

	bool allSame = std::all_of(tags.begin() + 1, tags.end(), [name4, pFrame0](const Tag &t) -> bool {
		const Frame *pFrame = t.frame_by_name4(name4);
		if (!pFrame)
			return false; // frame doesn't exist in this posterior tag
		
		// Compare the textual rendering of this frame with the first frame.
		return pFrame->as_simple_text() == pFrame0->as_simple_text();
	});

	return allSame ? pFrame0 : nullptr;
}

const BodyPicture* id3v2::same_apic_frame_across_all_tags(const std::vector<Tag> &tags) {
	if (const id3v2::Frame *pFrame = id3v2::same_frame_across_all_tags(L"APIC", tags); pFrame) { // do we have single APIC?
		if (auto pApic = dynamic_cast<const id3v2::BodyPicture*>(pFrame->body()); pApic) [[likely]] {
			return pApic;
		} else [[unlikely]] {
			throw ParsingError{L"APIC frame without BodyPicture."}; // should never happen
		}
	}
	return nullptr;
}
