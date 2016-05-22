package utils

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
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

func BytesToVec4f(bs []byte) [4]float32 {
	return [4]float32{
		math.Float32frombits(binary.LittleEndian.Uint32(bs[0:4])),
		math.Float32frombits(binary.LittleEndian.Uint32(bs[4:8])),
		math.Float32frombits(binary.LittleEndian.Uint32(bs[8:12])),
		math.Float32frombits(binary.LittleEndian.Uint32(bs[12:16])),
	}
}

func BytesToVec4i(bs []byte) [4]int32 {
	return [4]int32{
		int32(binary.LittleEndian.Uint32(bs[0:4])),
		int32(binary.LittleEndian.Uint32(bs[4:8])),
		int32(binary.LittleEndian.Uint32(bs[8:12])),
		int32(binary.LittleEndian.Uint32(bs[12:16])),
	}
}

func BytesToVec4u(bs []byte) [4]uint32 {
	return [4]uint32{
		binary.LittleEndian.Uint32(bs[0:4]),
		binary.LittleEndian.Uint32(bs[4:8]),
		binary.LittleEndian.Uint32(bs[8:12]),
		binary.LittleEndian.Uint32(bs[12:16]),
	}
}

func BytesToMat4f(bs []byte) [16]float32 {
	var v [16]float32
	for i := range v {
		v[i] = math.Float32frombits(binary.LittleEndian.Uint32(bs[i*4 : i*4+4]))
	}
	return v
}
