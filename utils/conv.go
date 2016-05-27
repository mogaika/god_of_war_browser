package utils

import (
	"bytes"
	"errors"
)

const SECTOR_SIZE = 0x800

var ErrHandlerNotFound = errors.New("File handler not found")

func BytesToString(bs []byte) string {
	n := bytes.IndexByte(bs, 0)
	if n < 0 {
		n = len(bs)
	}
	return string(bs[0:n])
}
