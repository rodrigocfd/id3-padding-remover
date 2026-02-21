#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod dlgs;
mod id3v2;
mod ids;
mod msgbox;

use winsafe::{self as w, co, prelude::*};

use dlgs::DlgMain;

fn main() {
	let _ole_guard = w::OleInitialize().unwrap();
	DlgMain::run_main().unwrap();
}
