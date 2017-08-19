package utils

import (
	"bytes"
	"encoding/binary"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const SECTOR_SIZE = 0x800

func BytesToString(bs []byte) string {
	n := bytes.IndexByte(bs, 0)
	if n < 0 {
		n = len(bs)
	}

	s, _, err := transform.Bytes(charmap.Windows1252.NewDecoder(), bs[0:n])
	if err != nil {
		panic(err)
	}
	return string(s)
}

func BytesStringLength(bs []byte) int {
	if l := bytes.IndexByte(bs, 0); l == -1 {
		return len(bs)
	} else {
		return l
	}
}

func StringToBytes(s string, bufSize int, nilTerminate bool) []byte {
	bs, _, err := transform.Bytes(charmap.Windows1252.NewEncoder(), []byte(s))
	if err != nil {
		panic(err)
	}
	if nilTerminate {
		bs = append(bs, 0)
	}
	if len(bs) < bufSize {
		r := make([]byte, bufSize)
		copy(r, bs)
		bs = r
	} else if len(bs) > bufSize {
		panic(bs)
	}
	return bs
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

func ReadBytes(out interface{}, raw []byte) {
	if err := binary.Read(bytes.NewReader(raw), binary.LittleEndian, out); err != nil {
		panic(err)
	}
}
