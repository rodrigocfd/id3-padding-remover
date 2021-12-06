use winsafe::{prelude::*, self as w, co};

use super::WndPicture;

impl WndPicture {
	pub(super) fn _events(&self) {
		self.wnd.on().wm_paint({
			let self2 = self.clone();
			move || {
				let mut ps = w::PAINTSTRUCT::default();
				let hdc = self2.wnd.hwnd().BeginPaint(&mut ps)?;

				if let Some(image) = self2.image.try_borrow()?.as_ref() {
					let rc_cli = self2.wnd.hwnd().GetClientRect()?; // *** NOT CRASHING ON ERROR ??? ***

					let ipic = w::idl::IPicture::from_slice(&image, true)?;
					let (width, height) = hdc.HiMetricToPixel(ipic.get_Width()?, ipic.get_Height()?);

					let hdc_mem = hdc.CreateCompatibleDC()?;
					let (_, hbmp_old) = ipic.SelectPicture(hdc_mem)?;
					hdc.SetStretchBltMode(co::STRETCH_MODE::HALFTONE)?;
					hdc.SetBrushOrgEx(w::POINT::new(0, 0))?;
					hdc.StretchBlt(w::POINT::new(0, 0), w::SIZE::new(rc_cli.right, rc_cli.bottom),
						hdc_mem, w::POINT::new(0, 0), w::SIZE::new(width, height), co::ROP::SRCCOPY)?;
					hdc_mem.SelectObjectBitmap(hbmp_old)?;
					hdc_mem.DeleteDC()?;
				}

				self2.wnd.hwnd().EndPaint(&ps);
				Ok(())
			}
		});
	}
}
