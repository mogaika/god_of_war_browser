package utils

import (
	"bytes"
	"errors"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const SECTOR_SIZE = 0x800

var ErrHandlerNotFound = errors.New("File handler not found")

func BytesToString(bs []byte) string {
	n := bytes.IndexByte(bs, 0)
	if n < 0 {
		n = len(bs)
	}

	s, _, _ := transform.Bytes(charmap.Windows1252.NewDecoder(), bs[0:n])
	return string(s)
}

func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func ReverseBytes(a []byte) []byte {
	r := make([]byte, len(a))
	j := len(r)
	for _, b := range a {
		j--
		r[j] = b
	}
	return r
}
