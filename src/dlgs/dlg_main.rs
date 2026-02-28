use std::cell::Cell;
use std::rc::Rc;
use winsafe::{self as w, gui};

use super::ids;
use crate::id3v2;

#[derive(Clone)]
pub struct DlgMain {
	pub(super) wnd: gui::WindowMain,
	pub(super) lst_files: gui::ListView<id3v2::Tag>,
	pub(super) cur_sort: Rc<Cell<(u32, bool)>>, // index, ascending
	pub(super) drop_target: w::IDropTarget,
}

impl DlgMain {
	/// Creates the main dialog and displays it, blocking until it's closed.
	#[must_use]
	pub fn run_main() -> w::AnyResult<i32> {
		use gui::{Horz as H, Vert as V};

		let wnd = gui::WindowMain::new_dlg(ids::DLG_MAIN, Some(ids::ICO_APP), Some(ids::ACC_MAIN));
		let lst_files = gui::ListView::new_dlg(
			&wnd,
			ids::LST_FILES,
			(H::Resize, V::Resize),
			Some(ids::MNU_FILE),
		);
		let cur_sort = Rc::new(Cell::new((0, true))); // 1st col, ascending
		let drop_target = w::IDropTarget::new_impl();

		let new_self = Self { wnd, lst_files, cur_sort, drop_target };
		new_self.events();
		new_self.wnd.run_main(None)
	}
}

/// Columns of the main MP3 list:
/// * title;
/// * width;
/// * text justification;
/// * ID3v2 frame ID
pub const LIST_COLS: &[(&str, i32, Option<gui::HeaderJustify>, &str)] = &[
	("File", 1, None, ""), // this column will fill the empty space
	("Pad", 50, Some(gui::HeaderJustify::Right), ""),
	("Art", 30, Some(gui::HeaderJustify::Center), ""),
	("RG", 30, Some(gui::HeaderJustify::Center), ""),
	("Artist", 90, None, "TPE1"),
	("T#", 30, Some(gui::HeaderJustify::Right), "TRCK"),
	("Title", 100, None, "TIT2"),
	("Album", 100, None, "TALB"),
	("Year", 40, Some(gui::HeaderJustify::Right), "TYER"),
	("Genre", 90, None, "TCON"),
	("Performer", 70, None, "TPE3"),
	("Composer", 70, None, "TCOM"),
	("Lyricist", 70, None, "TEXT"),
	("Orig. artist", 70, None, "TOPE"),
	("Comment", 70, None, "COMM"),
];
