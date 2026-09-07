#include "body.h"
using namespace id3v2;

BodyText::BodyText(std::span<const BYTE> src) {
	Encoding enc = misc::parse_enc(src);
	text = misc::parse_str(enc, src);
}

std::unique_ptr<Body> BodyText::clone() const {
	auto b = new BodyText{};
	b->text = text;
	return std::unique_ptr<Body>(b);
}

size_t BodyText::serialize_len() const {
	Encoding enc = misc::serialize_str_enc({text});
	size_t szText = misc::serialize_str_len(enc, text, misc::NullT::yes);
	return 1 + szText;
}

void BodyText::serialize(std::vector<BYTE> &dest) const {
	Encoding enc = misc::serialize_str_enc({text});
	dest.push_back(static_cast<BYTE>(enc));
	misc::serialize_str(enc, dest, text, misc::NullT::yes);
}


BodyUserText::BodyUserText(std::span<const BYTE> src) {
	Encoding enc = misc::parse_enc(src);
	descr = misc::parse_str(enc, src);
	text = misc::parse_str(enc, src);
}

std::unique_ptr<Body> BodyUserText::clone() const {
	auto u = new BodyUserText{};
	u->descr = descr;
	u->text = text;
	return std::unique_ptr<Body>(u);
}

size_t BodyUserText::serialize_len() const {
	Encoding enc = misc::serialize_str_enc({descr, text});
	size_t szDescr = misc::serialize_str_len(enc, descr, misc::NullT::yes);
	size_t szText = misc::serialize_str_len(enc, text, misc::NullT::yes);
	return 1 + szDescr + szText;
}

void BodyUserText::serialize(std::vector<BYTE> &dest) const {
	Encoding enc = misc::serialize_str_enc({descr, text});
	dest.push_back(static_cast<BYTE>(enc));
	misc::serialize_str(enc, dest, descr, misc::NullT::yes);
	misc::serialize_str(enc, dest, text, misc::NullT::yes);
}


BodyBinary::BodyBinary(std::span<const BYTE> src) {
	bin.insert(bin.end(), src.begin(), src.end());
}

std::unique_ptr<Body> BodyBinary::clone() const {
	auto b = new BodyBinary{};
	b->bin = bin;
	return std::unique_ptr<Body>(b);
}

size_t BodyBinary::serialize_len() const {
	return bin.size();
}

void BodyBinary::serialize(std::vector<BYTE> &dest) const {
	dest.insert(dest.end(), bin.begin(), bin.end());
}


BodyComment::BodyComment(std::span<const BYTE> src) {
	Encoding enc = misc::parse_enc(src);
	lang3 = misc::parse_str_iso88591(src);
	descr = misc::parse_str(enc, src);
	text = misc::parse_str(enc, src);
	if (text.empty()) // if no text, we actually have no descr
		text = std::move(descr);
}

std::unique_ptr<Body> BodyComment::clone() const {
	auto c = new BodyComment{};
	c->lang3 = lang3;
	c->descr = descr;
	c->text = text;
	return std::unique_ptr<Body>(c);
}

size_t BodyComment::serialize_len() const {
	Encoding enc = misc::serialize_str_enc({descr, text});
	size_t szDescr = misc::serialize_str_len(enc, descr, misc::NullT::yes);
	size_t szText = misc::serialize_str_len(enc, text, misc::NullT::yes);
	return 1 + 3 + szDescr + szText;
}

void BodyComment::serialize(std::vector<BYTE> &dest) const {
	Encoding enc = misc::serialize_str_enc({descr, text});
	dest.push_back(static_cast<BYTE>(enc));
	misc::serialize_str(Encoding::iso88591, dest, lang3, misc::NullT::no);
	misc::serialize_str(enc, dest, descr, misc::NullT::yes);
	misc::serialize_str(enc, dest, text, misc::NullT::yes);
}


BodyPicture::BodyPicture(std::span<const BYTE> src) {
	Encoding enc = misc::parse_enc(src);
	mime = misc::parse_str_iso88591(src);
	picType = static_cast<PicType>(src[0]);
	src = src.subspan(1);
	descr = misc::parse_str(enc, src);
	bin.insert(bin.end(), src.begin(), src.end());
}

std::unique_ptr<Body> BodyPicture::clone() const {
	auto p = new BodyPicture{};
	p->mime = mime;
	p->picType = picType;
	p->descr = descr;
	p->bin = bin;
	return std::unique_ptr<Body>(p);
}

std::wstring BodyPicture::as_simple_text() const {
	return wd::str::fmt(L"%s %s %s",
		id3v2::misc::fmt_pictype(picType),
		mime,
		wd::str::fmt_bytes(bin.size()));
}

size_t BodyPicture::serialize_len() const {
	Encoding enc = misc::serialize_str_enc({descr});
	size_t szMime = misc::serialize_str_len(Encoding::iso88591, mime, misc::NullT::yes);
	size_t szDescr = misc::serialize_str_len(enc, descr, misc::NullT::yes);
	return 1 + szMime + 1 + szDescr + bin.size();
}

void BodyPicture::serialize(std::vector<BYTE> &dest) const {
	Encoding enc = misc::serialize_str_enc({descr});
	dest.push_back(static_cast<BYTE>(enc));
	misc::serialize_str(Encoding::iso88591, dest, mime, misc::NullT::yes);
	dest.push_back(static_cast<BYTE>(picType));
	misc::serialize_str(enc, dest, descr, misc::NullT::yes);
	dest.insert(dest.end(), bin.begin(), bin.end());
}


BodyGeob::BodyGeob(std::span<const BYTE> src) {
	Encoding enc = misc::parse_enc(src);
	mime = misc::parse_str_iso88591(src);
	fileName = misc::parse_str(enc, src);
	descr = misc::parse_str(enc, src);
	encObj.insert(encObj.end(), src.begin(), src.end());
}

std::unique_ptr<Body> BodyGeob::clone() const {
	auto g = new BodyGeob{};
	g->mime = mime;
	g->fileName = fileName;
	g->descr = descr;
	g->encObj = encObj;
	return std::unique_ptr<Body>(g);
}

std::wstring BodyGeob::as_simple_text() const {
	std::wstring buf{};
	buf.reserve(mime.length() + fileName.length() + descr.length() + 20);

	if (!mime.empty())
		buf += mime + L" ";
	if (!fileName.empty())
		buf += fileName + L" ";
	if (!descr.empty())
		buf += descr + L" ";
	buf += wd::str::fmt_bytes(encObj.size());
	return buf;
}

size_t BodyGeob::serialize_len() const {
	Encoding enc = misc::serialize_str_enc({fileName, descr});
	size_t szMime = misc::serialize_str_len(Encoding::iso88591, mime, misc::NullT::yes);
	size_t szFileName = misc::serialize_str_len(enc, fileName, misc::NullT::yes);
	size_t szDescr = misc::serialize_str_len(enc, descr, misc::NullT::yes);
	return 1 + szMime + szFileName + szDescr + encObj.size();
}

void BodyGeob::serialize(std::vector<BYTE> &dest) const {
	Encoding enc = misc::serialize_str_enc({fileName, descr});
	dest.push_back(static_cast<BYTE>(enc));
	misc::serialize_str(Encoding::iso88591, dest, mime, misc::NullT::yes);
	misc::serialize_str(enc, dest, fileName, misc::NullT::yes);
	misc::serialize_str(enc, dest, descr, misc::NullT::yes);
	dest.insert(dest.end(), encObj.begin(), encObj.end());
}
