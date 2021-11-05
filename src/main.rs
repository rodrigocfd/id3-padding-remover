#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod id3v2;
mod ids;
mod util;
mod wnd_fields;
mod wnd_main;

use winsafe::{prelude::*, self as w};

fn main() {
	if let Err(e) = run_app() {
		util::prompt::err(w::HWND::NULL,
			"Oops...", Some("Uncaught error"), &e.to_string()).unwrap();
	}
}

fn run_app() -> w::ErrResult<i32> {
	wnd_main::WndMain::new()?.run()
}
