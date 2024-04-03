use winsafe::{self as w, prelude::*};

use super::WndPicture;

impl WndPicture {
	pub(super) fn events(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_paint(move || {
			let hdc = self2.wnd.hwnd().BeginPaint()?;

			Ok(())
		});
	}
}
