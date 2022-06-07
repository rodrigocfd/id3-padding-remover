use defer_lite::defer;
use winsafe::{prelude::*, self as w, co};

use super::WndPicture;

impl WndPicture {
	pub(super) fn _events(&self) {
		self.wnd.on().wm_paint({
			let self2 = self.clone();
			move || {
				let mut ps = w::PAINTSTRUCT::default();
				let hdc = self2.wnd.hwnd().BeginPaint(&mut ps)?;
				defer! { self2.wnd.hwnd().EndPaint(&ps); }

				if let Some(image) = self2.image.try_borrow()?.as_ref() {
					let rc_cli = self2.wnd.hwnd().GetClientRect()?;

					let ipic = w::IPicture::from_slice(&image, true)?;
					let sz_pic = ipic.size_px(Some(hdc))?;

					let hdc_mem = hdc.CreateCompatibleDC()?;
					defer! { hdc_mem.DeleteDC().expect("DeleteDC"); }

					let (_, hbmp_old) = ipic.SelectPicture(hdc_mem)?;
					defer! { hdc_mem.SelectObjectBitmap(hbmp_old).expect("SelectObjectBitmap"); }

					hdc.SetStretchBltMode(co::STRETCH_MODE::HALFTONE)?;
					hdc.SetBrushOrgEx(w::POINT::new(0, 0))?;
					hdc.StretchBlt(w::POINT::new(0, 0),
						w::SIZE::new(rc_cli.right, rc_cli.bottom),
						hdc_mem, w::POINT::new(0, 0), sz_pic, co::ROP::SRCCOPY)?;
				}

				Ok(())
			}
		});
	}
}
