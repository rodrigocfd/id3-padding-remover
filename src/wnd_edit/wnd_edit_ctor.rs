use std::cell::{Cell, RefCell};
use std::rc::Rc;
use std::sync::Arc;
use try_iterator::prelude::*;
use winsafe::{self as w, prelude::*, co, gui};

use crate::{id3v2, ids, wnd_picture::WndPicture};
use super::{FieldPack, GENRES, WndEdit};

impl FieldPack {
	fn new_edit(name4: &str, parent: &impl GuiParent, chk_id: u16) -> Self {
		Self {
			name4: name4.to_owned(),
			chk: gui::CheckBox::new_dlg(parent, chk_id, NN),
			txt: Arc::new(gui::Edit::new_dlg(parent, chk_id + 1, NN)),
		}
	}
}

//------------------------------------------------------------------------------

/// No horizontal or vertical changes.
const NN: (gui::Horz, gui::Vert) = (gui::Horz::None, gui::Vert::None);

impl WndEdit {
	/// Creates a new `WndEdit` object.
	#[must_use]
	pub fn new(
		parent: &impl GuiParent,
		sel_tags: Vec<Rc<RefCell<id3v2::Tag>>>,
	) -> w::AnyResult<Self>
	{
		let wnd = gui::WindowModal::new_dlg(parent, ids::DLG_EDIT);
		let btn_ok = gui::Button::new_dlg(&wnd, co::DLGID::OK.into(), NN);
		let btn_cancel = gui::Button::new_dlg(&wnd, co::DLGID::CANCEL.into(), NN);
		let field_packs = Rc::new(RefCell::new(vec![
			FieldPack::new_edit("TPE1", &wnd, ids::CHK_ARTIST),
			FieldPack::new_edit("TIT2", &wnd, ids::CHK_TITLE),
			FieldPack::new_edit("TIT3", &wnd, ids::CHK_SUBTITLE),
			FieldPack::new_edit("TALB", &wnd, ids::CHK_ALBUM),
			FieldPack::new_edit("TRCK", &wnd, ids::CHK_TRACK),
			FieldPack::new_edit("TYER", &wnd, ids::CHK_YEAR),
			FieldPack {
				name4: "TCON".to_owned(),
				chk: gui::CheckBox::new_dlg(&wnd, ids::CHK_GENRE, NN),
				txt: Arc::new(gui::ComboBox::new_dlg(&wnd, ids::CMB_GENRE, NN)),
			},
			FieldPack::new_edit("TPE3", &wnd, ids::CHK_PERFORMER),
			FieldPack::new_edit("TPUB", &wnd, ids::CHK_PUBLISHER),
			FieldPack::new_edit("TOPE", &wnd, ids::CHK_ORIG_ARTIST),
			FieldPack::new_edit("TOAL", &wnd, ids::CHK_ORIG_ALBUM),
			FieldPack::new_edit("TORY", &wnd, ids::CHK_ORIG_YEAR),
			FieldPack::new_edit("TCOM", &wnd, ids::CHK_COMPOSER),
			FieldPack::new_edit("TEXT", &wnd, ids::CHK_LYRICIST),
			FieldPack::new_edit("COMM", &wnd, ids::CHK_COMMENT),
		]));
		let wnd_pic = WndPicture::new(&wnd, sel_tags.clone(), (250, 22), (120, 120), NN)?;
		let btn_uncheck = gui::Button::new_dlg(&wnd, ids::BTN_UNCHECK_ALL, NN);
		let lst_frames = gui::ListView::new_dlg(&wnd, ids::LST_FRAMES, NN, None);
		let modal_return = Rc::new(Cell::new(false));

		let new_self = Self {
			wnd,
			btn_ok, btn_cancel, field_packs, wnd_pic,
			btn_uncheck, lst_frames, sel_tags, modal_return,
		};
		new_self.wm_events();
		Ok(new_self)
	}

	pub fn show(&self) -> w::AnyResult<bool> {
		self.wnd.show_modal()
			.map(|_| self.modal_return.get())
	}

	/// Initializes the `WndEdit` window.
	pub(super) fn on_init_dialog(&self) -> w::AnyResult<bool> {
		self.wnd.set_text(&format!(
			"Edit {} file{}",
			self.sel_tags.len(),
			if self.sel_tags.len() == 1 { "" } else { "s" },
		));
		self.fill_chks_and_txts()?;
		self.fill_listview_fields()?;
		Ok(true)
	}

	fn fill_chks_and_txts(&self) -> w::AnyResult<()> {
		self.field_packs.try_borrow()?
			.iter()
			.try_for_each(|field_pack| {
				if field_pack.name4 == "TCON" { // feed the genres to the combo
					field_pack.txt.as_any()
						.downcast_ref::<gui::ComboBox>()
						.expect("ComboBox downcast failed.")
						.items()
						.add(GENRES);
				}

				let maybe_idx_first_mp3 = self.sel_tags.iter() // index of first MP3 which has the field
					.try_position(|tag| {
						let has = tag.try_borrow()?.frame(&field_pack.name4).is_some();
						w::AnyResult::Ok(has)
					})?;

				match maybe_idx_first_mp3 {
					Some(idx_first) => { // at least 1 MP3 has this field
						let first_tag = self.sel_tags[idx_first].try_borrow()?;
						let first_frame = first_tag.frame(&field_pack.name4).unwrap();

						let frame_equal_in_all_mp3s = self.sel_tags.iter()
							.skip(idx_first + 1)
							.try_all(|tag| {
								let is_equal_to_1st = match tag.try_borrow()?.frame(&field_pack.name4) {
									None => false, // this MP3 doesn't have this field
									Some(frame) => frame == first_frame,
								};
								w::AnyResult::Ok(is_equal_to_1st)
							})?;

						if frame_equal_in_all_mp3s {
							field_pack.txt.set_text(&first_frame.body().to_string());
							field_pack.chk.set_check_state_and_trigger(gui::CheckState::Checked);
						} else {
							field_pack.chk.set_check_state_and_trigger(gui::CheckState::Unchecked);
						}
					},
					None => { // no MP3 has this field
						field_pack.chk.set_check_state_and_trigger(gui::CheckState::Unchecked);
					},
				}

				w::AnyResult::Ok(())
			})
	}

	fn fill_listview_fields(&self) -> w::AnyResult<()> {
		self.lst_frames.columns().add(&[
			("Frame", 56),
			("Value", 1),
		]);
		self.lst_frames.columns().get(1).set_width_to_fill();
		self.lst_frames.set_extended_style(true, co::LVS_EX::FULLROWSELECT | co::LVS_EX::GRIDLINES);

		if self.sel_tags.len() > 1 {
			self.lst_frames.items().add(&["", &format!("{} files...", self.sel_tags.len())], None, ());
		} else {
			let sel_tag = self.sel_tags[0].try_borrow()?;
			sel_tag.frames()
				.iter()
				.for_each(|frame| {
					self.lst_frames.items().add(&[frame.name4(), &frame.body().to_string()], None, ());
				});
		}

		Ok(())
	}
}
