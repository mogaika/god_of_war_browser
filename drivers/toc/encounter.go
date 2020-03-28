package toc

import (
	"io"
	"log"
)

type PakIndex int8

type Encounter struct {
	Offset int64
	Size   int64
	Pak    PakIndex // 0 - means pak1, ignored when used with PACK_ADDR_ABSOLUTE
}

type EncounterReaderWriter struct {
	pa *PaksArray
	e  Encounter
}

func (per *EncounterReaderWriter) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= per.e.Size {
		return 0, io.EOF
	}
	switch per.pa.addrType {
	case PACK_ADDR_ABSOLUTE:
		return per.pa.absoluteReadWriteAt(p, per.e.Offset+off, true)
	case PACK_ADDR_INDEX:
		return per.pa.paks[per.e.Pak].ReadAt(p, per.e.Offset+off)
	default:
		log.Panicf("Unknown addrType for reading: %v", per.pa.addrType)
		return 0, nil
	}
}

func (per *EncounterReaderWriter) WriteAt(p []byte, off int64) (n int, err error) {
	switch per.pa.addrType {
	case PACK_ADDR_ABSOLUTE:
		return per.pa.absoluteReadWriteAt(p, per.e.Offset+off, false)
	case PACK_ADDR_INDEX:
		return per.pa.paks[per.e.Pak].WriteAt(p, per.e.Offset+off)
	default:
		log.Panicf("Unknown addrType for reading: %v", per.pa.addrType)
		return 0, nil
	}
}

func (per *EncounterReaderWriter) Size() int64 {
	return per.e.Size
}
