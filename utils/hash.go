package utils

func GameStringHashNodes(str string, initial uint32) uint32 {
	hash := initial
	for _, c := range str {
		sym := uint32(byte(c))
		sym <<= 24
		sym >>= 24
		hash = (hash << 7) - hash + sym
	}

	return hash
}
