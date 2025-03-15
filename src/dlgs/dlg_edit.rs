use std::cell::{Cell, RefCell};
use std::rc::Rc;
use std::sync::Arc;
use winsafe::{self as w, co, gui, prelude::*};

use super::{Input, WndPicture};
use crate::{id3v2, ids};

#[derive(Clone)]
pub struct DlgEdit {
	pub(super) wnd: gui::WindowModal,
	pub(super) btn_ok: gui::Button,
	pub(super) btn_cancel: gui::Button,
	pub(super) inputs: Rc<RefCell<Vec<Input>>>,
	pub(super) wnd_pic: WndPicture,
	pub(super) btn_uncheck: gui::Button,
	pub(super) lst_frames: gui::ListView,
	pub(super) sel_tags: Vec<Rc<RefCell<id3v2::Tag>>>,
	pub(super) modal_return: Rc<Cell<bool>>,
}

impl DlgEdit {
	#[must_use]
	pub fn new(sel_tags: Vec<Rc<RefCell<id3v2::Tag>>>) -> w::AnyResult<Self> {
		let none2 = (gui::Horz::None, gui::Vert::None);

		let wnd = gui::WindowModal::new_dlg(ids::DLG_EDIT);
		let btn_ok = gui::Button::new_dlg(&wnd, co::DLGID::OK.into(), none2);
		let btn_cancel = gui::Button::new_dlg(&wnd, co::DLGID::CANCEL.into(), none2);
		let inputs = Rc::new(RefCell::new(vec![
			Input::new_edit("TPE1", &wnd, ids::CHK_ARTIST),
			Input::new_edit("TIT2", &wnd, ids::CHK_TITLE),
			Input::new_edit("TIT3", &wnd, ids::CHK_SUBTITLE),
			Input::new_edit("TALB", &wnd, ids::CHK_ALBUM),
			Input::new_edit("TRCK", &wnd, ids::CHK_TRACK),
			Input::new_edit("TYER", &wnd, ids::CHK_YEAR),
			Input {
				name4: "TCON".to_owned(),
				chk: gui::CheckBox::new_dlg(&wnd, ids::CHK_GENRE, none2),
				txt: Arc::new(gui::ComboBox::new_dlg(&wnd, ids::CMB_GENRE, none2)),
			},
			Input::new_edit("TPE3", &wnd, ids::CHK_PERFORMER),
			Input::new_edit("TPUB", &wnd, ids::CHK_PUBLISHER),
			Input::new_edit("TOPE", &wnd, ids::CHK_ORIG_ARTIST),
			Input::new_edit("TOAL", &wnd, ids::CHK_ORIG_ALBUM),
			Input::new_edit("TORY", &wnd, ids::CHK_ORIG_YEAR),
			Input::new_edit("TCOM", &wnd, ids::CHK_COMPOSER),
			Input::new_edit("TEXT", &wnd, ids::CHK_LYRICIST),
			Input::new_edit("COMM", &wnd, ids::CHK_COMMENT),
		]));
		let wnd_pic =
			WndPicture::new(&wnd, sel_tags.clone(), gui::dpi(440, 50), gui::dpi(200, 200), none2)?;
		let btn_uncheck = gui::Button::new_dlg(&wnd, ids::BTN_UNCHECK_ALL, none2);
		let lst_frames = gui::ListView::new_dlg(&wnd, ids::LST_FRAMES, none2, None);
		let modal_return = Rc::new(Cell::new(false));

		let new_self = Self {
			wnd,
			btn_ok,
			btn_cancel,
			inputs,
			wnd_pic,
			btn_uncheck,
			lst_frames,
			sel_tags,
			modal_return,
		};
		new_self.wm_events();
		Ok(new_self)
	}

	pub fn show(&self, parent: &impl GuiParent) -> w::AnyResult<bool> {
		self.wnd.show_modal(parent).map(|_| self.modal_return.get())
	}
}
