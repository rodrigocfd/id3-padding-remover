use winsafe as w;

pub fn timer_start() -> i64 {
	w::QueryPerformanceCounter().unwrap()
}

pub fn timer_end_ms(t0: i64) -> f64 {
	let freq = w::QueryPerformanceFrequency().unwrap();
	let t1 = w::QueryPerformanceCounter().unwrap();

	((t1 - t0) as f64 / freq as f64) * 1000.0
}
