#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use winsafe::{self as w, prelude::*, co};

mod dlgs;
mod id3v2;
mod ids;

use dlgs::DlgMain;

fn main() {
	if let Err(e) = (|| DlgMain::new().run())() {
		w::HWND::NULL.MessageBox(
			&e.to_string(), "Uncaught error", co::MB::ICONERROR).unwrap();
	}
}
