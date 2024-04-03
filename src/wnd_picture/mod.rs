use winsafe::gui;

mod wnd_picture_ctor;
mod wnd_picture_wm;

#[derive(Clone)]
pub struct WndPicture {
	wnd: gui::WindowControl,
}
