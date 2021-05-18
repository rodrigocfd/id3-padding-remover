use winsafe as w;
use winsafe::gui;

use crate::ids;
use super::WndMain;

impl WndMain {
	pub fn new() -> Self {
		let wnd = gui::WindowMain::new_dlg(ids::DLG_MAIN, Some(ids::ICO_FROG), None);
		let lst_files = gui::ListView::new_dlg(&wnd, ids::LST_FILES);
		let lst_frames = gui::ListView::new_dlg(&wnd, ids::LST_FRAMES);
		let resizer = gui::Resizer::new(&wnd);

		let selfc = Self { wnd, lst_files, lst_frames, resizer };
		selfc.events();
		selfc
	}

	pub fn run(&self) -> w::WinResult<()> {
		self.wnd.run_main(None)
	}
}
