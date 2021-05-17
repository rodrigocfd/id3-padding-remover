use winsafe::msg;
use super::WndMain;

impl WndMain {
	pub(super) fn events(&self) {
		self.wnd.on().wm_init_dialog({
			let selfc = self.clone();
			move |_: msg::wm::InitDialog| -> bool {
				selfc.lst_files.columns().add(&[
					("File", 0),
					("Padding", 60),
				]).unwrap();
				selfc.lst_files.columns().set_width_to_fill(0).unwrap();

				selfc.lst_frames.columns().add(&[
					("Frame", 65),
					("Value", 0),
				]).unwrap();
				selfc.lst_frames.columns().set_width_to_fill(1).unwrap();

				true
			}
		});
	}
}
