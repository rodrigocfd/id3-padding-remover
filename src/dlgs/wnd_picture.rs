use std::cell::{Cell, RefCell};
use std::rc::Rc;
use winsafe::{self as w, co, gui, prelude::*};

#[derive(Clone)]
pub struct WndPicture {
	pub(super) wnd: gui::WindowControl,
	pub(super) pic: Rc<RefCell<Option<w::IPicture>>>,
	pub(super) pic_err: Rc<Cell<Option<co::HRESULT>>>,
}

impl WndPicture {
	#[must_use]
	pub fn new(
		parent: &(impl GuiParent + 'static),
		position: (i32, i32),
		size: (i32, i32),
		resize_behavior: (gui::Horz, gui::Vert),
	) -> Self {
		use co::{WS, WS_EX};

		let wnd = gui::WindowControl::new(
			parent,
			gui::WindowControlOpts {
				position,
				size,
				resize_behavior,
				class_bg_brush: gui::Brush::Color(co::COLOR::BTNFACE),
				style: WS::CHILD | WS::VISIBLE | WS::CLIPCHILDREN | WS::CLIPSIBLINGS | WS::DISABLED,
				ex_style: WS_EX::LEFT | WS_EX::CLIENTEDGE,
				..Default::default()
			},
		);
		let pic = Rc::new(RefCell::new(None));
		let pic_err = Rc::new(Cell::new(None));

		let new_self = Self { wnd, pic, pic_err };
		new_self.events();
		new_self
	}

	fn events(&self) {
		self.wnd.on().wm_paint({
			let self2 = self.clone();
			move || self2.on_paint()
		});
	}
}
