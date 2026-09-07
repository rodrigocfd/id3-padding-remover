#pragma once
#include <array>
#include <memory>
#include "body.h"

namespace id3v2 {

	class Frame final {
	private:
		Frame() = default;
		Frame(const Frame&) = delete;
		Frame& operator=(const Frame&) = delete; // non-copyable
	public:
		Frame(Frame&&) = default;
		Frame& operator=(Frame&&) = default; // movable

		explicit Frame(std::span<const BYTE> src);
		Frame(wd::StrView name4, wd::StrView simpleText);
		Frame(const std::vector<BYTE> &picData, wd::StrView mimeType);

		[[nodiscard]] constexpr wd::StrView name4() const    { return _name4; }
		[[nodiscard]] constexpr size_t declared_size() const { return _declaredSize; }
		[[nodiscard]] const Body* body() const               { return _body.get(); }
		[[nodiscard]] Body* body()                           { return _body.get(); }

		[[nodiscard]] Frame clone() const;
		[[nodiscard]] size_t serialize_len() const;
		void serialize(std::vector<BYTE> &dest) const;

		[[nodiscard]] std::wstring as_simple_text() const { return _body->as_simple_text(); }
		void set_simple_text(wd::StrView text);

	private:
		std::wstring _name4{};
		size_t _declaredSize = 0; // used only at parsing
		std::array<BYTE, 2> _flags{};
		std::unique_ptr<Body> _body{}; // polymorphic
	};

}
