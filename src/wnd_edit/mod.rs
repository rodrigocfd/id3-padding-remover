use std::cell::{Cell, RefCell};
use std::rc::Rc;
use std::sync::Arc;
use winsafe::{gui, prelude::*};

use crate::{id3v2, wnd_picture::WndPicture};

mod wnd_edit_ctor;
mod wnd_edit_wm;

#[allow(dead_code)]
#[derive(Clone)]
pub struct WndEdit {
	wnd:          gui::WindowModal,
	btn_ok:       gui::Button,
	btn_cancel:   gui::Button,
	field_packs:  Rc<RefCell<Vec<FieldPack>>>,
	wnd_pic:      WndPicture,
	btn_uncheck:  gui::Button,
	lst_frames:   gui::ListView,
	sel_tags:     Vec<Rc<RefCell<id3v2::Tag>>>,
	modal_return: Rc<Cell<bool>>,
}

trait ChildFocus: GuiWindowText + GuiChildFocus {}
impl ChildFocus for gui::ComboBox {}
impl ChildFocus for gui::Edit {}

/// Known tag field identifier, checkbox and textbox.
#[derive(Clone)]
struct FieldPack {
	name4: String,
	chk:   gui::CheckBox,
	txt:   Arc<dyn ChildFocus>,
}

const GENRES: &[&str] = &[
	"Alternative rock",
	"Axé",
	"Black metal",
	"Bluegrass",
	"Blues",
	"Blues rock",
	"Brega",
	"Britpop",
	"Country",
	"Dance",
	"Death metal",
	"Disco",
	"Doom metal",
	"Folk metal",
	"Forró",
	"Glam metal",
	"Gothic metal",
	"Grindcore",
	"Groove metal",
	"Grunge",
	"Guitar rock",
	"Hard rock",
	"Hardcore",
	"Heavy metal",
	"Hip hop",
	"Indie rock",
	"Jazz",
	"Manguebeat",
	"Nu metal",
	"Pop",
	"Pop rock",
	"Post-punk",
	"Power ballad",
	"Power metal",
	"Progressive metal",
	"Progressive rock",
	"Psychedelic rock",
	"Punk rock",
	"R&B",
	"Rap metal",
	"Reggae",
	"Rock",
	"Samba",
	"Sertanejo",
	"Smooth jazz",
	"Soul",
	"Southern rock",
	"Symphonic metal",
	"Synthpop",
	"Thrash metal",
	"Viking metal",
];
