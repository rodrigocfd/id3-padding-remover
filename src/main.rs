#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod dlgs;
mod id3v2;
mod ids;
mod msgbox;

use winsafe::{self as w, co, prelude::*};

use dlgs::DlgMain;

fn main() {
	if let Err(e) = (|| {
		let _ole_guard = w::OleInitialize()?;
		DlgMain::run_main()
	})() {
		w::HWND::NULL
			.MessageBox(&e.to_string(), "Uncaught error", co::MB::ICONERROR)
			.unwrap();
	}
}
