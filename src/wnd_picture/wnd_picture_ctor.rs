use winsafe::{self as w, prelude::*, co, gui};

use super::WndPicture;

impl WndPicture {
	/// Creates a new `WndPicture` object.
	#[must_use]
	pub fn new(
		parent: &impl GuiParent,
		position: (i32, i32),
		size: (u32, u32),
		resize_behavior: (gui::Horz, gui::Vert),
	) -> Self
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

		let new_self = Self { wnd };
		new_self.events();
		new_self
	}
}
