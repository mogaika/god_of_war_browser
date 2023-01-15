package anm

import (
	"math/bits"
)

type AnimInstanceHeader struct {
}

type DataBitMap struct {
	PairedElementsCount uint8    // count of 1 bits in entire words array
	DataOffset          uint16   // offset to element
	Bitmap              []uint16 // bit mask of values
}

type AnimStateDescrHeader struct {
	BaseTargetDataIndex   uint16
	FlagsProbably         byte
	HowMany64kbWeNeedSkip byte // high byte of 3 byte offset value
}

func (asdh *AnimStateDescrHeader) FromBuf(descrStateBuf []byte) {
	asdh.BaseTargetDataIndex = u16(descrStateBuf, 0)
	asdh.FlagsProbably = descrStateBuf[2]
	asdh.HowMany64kbWeNeedSkip = descrStateBuf[3]
}

type AnimStateSubstream struct {
	Manager AnimSamplesManager
	Samples map[int]interface{}
}

var defaultDataBitMap = []byte{01, 01, 00, 00, 01, 00, 00, 00}

func NewDataBitMapFromBuf(b []byte) DataBitMap {
	dbm := DataBitMap{
		PairedElementsCount: b[1],
		DataOffset:          u16(b, 2),
		Bitmap:              make([]uint16, b[0]),
	}
	for i := range dbm.Bitmap {
		dbm.Bitmap[i] = u16(b, uint32(4+i*2))
	}
	return dbm
}

func (dbm *DataBitMap) Iterate(f func(bitIndex, iteration int)) {
	iteration := 0
	for iBitMapWord, bitMapWord := range dbm.Bitmap {
		for bitmask := bitMapWord; bitmask != 0; {
			// take lowest bit index, indexation from zero (if zero, then lowest bit was 1 and so)
			bitIndex := bits.TrailingZeros16(bitmask)
			bitmask = bitmaskZeroBitsShift(bitmask)
			f(iBitMapWord*16+bitIndex, iteration)
			iteration++
		}
	}
}

type AnimSamplesManager struct {
	Count        uint16
	Offset       uint16
	DatasCount3  uint16 // mapped count of samples to scale a little?
	OffsetToData uint16 // low 2bytes of 3bytes offset value
}

func (m *AnimSamplesManager) FromBuf(b []byte) *AnimSamplesManager {
	m.Count = u16(b, 0)
	m.Offset = u16(b, 2)
	m.DatasCount3 = u16(b, 4)
	m.OffsetToData = u16(b, 6)
	return m
}

func NewAnimSamplesManagerFromBuf(b []byte) *AnimSamplesManager {
	return new(AnimSamplesManager).FromBuf(b)
}
