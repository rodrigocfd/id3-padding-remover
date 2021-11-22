use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use winsafe::{prelude::*, self as w, gui};

use crate::util;
use crate::wnd_fields::WndFields;
use super::{ids, WndMain};

impl WndMain {
	pub fn new() -> w::ErrResult<Self> {
		use gui::{Horz as H, ListView, Vert as V, WindowMain};

		let mnu_mp3s = w::HINSTANCE::NULL
			.LoadMenu(w::IdStr::Id(ids::MNU_MP3S))?
			.GetSubMenu(0).unwrap();
		let mnu_frames = w::HINSTANCE::NULL
			.LoadMenu(w::IdStr::Id(ids::MNU_FRAMES))?
			.GetSubMenu(0).unwrap();

		let tags_cache = Arc::new(Mutex::new(HashMap::default()));

		let wnd        = WindowMain::new_dlg(ids::DLG_MAIN, Some(ids::ICO_FROG), Some(ids::ACC_MAIN));
		let lst_mp3s   = ListView::new_dlg(&wnd, ids::LST_MP3S, (H::Resize, V::Resize), Some(mnu_mp3s));
		let wnd_fields = WndFields::new(&wnd, tags_cache.clone(), w::POINT::new(292, 4), (H::Repos, V::None));
		let lst_frames = ListView::new_dlg(&wnd, ids::LST_FRAMES, (H::Repos, V::Resize), Some(mnu_frames));

		let new_self = Self {
			wnd, lst_mp3s, wnd_fields, lst_frames,
			tags_cache,
			app_name: util::app_name_from_res()?,
		};
		new_self._events()?;
		new_self._menu_events();
		Ok(new_self)
	}

	pub fn run(&self) -> w::ErrResult<i32> {
		self.wnd.run_main(None)
	}
}
