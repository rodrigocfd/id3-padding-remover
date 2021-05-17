use winsafe::gui;

mod wnd_main_events;
mod wnd_main_funcs;

#[derive(Clone)]
pub struct WndMain {
	wnd:        gui::WindowMain,
	lst_files:  gui::ListView,
	lst_frames: gui::ListView,
}
