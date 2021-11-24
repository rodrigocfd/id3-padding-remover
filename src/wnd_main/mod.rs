use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use winsafe::gui;

use crate::id3v2::Tag;
use crate::wnd_fields::WndFields;

mod ids;
mod mp3_error;
mod wnd_main_events;
mod wnd_main_funcs;
mod wnd_main_menu_frames;
mod wnd_main_menu_mp3;
mod wnd_main_privs;

/// Main application window.
#[derive(Clone)]
pub struct WndMain {
	wnd:        gui::WindowMain,
	lst_mp3s:   gui::ListView,
	wnd_fields: WndFields,
	lst_frames: gui::ListView,
	tags_cache: Arc<Mutex<HashMap<String, Tag>>>,
	app_name:   String,
}

/// Did the event happened before the file item was deleted?
#[derive(PartialEq, Eq, Clone, Copy)]
pub enum PreDelete { Yes, No }

/// Operation to be performed asynchronously on a batch of MP3 tags.
#[derive(PartialEq, Eq, Clone, Copy)]
pub enum TagOp { Load, SaveAndLoad }

/// When renaming a file, should it be prefixed with the track number?
#[derive(PartialEq, Eq, Clone, Copy)]
pub enum PrefixWithTrack { Yes, No }

/// Specifies which frames should be removed from MP3 tags.
#[derive(PartialEq, Eq, Clone, Copy)]
pub enum WhatFrame { Replg, ReplgArt }
