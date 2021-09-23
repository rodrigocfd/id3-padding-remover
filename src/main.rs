#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod id3v2;
mod ids;
mod util;
mod wnd_main;
mod wnd_modify;

use winsafe as w;
use wnd_main::WndMain;

fn main() {
	if let Err(e) = WndMain::new().run() {
		util::prompt::err(w::HWND::NULL,
			"Oops...", Some("Uncaught error"), &e.to_string());
	}
}
