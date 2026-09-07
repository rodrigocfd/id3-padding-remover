#include "frame.h"
using namespace id3v2;

Frame::Frame(std::span<const BYTE> src) {
	// Parse the 10-byte frame header.
	std::span<const BYTE> srcName4{src.subspan(0, 4)};
	_name4 = misc::parse_str_iso88591(srcName4);
	_declaredSize = misc::parse_dword_be(src.subspan(4)) + 10; // also count 10-byte tag header
	_flags = {src[8], src[9]};

	if (_declaredSize > src.size()) [[unlikely]] {
		_declaredSize = src.size(); // if serialized with error, be complacent
	}

	src = src.subspan(10, _declaredSize - 10); // skip frame header, truncate to declared frame size

	if (_name4 == L"COMM") {
		_body = std::make_unique<BodyComment>(src);
	} else if (_name4 == L"APIC") {
		_body = std::make_unique<BodyPicture>(src);
	} else if (_name4 == L"GEOB") {
		_body = std::make_unique<BodyGeob>(src);
	} else if (_name4 == L"TXXX") {
		_body = std::make_unique<BodyUserText>(src);
	} else if (_name4[0] == L'T' && _name4 != L"TDAT") [[likely]] {
		_body = std::make_unique<BodyText>(src);
	} else { // everything else is treated as raw binary
		_body = std::make_unique<BodyBinary>(src);
	}
}

Frame::Frame(wd::StrView name4, wd::StrView simpleText) {
	_name4 = name4;
	if (_name4 == L"COMM") {
		_body = std::make_unique<BodyComment>(simpleText);
	} else if (_name4[0] == L'T' && _name4 != L"TDAT") [[likely]] {
		_body = std::make_unique<BodyText>(simpleText);
	} else {
		throw ParsingError{L"Invalid body type for creating simple text."};
	}
}

Frame::Frame(const std::vector<BYTE> &picData, wd::StrView mimeType) {
	_name4 = L"APIC";
	_body = std::make_unique<BodyPicture>(picData, mimeType);
}

Frame Frame::clone() const {
	Frame f{};
	f._name4 = _name4;
	f._declaredSize = _declaredSize;
	f._flags = _flags;
	f._body = _body->clone();
	return f;
}

size_t Frame::serialize_len() const {
	return 10 + _body->serialize_len(); // start with 10-byte frame header
}

void Frame::serialize(std::vector<BYTE> &dest) const {
	misc::serialize_str(Encoding::iso88591, dest, _name4, misc::NullT::no);
	misc::serialize_dword_be(dest, static_cast<DWORD>(_body->serialize_len())); // don't count 10-byte header size
	dest.insert(dest.end(), _flags.begin(), _flags.end());
	_body->serialize(dest);
}

void Frame::set_simple_text(wd::StrView text) {
	if (auto t = dynamic_cast<BodyText*>(_body.get()); t) {
		t->text = text;
	} else if (auto u = dynamic_cast<BodyUserText*>(_body.get()); u) {
		u->descr.clear();
		u->text = text;
	} else if (auto c = dynamic_cast<BodyComment*>(_body.get()); c) {
		c->lang3 = L"eng";
		c->descr.clear();
		c->text = text;
	} else {
		throw ParsingError{L"Invalid body type for writing simple text."};
	}
}
