pub fn SynchSafeEncode(mut n: u32) -> u32 {
	let mut out: u32 = 0;
	let mut mask: u32 = 0x7f;

	while (mask ^ 0x7fff_ffff) != 0 {
		out = n & !mask;
		out <<= 1;
		out |= n & mask;
		mask = ((mask + 1) << 8) - 1;
		n = out;
	}

	out
}

pub fn SynchSafeDecode(n: u32) -> u32 {
	let mut out: u32 = 0;
	let mut mask: u32 = 0x7f00_0000;

	while mask != 0 {
		out >>= 1;
		out |= n & mask;
		mask >>= 8;
	}

	out
}
