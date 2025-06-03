use std::cell::Cell;
use std::rc::Rc;
use winsafe::{self as w, bind, bind_ig, bind_p, co, gui, prelude::*};

use crate::{id3v2, ids};

#[derive(Clone)]
pub struct DlgMain {
	pub(super) wnd: gui::WindowMain,
	pub(super) lst_files: gui::ListView<id3v2::Tag>,
	pub(super) cur_sort: Rc<Cell<(u32, bool)>>, // index, ascending
	pub(super) drop_target: w::IDropTarget,
}

impl DlgMain {
	#[must_use]
	pub fn new() -> Self {
		use gui::{Horz as H, Vert as V};

		let wnd = gui::WindowMain::new_dlg(ids::DLG_MAIN, Some(ids::ICO_APP), Some(ids::ACC_MAIN));
		let lst_files = gui::ListView::new_dlg(
			&wnd,
			ids::LST_FILES,
			(H::Resize, V::Resize),
			Some(ids::MNU_FILE),
		);
		let cur_sort = Rc::new(Cell::new((0, true))); // 1st col, ascending
		let drop_target = w::IDropTarget::new_impl();

		let new_self = Self { wnd, lst_files, cur_sort, drop_target };
		new_self.events();
		new_self
	}

	pub fn run(&self) -> w::AnyResult<i32> {
		self.wnd.run_main(None)
	}

	fn events(&self) {
		self.wnd
			.on()
			.wm_init_dialog(bind_ig!(self, Self::on_init_dialog))
			.wm_size(bind_p!(self, Self::on_size))
			.wm_init_menu_popup(bind_p!(self, Self::on_init_menu_popup))
			.wm_command_acc_menu(ids::MNU_FILE_OPEN, bind!(self, Self::on_menu_file_open))
			.wm_command_acc_menu(ids::MNU_FILE_EDIT, bind!(self, Self::on_menu_file_edit))
			.wm_command_acc_menu(ids::MNU_FILE_REMOVE, bind!(self, Self::on_menu_file_remove))
			.wm_command_acc_menu(ids::MNU_FILE_REWRITE, bind!(self, Self::on_menu_file_rewrite))
			.wm_command_acc_menu(ids::MNU_FILE_RENAME_TAT, {
				let self2 = self.clone();
				move || self2.rename(true)
			})
			.wm_command_acc_menu(ids::MNU_FILE_RENAME_AT, {
				let self2 = self.clone();
				move || self2.rename(false)
			})
			.wm_command_acc_menu(ids::MNU_FILE_REMRG, {
				let self2 = self.clone();
				move || self2.remove_rg_art(false)
			})
			.wm_command_acc_menu(ids::MNU_FILE_REMRGART, {
				let self2 = self.clone();
				move || self2.remove_rg_art(true)
			})
			.wm_command_acc_menu(ids::MNU_FILE_ABOUT, bind!(self, Self::on_menu_file_about));

		self.lst_files
			.on()
			.lvn_item_changed(bind_ig!(self, Self::on_lst_files_item_changed))
			.lvn_key_down(bind_p!(self, Self::on_lst_files_key_down))
			.nm_dbl_clk(bind_ig!(self, Self::on_menu_file_edit))
			.lvn_delete_item(bind_ig!(self, Self::on_lst_files_delete_item));

		self.lst_files
			.header()
			.unwrap()
			.on()
			.hdn_item_click(bind_p!(self, Self::on_header_item_click));

		self.drop_target.Drop({
			let self2 = self.clone();
			move |d: &w::IDataObject,
			      _: co::MK,
			      _: w::POINT,
			      _: &mut co::DROPEFFECT|
			      -> w::AnyResult<()> { self2.on_drop_target(d) }
		});
	}
}

pub const LIST_COLS: &[(&str, i32, &str)] = &[
	("File", 1, ""), // to fill the remaining space
	("Pad", 50, ""),
	("Art", 30, ""),
	("RG", 30, ""),
	("Artist", 90, "TPE1"),
	("T#", 30, "TRCK"),
	("Title", 100, "TIT2"),
	("Album", 100, "TALB"),
	("Year", 40, "TYER"),
	("Genre", 90, "TCON"),
	("Performer", 70, "TPE3"),
	("Composer", 70, "TCOM"),
	("Lyricist", 70, "TEXT"),
	("Orig. artist", 70, "TOPE"),
	("Comment", 70, "COMM"),
];
