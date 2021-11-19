use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use winsafe::gui;

use crate::id3v2::Tag;
use crate::wnd_fields::WndFields;

mod ids;
mod wnd_main_events;
mod wnd_main_funcs;
mod wnd_main_menu;
mod wnd_main_privs;

#[derive(Clone)]
pub struct WndMain {
	wnd:        gui::WindowMain,
	lst_mp3s:   gui::ListView,
	wnd_fields: WndFields,
	lst_frames: gui::ListView,
	tags_cache: Arc<Mutex<HashMap<String, Tag>>>,
	app_name:   String,
}

pub enum PreDelete { Yes, No }
