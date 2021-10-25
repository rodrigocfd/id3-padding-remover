use winsafe::{self as w, WinResult};

pub struct Timer(i64);

impl Timer {
	pub fn start() -> WinResult<Self> {
		Ok( Self(w::QueryPerformanceCounter()?) )
	}

	pub fn now_ms(&self) -> WinResult<f64> {
		let freq = w::QueryPerformanceFrequency()?;
		let t1 = w::QueryPerformanceCounter()?;
		Ok( ((t1 - self.0) as f64 / freq as f64) * 1000.0 )
	}
}

pub fn clear_diacritics(s: &str) -> String {
	const SRC: &str = "ÁáÀàÃãÂâÄäÉéÈèÊêËëÍíÌìÎîÏïÓóÒòÕõÔôÖöÚúÙùÛûÜüÇçÅåÐðÑñØøÝýÿ";
	const DST: &str = "AaAaAaAaAaEeEeEeEeIiIiIiIiOoOoOoOoOoUuUuUuUuCcAaDdNnOoYyy";

	let mut out = String::with_capacity(s.len());

	for (_, ch) in s.chars().enumerate() {
		let mut replaced = false;

		for (diac_idx, diac_ch) in SRC.chars().enumerate() {
			if ch == diac_ch {
				out.push(DST.chars().nth(diac_idx).unwrap());
				replaced = true;
				break;
			}
		}

		if !replaced {
			out.push(ch);
		}
	}

	out
}

pub fn format_bytes(num_bytes: usize) -> String {
	if num_bytes < 1024 {
		format!("{} bytes", num_bytes)
	} else if num_bytes < 1024 * 1024 {
		format!("{:.2} KB", (num_bytes as f64) / 1024.0)
	} else if num_bytes < 1024 * 1024 * 1024 {
		format!("{:.2} MB", (num_bytes as f64) / 1024.0 / 1024.0)
	} else if num_bytes < 1024 * 1024 * 1024 * 1024 {
		format!("{:.2} GB", (num_bytes as f64) / 1024.0 / 1024.0 / 1024.0)
	} else if num_bytes < 1024 * 1024 * 1024 * 1024 * 1024 {
		format!("{:.2} TB", (num_bytes as f64) / 1024.0 / 1024.0 / 1024.0 / 1024.0)
	} else {
		format!("{:.2} PB", (num_bytes as f64) / 1024.0 / 1024.0 / 1024.0 / 1024.0 / 1024.0)
	}
}

pub fn app_name_from_res() -> WinResult<String> {
	let exe_name = w::HINSTANCE::NULL.GetModuleFileName()?;
	let res_info = w::ResourceInfo::read_from(&exe_name)?;
	let (lang, cp) = res_info.langs_and_code_pages().unwrap()[0];
	Ok(res_info.product_name(lang, cp).unwrap())
}

pub mod prompt {
	use winsafe::{self as w, co, HWND, WinResult};

	pub fn err(
		hwnd: HWND, title: &str, instruc: Option<&str>, body: &str) -> WinResult<co::DLGID>
	{
		base(hwnd, title, instruc, body, co::TDCBF::OK, co::TD_ICON::ERROR)
	}

	pub fn info(
		hwnd: HWND, title: &str, instruc: Option<&str>, body: &str) -> WinResult<co::DLGID>
	{
		base(hwnd, title, instruc, body, co::TDCBF::OK, co::TD_ICON::INFORMATION)
	}

	pub fn ok_cancel(
		hwnd: HWND, title: &str, instruc: Option<&str>, body: &str) -> WinResult<co::DLGID>
	{
		base(hwnd, title, instruc, body, co::TDCBF::OK | co::TDCBF::CANCEL, co::TD_ICON::WARNING)
	}

	fn base(hwnd: HWND, title: &str, instruc: Option<&str>,
		body: &str, btns: co::TDCBF, ico: co::TD_ICON) -> WinResult<co::DLGID>
	{
		let mut tdc = w::TASKDIALOGCONFIG::default();
		tdc.hwndParent = hwnd;
		tdc.dwFlags = co::TDF::ALLOW_DIALOG_CANCELLATION;
		tdc.dwCommonButtons = btns;
		tdc.set_pszMainIcon(w::IconIdTdicon::Tdicon(ico));

		let mut title = w::WString::from_str(title);
		tdc.set_pszWindowTitle(Some(&mut title));

		let mut instruc = instruc.map(|s| w::WString::from_str(s));
		tdc.set_pszMainInstruction(instruc.as_mut().map(|s| s));

		let mut body = w::WString::from_str(body);
		tdc.set_pszContent(Some(&mut body));

		let (res, _) = w::TaskDialogIndirect(&mut tdc, None)?;
		Ok(res)
	}
}