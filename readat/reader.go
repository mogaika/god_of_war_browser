package readat

import (
	"encoding/binary"
	"io"
	"log"
	"math"
)

type Reader struct {
	source io.ReaderAt
	offset int64
}

func NewReader(source io.ReaderAt, offset int64) *Reader {
	return &Reader{
		source: source,
		offset: offset,
	}
}

func (r *Reader) Offset() int64 {
	return r.offset
}

func (r *Reader) SubReader(offset int64) *Reader {
	return &Reader{
		source: r.source,
		offset: r.offset + offset,
	}
}

func (r *Reader) ReadAt(p []byte, off int64) (n int, err error) {
	//log.Printf("Read %d at 0x%x+0x%x=0x%x", len(p), r.offset, off, r.offset+off)
	_ = log.Printf
	return r.source.ReadAt(p, r.offset+off)
}

func (r *Reader) ReadAtP(p []byte, off int64) (n int) {
	if n, err := r.ReadAt(p, off); err != nil {
		panic(err)
	} else {
		return n
	}
}

func (r *Reader) ReadAtBP(size int64, off int64) ([]byte, int) {
	buffer := make([]byte, size)
	return buffer, r.ReadAtP(buffer, off)
}

func (r *Reader) ReadU64LE(off int64) uint64 {
	var b [8]byte
	r.ReadAtP(b[:], off)
	return binary.LittleEndian.Uint64(b[:])
}
func (r *Reader) ReadI64LE(off int64) int64 { return int64(r.ReadU64LE(off)) }

func (r *Reader) ReadU64BE(off int64) uint64 {
	var b [8]byte
	r.ReadAtP(b[:], off)
	return binary.BigEndian.Uint64(b[:])
}
func (r *Reader) ReadI64BE(off int64) int64 { return int64(r.ReadU64BE(off)) }

func (r *Reader) ReadU32LE(off int64) uint32 {
	var b [4]byte
	r.ReadAtP(b[:], off)
	return binary.LittleEndian.Uint32(b[:])
}
func (r *Reader) ReadI32LE(off int64) int32 { return int32(r.ReadU32LE(off)) }

func (r *Reader) ReadU32BE(off int64) uint32 {
	var b [4]byte
	r.ReadAtP(b[:], off)
	return binary.BigEndian.Uint32(b[:])
}
func (r *Reader) ReadI32BE(off int64) int32 { return int32(r.ReadU32BE(off)) }

func (r *Reader) ReadU16LE(off int64) uint16 {
	var b [2]byte
	r.ReadAtP(b[:], off)
	return binary.LittleEndian.Uint16(b[:])
}
func (r *Reader) ReadI16LE(off int64) int16 { return int16(r.ReadU16LE(off)) }

func (r *Reader) ReadU16BE(off int64) uint16 {
	var b [2]byte
	r.ReadAtP(b[:], off)
	return binary.BigEndian.Uint16(b[:])
}
func (r *Reader) ReadI16BE(off int64) int16 { return int16(r.ReadU16BE(off)) }

func (r *Reader) ReadU8(off int64) uint8 {
	var b [1]byte
	r.ReadAtP(b[:], off)
	return b[0]
}
func (r *Reader) ReadI8(off int64) int8 { return int8(r.ReadU8(off)) }

func (r *Reader) ReadF32LE(off int64) float32 { return math.Float32frombits(r.ReadU32LE(off)) }
func (r *Reader) ReadF32BE(off int64) float32 { return math.Float32frombits(r.ReadU32BE(off)) }

func (r *Reader) ReadF64LE(off int64) float64 { return math.Float64frombits(r.ReadU64LE(off)) }
func (r *Reader) ReadF64BE(off int64) float64 { return math.Float64frombits(r.ReadU64BE(off)) }
