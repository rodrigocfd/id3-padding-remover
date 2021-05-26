use winsafe as w;
use winsafe::co;

pub enum MappedFileAccess {
	Read,
	ReadWrite,
}

pub struct MappedFile {
	access: MappedFileAccess,
	hfile:  w::HFILE,
	hmap:   w::HFILEMAP,
	hview:  w::HFILEMAPVIEW,
	size:   usize,
}

impl Drop for MappedFile {
	fn drop(&mut self) {
		if !self.hview.is_null() { self.hview.UnmapViewOfFile().unwrap(); }
		if !self.hmap.is_null() { self.hmap.CloseHandle().unwrap(); }
		if !self.hfile.is_null() { self.hfile.CloseHandle().unwrap(); }
	}
}

impl MappedFile {
	pub fn open(file_path: &str, access: MappedFileAccess) -> w::WinResult<Self> {
		let (hfile, _) = w::HFILE::CreateFile(
			file_path,
			match access {
				MappedFileAccess::Read => co::GENERIC::READ,
				MappedFileAccess::ReadWrite => co::GENERIC::READ | co::GENERIC::WRITE,
			},
			match access {
				MappedFileAccess::Read => co::FILE_SHARE::READ,
				MappedFileAccess::ReadWrite => co::FILE_SHARE::NONE,
			},
			None,
			match access {
				MappedFileAccess::Read => co::DISPOSITION::OPEN_EXISTING,
				MappedFileAccess::ReadWrite => co::DISPOSITION::OPEN_ALWAYS
			},
			co::FILE_ATTRIBUTE::NORMAL,
			None,
		)?;

		let mut new_self = Self {
			access,
			hfile,
			hmap: w::HFILEMAP::NULL,
			hview: w::HFILEMAPVIEW::NULL,
			size: 0,
		};
		new_self.map_in_memory()?;
		Ok(new_self)
	}

	fn map_in_memory(&mut self) -> w::WinResult<()> {
		self.hmap = self.hfile.CreateFileMapping(
			None,
			match self.access {
				MappedFileAccess::Read => co::PAGE::READONLY,
				MappedFileAccess::ReadWrite => co::PAGE::READWRITE,
			},
			None,
			None,
		)?;

		self.hview = self.hmap.MapViewOfFile(
			match self.access {
				MappedFileAccess::Read => co::FILE_MAP::READ,
				MappedFileAccess::ReadWrite => co::FILE_MAP::READ | co::FILE_MAP::WRITE,
			},
			0,
			None,
		)?;

		self.size = self.hfile.GetFileSizeEx()?;
		Ok(())
	}

	pub fn size(&self) -> usize {
		self.size
	}

	pub fn as_slice(&self) -> &[u8] {
		self.hview.as_slice(self.size)
	}

	pub fn as_mut_slice(&mut self) -> &mut [u8] {
		self.hview.as_mut_slice(self.size)
	}

	pub fn resize(&mut self, num_bytes: usize) -> w::WinResult<()> {
		self.hview.UnmapViewOfFile()?;
		self.hmap.CloseHandle()?;

		self.hfile.SetFilePointerEx(num_bytes as _, co::FILE_STARTING_POINT::BEGIN)?;
		self.hfile.SetEndOfFile()?;
		self.hfile.SetFilePointerEx(0, co::FILE_STARTING_POINT::BEGIN)?;

		self.map_in_memory()
	}
}