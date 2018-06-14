package psarc

import (
	"encoding/binary"

	"github.com/mogaika/god_of_war_browser/utils"
)

const (
	RAW_HEADER_SIZE = 0x20
	RAW_ENTRY_SIZE  = 30
)

type Header struct {
	MagicNumber       uint32
	VersionNumber     uint32
	CompressionMethod [4]byte
	TotalTOCSize      uint32
	TOCEntrySize      uint32
	NumFiles          uint32
	BlockSize         uint32
	ArchiveFlags      uint32
}

func (h *Header) FromBuf(b []byte) {
	h.MagicNumber = binary.BigEndian.Uint32(b[0:])
	h.VersionNumber = binary.BigEndian.Uint32(b[4:])
	copy(h.CompressionMethod[:4], b[8:0xc])
	h.TotalTOCSize = binary.BigEndian.Uint32(b[0xc:])
	h.TOCEntrySize = binary.BigEndian.Uint32(b[0x10:])
	h.NumFiles = binary.BigEndian.Uint32(b[0x14:])
	h.BlockSize = binary.BigEndian.Uint32(b[0x18:])
	h.ArchiveFlags = binary.BigEndian.Uint32(b[0x1c:])
}

type Entry struct {
	MD5            [16]byte
	BlockListStart uint32
	OriginalSize   int64
	StartOffset    int64
	Name           string
}

func (e *Entry) FromBuf(b []byte) {
	copy(e.MD5[:], b[:16])
	e.BlockListStart = binary.BigEndian.Uint32(b[16:])
	e.OriginalSize = int64(utils.Read40bitUint(binary.BigEndian, b[20:]))
	e.StartOffset = int64(utils.Read40bitUint(binary.BigEndian, b[25:]))
}
