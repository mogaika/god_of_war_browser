package anm

import (
	"encoding/binary"
	"math"
	"math/bits"

	"github.com/mogaika/god_of_war_browser/utils"
)

type AnimInterpolationSettings struct {
	// TODO: rename struct
	CountOfWords        uint8    // count of words after this element
	PairedElementsCount uint8    // elements count?, also count of 1 bits in entire words array
	OffsetToElement     uint16   // offset to element
	Words               []uint16 // bit mask of values or something
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

func NewAnimInterpolationSettingsFromBuf(b []byte) AnimInterpolationSettings {
	ais := AnimInterpolationSettings{
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

type AnimState8Texturepos struct {
	Descr                 AnimStateDescrHeader
	Stream                AnimStateSubstream
	InterpolationSettings AnimInterpolationSettings
}

var defaultInterpolationSettingsForSingleElement = []byte{01, 01, 00, 00, 01, 00, 00, 00}

func AnimState8TextureposFromBuf(dtype *AnimDatatype, buf []byte, index int) *AnimState8Texturepos {
	stateBuf := buf[index*0xc:]

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

	interpolationSettingsBuf := defaultInterpolationSettingsForSingleElement
	if a.Descr.FlagsProbably&2 != 0 {
		interpolationSettingsBuf = stateData
	}
	a.InterpolationSettings = NewAnimInterpolationSettingsFromBuf(interpolationSettingsBuf)

	if len(a.InterpolationSettings.Words) != 1 {
		panic("DATATYPE_TEXUREPOS Unsuported len(word) != 1")
	}

	animTargetDataIndex := a.Descr.BaseTargetDataIndex
	animDataOffset := uint32(a.InterpolationSettings.OffsetToElement)
	animDataStep := uint32(a.InterpolationSettings.PairedElementsCount) * 4
	for animIterator := a.InterpolationSettings.Words[0]; animIterator != 0; animIterator = ((animIterator - 1) / 2) * 2 {
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

type AnimState0Skinning struct {
	// Every anim act hold its own copy of encoded quaterion array for every joint
	// Then they blended together
	RotationDescr AnimStateDescrHeader
	PositionDescr AnimStateDescrHeader

	Stream          AnimStateSubstream
	SubStreamsRough []AnimStateSubstream
	SubStreamsAdd   []AnimStateSubstream

	RotationInterpolation AnimInterpolationSettings
}

func bitmaskZeroBitsShift(bitmask uint16) uint16 {
	bitmaski32 := int32(bitmask)
	return uint16(((bitmaski32 ^ (bitmaski32 & (-bitmaski32))) << 16) >> 16)
}

func (a *AnimState0Skinning) getInterpolationSettingsOffset(descr *AnimStateDescrHeader, stateData []byte) []byte {
	interpolationSettingsBuf := defaultInterpolationSettingsForSingleElement
	if descr.FlagsProbably&2 != 0 {
		interpolationSettingsBuf = stateData[2+int(stateData[1])*8:]
	}
	return interpolationSettingsBuf
}

func (a *AnimState0Skinning) GetInterpolationSettings(descr *AnimStateDescrHeader, stateData []byte) AnimInterpolationSettings {
	return NewAnimInterpolationSettingsFromBuf(a.getInterpolationSettingsOffset(descr, stateData))
}

func (a *AnimState0Skinning) GetShiftsArray(descr *AnimStateDescrHeader, interpolation *AnimInterpolationSettings, stateData []byte) []int8 {
	shifts := make([]int8, interpolation.PairedElementsCount)
	if interpolation.PairedElementsCount == 1 {
		shifts[0] = int8(descr.FlagsProbably) >> 4
	} else {
		s5ArrayRaw := a.getInterpolationSettingsOffset(descr, stateData)[interpolation.CountOfWords*2+4:]
		for i := range shifts {
			shifts[i] = int8(s5ArrayRaw[i])
		}
	}
	return shifts
}

func AnimState0SkinningFromBuf(dtype *AnimDatatype, buf []byte, stateIndex int, _l *utils.Logger) *AnimState0Skinning {
	stateBuf := buf[stateIndex*0xc:]
	a := &AnimState0Skinning{}
	a.RotationDescr.FromBuf(stateBuf)
	a.Stream.Manager.FromBuf(stateBuf[4:])
	return a
}

func (a *AnimState0Skinning) ParseRotations(dtype *AnimDatatype, buf []byte, stateIndex int, _l *utils.Logger) {
	stateBuf := buf[stateIndex*0xc:]
	stateData := stateBuf[(uint32(a.RotationDescr.HowMany64kbWeNeedSkip)<<16)+uint32(a.Stream.Manager.OffsetToData):]
	_l.Printf(">>>>>>>> STATE %d (baseDataIndex: %d, flags 0x%.2x, sm: %+v) >>>>>>>>>>>>>>>",
		stateIndex, a.RotationDescr.BaseTargetDataIndex, a.RotationDescr.FlagsProbably, a.Stream.Manager)

	a.RotationInterpolation = a.GetInterpolationSettings(&a.RotationDescr, stateData)

	// Parse rotations
	if a.Stream.Manager.Count == 0 {
		stateDataFirstByte := stateData[0]
		stateDataSecondByte := stateData[1]
		stateDataArrayBuf := stateData[2:]

		_l.Printf("   ! DATA: state (fb: %d, subs count) (sb: %d, total subs)", stateDataFirstByte, stateDataSecondByte)
		_l.Printf("   ! INTERPOLATION (default: %v)  %+v", a.RotationDescr.FlagsProbably&2 == 0, a.RotationInterpolation)

		a.SubStreamsAdd = make([]AnimStateSubstream, stateDataFirstByte)
		for iRotationSubDm := 0; iRotationSubDm < int(stateDataFirstByte); iRotationSubDm++ {
			subStream := &a.SubStreamsAdd[iRotationSubDm]
			subSm := &subStream.Manager
			subSm.FromBuf(stateDataArrayBuf[iRotationSubDm*8:])
			subStream.Samples = make(map[int]interface{})

			shifts := a.GetShiftsArray(&a.RotationDescr, &a.RotationInterpolation, stateData)
			_l.Printf("      - SDFB %d: %+v,  s5Array: %v", iRotationSubDm, *subSm, shifts)
			indx_t7 := 0
			shiftAmountMap := make(map[int]int32)

			// t3, a2 := range ...
			for iInterpolationWord, interpolationWord := range a.RotationInterpolation.Words {

				// ((x ^ (x & (-x))) << 16) >> 16
				// this bit magic "eats" lower non-zero bit
				// and must be applyed to int32 because of signed shifts
				for bitmask := interpolationWord; bitmask != 0; {
					// take lowest bit index, indexation from zero (if zero, then lowest bit was 1 and so)
					curBitIndex := bits.TrailingZeros16(bitmask)
					bitmask = bitmaskZeroBitsShift(bitmask)

					shiftAmount := shifts[indx_t7]
					/*
						// $s0  probably position between frames in range [0.0,1.0] but in fixedpoint [0,16384]
						frame_time_fixedpoint := (1 << 14)
						var mult int32
						if shiftAmount >= 0 {
							mult = int32(frame_time_fixedpoint >> uint(shiftAmount))
						} else {
							mult = int32(frame_time_fixedpoint << uint(-shiftAmount))
						}
						_ = mult
					*/

					frames := make([]int8, subSm.Count)
					frameStep := int(a.RotationInterpolation.PairedElementsCount) * 1 // sizeof byte

					for iFrame := range frames {
						index := frameStep*iFrame + int(a.RotationInterpolation.OffsetToElement) + int(subSm.OffsetToData) + indx_t7
						frames[iFrame] = int8(stateBuf[index+(int(a.RotationDescr.HowMany64kbWeNeedSkip)<<16)])
					}

					dataIndex := int(a.RotationDescr.BaseTargetDataIndex) + iInterpolationWord*16 + curBitIndex
					subStream.Samples[dataIndex] = frames
					shiftAmountMap[dataIndex] = int32(shiftAmount)

					_l.Printf("          time shift %d frames (additive) for index %d (%d): %v", shiftAmount, dataIndex, curBitIndex, frames)

					indx_t7++
				}
			}
			subStream.Samples[-100] = shiftAmountMap
		}

		a.SubStreamsRough = make([]AnimStateSubstream, stateDataSecondByte-stateDataFirstByte)
		for iOtherSubDm := int(stateDataFirstByte); iOtherSubDm < int(stateDataSecondByte); iOtherSubDm++ {
			subStream := &a.SubStreamsRough[iOtherSubDm-int(stateDataFirstByte)]
			subSm := &subStream.Manager
			subSm.FromBuf(stateDataArrayBuf[iOtherSubDm*8:])
			subStream.Samples = make(map[int]interface{})

			_l.Printf("      - halfs %d: %+v", iOtherSubDm, *subSm)
			indx_t7 := 0

			for iInterpolationWord, interpolationWord := range a.RotationInterpolation.Words {
				for bitmask := interpolationWord; bitmask != 0; {
					curBitIndex := bits.TrailingZeros16(bitmask)
					bitmask = bitmaskZeroBitsShift(bitmask)

					frames := make([]int16, subSm.Count)
					frameStep := int(a.RotationInterpolation.PairedElementsCount) * 2 // 16 bit

					for iFrame := range frames {
						offset := frameStep*iFrame + int(a.RotationInterpolation.OffsetToElement) + int(subSm.OffsetToData) + indx_t7*2

						frames[iFrame] = int16(binary.LittleEndian.Uint16(stateBuf[offset+(int(a.RotationDescr.HowMany64kbWeNeedSkip)<<16):]))
					}

					dataIndex := int(a.RotationDescr.BaseTargetDataIndex) + iInterpolationWord*16 + curBitIndex
					subStream.Samples[dataIndex] = frames

					_l.Printf("              frames (halfs) for index %d (%d): %v", dataIndex, curBitIndex, frames)
					indx_t7++
				}
			}
		}

		_ = utils.LogDump
	} else if a.Stream.Manager.Count != 0 {
		a.Stream.Samples = make(map[int]interface{})

		_l.Printf("       RAW %+v  flag %v", a.RotationInterpolation, a.RotationDescr.FlagsProbably&1 != 0)
		if a.RotationDescr.FlagsProbably&1 == 0 {
			animDataArrayIndex := 0
			for iInterpolationWord, interpolationWord := range a.RotationInterpolation.Words {
				for bitmask := interpolationWord; bitmask != 0; {
					curBitIndex := bits.TrailingZeros16(bitmask)
					bitmask = bitmaskZeroBitsShift(bitmask)

					frames := make([]int32, a.Stream.Manager.Count)
					frameStep := int(a.RotationInterpolation.PairedElementsCount) * 2 // 16 bit

					for iFrame := range frames {
						offset := frameStep*iFrame + int(a.RotationInterpolation.OffsetToElement) + animDataArrayIndex*2
						frames[iFrame] = int32(int16(binary.LittleEndian.Uint16(stateData[offset:])))
					}

					dataIndex := int(a.RotationDescr.BaseTargetDataIndex) + iInterpolationWord*16 + curBitIndex
					a.Stream.Samples[dataIndex] = frames
					animDataArrayIndex++
					_l.Printf("        frames (RAW) for index %d (%d): %v", dataIndex, curBitIndex, frames)
				}
			}
		} else {
			_l.Printf("       RAW ADDITIVE %+v  flag %v", a.RotationInterpolation, a.RotationDescr.FlagsProbably&1 != 0)
			//panic("not implemented")
		}
	}
}

func (a *AnimState0Skinning) ParsePositions(dtype *AnimDatatype, buf []byte, stateIndex int, _l *utils.Logger, rawAct []byte) {
	stateData := rawAct[binary.LittleEndian.Uint32(rawAct[0x80:])+uint32(stateIndex*0xc):]
	a.PositionDescr.FromBuf(stateData)
	utils.LogDump(stateData[:0x100])

}
