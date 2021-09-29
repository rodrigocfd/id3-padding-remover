use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use winsafe::gui;

use crate::id3v2::Tag;

mod wnd_main_events;
mod wnd_main_funcs;
mod wnd_main_menu;

#[derive(Clone)]
pub struct WndMain {
	wnd:       gui::WindowMain,
	lst_files: gui::ListView,

	chk_artist:   gui::CheckBox,
	txt_artist:   gui::Edit,
	chk_title:    gui::CheckBox,
	txt_title:    gui::Edit,
	chk_album:    gui::CheckBox,
	txt_album:    gui::Edit,
	chk_track:    gui::CheckBox,
	txt_track:    gui::Edit,
	chk_date:     gui::CheckBox,
	txt_date:     gui::Edit,
	chk_genre:    gui::CheckBox,
	cmb_genre:    gui::ComboBox,
	chk_composer: gui::CheckBox,
	txt_composer: gui::Edit,
	chk_comment:  gui::CheckBox,
	txt_comment:  gui::Edit,
	btn_save:     gui::Button,

	lst_frames: gui::ListView,
	resizer:    gui::Resizer,
	tags_cache: Rc<RefCell<HashMap<String, Tag>>>,
	app_name:   String,
}

pub enum PreDelete { Yes, No }
