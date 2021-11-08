use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use winsafe::{prelude::*, self as w, gui};

use crate::ids::main as id;
use crate::util;
use crate::wnd_fields::WndFields;
use super::WndMain;

impl WndMain {
	pub fn new() -> w::ErrResult<Self> {
		use gui::{Horz as H, ListView, Vert as V, WindowMain};

		let context_menu = w::HINSTANCE::NULL
			.LoadMenu(w::IdStr::Id(id::MNU_FILE))?
			.GetSubMenu(0).unwrap();

		let tags_cache = Rc::new(RefCell::new(HashMap::default()));

		let wnd        = WindowMain::new_dlg(id::DLG_MAIN, Some(id::ICO_FROG), Some(id::ACT_MAIN));
		let lst_files  = ListView::new_dlg(&wnd, id::LST_FILES, H::Resize, V::Resize, Some(context_menu));
		let wnd_fields = WndFields::new(&wnd, tags_cache.clone(), w::POINT::new(496, 8), H::Repos, V::None);
		let lst_frames = ListView::new_dlg(&wnd, id::LST_FRAMES, H::Repos, V::Resize, None);

		let new_self = Self {
			wnd, lst_files, wnd_fields, lst_frames,
			tags_cache,
			app_name: util::app_name_from_res()?,
		};
		new_self._events();
		new_self._menu_events();
		Ok(new_self)
	}

	pub fn run(&self) -> w::ErrResult<i32> {
		self.wnd.run_main(None)
	}
}
