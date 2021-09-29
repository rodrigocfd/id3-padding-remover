//! Resource IDs.

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
	pub const CHK_ARTIST:         u16 = 1002;
	pub const TXT_ARTIST:         u16 = 1003;
	pub const CHK_TITLE:          u16 = 1004;
	pub const TXT_TITLE:          u16 = 1005;
	pub const CHK_ALBUM:          u16 = 1006;
	pub const TXT_ALBUM:          u16 = 1007;
	pub const CHK_TRACK:          u16 = 1008;
	pub const TXT_TRACK:          u16 = 1009;
	pub const CHK_DATE:           u16 = 1010;
	pub const TXT_DATE:           u16 = 1011;
	pub const CHK_GENRE:          u16 = 1012;
	pub const CMB_GENRE:          u16 = 1013;
	pub const CHK_COMPOSER:       u16 = 1014;
	pub const TXT_COMPOSER:       u16 = 1015;
	pub const CHK_COMMENT:        u16 = 1016;
	pub const TXT_COMMENT:        u16 = 1017;
	pub const BTN_SAVE:           u16 = 1018;

	pub const LST_FRAMES:         u16 = 1019;
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
