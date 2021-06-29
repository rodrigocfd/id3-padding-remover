use winsafe as w;

pub fn timer_start() -> i64 {
	w::QueryPerformanceCounter().unwrap()
}

pub fn timer_end_ms(t0: i64) -> f64 {
	let freq = w::QueryPerformanceFrequency().unwrap();
	let t1 = w::QueryPerformanceCounter().unwrap();

	((t1 - t0) as f64 / freq as f64) * 1000.0
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
