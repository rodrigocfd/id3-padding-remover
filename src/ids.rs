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
	pub const LST_FRAMES:         u16 = 1002;
}

pub mod fields {
	pub const DLG_FIELDS:   u16 = 2000;
	pub const CHK_ARTIST:   u16 = 2001;
	pub const TXT_ARTIST:   u16 = 2002;
	pub const CHK_TITLE:    u16 = 2003;
	pub const TXT_TITLE:    u16 = 2004;
	pub const CHK_ALBUM:    u16 = 2005;
	pub const TXT_ALBUM:    u16 = 2006;
	pub const CHK_TRACK:    u16 = 2007;
	pub const TXT_TRACK:    u16 = 2008;
	pub const CHK_DATE:     u16 = 2009;
	pub const TXT_DATE:     u16 = 2010;
	pub const CHK_GENRE:    u16 = 2011;
	pub const CMB_GENRE:    u16 = 2012;
	pub const CHK_COMPOSER: u16 = 2013;
	pub const TXT_COMPOSER: u16 = 2014;
	pub const CHK_COMMENT:  u16 = 2015;
	pub const TXT_COMMENT:  u16 = 2016;
	pub const BTN_SAVE:     u16 = 2017;
}

pub mod modify {
	pub const DLG_MODIFY:      u16 = 3000;
	pub const CHK_REM_PADDING: u16 = 3001;
	pub const CHK_REM_ALBUM:   u16 = 3002;
	pub const CHK_REM_RG:      u16 = 3003;
	pub const CHK_PREFIX_YEAR: u16 = 3004;
	pub const BTN_OK:          u16 = 3005;
	pub const BTN_CANCEL:      u16 = 3006;
}
