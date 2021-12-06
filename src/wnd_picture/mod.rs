use std::cell::RefCell;
use std::rc::Rc;
use winsafe::gui;

mod wnd_picture_events;
mod wnd_picture_funcs;

/// Child window which renders a picture.
#[derive(Clone)]
pub struct WndPicture {
	wnd:   gui::WindowControl,
	image: Rc<RefCell<Option<Vec<u8>>>>,
}
