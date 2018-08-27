package crypto

func GetMaskLo(seed uint32) uint32 {
	return (((^seed) ^ 0x00001357) << 16) | (((seed) ^ 0xffffaaaa) & 0x0000ffff)
}

func GetMaskHi(seed uint32) uint32 {
	return (((seed) ^ 0x43210000) >> 16) | (((^seed) ^ 0xabcdffff) & 0xffff0000)
}
