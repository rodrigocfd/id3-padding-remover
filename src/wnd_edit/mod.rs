use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;
use std::sync::Arc;
use winsafe::{gui, prelude::*};

use crate::id3v2;

mod wnd_edit_ctor;
mod wnd_edit_wm;

#[derive(Clone)]
pub struct WndEdit {
	wnd:          gui::WindowModal,
	btn_ok:       gui::Button,
	btn_cancel:   gui::Button,
	fld_artist:   Field,
	fld_title:    Field,
	fld_subtitle: Field,
	fld_album:    Field,
	fld_track:    Field,
	fld_year:     Field,
	fld_genre:    Field,

	/// Each tag is indexed by its file path.
	all_tags: Rc<RefCell<HashMap<String, id3v2::Tag>>>,
	selected_paths: Vec<String>,
}

#[derive(Clone)]
struct Field {
	chk: gui::CheckBox,
	txt: Arc<dyn GuiWindowText>,
}
