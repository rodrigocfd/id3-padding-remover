use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use std::sync::Arc;
use winsafe::{self as w, prelude::*, co, gui};

use crate::{id3v2, ids};
use super::{Field, WndEdit};

/// No horizontal or vertical changes.
const NN: (gui::Horz, gui::Vert) = (gui::Horz::None, gui::Vert::None);

impl Field {
	fn new_edit(parent: &impl GuiParent, chk_id: u16) -> Self {
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
			fld_artist: Field::new_edit(&wnd, ids::CHK_ARTIST),
			fld_title: Field::new_edit(&wnd, ids::CHK_TITLE),
			fld_subtitle: Field::new_edit(&wnd, ids::CHK_SUBTITLE),
			fld_album: Field::new_edit(&wnd, ids::CHK_ALBUM),
			fld_track: Field::new_edit(&wnd, ids::CHK_TRACK),
			fld_year: Field::new_edit(&wnd, ids::CHK_YEAR),
			fld_genre: Field {
				chk: gui::CheckBox::new_dlg(&wnd, ids::CHK_GENRE, NN),
				txt: Arc::new(gui::ComboBox::new_dlg(&wnd, ids::CMB_GENRE, NN)),
			},
			fld_composer: Field::new_edit(&wnd, ids::CHK_COMPOSER),
			fld_lyricist: Field::new_edit(&wnd, ids::CHK_LYRICIST),
			fld_comment: Field::new_edit(&wnd, ids::CHK_COMMENT),
			fld_performer: Field::new_edit(&wnd, ids::CHK_PERFORMER),
			fld_publisher: Field::new_edit(&wnd, ids::CHK_PUBLISHER),
			fld_orig_artist: Field::new_edit(&wnd, ids::CHK_ORIG_ARTIST),
			fld_orig_album: Field::new_edit(&wnd, ids::CHK_ORIG_ALBUM),
			fld_orig_year: Field::new_edit(&wnd, ids::CHK_ORIG_YEAR),
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
