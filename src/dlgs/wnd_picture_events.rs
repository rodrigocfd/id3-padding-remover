use winsafe::{self as w, evt, prelude::*};

use super::WndPicture;

impl WndPicture {
	pub(super) fn events(&self) {
		evt!(self, wm_paint, on_paint);
	}

	fn on_paint(&self) -> w::AnyResult<()> {
		let hdc = self.wnd.hwnd().BeginPaint()?;
		if let Some(pic) = &*self.pic.try_borrow()? {
			pic.Render(&hdc, None, None, None, None, None)?;
		}
		Ok(())
	}
}
