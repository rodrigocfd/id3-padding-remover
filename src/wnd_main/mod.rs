use winsafe::{prelude::*, gui};

mod ctor;
mod events;

#[derive(Clone)]
pub struct WndMain {
	wnd:       gui::WindowMain,
	lst_files: gui::ListView,
}
