package anm

import (
	"bytes"
	"encoding/binary"
	"math"

	"github.com/mogaika/god_of_war_browser/utils"
)

type AnimState0Skinning struct {
	// Every anim act hold its own copy of encoded quaterion array for every joint
	// Then they blended together
	PositionDescr           AnimStateDescrHeader
	PositionStream          AnimStateSubstream
	PositionSubStreamsRough []AnimStateSubstream
	PositionSubStreamsAdd   []AnimStateSubstream
	PositionDataBitMap      DataBitMap

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

func (a *AnimState0Skinning) GetShiftsArray(descr *AnimStateDescrHeader, stateData []byte) []int8 {
	dataBitMap := a.GetDataBitMap(descr, stateData)
	shifts := make([]int8, dataBitMap.PairedElementsCount)
	if dataBitMap.PairedElementsCount == 1 {
		shifts[0] = int8(descr.FlagsProbably) >> 4
	} else {
		s5ArrayRaw := a.getDataBitMapOffset(descr, stateData)[len(dataBitMap.Bitmap)*2+4:]
		if err := binary.Read(bytes.NewBuffer(s5ArrayRaw), binary.LittleEndian, shifts); err != nil {
			panic(err)
		}
	}
	return shifts
}

func AnimState0SkinningFromBuf(buf []byte, stateIndex int, _l *utils.Logger) *AnimState0Skinning {
	return &AnimState0Skinning{}
}

func shiftToCoeff(shift int8) float32 {
	if shift == 0 {
		return 1.0
	} else if shift <= 0 {
		coeef := int(1) << uint(-shift)
		return float32(coeef)
	} else {
		coeef := int(1) << uint(shift)
		return 1.0 / float32(coeef)
	}
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

		bitMapOffset := stateData[2+int(stateDataSecondByte)*8:]
		a.RotationDataBitMap = a.GetDataBitMap(&a.RotationDescr, bitMapOffset)

		_l.Printf("   ! DATA: state (fb: %d, subs count) (sb: %d, total subs)", stateDataFirstByte, stateDataSecondByte)
		_l.Printf("   ! INTERPOLATION (default: %v)  %+v", a.RotationDescr.FlagsProbably&2 == 0, a.RotationDataBitMap)

		a.RotationSubStreamsAdd = make([]AnimStateSubstream, stateDataFirstByte)
		for iAddSubDm := 0; iAddSubDm < int(stateDataFirstByte); iAddSubDm++ {
			subStream := &a.RotationSubStreamsAdd[iAddSubDm]
			subStream.Manager.FromBuf(stateDataArrayBuf[iAddSubDm*8:])
			subStream.Samples = make(map[int]interface{})

			shifts := a.GetShiftsArray(&a.RotationDescr, bitMapOffset)
			_l.Printf("      - SDFB %d: %+v,  s5Array: %v", iAddSubDm, subStream.Manager, shifts)
			parseFramesRotationAdd(stateBuf, subStream, &a.RotationDataBitMap, &a.RotationDescr, true, shifts)
		}

		a.RotationSubStreamsRough = make([]AnimStateSubstream, stateDataSecondByte-stateDataFirstByte)
		for iRawSubDm := int(stateDataFirstByte); iRawSubDm < int(stateDataSecondByte); iRawSubDm++ {
			subStream := &a.RotationSubStreamsRough[iRawSubDm-int(stateDataFirstByte)]
			subStream.Manager.FromBuf(stateDataArrayBuf[iRawSubDm*8:])
			subStream.Samples = make(map[int]interface{})

			_l.Printf("      - halfs %d: %+v", iRawSubDm, subStream.Manager)
			parseFramesRotationRaw(stateBuf, subStream, &a.RotationDataBitMap, &a.RotationDescr, true)
		}
	} else {
		a.RotationDataBitMap = a.GetDataBitMap(&a.RotationDescr, stateData)
		a.RotationStream.Samples = make(map[int]interface{})

		_l.Printf("       RAW %+v  flag %v", a.RotationDataBitMap, a.RotationDescr.FlagsProbably&1 != 0)
		if a.RotationDescr.FlagsProbably&1 == 0 {
			_l.Printf("       RAW RAW %+v", a.RotationDataBitMap)
			parseFramesRotationRaw(stateData, &a.RotationStream, &a.RotationDataBitMap, &a.RotationDescr, false)
		} else {
			shifts := a.GetShiftsArray(&a.RotationDescr, stateData)
			_l.Printf("       RAW ADDITIVE %+v shifts: %v", a.RotationDataBitMap, shifts)
			parseFramesRotationAdd(stateData, &a.RotationStream, &a.RotationDataBitMap, &a.RotationDescr, false, shifts)
		}
	}
}

// POOSIIIIIIIITIIIIIIIIIOOOOOOONNNNNNNSSSSS
func (a *AnimState0Skinning) ParsePositions(buf []byte, stateIndex int, _l *utils.Logger, rawAct []byte) {
	stateBuf := rawAct[binary.LittleEndian.Uint32(rawAct[0x80:])+uint32(stateIndex*0xc):]
	a.PositionDescr.FromBuf(stateBuf)
	a.PositionStream.Manager.FromBuf(stateBuf[4:])
	stateData := stateBuf[(uint32(a.PositionDescr.HowMany64kbWeNeedSkip)<<16)+uint32(a.PositionStream.Manager.OffsetToData):]

	_l.Printf(">>>>>>>> POSITION PARSING >>>>>>>>>>>>>>>>>>>>> POSITION PARSING >>>>>>>>>>>>>>>>")
	_l.Printf(">>>>>>>> STATE %d (baseDataIndex: %d, flags 0x%.2x, sm: %+v) >>>>>>>>>>>>>>>",
		stateIndex, a.PositionDescr.BaseTargetDataIndex, a.PositionDescr.FlagsProbably, a.PositionStream.Manager)

	_ = stateData
	if a.PositionStream.Manager.Count == 0 {
		stateDataFirstByte := stateData[0]
		stateDataSecondByte := stateData[1]
		stateDataArrayBuf := stateData[2:]

		bitMapOffset := stateData[2+int(stateDataSecondByte)*8:]
		a.PositionDataBitMap = a.GetDataBitMap(&a.PositionDescr, bitMapOffset)

		a.PositionSubStreamsAdd = make([]AnimStateSubstream, stateDataFirstByte)
		for iAddSubDm := 0; iAddSubDm < int(stateDataFirstByte); iAddSubDm++ {
			subStream := &a.PositionSubStreamsAdd[iAddSubDm]
			subStream.Manager.FromBuf(stateDataArrayBuf[iAddSubDm*8:])
			subStream.Samples = make(map[int]interface{})
			shifts := a.GetShiftsArray(&a.PositionDescr, bitMapOffset)
			parseFramesPositionAdd(stateBuf, subStream, &a.PositionDataBitMap, &a.PositionDescr, true, shifts)
			//utils.LogDump(subStream)
		}

		a.PositionSubStreamsRough = make([]AnimStateSubstream, stateDataSecondByte-stateDataFirstByte)
		for iRawSubDm := int(stateDataFirstByte); iRawSubDm < int(stateDataSecondByte); iRawSubDm++ {
			subStream := &a.PositionSubStreamsRough[iRawSubDm-int(stateDataFirstByte)]
			subStream.Manager.FromBuf(stateDataArrayBuf[iRawSubDm*8:])
			subStream.Samples = make(map[int]interface{})
			parseFramesPositionRaw(stateBuf, subStream, &a.PositionDataBitMap, &a.PositionDescr, true)
			//utils.LogDump(subStream)
		}

	} else {
		a.PositionDataBitMap = a.GetDataBitMap(&a.PositionDescr, stateData)
		a.PositionStream.Samples = make(map[int]interface{})

		_l.Printf("       RAW %+v  flag %v", a.PositionDataBitMap, a.PositionDescr.FlagsProbably&1 != 0)
		if a.PositionDescr.FlagsProbably&1 == 0 {
			parseFramesPositionRaw(stateData, &a.PositionStream, &a.PositionDataBitMap, &a.PositionDescr, false)
		} else {
			shifts := a.GetShiftsArray(&a.PositionDescr, stateData)
			parseFramesPositionAdd(stateData, &a.PositionStream, &a.PositionDataBitMap, &a.PositionDescr, false, shifts)
		}
	}
}

func parseFramesRotationRaw(buf []byte, subStream *AnimStateSubstream, bitMap *DataBitMap,
	descr *AnimStateDescrHeader, useAdditionalOffset bool) {
	additionalOffset := 0
	if useAdditionalOffset {
		additionalOffset = (int(descr.HowMany64kbWeNeedSkip) << 16) + int(subStream.Manager.OffsetToData)
	}
	const elementSize = 2
	bitMap.Iterate(func(bitIndex, iteration int) {
		frames := make([]float32, subStream.Manager.Count)
		frameStep := int(bitMap.PairedElementsCount) * elementSize
		for iFrame := range frames {
			offset := frameStep*iFrame + int(bitMap.DataOffset) + iteration*elementSize + additionalOffset
			frames[iFrame] = float32(int16(binary.LittleEndian.Uint16(buf[offset:])))
		}
		subStream.Samples[int(descr.BaseTargetDataIndex)+bitIndex] = frames
	})
}

func parseFramesRotationAdd(buf []byte, subStream *AnimStateSubstream, bitMap *DataBitMap,
	descr *AnimStateDescrHeader, useAdditionalOffset bool, shifts []int8) {
	additionalOffset := 0
	if useAdditionalOffset {
		additionalOffset = (int(descr.HowMany64kbWeNeedSkip) << 16) + int(subStream.Manager.OffsetToData)
	}
	const elementSize = 1
	bitMap.Iterate(func(bitIndex, iteration int) {
		frames := make([]float32, subStream.Manager.Count)
		frameStep := int(bitMap.PairedElementsCount) * elementSize
		for iFrame := range frames {
			offset := frameStep*iFrame + int(bitMap.DataOffset) + iteration*elementSize + additionalOffset
			frames[iFrame] = float32(int8(buf[offset])) * shiftToCoeff(shifts[iteration])
		}
		subStream.Samples[int(descr.BaseTargetDataIndex)+bitIndex] = frames
	})
	subStream.Samples[-100] = true
}

func parseFramesPositionRaw(buf []byte, subStream *AnimStateSubstream, bitMap *DataBitMap,
	descr *AnimStateDescrHeader, useAdditionalOffset bool) {
	additionalOffset := 0
	if useAdditionalOffset {
		additionalOffset = (int(descr.HowMany64kbWeNeedSkip) << 16) + int(subStream.Manager.OffsetToData)
	}
	const elementSize = 4
	bitMap.Iterate(func(bitIndex, iteration int) {
		frames := make([]float32, subStream.Manager.Count)
		frameStep := int(bitMap.PairedElementsCount) * elementSize
		for iFrame := range frames {
			offset := frameStep*iFrame + int(bitMap.DataOffset) + iteration*elementSize + additionalOffset
			frames[iFrame] = math.Float32frombits(binary.LittleEndian.Uint32(buf[offset:]))
		}
		subStream.Samples[int(descr.BaseTargetDataIndex)+bitIndex] = frames
	})
}

func parseFramesPositionAdd(buf []byte, subStream *AnimStateSubstream, bitMap *DataBitMap,
	descr *AnimStateDescrHeader, useAdditionalOffset bool, shifts []int8) {
	additionalOffset := 0
	if useAdditionalOffset {
		additionalOffset = (int(descr.HowMany64kbWeNeedSkip) << 16) + int(subStream.Manager.OffsetToData)
	}
	const elementSize = 2
	bitMap.Iterate(func(bitIndex, iteration int) {
		frames := make([]float32, subStream.Manager.Count)
		frameStep := int(bitMap.PairedElementsCount) * elementSize
		for iFrame := range frames {
			offset := frameStep*iFrame + int(bitMap.DataOffset) + iteration*elementSize + additionalOffset
			frames[iFrame] = float32(int16(binary.LittleEndian.Uint16(buf[offset:]))) * shiftToCoeff(shifts[iteration]) / 256.0
			//log.Println(shifts[iteration], shiftToCoeff(shifts[iteration]), frames[iFrame], int16(binary.LittleEndian.Uint16(buf[offset:])))
		}
		subStream.Samples[int(descr.BaseTargetDataIndex)+bitIndex] = frames
	})
	subStream.Samples[-100] = true
}
