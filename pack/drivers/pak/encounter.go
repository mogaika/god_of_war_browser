package pak

import "log"

type PakIndex int8

type Encounter struct {
	Offset int64
	Size   int64
	Pak    PakIndex // 0 - means pak1, ignored when used with PACK_ADDR_ABSOLUTE
}

type PakEncounterReader struct {
	pa *PaksArray
	e  Encounter
}

func (per *PakEncounterReader) ReadAt(p []byte, off int64) (n int, err error) {
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

func (per *PakEncounterReader) Size() int64 {
	return per.e.Size
}
