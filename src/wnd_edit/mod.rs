use winsafe::gui;

mod wnd_edit_ctor;
mod wnd_edit_wm;

#[derive(Clone)]
pub struct WndEdit {
	wnd:        gui::WindowModal,
	btn_ok:     gui::Button,
	btn_cancel: gui::Button,
}
