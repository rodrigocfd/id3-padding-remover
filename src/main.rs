#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod id3v2;
mod ids;
mod wnd_edit;
mod wnd_main;

use winsafe::{self as w, prelude::*, co};
use wnd_main::WndMain;

fn main() {
	if let Err(e) = (|| WndMain::new().run())() {
		w::HWND::NULL.MessageBox(
			&e.to_string(), "Uncaught error", co::MB::ICONERROR).unwrap();
	}
}
