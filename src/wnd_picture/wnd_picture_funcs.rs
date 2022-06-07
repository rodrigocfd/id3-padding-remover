use std::cell::RefCell;
use std::rc::Rc;
use winsafe::{prelude::*, self as w, co, gui};

use super::WndPicture;

impl WndPicture {
	pub fn new(parent: &impl GuiParent,
		position: w::POINT, size: w::SIZE,
		resize_behavior: (gui::Horz, gui::Vert)) -> Self
	{
		use co::{WS, WS_EX};

		let wnd = gui::WindowControl::new(parent, gui::WindowControlOpts {
			size,
			position,
			style: WS::CHILD | WS::VISIBLE | WS::CLIPCHILDREN | WS::CLIPSIBLINGS | WS::DISABLED,
			ex_style: WS_EX::LEFT | WS_EX::CLIENTEDGE,
			horz_resize: resize_behavior.0,
			vert_resize: resize_behavior.1,
			..gui::WindowControlOpts::default()
		});

		let image = Rc::new(RefCell::new(None));

		let new_self = Self { wnd, image };
		new_self._events();
		new_self
	}

	pub fn enable(&self, enable: bool) {
		self.wnd.hwnd().EnableWindow(enable);
	}

	pub fn feed(&self, image: Option<&[u8]>) -> w::ErrResult<()> {
		*self.image.try_borrow_mut()? = image.map_or(None, |image| Some(image.to_vec()));
		self.wnd.hwnd().InvalidateRect(None, true)?;
		Ok(())
	}
}
