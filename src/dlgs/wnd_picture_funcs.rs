use std::cell::RefCell;
use std::rc::Rc;
use try_iterator::prelude::*;
use winsafe::{self as w};

use super::WndPicture;
use crate::id3v2;

impl WndPicture {
	pub(super) fn load_picture(
		sel_tags: Vec<Rc<RefCell<id3v2::Tag>>>,
	) -> w::AnyResult<Option<w::IPicture>> {
		let maybe_idx_first_mp3 = sel_tags
			.iter() // index of first MP3 which has APIC
			.try_position(|tag| {
				let has = tag.try_borrow()?.frame("APIC").is_some();
				w::AnyResult::Ok(has)
			})?;

		match maybe_idx_first_mp3 {
			Some(idx_first) => {
				// at last 1 MP3 has APIC
				let first_tag = sel_tags[idx_first].try_borrow()?;
				let first_apic = first_tag.frame("APIC").unwrap();

				let apic_equal_in_all_mp3s =
					sel_tags.iter().skip(idx_first + 1).try_all(|tag| {
						let is_equal_to_1st = match tag.try_borrow()?.frame("APIC") {
							None => false, // this MP3 has no APIC
							Some(apic) => apic == first_apic,
						};
						w::AnyResult::Ok(is_equal_to_1st)
					})?;

				if apic_equal_in_all_mp3s {
					let id3v2::Body::Picture(apic_data) = first_apic.body() else {
						panic!("APIC fail.")
					};
					let stream = w::SHCreateMemStream(&apic_data.data)?;
					let ipic = w::OleLoadPicture(&stream, None, true)?;
					Ok(Some(ipic))
				} else {
					Ok(None)
				}
			},
			None => Ok(None), // no MP3 has APIC
		}
	}
}
