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
