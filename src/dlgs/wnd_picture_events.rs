use winsafe::prelude::*;

use super::WndPicture;

impl WndPicture {
	pub(super) fn events(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_paint(move || {
			let hdc = self2.wnd.hwnd().BeginPaint()?;
			if let Some(pic) = &*self2.pic.try_borrow()? {
				pic.Render(&hdc, None, None, None, None, None)?;
			}
			Ok(())
		});
	}
}
