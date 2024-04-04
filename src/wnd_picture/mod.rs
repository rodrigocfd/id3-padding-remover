use winsafe::{self as w, gui};

mod wnd_picture_ctor;
mod wnd_picture_wm;

#[derive(Clone)]
pub struct WndPicture {
	wnd:  gui::WindowControl,
	ipic: Option<w::IPicture>,
}
