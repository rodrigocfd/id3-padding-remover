use std::cell::RefCell;
use std::rc::Rc;
use try_iterator::prelude::*;
use winsafe::{self as w, prelude::*, co, gui};

use crate::id3v2;
use super::WndPicture;

impl WndPicture {
	/// Creates a new `WndPicture` object.
	#[must_use]
	pub fn new(
		parent: &impl GuiParent,
		sel_tags: Vec<Rc<RefCell<id3v2::Tag>>>,
		position: (i32, i32),
		size: (u32, u32),
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

		let new_self = Self { wnd };
		new_self.events();
		new_self.load_picture_if_due(sel_tags)?;
		Ok(new_self)
	}

	fn load_picture_if_due(&self, sel_tags: Vec<Rc<RefCell<id3v2::Tag>>>) -> w::AnyResult<()> {
		let maybe_idx_first = sel_tags.iter()
			.try_position(|tag| {
				let has = tag.try_borrow()?.apic().is_some();
				w::AnyResult::Ok(has)
			})?;

		if let Some(idx_first) = maybe_idx_first { // at last 1 MP3 has APIC
			let val_equal_in_all_mp3s = sel_tags.iter()
				.skip(1)
				.try_all(|tag| {
					let is_equal_to_1st = sel_tags[idx_first].try_borrow()?.apic().unwrap()
						== tag.try_borrow()?.apic().unwrap();
					w::AnyResult::Ok(is_equal_to_1st)
				})?;

			if val_equal_in_all_mp3s {
				println!("YES");
			}
		}

		Ok(())
	}
}
