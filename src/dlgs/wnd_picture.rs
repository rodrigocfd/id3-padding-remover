use std::cell::RefCell;
use std::rc::Rc;
use winsafe::{self as w, prelude::*, co, gui};

use crate::id3v2;

#[derive(Clone)]
pub struct WndPicture {
	pub(super) wnd:  gui::WindowControl,
	pub(super) ipic: Option<w::IPicture>,
}

impl WndPicture {
	#[must_use]
	pub fn new(
		parent: &(impl GuiParent + 'static),
		sel_tags: Vec<Rc<RefCell<id3v2::Tag>>>,
		position: (i32, i32),
		size: (i32, i32),
		resize_behavior: (gui::Horz, gui::Vert),
	) -> w::AnyResult<Self>
	{
		use co::{WS, WS_EX};

		let wnd = gui::WindowControl::new(parent, gui::WindowControlOpts {
			position,
			size,
			style: WS::CHILD | WS::VISIBLE | WS::CLIPCHILDREN | WS::CLIPSIBLINGS | WS::DISABLED,
			ex_style: WS_EX::LEFT | WS_EX::CLIENTEDGE,
			resize_behavior,
			..gui::WindowControlOpts::default()
		});
		let ipic = Self::load_picture(sel_tags)?;

		let new_self = Self { wnd, ipic };
		new_self.events();
		Ok(new_self)
	}
}
