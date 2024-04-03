use std::cell::RefCell;
use std::rc::Rc;
use std::sync::Arc;
use try_iterator::prelude::*;
use winsafe::{self as w, prelude::*, co, gui};

use crate::{id3v2, ids, wnd_picture::WndPicture};
use super::{FieldPack, WndEdit};

impl FieldPack {
	fn new_edit(field: id3v2::Field, parent: &impl GuiParent, chk_id: u16) -> Self {
		Self {
			field,
			chk: gui::CheckBox::new_dlg(parent, chk_id, NN),
			txt: Arc::new(gui::Edit::new_dlg(parent, chk_id + 1, NN)),
		}
	}
}

//------------------------------------------------------------------------------

/// No horizontal or vertical changes.
pub const NN: (gui::Horz, gui::Vert) = (gui::Horz::None, gui::Vert::None);

impl WndEdit {
	/// Creates a new `WndEdit` object.
	#[must_use]
	pub fn new(
		parent: &impl GuiParent,
		sel_tags: Vec<Rc<RefCell<id3v2::Tag>>>,
	) -> Self
	{
		let wnd = gui::WindowModal::new_dlg(parent, ids::DLG_EDIT);
		let btn_ok = gui::Button::new_dlg(&wnd, co::DLGID::OK.into(), NN);
		let btn_cancel = gui::Button::new_dlg(&wnd, co::DLGID::CANCEL.into(), NN);
		let field_packs = Rc::new(RefCell::new(vec![
			FieldPack::new_edit(id3v2::Field::Artist, &wnd, ids::CHK_ARTIST),
			FieldPack::new_edit(id3v2::Field::Title, &wnd, ids::CHK_TITLE),
			FieldPack::new_edit(id3v2::Field::Subtitle, &wnd, ids::CHK_SUBTITLE),
			FieldPack::new_edit(id3v2::Field::Album, &wnd, ids::CHK_ALBUM),
			FieldPack::new_edit(id3v2::Field::Track, &wnd, ids::CHK_TRACK),
			FieldPack::new_edit(id3v2::Field::Year, &wnd, ids::CHK_YEAR),
			FieldPack {
				field: id3v2::Field::Genre,
				chk: gui::CheckBox::new_dlg(&wnd, ids::CHK_GENRE, NN),
				txt: Arc::new(gui::ComboBox::new_dlg(&wnd, ids::CMB_GENRE, NN)),
			},
			FieldPack::new_edit(id3v2::Field::Composer, &wnd, ids::CHK_COMPOSER),
			FieldPack::new_edit(id3v2::Field::Lyricist, &wnd, ids::CHK_LYRICIST),
			FieldPack::new_edit(id3v2::Field::Comment, &wnd, ids::CHK_COMMENT),
			FieldPack::new_edit(id3v2::Field::Performer, &wnd, ids::CHK_PERFORMER),
			FieldPack::new_edit(id3v2::Field::Publisher, &wnd, ids::CHK_PUBLISHER),
			FieldPack::new_edit(id3v2::Field::OrigArtist, &wnd, ids::CHK_ORIG_ARTIST),
			FieldPack::new_edit(id3v2::Field::OrigAlbum, &wnd, ids::CHK_ORIG_ALBUM),
			FieldPack::new_edit(id3v2::Field::OrigYear, &wnd, ids::CHK_ORIG_YEAR),
		]));
		let wnd_pic = WndPicture::new(&wnd, (466, 10), (90, 90), NN);

		let new_self = Self { wnd, btn_ok, btn_cancel, field_packs, wnd_pic, sel_tags };
		new_self.wm_events();
		new_self
	}

	pub fn show(&self) -> w::AnyResult<()> {
		self.wnd.show_modal()
			.map(|_| ())
	}

	/// Initializes the `WndEdit` window.
	pub(super) fn init_dialog(&self) -> w::AnyResult<bool> {
		self.field_packs.try_borrow()?
			.iter()
			.try_for_each(|field_pack| {
				if self.sel_tags.len() == 1 { // just 1 MP3 being edited?
					self.wnd.set_text("Edit tag");
					match &self.sel_tags[0].try_borrow()?.known_field(field_pack.field) {
						Some(field) => { // the MP3 has this field
							field_pack.txt.set_text(field);
							field_pack.chk.set_check_state_and_trigger(gui::CheckState::Checked);
						},
						None => { // the MP3 doesn't have this field
							field_pack.chk.set_check_state_and_trigger(gui::CheckState::Unchecked);
						},
					}
				} else { // multiple MP3s being edited
					self.wnd.set_text(&format!("Edit {} tags", self.sel_tags.len()));
					let maybe_idx_first = self.sel_tags.iter() // index of first MP3 which has the field
						.try_position(|tag| {
							let has = tag.try_borrow()?.has_known_field(field_pack.field);
							w::AnyResult::Ok(has)
						})?;

					match maybe_idx_first {
						Some(idx_first) => { // at least 1 MP3 has this field
							let first_val = self.sel_tags[idx_first]
								.try_borrow()?
								.known_field(field_pack.field).unwrap();
							let val_equal_in_all_mp3s = self.sel_tags.iter()
								.skip(1)
								.try_all(|tag| {
									let is_all = tag.try_borrow()?
										.known_field(field_pack.field)
										.unwrap_or_default() == first_val;
									w::AnyResult::Ok(is_all)
								})?;

							if val_equal_in_all_mp3s {
								field_pack.txt.set_text(&first_val);
								field_pack.chk.set_check_state_and_trigger(gui::CheckState::Checked);
							} else {
								field_pack.chk.set_check_state_and_trigger(gui::CheckState::Unchecked);
							}
						},
						None => { // no MP3 has this field
							field_pack.chk.set_check_state_and_trigger(gui::CheckState::Unchecked);
						},
					}
				}

				w::AnyResult::Ok(())
			})?;
		Ok(true)
	}
}
