package utils

import (
	"bytes"
	"encoding/binary"

	"github.com/mogaika/god_of_war_browser/config"

	"golang.org/x/text/transform"
)

const SECTOR_SIZE = 0x800

func GetRequiredSectorsCount(size int64) int64 {
	return (size + SECTOR_SIZE - 1) / SECTOR_SIZE
}

func BytesToString(bs []byte) string {
	n := bytes.IndexByte(bs, 0)
	if n < 0 {
		n = len(bs)
	}

	s, _, err := transform.Bytes(config.GetEncoding().NewDecoder(), bs[0:n])
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

func StringToBytesBuffer(s string, bufSize int, nilTerminate bool) []byte {
	bs, _, err := transform.Bytes(config.GetEncoding().NewEncoder(), []byte(s))
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

func StringToBytes(s string, nilTerminate bool) []byte {
	bs, _, err := transform.Bytes(config.GetEncoding().NewEncoder(), []byte(s))
	if err != nil {
		panic(err)
	}

	if nilTerminate {
		bs = append(bs, 0)
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

func AsBytes(data interface{}) []byte {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.LittleEndian, data); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func Read40bitUint(o binary.ByteOrder, bin []byte) uint64 {
	var buf [8]byte
	if o == binary.LittleEndian {
		copy(buf[0:], bin[:5])
	} else {
		copy(buf[3:], bin[:5])
	}
	return o.Uint64(buf[:])
}

func Read24bitUint(o binary.ByteOrder, bin []byte) uint32 {
	var buf [4]byte
	if o == binary.LittleEndian {
		copy(buf[0:], bin[:3])
	} else {
		copy(buf[1:], bin[:3])
	}
	return o.Uint32(buf[:])
}
