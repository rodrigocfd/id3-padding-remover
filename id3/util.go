package id3

func synchSafeEncode(n uint32) uint32 {
	out, mask := uint32(0), uint32(0x7F)
	for (mask ^ 0x7FFFFFFF) != 0 {
		out = n & ^mask
		out <<= 1
		out |= n & mask
		mask = ((mask + 1) << 8) - 1
		n = out
	}
	return out
}

func synchSafeDecode(n uint32) uint32 {
	out, mask := uint32(0), uint32(0x7F000000)
	for mask != 0 {
		out >>= 1
		out |= n & mask
		mask >>= 8
	}
	return out
}
