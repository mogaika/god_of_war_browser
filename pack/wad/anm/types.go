package anm

import (
	"github.com/mogaika/god_of_war_browser/utils"
)

type AnimInterpolationSettings struct {
	CountOfWords        uint8  // count of words after this element
	PairedElementsCount uint8  // paired elements count?
	OffsetOrSize        uint16 // offset of data
	Words               []uint16
}

func NewAnimInterpolationSettingsFromBuf(b []byte) *AnimInterpolationSettings {
	ais := &AnimInterpolationSettings{
		CountOfWords:        b[0],
		PairedElementsCount: b[1],
		OffsetOrSize:        u16(b, 2),
	}
	ais.Words = make([]uint16, ais.CountOfWords)
	for i := range ais.Words {
		ais.Words[i] = u16(b, uint32(4+i*2))
	}
	return ais
}

type AnimState8Texturepos struct {
	ChangedDataIndex uint16
	Byte2            byte
	Byte3            byte
	DatasCount       uint32
	Type             string
	Data             []float32
	OffsetToData     uint16
}

func AnimState8TextureposFromBuf(dtype *AnimDatatype, buf []byte) *AnimState8Texturepos {
	a := &AnimState8Texturepos{
		ChangedDataIndex: u16(buf, 0),
		Byte2:            buf[2],
		Byte3:            buf[3],
		DatasCount:       u32(buf, 4),
		OffsetToData:     u16(buf, 0xa),
	}

	// dtype.param1 contain layer id
	// Game checks that dtype.Param1 != 0, otherwise this is material color special case
	// Also game checks dtype.Param1 & 0x80 != 0, otherwise it is similar material animation
	// But color animation have Type == 3, not 8, so we can skip this checks

	var interpolationSettings *AnimInterpolationSettings

	if a.Byte2&2 == 0 {
		// interpolation parameters inside binary
		// 01 01 00 00 01 00
		interpolationSettings = NewAnimInterpolationSettingsFromBuf([]byte{01, 01, 00, 00, 01, 00})
	} else {
		strangeOffset := uint32(a.Byte3) << 16
		offsetToDatas := strangeOffset + uint32(a.OffsetToData)
		interpolationSettings = NewAnimInterpolationSettingsFromBuf(buf[offsetToDatas:])
	}
	utils.LogDump(interpolationSettings, buf)
	/*
		a.Data = make([]float32, a.DatasCount)
		for i := range a.Data {
			a.Data[i] = math.Float32frombits(u32(buf, uint32(0xc+i*4)))
		}
	*/
	return a
}
