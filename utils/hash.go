package utils

func GameStringHashNodes(str string, initial uint32) uint32 {
	hash := initial
	for _, c := range str {
		hash = hash*127 + uint32(byte(c))
	}
	return hash
}
