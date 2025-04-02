use winsafe::{self as w, prelude::*};

use super::WndPicture;

impl WndPicture {
	pub(super) fn on_paint(&self) -> w::AnyResult<()> {
		let hdc = self.wnd.hwnd().BeginPaint()?;
		if let Some(pic) = &*self.pic.try_borrow()? {
			pic.Render(&hdc, None, None, None, None, None)?;
		}
		Ok(())
	}
}
