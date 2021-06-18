package util

func SynchSafeEncode(n uint32) uint32 {
	out, mask := uint32(0), uint32(0x7f)
	for (mask ^ 0x7fff_ffff) != 0 {
		out = n & ^mask
		out <<= 1
		out |= n & mask
		mask = ((mask + 1) << 8) - 1
		n = out
	}
	return out
}

func SynchSafeDecode(n uint32) uint32 {
	out, mask := uint32(0), uint32(0x7f00_0000)
	for mask != 0 {
		out >>= 1
		out |= n & mask
		mask >>= 8
	}
	return out
}

func IsSliceZeroed(blob []byte) bool {
	for _, b := range blob {
		if b != 0x00 {
			return false
		}
	}
	return true
}
