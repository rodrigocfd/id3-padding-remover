use std::cell::Cell;
use std::rc::Rc;
use winsafe::{self as w, prelude::*, co, gui};

use crate::ids;
use super::{LIST_COLS, WndMain};

impl WndMain {
	/// Creates a new `WndMain` object.
	#[must_use]
	pub fn new() -> Self {
		use gui::{Horz as H, Vert as V};

		let wnd = gui::WindowMain::new_dlg(ids::DLG_MAIN, Some(ids::ICO_APP), Some(ids::ACC_MAIN));
		let lst_files = gui::ListView::new_dlg(&wnd, ids::LST_FILES, (H::Resize, V::Resize), Some(ids::MNU_MAIN));
		let lst_files_h = gui::Header::from_list_view(&lst_files);
		let cur_sort_col = Rc::new(Cell::new((0xffff_ffff, false)));

		let new_self = Self { wnd, lst_files, lst_files_h, cur_sort_col };
		new_self.wm_events();
		new_self.list_events();
		new_self
	}

	/// Runs the `WndMain` as the main application window.
	pub fn run(&self) -> w::AnyResult<i32> {
		self.wnd.run_main(None)
	}

	/// Initializes the `WndMain` window.
	pub(super) fn on_init_dialog(&self) -> w::AnyResult<bool> {
		self.update_num_files(self.lst_files.items().count());

		self.lst_files.set_image_list(co::LVSIL::SMALL, {
			let il = w::HIMAGELIST::Create(w::SIZE::new(16, 16), co::ILC::COLOR32, 1, 1)?;
			il.add_icons_from_shell(&["mp3"])?;
			il
		});
		self.lst_files.context_menu().unwrap().SetMenuDefaultItem(w::IdPos::Id(ids::MNU_MAIN_EDIT))?;
		self.lst_files.set_extended_style(true, co::LVS_EX::FULLROWSELECT);
		self.lst_files.columns().add(
			&LIST_COLS.iter()
				.map(|(title, cx, _)| (*title, *cx))
				.collect::<Vec<_>>(),
		);

		self.sort_list(0, true); // sort by path initially
		Ok(true)
	}
}
