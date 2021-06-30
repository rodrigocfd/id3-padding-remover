use winsafe::{co, msg};

use crate::id3v2::Tag;
use crate::util;
use super::WndModify;

impl WndModify {
	pub(super) fn events(&self) {
		self.wnd.on().wm_init_dialog({
			let self2 = self.clone();
			move |_: msg::wm::InitDialog| {
				self2.wnd.hwnd().SetWindowText(
					&format!("Modify {} file(s)", self2.files.len()),
				).unwrap();

				self2.chk_rem_padding.set_check(true);

				true
			}
		});

		self.wnd.on().wm_command_accel_menu(co::DLGID::CANCEL.into(), {
			let wnd = self.wnd.clone();
			move || {
				wnd.hwnd().EndDialog(0).unwrap(); // close on ESC
			}
		});

		self.chk_rem_padding.on().bn_clicked({
			let self2 = self.clone();
			move || self2.enable_disable_rem_padding()
		});

		self.chk_rem_album.on().bn_clicked({
			let self2 = self.clone();
			move || self2.enable_disable_rem_padding()
		});

		self.chk_rem_rg.on().bn_clicked({
			let self2 = self.clone();
			move || self2.enable_disable_rem_padding()
		});

		self.chk_prefix_year.on().bn_clicked({
			let self2 = self.clone();
			move || self2.enable_disable_rem_padding()
		});

		self.btn_ok.on().bn_clicked({
			let self2 = self.clone();
			move || {
				self2.wnd.hwnd().EnumChildWindows(|hchild| { // disable all children
					hchild.EnableWindow(false);
					true
				});

				let mut tags_cache = self2.tags_cache.borrow_mut();

				for file in self2.files.iter() { // execute the chosen operations on each file
					let mut tag = tags_cache.get_mut(file).unwrap();

					if self2.chk_rem_album.is_checked() {
						tag.frames_mut().retain(|f| f.name4() != "APIC");
					}
					if self2.chk_rem_rg.is_checked() {
						self2.remove_replay_gain(&mut tag);
					}
					if self2.chk_prefix_year.is_checked() {
						if let Err(err) = self2.prefix_year(&mut tag, file) {
							util::msg::err(self2.wnd.hwnd(), "Operation error", &err.to_string());
							self2.wnd.hwnd().EndDialog(0).unwrap(); // close after error
						}
					}
				}

				let t0 = util::timer_start();

				for file in self2.files.iter() {
					let tag = tags_cache.get_mut(file).unwrap();
					tag.write(file).unwrap();        // save tag to file, no padding is written
					*tag = Tag::read(file).unwrap(); // load tag back from file
				}

				util::msg::info(self2.wnd.hwnd(), "Operation successful",
					&format!("{} file(s) processed in {:.2} ms.",
						self2.files.len(), util::timer_end_ms(t0)));

				self2.wnd.hwnd().EndDialog(0).unwrap(); // close after process is finished
			}
		});

		self.btn_cancel.on().bn_clicked({
			let wnd = self.wnd.clone();
			move || {
				wnd.hwnd().EndDialog(0).unwrap(); // close on Cancel
			}
		});
	}
}
