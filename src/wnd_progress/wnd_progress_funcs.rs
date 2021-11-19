use std::sync::{Arc, Mutex};
use winsafe::{prelude::*, self as w, gui, msg};

use super::{ids, WndProgress};

impl WndProgress {
	pub fn new<F>(parent: &impl Parent, job: F) -> Self
		where F: Fn() -> w::ErrResult<()> + Send + 'static,
	{
		use gui::{Horz, Vert};

		let wnd     = gui::WindowModal::new_dlg(parent, ids::DLG_RUN);
		let pro_run = gui::ProgressBar::new_dlg(&wnd, ids::PRO_RUN, (Horz::None, Vert::None));
		let job_cb  = Arc::new(Mutex::new(job));

		let new_self = Self { wnd, pro_run, job_cb };
		new_self._events();
		new_self
	}

	pub fn show(&self) -> w::ErrResult<i32> {
		self.wnd.show_modal()
			.map_err(|e| e.into())
	}

	fn _events(&self) {
		self.wnd.on().wm_init_dialog({
			let self2 = self.clone();
			move |_| {
				self2.pro_run.set_marquee(true);

				self2.wnd.spawn_new_thread({ // run a new thread immediately
					let self2 = self2.clone();
					move || {
						let cb_ret = self2.job_cb.lock().unwrap()(); // execute user callback

						self2.wnd.run_ui_thread({ // return to UI thread and close modal
							let self2 = self2.clone();
							move || {
								self2.wnd.hwnd().SendMessage(msg::wm::Close {});
								Ok(())
							}
						});

						cb_ret // user closure erros will go down through library pipes
					}
				});

				Ok(true)
			}
		});
	}
}
