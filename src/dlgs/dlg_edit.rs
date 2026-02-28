use std::cell::{Cell, RefCell};
use std::rc::Rc;
use winsafe::{self as w, gui, prelude::*};

use super::{Input, WndPicture, ids};
use crate::id3v2;

#[derive(Clone)]
pub struct DlgEdit {
	pub(super) wnd: gui::WindowModal,
	pub(super) inputs: Vec<Input>, // all checkbox + textbox for the fields
	pub(super) chk_pic: gui::CheckBox,
	pub(super) wnd_pic: WndPicture,
	pub(super) lbl_pic: gui::Label,
	pub(super) lst_frames: gui::ListView<id3v2::Frame>,
	pub(super) btn_uncheck_all: gui::Button,
	pub(super) btn_check_filled: gui::Button,
	pub(super) sel_tags: Rc<RefCell<Vec<id3v2::Tag>>>, // these will be modified and then returned
	pub(super) user_clicked_ok: Rc<Cell<bool>>,
}

impl DlgEdit {
	/// Creates the dialog and displays it, blocking until it's closed.
	///
	/// Takes ownership of the tags, and returns the modified tags.
	#[must_use]
	pub fn show(
		parent: &(impl GuiParent + 'static),
		sel_tags: Vec<id3v2::Tag>,
	) -> w::AnyResult<Option<Vec<id3v2::Tag>>> {
		let no_res = (gui::Horz::None, gui::Vert::None);

		let wnd = gui::WindowModal::new_dlg(ids::DLG_EDIT);
		let inputs = (ids::CHK_ARTIST..=ids::CHK_COMMENT)
			.step_by(2)
			.zip(FIELDS_NAME4.iter())
			.map(|(chk_id, name4)| Input::new(&wnd, *name4, chk_id))
			.collect::<Vec<_>>();
		let chk_picture = gui::CheckBox::new_dlg(&wnd, ids::CHK_PICTURE, no_res);
		let wnd_pic = WndPicture::new(&wnd, gui::dpi(420, 60), gui::dpi(200, 200), no_res);
		let lbl_pic = gui::Label::new_dlg(&wnd, ids::LBL_PICTURE, no_res);
		let lst_frames =
			gui::ListView::new_dlg(&wnd, ids::LST_FRAMES, no_res, Some(ids::MNU_FRAMES));
		let btn_uncheck_all = gui::Button::new_dlg(&wnd, ids::BTN_UNCHECK_ALL, no_res);
		let btn_check_filled = gui::Button::new_dlg(&wnd, ids::BTN_CHECK_FILLED, no_res);
		let sel_tags = Rc::new(RefCell::new(sel_tags));
		let user_clicked_ok = Rc::new(Cell::new(false));

		let new_self = Self {
			wnd,
			inputs,
			chk_pic: chk_picture,
			wnd_pic,
			lbl_pic,
			lst_frames,
			btn_uncheck_all,
			btn_check_filled,
			sel_tags,
			user_clicked_ok,
		};
		new_self.events();
		new_self.show_modal(parent)
	}

	fn show_modal(&self, parent: &impl GuiParent) -> w::AnyResult<Option<Vec<id3v2::Tag>>> {
		self.wnd.show_modal(parent)?;
		if self.user_clicked_ok.get() {
			let edited_tags = self.sel_tags.replace(Vec::new());
			Ok(Some(edited_tags)) // user clicked OK
		} else {
			Ok(None) // user clicked Cancel
		}
	}
}

/// Frame names for each Input, in order.
pub const FIELDS_NAME4: &[&str] = &[
	"TPE1", "TIT2", "TIT3", "TALB", "TRCK", "TYER", "TCON", "TPE3", "TPUB", "TOPE", "TOAL", "TORY",
	"TCOM", "TEXT", "COMM",
];
