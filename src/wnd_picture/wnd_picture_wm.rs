use winsafe::{self as w, prelude::*, co};

use super::WndPicture;

impl WndPicture {
	pub(super) fn events(&self) {
		let self2 = self.clone();
		self.wnd.on().wm_paint(move || {
			let hdc = self2.wnd.hwnd().BeginPaint()?;

			if let Some(ipic) = &self2.ipic {
				let hdc_mem = hdc.CreateCompatibleDC()?;
				ipic.SelectPicture(&hdc_mem)?;

				let rc_cli = self2.wnd.hwnd().GetClientRect()?;
				let sz_pic_px = hdc.HiMetricToPixel(ipic.get_Width()?, ipic.get_Height()?);

				hdc.SetStretchBltMode(co::STRETCH_MODE::HALFTONE)?;
				hdc.SetBrushOrgEx(w::POINT::default())?;
				hdc.StretchBlt(
					w::POINT::default(),
					w::SIZE::new(rc_cli.right, rc_cli.bottom),
					&hdc_mem,
					w::POINT::default(),
					w::SIZE::new(sz_pic_px.0, sz_pic_px.1),
					co::ROP::SRCCOPY,
				)?;
			}

			Ok(())
		});
	}
}
