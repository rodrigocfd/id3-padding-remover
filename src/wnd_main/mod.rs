use winsafe::{prelude::*, gui};

mod context_menu;
mod ctor;
mod events;
mod funcs;

#[derive(Clone)]
pub struct WndMain {
	wnd:       gui::WindowMain,
	lst_files: gui::ListView,
}
