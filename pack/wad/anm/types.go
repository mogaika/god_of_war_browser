package anm

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/bits"

	"github.com/mogaika/god_of_war_browser/utils"
)

type AnimInterpolationSettings struct {
	CountOfWords        uint8    // count of words after this element
	PairedElementsCount uint8    // elements count?, also count of 1 bits in entire words array
	OffsetToElement     uint16   // offset to element
	Words               []uint16 // bit mask of values or something
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

type AnimDuplicationManager struct {
	// if DatasCount1 == 0,
	// then there is others managers inside state
	// for mat layer uv animation this is count of datas
	// even if datascount3 defined
	DatasCount1  uint16 // count of shits
	DatasCount2  uint16 // delay, offset
	DatasCount3  uint16 //
	OffsetToData uint16 // low 2 bytes of 3 byte offset value
}

func (m *AnimDuplicationManager) FromBuf(b []byte) *AnimDuplicationManager {
	m.DatasCount1 = u16(b, 0)
	m.DatasCount2 = u16(b, 2)
	m.DatasCount3 = u16(b, 4)
	m.OffsetToData = u16(b, 6)
	return m
}

func NewAnimDuplicationManagerFromBuf(b []byte) *AnimDuplicationManager {
	return new(AnimDuplicationManager).FromBuf(b)
}

type AnimState8Texturepos struct {
	BaseTargetDataIndex   uint16
	FlagsProbably         byte
	HowMany64kbWeNeedSkip byte // high byte of 3 byte offset value
	Dm                    AnimDuplicationManager

	InterpolationSettings *AnimInterpolationSettings
	Stream                map[string][]float32
}

var defaultInterpolationSettingsForSingleElement = []byte{01, 01, 00, 00, 01, 00, 00, 00}

func AnimState8TextureposFromBuf(dtype *AnimDatatype, buf []byte, index int) *AnimState8Texturepos {
	stateBuf := buf[index*0xc:]

	a := &AnimState8Texturepos{
		BaseTargetDataIndex:   u16(stateBuf, 0),
		FlagsProbably:         stateBuf[2],
		HowMany64kbWeNeedSkip: stateBuf[3],

		Stream: make(map[string][]float32),
	}
	a.Dm.FromBuf(stateBuf[4:])

	if a.Dm.DatasCount1 == 0 {
		panic("ERROR: DATATYPE_TEXUREPOS DatasCount1 == 0")
		// Actually we must parse AnimDuplicationManager inside state
		// But I cannot find any of texture animation that use this behaviour
	}

	if a.Dm.DatasCount2 != 0 || a.Dm.DatasCount3 != 0 {
		// panic("ERROR: DATATYPE_TEXUREPOS DatasCount2 != 0 || DatasCount3 != 0")
		// R_MEW.WAD => eyebeam1
	}

	// dtype.param1 contain layer id
	// Game checks that dtype.Param1 != 0, otherwise this is material color special case
	// Also game checks dtype.Param1 & 0x80 != 0, otherwise it is similar material animation
	// But color animation have Type == 3, not 8, so we can skip this checks

	stateData := stateBuf[(uint32(a.HowMany64kbWeNeedSkip)<<16)+uint32(a.Dm.OffsetToData):]

	interpolationSettingsBuf := defaultInterpolationSettingsForSingleElement
	if a.FlagsProbably&2 != 0 {
		interpolationSettingsBuf = stateData
	}
	a.InterpolationSettings = NewAnimInterpolationSettingsFromBuf(interpolationSettingsBuf)

	if len(a.InterpolationSettings.Words) != 1 {
		panic("DATATYPE_TEXUREPOS Unsuported len(word) != 1")
	}

	animTargetDataIndex := a.BaseTargetDataIndex
	animDataOffset := uint32(a.InterpolationSettings.OffsetToElement)
	animDataStep := uint32(a.InterpolationSettings.PairedElementsCount) * 4
	for animIterator := a.InterpolationSettings.Words[0]; animIterator != 0; animIterator = ((animIterator - 1) / 2) * 2 {
		floatsDataArray := make([]float32, a.Dm.DatasCount1)

		for j := range floatsDataArray {
			floatsDataArray[j] = math.Float32frombits(u32(stateData, animDataOffset+uint32(j)*animDataStep))
		}

		switch animTargetDataIndex {
		case 0:
			a.Stream["U"] = floatsDataArray
		case 1:
			a.Stream["V"] = floatsDataArray
		default:
			panic("DATATYPE_TEXUREPOS Unknown data index")
		}
		animTargetDataIndex += 1
		animDataOffset += 4
	}

	return a
}

type AnimState0Skinning struct {
	// Every anim act hold its own copy of encoded quaterion array for every joint
	// Then they blended together
	BaseTargetDataIndex   uint16
	FlagsProbably         byte
	HowMany64kbWeNeedSkip byte
	Dm                    AnimDuplicationManager

	Stream                map[string][]float32
	InterpolationSettings *AnimInterpolationSettings
	SubDms                []AnimDuplicationManager
}

func AnimState0SkinningFromBuf(dtype *AnimDatatype, buf []byte, stateIndex int, _l *utils.Logger) *AnimState0Skinning {
	stateBuf := buf[stateIndex*0xc:]

	a := &AnimState0Skinning{
		BaseTargetDataIndex:   u16(stateBuf, 0),
		FlagsProbably:         stateBuf[2],
		HowMany64kbWeNeedSkip: stateBuf[3],
		Stream:                make(map[string][]float32),
	}
	a.Dm.FromBuf(stateBuf[4:])

	stateData := stateBuf[(uint32(a.HowMany64kbWeNeedSkip)<<16)+uint32(a.Dm.OffsetToData):]
	_l.Printf(">>>>>>>> STATE %d (baseDataIndex: %d, flags 0x%.2x, dm: %+v) >>>>>>>>>>>>>>>",
		stateIndex, a.BaseTargetDataIndex, a.FlagsProbably, a.Dm)

	interpolationSettingsBuf := defaultInterpolationSettingsForSingleElement

	if a.Dm.DatasCount1 == 0 {
		stateDataFirstByte := stateData[0]
		stateDataSecondByte := stateData[1]
		stateDataArrayBuf := stateData[2:]

		if a.FlagsProbably&2 != 0 {
			interpolationSettingsBuf = stateDataArrayBuf[int(stateDataSecondByte)*8:]
		}
		a.InterpolationSettings = NewAnimInterpolationSettingsFromBuf(interpolationSettingsBuf)

		_l.Printf("   ! DATA: state (fb: %d, subs count) (sb: %d, total subs)", stateDataFirstByte, stateDataSecondByte)
		_l.Printf("   ! INTERPOLATION (default: %v)  %+v", a.FlagsProbably&2 == 0, *a.InterpolationSettings)

		a.SubDms = make([]AnimDuplicationManager, stateDataSecondByte)

		for iRotationSubDm := 0; iRotationSubDm < int(stateDataFirstByte); iRotationSubDm++ {
			a.SubDms[iRotationSubDm].FromBuf(stateDataArrayBuf[iRotationSubDm*8:])
			adManager := &a.SubDms[iRotationSubDm]

			// next DatasCount2 = prev DatasCount1 + prev DataCount2 + prev DataCount - 1

			var s5Array []byte // actualy this is int8
			if a.InterpolationSettings.PairedElementsCount == 1 {
				s5Array = make([]byte, 1)
				// must be signed shift
				s5Array[0] = byte(int8(a.FlagsProbably) >> 4)
			} else {
				// must correspond to setted bits count in interpolationSettings.words array probably
				s5Array = interpolationSettingsBuf[a.InterpolationSettings.CountOfWords*2+4:]
			}
			s5Array = s5Array[:a.InterpolationSettings.PairedElementsCount]

			_l.Printf("      - SDFB %d: %+v,  s5Array: %v", iRotationSubDm, *adManager, s5Array)
			indx_t7 := 0

			// t3, a2 := range ...
			for iInterpolationWord, interpolationWord := range a.InterpolationSettings.Words {

				// ((x ^ (x & (-x))) << 16) >> 16
				// this bit magic "eats" lower non-zero bit
				// and must be applyed to int32 because of signed shifts

				for bitmask := interpolationWord; bitmask != 0; {

					// take lowest bit index, indexation from zero (if zero, then lowest bit was 1 and so)
					curBitIndex := bits.TrailingZeros16(bitmask)

					// update bitmask value
					bitmaski32 := int32(bitmask)
					bitmask = uint16(((bitmaski32 ^ (bitmaski32 & (-bitmaski32))) << 16) >> 16)

					shiftAmount := int8(s5Array[indx_t7])

					/*
						// $s0  probably position between frames in range [0.0,1.0] but in fixedpoint [0,16384]
						frame_time_fixedpoint := (1 << 14)
						var mult int32
						if a1 >= 0 {
							mult = int32(frame_time_fixedpoint >> uint(a1))
						} else {
							mult = int32(frame_time_fixedpoint << uint(-a1))
						}
						_ = mult
					*/

					//frames := make([]int8, a.Dm.DatasCount1+a.Dm.DatasCount2+a.Dm.DatasCount3)
					frames := make([]int8, adManager.DatasCount1)

					frameStep := int(a.InterpolationSettings.PairedElementsCount) * 1 // sizeof byte

					for iFrame := range frames {
						index := frameStep*iFrame + int(a.InterpolationSettings.OffsetToElement) + int(adManager.OffsetToData) + indx_t7
						frames[iFrame] = int8(stateBuf[index+(int(a.HowMany64kbWeNeedSkip)<<16)])
					}

					dataIndex := int(a.BaseTargetDataIndex) + iInterpolationWord*16 + curBitIndex

					_l.Printf("          time shift %d frames (additive) for index %d (%d): %v", shiftAmount, dataIndex, curBitIndex, frames)

					indx_t7++
				}
			}
		}

		for iOtherSubDm := int(stateDataFirstByte); iOtherSubDm < int(stateDataSecondByte); iOtherSubDm++ {
			a.SubDms[iOtherSubDm].FromBuf(stateDataArrayBuf[iOtherSubDm*8:])
			adManager := &a.SubDms[iOtherSubDm]

			_l.Printf("      - RDTDM %d: %+v", iOtherSubDm, *adManager)
			indx_t7 := 0

			for iInterpolationWord, interpolationWord := range a.InterpolationSettings.Words {
				for bitmask := interpolationWord; bitmask != 0; {
					curBitIndex := bits.TrailingZeros16(bitmask)

					// update bitmask value
					bitmaski32 := int32(bitmask)
					bitmask = uint16(((bitmaski32 ^ (bitmaski32 & (-bitmaski32))) << 16) >> 16)

					//frames := make([]int16, a.Dm.DatasCount1+a.Dm.DatasCount2+a.Dm.DatasCount3)
					frames := make([]int16, adManager.DatasCount1)
					frameStep := int(a.InterpolationSettings.PairedElementsCount) * 2 // 16 bit

					for iFrame := range frames {
						offset := frameStep*iFrame + int(a.InterpolationSettings.OffsetToElement) + int(adManager.OffsetToData) + indx_t7*2

						frames[iFrame] = int16(binary.LittleEndian.Uint16(stateBuf[offset+(int(a.HowMany64kbWeNeedSkip)<<16):]))
					}

					dataIndex := int(a.BaseTargetDataIndex) + iInterpolationWord*16 + curBitIndex

					_l.Printf("              frames (REAL) for index %d (%d): %v", dataIndex, curBitIndex, frames)
					indx_t7++
				}
			}
		}

		_ = utils.LogDump
	} else if a.Dm.DatasCount1 != 0 {
		if a.FlagsProbably&2 != 0 {
			interpolationSettingsBuf = stateData[:]
		}
		a.InterpolationSettings = NewAnimInterpolationSettingsFromBuf(interpolationSettingsBuf)

		_l.Printf("       JUST DO IT %+v", *a.InterpolationSettings)
		if a.FlagsProbably&1 == 0 {
			animDataArrayIndex := 0
			for iInterpolationWord, interpolationWord := range a.InterpolationSettings.Words {
				for bitmask := interpolationWord; bitmask != 0; {
					curBitIndex := bits.TrailingZeros16(bitmask)

					// update bitmask value
					bitmaski32 := int32(bitmask)
					bitmask = uint16(((bitmaski32 ^ (bitmaski32 & (-bitmaski32))) << 16) >> 16)

					frames := make([]float32, a.Dm.DatasCount1)
					frameStep := int(a.InterpolationSettings.PairedElementsCount) * 2 // 16 bit

					for iFrame := range frames {
						offset := frameStep*iFrame + int(a.InterpolationSettings.OffsetToElement) + animDataArrayIndex*2
						frames[iFrame] = float32(binary.LittleEndian.Uint16(stateData[offset:])) / 65536.0
					}

					dataIndex := int(a.BaseTargetDataIndex) + iInterpolationWord*16 + curBitIndex
					a.Stream[fmt.Sprintf("%d", dataIndex)] = frames
					animDataArrayIndex++
					_l.Printf("        frames (RAW) for index %d (%d): %v", dataIndex, curBitIndex, frames)
				}
			}
		} else {
			//panic("not implemented")
		}
	}

	return a
}
