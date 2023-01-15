package anm

import (
	"math"
)

type AnimState8Texturepos struct {
	Descr      AnimStateDescrHeader
	Stream     AnimStateSubstream
	DataBitMap DataBitMap
}

func AnimState8TextureposFromBuf(stateBuf []byte) *AnimState8Texturepos {
	a := &AnimState8Texturepos{
		Stream: AnimStateSubstream{
			Samples: make(map[int]interface{}),
		},
	}
	a.Descr.FromBuf(stateBuf)
	a.Stream.Manager.FromBuf(stateBuf[4:])

	if a.Stream.Manager.Count == 0 {
		panic("ERROR: DATATYPE_TEXUREPOS Count == 0")
		// Actually we must parse AnimDuplicationManager inside state
		// But I cannot find any of texture animation that use this behaviour
	}

	if a.Stream.Manager.Offset != 0 || a.Stream.Manager.DatasCount3 != 0 {
		// panic("ERROR: DATATYPE_TEXUREPOS Offset != 0 || DatasCount3 != 0")
		// R_MEW.WAD => eyebeam1
	}

	// dtype.param1 contain layer id
	// Game checks that dtype.Param1 != 0, otherwise this is material color special case
	// Also game checks dtype.Param1 & 0x80 != 0, otherwise it is similar material animation
	// But color animation have Type == 3, not 8, so we can skip this checks

	stateData := stateBuf[(uint32(a.Descr.HowMany64kbWeNeedSkip)<<16)+uint32(a.Stream.Manager.OffsetToData):]

	dataBitMapBuf := defaultDataBitMap
	if a.Descr.FlagsProbably&2 != 0 {
		dataBitMapBuf = stateData
	}
	a.DataBitMap = NewDataBitMapFromBuf(dataBitMapBuf)

	if len(a.DataBitMap.Bitmap) != 1 {
		panic("DATATYPE_TEXUREPOS Unsuported len(word) != 1")
	}

	animTargetDataIndex := a.Descr.BaseTargetDataIndex
	animDataOffset := uint32(a.DataBitMap.DataOffset)
	animDataStep := uint32(a.DataBitMap.PairedElementsCount) * 4
	for animIterator := a.DataBitMap.Bitmap[0]; animIterator != 0; animIterator = ((animIterator - 1) / 2) * 2 {
		floatsDataArray := make([]float32, a.Stream.Manager.Count)

		for j := range floatsDataArray {
			floatsDataArray[j] = math.Float32frombits(u32(stateData, animDataOffset+uint32(j)*animDataStep))
		}

		if animTargetDataIndex == 0 || animTargetDataIndex == 1 {
			a.Stream.Samples[int(animTargetDataIndex)] = floatsDataArray
		} else {
			panic("DATATYPE_TEXUREPOS Unknown data index ")
		}
		animTargetDataIndex += 1
		animDataOffset += 4
	}

	return a
}
