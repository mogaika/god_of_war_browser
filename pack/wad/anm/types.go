package anm

import (
	"math"

	"github.com/mogaika/god_of_war_browser/utils"
)

type AnimInterpolationSettings struct {
	CountOfWords        uint8  // count of words after this element
	PairedElementsCount uint8  // paired elements count?
	OffsetToElement     uint16 // offset to element
	Words               []uint16
}

func NewAnimInterpolationSettingsFromBuf(b []byte) *AnimInterpolationSettings {
	ais := &AnimInterpolationSettings{
		CountOfWords:        b[0],
		PairedElementsCount: b[1],
		OffsetToElement:     u16(b, 2),
	}
	ais.Words = make([]uint16, ais.CountOfWords)
	for i := range ais.Words {
		ais.Words[i] = u16(b, uint32(4+i*2))
	}
	return ais
}

type AnimState8Texturepos struct {
	BaseTargetDataIndex   uint16
	FlagsProbably         byte
	HowMany64kbWeNeedSkip byte
	DatasCount            uint16
	OffsetToData          uint16
	Stream                map[string][]float32
}

var defaultInterpolationSettingsForSingleElement = []byte{01, 01, 00, 00, 01, 00}

func AnimState8TextureposFromBuf(dtype *AnimDatatype, buf []byte) *AnimState8Texturepos {
	a := &AnimState8Texturepos{
		BaseTargetDataIndex:   u16(buf, 0),
		FlagsProbably:         buf[2],
		HowMany64kbWeNeedSkip: buf[3],
		DatasCount:            u16(buf, 4),
		OffsetToData:          u16(buf, 0xa),
		Stream:                make(map[string][]float32),
	}

	// dtype.param1 contain layer id
	// Game checks that dtype.Param1 != 0, otherwise this is material color special case
	// Also game checks dtype.Param1 & 0x80 != 0, otherwise it is similar material animation
	// But color animation have Type == 3, not 8, so we can skip this checks

	var interpolationSettingsBuf []byte
	if a.FlagsProbably&2 != 0 {
		interpolationSettingsBuf = buf[(uint32(a.HowMany64kbWeNeedSkip)<<16)+uint32(a.OffsetToData):]
	} else {
		interpolationSettingsBuf = defaultInterpolationSettingsForSingleElement
	}
	interpolationSettings := NewAnimInterpolationSettingsFromBuf(interpolationSettingsBuf)

	if len(interpolationSettings.Words) != 1 {
		panic("Unsuported len(word) != 1")
	}

	animTargetDataIndex := a.BaseTargetDataIndex
	animDataOffset := (uint32(a.HowMany64kbWeNeedSkip) << 16) + uint32(a.OffsetToData) + uint32(interpolationSettings.OffsetToElement)
	animDataStep := uint32(interpolationSettings.PairedElementsCount) * 4
	for animIterator := interpolationSettings.Words[0]; animIterator != 0; animIterator = ((animIterator - 1) / 2) * 2 {
		floatsDataArray := make([]float32, a.DatasCount)

		for j := range floatsDataArray {
			floatsDataArray[j] = math.Float32frombits(u32(buf, animDataOffset+uint32(j)*animDataStep))
		}

		switch animTargetDataIndex {
		case 0:
			a.Stream["U"] = floatsDataArray
		case 1:
			a.Stream["V"] = floatsDataArray
		default:
			panic("Unknown data index")
		}
		animTargetDataIndex += 1
		animDataOffset += 4
	}

	utils.LogDump(interpolationSettings, buf[:32])
	/*
		a.Data = make([]float32, a.DatasCount)
		for i := range a.Data {
			a.Data[i] = math.Float32frombits(u32(buf, uint32(0xc+i*4)))
		}
	*/
	return a
}
