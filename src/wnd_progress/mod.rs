use std::sync::{Arc, Mutex};
use winsafe::{self as w, gui};

mod ids;
mod wnd_progress_funcs;

/// Modal window with a progress bar and executes a closure asynchronously.
#[derive(Clone)]
pub struct WndProgress {
	wnd:     gui::WindowModal,
	pro_run: gui::ProgressBar,
	job_cb:  Arc<Mutex<dyn Fn() -> w::ErrResult<()> + Send>>,
}
