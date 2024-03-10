use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use std::sync::Arc;
use winsafe::{self as w, prelude::*, co, gui};

use crate::{id3v2, ids};
use super::{Field, WndEdit};

const NN: (gui::Horz, gui::Vert) = (gui::Horz::None, gui::Vert::None);

impl Field {
	fn new(parent: &impl GuiParent, chk_id: u16) -> Self {
		Self {
			chk: gui::CheckBox::new_dlg(parent, chk_id, NN),
			txt: Arc::new(gui::Edit::new_dlg(parent, chk_id + 1, NN)),
		}
	}
}

impl WndEdit {
	/// Creates a new `WndEdit` object.
	#[must_use]
	pub fn new(
		parent: &impl GuiParent,
		all_tags: Rc<RefCell<HashMap<String, id3v2::Tag>>>,
		selected_paths: Vec<String>,
	) -> Self
	{
		let wnd = gui::WindowModal::new_dlg(parent, ids::DLG_EDIT);

		let new_self = Self {
			wnd: wnd.clone(),
			btn_ok: gui::Button::new_dlg(&wnd, co::DLGID::OK.into(), NN),
			btn_cancel: gui::Button::new_dlg(&wnd, co::DLGID::CANCEL.into(), NN),
			fld_artist: Field::new(&wnd, ids::CHK_ARTIST),
			fld_title: Field::new(&wnd, ids::CHK_TITLE),
			fld_subtitle: Field::new(&wnd, ids::CHK_SUBTITLE),
			fld_album: Field::new(&wnd, ids::CHK_ALBUM),
			fld_track: Field::new(&wnd, ids::CHK_TRACK),
			fld_year: Field::new(&wnd, ids::CHK_YEAR),
			fld_genre: Field {
				chk: gui::CheckBox::new_dlg(&wnd, ids::CHK_GENRE, NN),
				txt: Arc::new(gui::ComboBox::new_dlg(&wnd, ids::CMB_GENRE, NN)),
			},
			all_tags,
			selected_paths,
		};
		new_self.wm_events();
		new_self
	}

	pub fn show(&self) -> w::AnyResult<()> {
		self.wnd.show_modal()
			.map(|_| ())
	}

	/// Initializes the `WndEdit` window.
	pub(super) fn init_dialog(&self) -> w::AnyResult<bool> {

		Ok(true)
	}
}
