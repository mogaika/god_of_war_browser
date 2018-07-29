package anm

import (
	"encoding/binary"
	"math/bits"

	"github.com/mogaika/god_of_war_browser/utils"
)

type AnimState0Skinning struct {
	// Every anim act hold its own copy of encoded quaterion array for every joint
	// Then they blended together
	PositionDescr      AnimStateDescrHeader
	PositionStream     AnimStateSubstream
	PositionDataBitMap DataBitMap

	RotationDescr           AnimStateDescrHeader
	RotationStream          AnimStateSubstream
	RotationSubStreamsRough []AnimStateSubstream
	RotationSubStreamsAdd   []AnimStateSubstream
	RotationDataBitMap      DataBitMap
}

func bitmaskZeroBitsShift(bitmask uint16) uint16 {
	bitmaski32 := int32(bitmask)
	return uint16(((bitmaski32 ^ (bitmaski32 & (-bitmaski32))) << 16) >> 16)
}

func (a *AnimState0Skinning) getDataBitMapOffset(descr *AnimStateDescrHeader, stateData []byte) []byte {
	dataBitMap := defaultDataBitMap
	if descr.FlagsProbably&2 != 0 {
		dataBitMap = stateData
	}
	return dataBitMap
}

func (a *AnimState0Skinning) GetDataBitMap(descr *AnimStateDescrHeader, stateData []byte) DataBitMap {
	return NewDataBitMapFromBuf(a.getDataBitMapOffset(descr, stateData))
}

func (a *AnimState0Skinning) GetShiftsArray(descr *AnimStateDescrHeader, dataBitMap *DataBitMap, stateData []byte) []int8 {
	shifts := make([]int8, dataBitMap.PairedElementsCount)
	if dataBitMap.PairedElementsCount == 1 {
		shifts[0] = int8(descr.FlagsProbably) >> 4
	} else {
		s5ArrayRaw := a.getDataBitMapOffset(descr, stateData)[len(dataBitMap.Bitmap)*2+4:]
		for i := range shifts {
			shifts[i] = int8(s5ArrayRaw[i])
		}
	}
	return shifts
}

func AnimState0SkinningFromBuf(buf []byte, stateIndex int, _l *utils.Logger) *AnimState0Skinning {
	return &AnimState0Skinning{}
}

func (a *AnimState0Skinning) ParseRotations(buf []byte, stateIndex int, _l *utils.Logger) {
	stateBuf := buf[stateIndex*0xc:]

	a.RotationDescr.FromBuf(stateBuf)
	a.RotationStream.Manager.FromBuf(stateBuf[4:])

	stateData := stateBuf[(uint32(a.RotationDescr.HowMany64kbWeNeedSkip)<<16)+uint32(a.RotationStream.Manager.OffsetToData):]
	_l.Printf(">>>>>>>> STATE %d (baseDataIndex: %d, flags 0x%.2x, sm: %+v) >>>>>>>>>>>>>>>",
		stateIndex, a.RotationDescr.BaseTargetDataIndex, a.RotationDescr.FlagsProbably, a.RotationStream.Manager)

	// Parse rotations
	if a.RotationStream.Manager.Count == 0 {
		stateDataFirstByte := stateData[0]
		stateDataSecondByte := stateData[1]
		stateDataArrayBuf := stateData[2:]

		a.RotationDataBitMap = a.GetDataBitMap(&a.RotationDescr, stateData[2+int(stateDataSecondByte)*8:])

		_l.Printf("   ! DATA: state (fb: %d, subs count) (sb: %d, total subs)", stateDataFirstByte, stateDataSecondByte)
		_l.Printf("   ! INTERPOLATION (default: %v)  %+v", a.RotationDescr.FlagsProbably&2 == 0, a.RotationDataBitMap)

		a.RotationSubStreamsAdd = make([]AnimStateSubstream, stateDataFirstByte)
		for iRotationSubDm := 0; iRotationSubDm < int(stateDataFirstByte); iRotationSubDm++ {
			subStream := &a.RotationSubStreamsAdd[iRotationSubDm]
			subSm := &subStream.Manager
			subSm.FromBuf(stateDataArrayBuf[iRotationSubDm*8:])
			subStream.Samples = make(map[int]interface{})

			shifts := a.GetShiftsArray(&a.RotationDescr, &a.RotationDataBitMap, stateData)
			_l.Printf("      - SDFB %d: %+v,  s5Array: %v", iRotationSubDm, *subSm, shifts)
			indx_t7 := 0
			shiftAmountMap := make(map[int]int32)

			// t3, a2 := range ...
			for iBitMapWord, bitMapWord := range a.RotationDataBitMap.Bitmap {

				// ((x ^ (x & (-x))) << 16) >> 16
				// this bit magic "eats" lower non-zero bit
				// and must be applyed to int32 because of signed shifts
				for bitmask := bitMapWord; bitmask != 0; {
					// take lowest bit index, indexation from zero (if zero, then lowest bit was 1 and so)
					curBitIndex := bits.TrailingZeros16(bitmask)
					bitmask = bitmaskZeroBitsShift(bitmask)

					shiftAmount := shifts[indx_t7]

					frames := make([]int8, subSm.Count)
					frameStep := int(a.RotationDataBitMap.PairedElementsCount) * 1 // sizeof byte

					for iFrame := range frames {
						index := frameStep*iFrame + int(a.RotationDataBitMap.DataOffset) + int(subSm.OffsetToData) + indx_t7
						frames[iFrame] = int8(stateBuf[index+(int(a.RotationDescr.HowMany64kbWeNeedSkip)<<16)])
					}

					dataIndex := int(a.RotationDescr.BaseTargetDataIndex) + iBitMapWord*16 + curBitIndex
					subStream.Samples[dataIndex] = frames
					shiftAmountMap[dataIndex] = int32(shiftAmount)
					// TODO: http://127.0.0.1:8000/#/R_PERM.WAD/962 minigamecircle broken
					_l.Printf("          time shift %d frames (additive) for index %d (%d): %v", shiftAmount, dataIndex, curBitIndex, frames)

					indx_t7++
				}
			}
			subStream.Samples[-100] = shiftAmountMap
		}

		a.RotationSubStreamsRough = make([]AnimStateSubstream, stateDataSecondByte-stateDataFirstByte)
		for iOtherSubDm := int(stateDataFirstByte); iOtherSubDm < int(stateDataSecondByte); iOtherSubDm++ {
			subStream := &a.RotationSubStreamsRough[iOtherSubDm-int(stateDataFirstByte)]
			subSm := &subStream.Manager
			subSm.FromBuf(stateDataArrayBuf[iOtherSubDm*8:])
			subStream.Samples = make(map[int]interface{})

			_l.Printf("      - halfs %d: %+v", iOtherSubDm, *subSm)
			indx_t7 := 0

			for iBitMapWord, bitMapWord := range a.RotationDataBitMap.Bitmap {
				for bitmask := bitMapWord; bitmask != 0; {
					curBitIndex := bits.TrailingZeros16(bitmask)
					bitmask = bitmaskZeroBitsShift(bitmask)

					frames := make([]int16, subSm.Count)
					frameStep := int(a.RotationDataBitMap.PairedElementsCount) * 2 // 16 bit

					for iFrame := range frames {
						offset := frameStep*iFrame + int(a.RotationDataBitMap.DataOffset) + int(subSm.OffsetToData) + indx_t7*2

						frames[iFrame] = int16(binary.LittleEndian.Uint16(stateBuf[offset+(int(a.RotationDescr.HowMany64kbWeNeedSkip)<<16):]))
					}

					dataIndex := int(a.RotationDescr.BaseTargetDataIndex) + iBitMapWord*16 + curBitIndex
					subStream.Samples[dataIndex] = frames

					_l.Printf("              frames (halfs) for index %d (%d): %v", dataIndex, curBitIndex, frames)
					indx_t7++
				}
			}
		}

		_ = utils.LogDump
	} else if a.RotationStream.Manager.Count != 0 {
		a.RotationDataBitMap = a.GetDataBitMap(&a.RotationDescr, stateData)

		a.RotationStream.Samples = make(map[int]interface{})

		_l.Printf("       RAW %+v  flag %v", a.RotationDataBitMap, a.RotationDescr.FlagsProbably&1 != 0)
		if a.RotationDescr.FlagsProbably&1 == 0 {
			animDataArrayIndex := 0
			for iBitMapWord, bitMapWord := range a.RotationDataBitMap.Bitmap {
				for bitmask := bitMapWord; bitmask != 0; {
					curBitIndex := bits.TrailingZeros16(bitmask)
					bitmask = bitmaskZeroBitsShift(bitmask)

					frames := make([]int32, a.RotationStream.Manager.Count)
					frameStep := int(a.RotationDataBitMap.PairedElementsCount) * 2 // 16 bit

					for iFrame := range frames {
						// TODO: deside about animDataArrayIndex*2
						offset := frameStep*iFrame + int(a.RotationDataBitMap.DataOffset) + animDataArrayIndex*2
						frames[iFrame] = int32(int16(binary.LittleEndian.Uint16(stateData[offset:])))
					}

					dataIndex := int(a.RotationDescr.BaseTargetDataIndex) + iBitMapWord*16 + curBitIndex
					a.RotationStream.Samples[dataIndex] = frames
					animDataArrayIndex++
					_l.Printf("        frames (RAW) for index %d (%d): %v", dataIndex, curBitIndex, frames)
				}
			}
		} else {
			_l.Printf("       RAW ADDITIVE %+v  flag %v", a.RotationDataBitMap, a.RotationDescr.FlagsProbably&1 != 0)
			//panic("not implemented")
		}
	}
}

func (a *AnimState0Skinning) ParsePositions(buf []byte, stateIndex int, _l *utils.Logger, rawAct []byte) {
	stateBuf := rawAct[binary.LittleEndian.Uint32(rawAct[0x80:])+uint32(stateIndex*0xc):]
	a.PositionDescr.FromBuf(stateBuf)
	a.PositionStream.Manager.FromBuf(stateBuf[4:])
	stateData := stateBuf[(uint32(a.PositionDescr.HowMany64kbWeNeedSkip)<<16)+uint32(a.PositionStream.Manager.OffsetToData):]

	_ = stateData
	/*
		a.PositionInterpolation = a.GetInterpolationSettings(&a.PositionDescr, stateData)
		_l.Printf("       POSITION Manager %+v", a.PositionStream.Manager)
		_l.Printf("       POSITION Interpolation %+v", a.PositionInterpolation)
		if a.PositionStream.Manager.Count != 0 {
			if a.PositionDescr.FlagsProbably&1 != 0 {
				_l.Printf("       POSITION ADDITIVE %+v ", a.PositionDescr)
				// TODO:
			} else {
				_l.Printf("       POSITION RAW %+v", a.PositionDescr)
				a.PositionStream.Samples = make(map[int]interface{})
				animDataArrayIndex := 0
				for iInterpolationWord, interpolationWord := range a.PositionInterpolation.Words {
					for bitmask := interpolationWord; bitmask != 0; {
						curBitIndex := bits.TrailingZeros16(bitmask)
						bitmask = bitmaskZeroBitsShift(bitmask)

						frames := make([]float32, a.PositionStream.Manager.Count)
						frameStep := int(a.PositionInterpolation.PairedElementsCount) * 4 // float32

						for iFrame := range frames {
							offset := frameStep*iFrame + int(a.PositionInterpolation.OffsetToElement) //+ animDataArrayIndex*2
							frames[iFrame] = math.Float32frombits(binary.LittleEndian.Uint32(stateData[offset:]))
						}

						dataIndex := int(a.PositionDescr.BaseTargetDataIndex) + iInterpolationWord*16 + curBitIndex
						a.PositionStream.Samples[dataIndex] = frames
						animDataArrayIndex++
						_l.Printf("        frames (RAW) for index %d (%d): %v", dataIndex, curBitIndex, frames)
					}
				}
			}
		}
	*/
}
