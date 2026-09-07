#pragma once
#include <memory>
#include "../windlg/lib.hpp"
#include "misc.h"

namespace id3v2 {

	struct Body {
		virtual ~Body() = default;
		Body() = default;

		Body(const Body&) = delete;
		Body& operator=(const Body&) = delete; // non-copyable
		Body(Body&&) = default;
		Body& operator=(Body&&) = default; // movable

		virtual std::unique_ptr<Body> clone() const = 0;
		[[nodiscard]] virtual std::wstring as_simple_text() const = 0;
		[[nodiscard]] virtual size_t serialize_len() const = 0;
		virtual void serialize(std::vector<BYTE> &dest) const = 0;
	};

	struct BodyText final : public Body {
	private:
		BodyText() = default;
	public:
		explicit BodyText(std::span<const BYTE> src);
		explicit BodyText(wd::StrView simpleText)
			: text{simpleText} { }

		std::unique_ptr<Body> clone() const override;
		[[nodiscard]] std::wstring as_simple_text() const override { return text; }
		[[nodiscard]] size_t serialize_len() const override;
		void serialize(std::vector<BYTE> &dest) const override;

		std::wstring text{};
	};

	struct BodyUserText final : public Body {
	private:
		BodyUserText() = default;
	public:
		explicit BodyUserText(std::span<const BYTE> src);

		std::unique_ptr<Body> clone() const override;
		[[nodiscard]] std::wstring as_simple_text() const override { return descr + L" " + text; }
		[[nodiscard]] size_t serialize_len() const override;
		void serialize(std::vector<BYTE> &dest) const override;

		std::wstring descr{}, text{};
	};

	struct BodyBinary final : public Body {
	private:
		BodyBinary() = default;
	public:
		explicit BodyBinary(std::span<const BYTE> src);

		std::unique_ptr<Body> clone() const override;
		[[nodiscard]] std::wstring as_simple_text() const override { return wd::str::fmt_bytes(bin.size()); }
		[[nodiscard]] size_t serialize_len() const override;
		void serialize(std::vector<BYTE> &dest) const override;

		std::vector<BYTE> bin{};
	};

	struct BodyComment final : public Body {
	private:
		BodyComment() = default;
	public:
		explicit BodyComment(std::span<const BYTE> src);
		explicit BodyComment(wd::StrView simpleText)
			: lang3{L"eng"}, text{simpleText} { }

		std::unique_ptr<Body> clone() const override;
		[[nodiscard]] std::wstring as_simple_text() const override { return text; }
		[[nodiscard]] size_t serialize_len() const override;
		void serialize(std::vector<BYTE> &dest) const override;

		std::wstring lang3{}, descr{}, text{};
	};

	struct BodyPicture final : public Body {
	private:
		BodyPicture() = default;
	public:
		explicit BodyPicture(std::span<const BYTE> src);
		BodyPicture(const std::vector<BYTE> &binData, wd::StrView mimeType)
			: mime{mimeType}, picType{PicType::cover_front}, bin{binData} { }

		std::unique_ptr<Body> clone() const override;
		[[nodiscard]] std::wstring as_simple_text() const override;
		[[nodiscard]] size_t serialize_len() const override;
		void serialize(std::vector<BYTE> &dest) const override;

		std::wstring mime{};
		PicType picType = PicType::other;
		std::wstring descr{};
		std::vector<BYTE> bin{};
	};

	struct BodyGeob final : public Body {
	private:
		BodyGeob() = default;
	public:
		explicit BodyGeob(std::span<const BYTE> src);

		std::unique_ptr<Body> clone() const override;
		[[nodiscard]] std::wstring as_simple_text() const override;
		[[nodiscard]] size_t serialize_len() const override;
		void serialize(std::vector<BYTE> &dest) const override;

		std::wstring mime{}, fileName{}, descr{};
		std::vector<BYTE> encObj{};
	};

}
