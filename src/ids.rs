//! Resource IDs.

pub const APP_TITLE: &str = "ID3 Fit";

pub mod main {
	pub const ICO_FROG:           u16 = 101;

	pub const MNU_MAIN:           u16 = 200;
	pub const MNU_FILE_OPEN:      u16 = 201;
	pub const MNU_FILE_EXCSEL:    u16 = 202;
	pub const MNU_FILE_MODIFY:    u16 = 203;
	pub const MNU_FILE_CLR_DIACR: u16 = 204;
	pub const MNU_FILE_ABOUT:     u16 = 205;

	pub const ACT_MAIN:           u16 = 300;

	pub const DLG_MAIN:           u16 = 1000;
	pub const LST_FILES:          u16 = 1001;
	pub const LST_FRAMES:         u16 = 1002;
}

pub mod modify {
	pub const DLG_MODIFY:      u16 = 2000;
	pub const CHK_REM_PADDING: u16 = 2001;
	pub const CHK_REM_ALBUM:   u16 = 2002;
	pub const CHK_REM_RG:      u16 = 2003;
	pub const CHK_PREFIX_YEAR: u16 = 2004;
	pub const BTN_OK:          u16 = 2005;
	pub const BTN_CANCEL:      u16 = 2006;
}
