use std::cell::Cell;
use std::rc::Rc;
use winsafe::{self as w, gui};

use crate::{id3v2, ids};

#[derive(Clone)]
pub struct DlgMain {
	pub(super) wnd: gui::WindowMain,
	pub(super) lst_files: gui::ListView<id3v2::Tag>,
	pub(super) cur_sort_col: Rc<Cell<(u32, bool)>>, // index, reversed
}

impl DlgMain {
	#[must_use]
	pub fn new() -> Self {
		use gui::{Horz as H, Vert as V};

		let wnd = gui::WindowMain::new_dlg(ids::DLG_MAIN, Some(ids::ICO_APP), Some(ids::ACC_MAIN));
		let lst_files = gui::ListView::new_dlg(
			&wnd,
			ids::LST_FILES,
			(H::Resize, V::Resize),
			Some(ids::MNU_MAIN),
		);
		let cur_sort_col = Rc::new(Cell::new((0xffff_ffff, false)));

		let new_self = Self { wnd, lst_files, cur_sort_col };
		new_self.events();
		new_self
	}

	pub fn run(&self) -> w::AnyResult<i32> {
		self.wnd.run_main(None)
	}
}

pub const LIST_COLS: &[(&str, i32, &str)] = &[
	("File", 400, ""),
	("Pad", 50, ""),
	("Art", 30, ""),
	("RG", 30, ""),
	("Artist", 90, "TPE1"),
	("T#", 30, "TRCK"),
	("Title", 100, "TIT2"),
	("Album", 100, "TALB"),
	("Year", 40, "TYER"),
	("Genre", 90, "TCON"),
	("Performer", 70, "TPE3"),
	("Composer", 70, "TCOM"),
	("Lyricist", 70, "TEXT"),
	("Orig. artist", 70, "TOPE"),
	("Comment", 70, "COMM"),
];
