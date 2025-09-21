use winsafe::{self as w};

use crate::id3v2::*;

/// Polymorphic data of a frame.
#[derive(Clone, PartialEq, Eq)]
pub enum Body {
	Text(Text),
	UserText(UserText),
	Binary(Binary),
	Comment(Comment),
	Picture(Picture),
	Geob(Geob),
}

/// Implemented by each variant of `Body`.
pub(super) trait BodyVariant: Sized + std::fmt::Display {
	#[must_use]
	fn parse(src: &[u8]) -> w::AnyResult<Self>;

	#[must_use]
	fn serialize_size(&self) -> usize;

	fn serialize(&self, dest: &mut Vec<u8>);

	#[must_use]
	fn as_editable_str(&self) -> w::AnyResult<String>;

	fn set_editable_str(&mut self, text: &str) -> w::AnyResult<()>;
}

impl std::fmt::Display for Body {
	fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
		use Body::*;
		match self {
			Text(t) => t.fmt(f),
			UserText(ut) => ut.fmt(f),
			Binary(b) => b.fmt(f),
			Comment(c) => c.fmt(f),
			Picture(p) => p.fmt(f),
			Geob(g) => g.fmt(f),
		}
	}
}

impl Body {
	/// Attempts to create a body from a simple text.
	#[must_use]
	pub(super) fn new_from_editable_str(name4: &str, text: &str) -> w::AnyResult<Self> {
		if name4 == "COMM" {
			Ok(Self::Comment(Comment {
				lang3: "eng".to_owned(), // default lang
				descr: String::new(),
				text: text.to_owned(),
			}))
		} else if name4.starts_with('T') && name4 != "TDAT" {
			Ok(Self::Text(Text { text: text.to_owned() })) // simplest case
		} else {
			Err(format!("Cannot create a single-text {name4} frame.").into())
		}
	}

	/// Parses the body from raw binary data.
	#[must_use]
	pub(super) fn parse(name4: &str, src: &[u8]) -> w::AnyResult<Self> {
		Ok(if name4 == "COMM" {
			Self::Comment(Comment::parse(src)?)
		} else if name4 == "APIC" {
			Self::Picture(Picture::parse(src)?)
		} else if name4 == "GEOB" {
			Self::Geob(Geob::parse(src)?)
		} else if name4 == "TXXX" {
			Self::UserText(UserText::parse(src)?)
		} else if name4.starts_with('T') && name4 != "TDAT" {
			Self::Text(Text::parse(src)?)
		} else {
			Self::Binary(Binary::parse(src)?) // everything else is trated as raw binary
		})
	}

	/// Size in bytes of the body when serialized.
	#[must_use]
	pub(super) fn serialize_size(&self) -> usize {
		use Body::*;
		match self {
			Text(t) => t.serialize_size(),
			UserText(ut) => ut.serialize_size(),
			Binary(b) => b.serialize_size(),
			Comment(c) => c.serialize_size(),
			Picture(p) => p.serialize_size(),
			Geob(g) => g.serialize_size(),
		}
	}

	/// Serializes the body into a Vec.
	pub(super) fn serialize(&self, dest: &mut Vec<u8>) {
		use Body::*;
		match self {
			Text(t) => t.serialize(dest),
			UserText(ut) => ut.serialize(dest),
			Binary(b) => b.serialize(dest),
			Comment(c) => c.serialize(dest),
			Picture(p) => p.serialize(dest),
			Geob(g) => g.serialize(dest),
		}
	}

	/// Tries to return the value as an editable string, or an error if not
	/// possible.
	#[must_use]
	pub(super) fn as_editable_str(&self) -> w::AnyResult<String> {
		use Body::*;
		match self {
			Text(t) => t.as_editable_str(),
			UserText(ut) => ut.as_editable_str(),
			Binary(b) => b.as_editable_str(),
			Comment(c) => c.as_editable_str(),
			Picture(p) => p.as_editable_str(),
			Geob(g) => g.as_editable_str(),
		}
	}

	/// Tries to set the value as an editable string, returning an error if not
	/// possible.
	pub(super) fn set_editable_str(&mut self, text: &str) -> w::AnyResult<()> {
		use Body::*;
		match self {
			Text(t) => t.set_editable_str(text),
			UserText(ut) => ut.set_editable_str(text),
			Binary(b) => b.set_editable_str(text),
			Comment(c) => c.set_editable_str(text),
			Picture(p) => p.set_editable_str(text),
			Geob(g) => g.set_editable_str(text),
		}
	}
}
