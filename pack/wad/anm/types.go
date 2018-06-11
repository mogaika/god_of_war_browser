package anm

import (
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
	DatasCount1  uint16 // Datas count intro part?
	DatasCount2  uint16 // Datas count outro part?
	DatasCount3  uint16 // Datas count repeated part?, can be bigger than DatasCount1
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

func AnimState0SkinningFromBuf(dtype *AnimDatatype, buf []byte, stateIndex int) *AnimState0Skinning {
	stateBuf := buf[stateIndex*0xc:]

	a := &AnimState0Skinning{
		BaseTargetDataIndex:   u16(stateBuf, 0),
		FlagsProbably:         stateBuf[2],
		HowMany64kbWeNeedSkip: stateBuf[3],
		Stream:                make(map[string][]float32),
	}
	a.Dm.FromBuf(stateBuf[4:])

	stateData := stateBuf[(uint32(a.HowMany64kbWeNeedSkip)<<16)+uint32(a.Dm.OffsetToData):]
	if a.Dm.DatasCount1 == 0 {
		interpolationSettingsBuf := defaultInterpolationSettingsForSingleElement

		stateDataFirstByte := stateData[0]
		stateDataSecondByte := stateData[1]
		stateDataArrayBuf := stateData[2:]

		if a.FlagsProbably&2 != 0 {
			interpolationSettingsBuf = stateData[2+int(stateDataSecondByte)*8:]
		}

		a.InterpolationSettings = NewAnimInterpolationSettingsFromBuf(interpolationSettingsBuf)

		a.SubDms = make([]AnimDuplicationManager, stateDataFirstByte)

		if stateDataFirstByte != 0 {
			// cycle over strange data
			for iSDFB := 0; iSDFB < int(stateDataFirstByte); iSDFB++ {
				a.SubDms[iSDFB].FromBuf(stateDataArrayBuf[iSDFB*8:])
				adManager := &a.SubDms[iSDFB]

				var s5Array []byte // actualy this is int8
				if a.InterpolationSettings.PairedElementsCount != 1 {
					s5Array = make([]byte, 1)
					// must be signed shift
					s5Array[0] = byte(int8(a.FlagsProbably) >> 4)
				} else {
					// must correspond to setted bits count in interpolationSettings.words array probably
					s5Array = interpolationSettingsBuf[a.InterpolationSettings.CountOfWords*2+4:]
				}

				_ = s5Array
				_ = adManager

				// utils.LogDump("OUR TEST", adManager, a.InterpolationSettings)

				unkoff_t8 := a.InterpolationSettings.OffsetToElement
				unkoff_t9 := a.InterpolationSettings.PairedElementsCount

				indx_t7 := 0

				_ = unkoff_t8
				_ = unkoff_t9

				// probably
				// (next frame time - prev frame time) * samples per second * 16384
				s0 := 0

				// s7 or t1 later
				jointDataOffset := a.BaseTargetDataIndex
				// s3 - offset to elements, with step of a.InterpolationSettings.PairedElementsCount
				// (also + a.InterpolationSettings.OffsetToElement )
				// s3elementOffset :=

				// t3, a2 := range ...
				for _, interpolationWord := range a.InterpolationSettings.Words {

					// ((x ^ (x & (-x))) << 16) >> 16
					// this bit magic "eats" lower non-zero bit
					// and must be applyed to int32 because of signed shifts

					for bitmask := interpolationWord; bitmask != 0; {

						// lowest bit index, indexation from zero (if zero, then lowest bit was 1 and so)
						curBitIndex := bits.TrailingZeros16(bitmask)

						_ = curBitIndex

						// update bitmask value
						bitmaski32 := int32(bitmask)
						bitmask = uint16(((bitmaski32 ^ (bitmaski32 & (-bitmaski32))) << 16) >> 16)

						/*
							a1 := int8(s5Array[indx_t7])
							var unkmult int
							if a1 >= 0 {
								unkmult = s0 >> uint(a1)
							} else {
								unkmult = s0 << uint(-a1)
							}

							_ = unkmult
						*/
						_ = s0
						indx_t7 += 1
					}
					jointDataOffset += 0x10
				}
			}
		}

		_ = utils.LogDump
	}

	return a
}
