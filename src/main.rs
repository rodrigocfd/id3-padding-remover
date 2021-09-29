#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod id3v2;
mod ids;
mod util;
mod wnd_fields;
mod wnd_main;
mod wnd_modify;

use winsafe as w;

fn main() {
	if let Err(e) = run_app() {
		util::prompt::err(w::HWND::NULL,
			"Oops...", Some("Uncaught error"), &e.to_string()).unwrap();
	}
}

fn run_app() -> w::BoxResult<i32> {
	wnd_main::WndMain::new()?.run()
}
