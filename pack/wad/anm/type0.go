package anm

import (
	"bytes"
	"encoding/binary"
	"log"
	"math"

	"github.com/mogaika/god_of_war_browser/utils"
)

type AnimStateSkinningTracks struct {
	Position    []*AnimStateSkinningAttributeTrack
	Rotation    []*AnimStateSkinningAttributeTrack
	Scale       []*AnimStateSkinningAttributeTrack
	Attachments []*AnimStateSkinningAttributeTrack
}

type AnimStateSkinningAttributeTrack struct {
	// Every anim clip hold its own copy of encoded quaterion array for every joint
	// Then they blended together
	Descr           AnimStateDescrHeader
	Stream          AnimStateSubstream
	SubStreamsRough []AnimStateSubstream
	SubStreamsAdd   []AnimStateSubstream
	DataBitMap      DataBitMap
}

func bitmaskZeroBitsShift(bitmask uint16) uint16 {
	bitmaski32 := int32(bitmask)
	return uint16(((bitmaski32 ^ (bitmaski32 & (-bitmaski32))) << 16) >> 16)
}

func (a *AnimStateSkinningAttributeTrack) getDataBitMapOffset(descr *AnimStateDescrHeader, stateData []byte) []byte {
	dataBitMap := defaultDataBitMap
	if descr.FlagsProbably&2 != 0 {
		dataBitMap = stateData
	}
	return dataBitMap
}

func (a *AnimStateSkinningAttributeTrack) GetDataBitMap(descr *AnimStateDescrHeader, stateData []byte) DataBitMap {
	return NewDataBitMapFromBuf(a.getDataBitMapOffset(descr, stateData))
}

func (a *AnimStateSkinningAttributeTrack) GetShiftsArray(descr *AnimStateDescrHeader, stateData []byte) []int8 {
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

func ParseSkinningAttributeTrackRotation(stateBuf []byte, _l *utils.Logger) *AnimStateSkinningAttributeTrack {
	track := &AnimStateSkinningAttributeTrack{}
	track.Descr.FromBuf(stateBuf)
	track.Stream.Manager.FromBuf(stateBuf[4:])

	stateData := stateBuf[(uint32(track.Descr.HowMany64kbWeNeedSkip)<<16)+uint32(track.Stream.Manager.OffsetToData):]
	_l.Printf(">>>>>>>> STATE (baseDataIndex: %d, flags 0x%.2x, sm: %+v) >>>>>>>>>>>>>>>",
		track.Descr.BaseTargetDataIndex, track.Descr.FlagsProbably, track.Stream.Manager)

	// Parse rotations
	if track.Stream.Manager.Count == 0 {
		stateDataFirstByte := stateData[0]
		stateDataSecondByte := stateData[1]
		stateDataArrayBuf := stateData[2:]

		bitMapOffset := stateData[2+int(stateDataSecondByte)*8:]
		track.DataBitMap = track.GetDataBitMap(&track.Descr, bitMapOffset)

		_l.Printf("   ! DATA: state (fb: %d, subs count) (sb: %d, total subs)", stateDataFirstByte, stateDataSecondByte)
		_l.Printf("   ! INTERPOLATION (default: %v)  %+v", track.Descr.FlagsProbably&2 == 0, track.DataBitMap)

		track.SubStreamsAdd = make([]AnimStateSubstream, stateDataFirstByte)
		for iAddSubDm := 0; iAddSubDm < int(stateDataFirstByte); iAddSubDm++ {
			subStream := &track.SubStreamsAdd[iAddSubDm]
			subStream.Manager.FromBuf(stateDataArrayBuf[iAddSubDm*8:])
			subStream.Samples = make(map[int]interface{})

			shifts := track.GetShiftsArray(&track.Descr, bitMapOffset)
			_l.Printf("      - SDFB %d: %+v,  s5Array: %v", iAddSubDm, subStream.Manager, shifts)
			parseFramesRotationAdd(stateBuf, subStream, &track.DataBitMap, &track.Descr, true, shifts)
		}

		track.SubStreamsRough = make([]AnimStateSubstream, stateDataSecondByte-stateDataFirstByte)
		for iRawSubDm := int(stateDataFirstByte); iRawSubDm < int(stateDataSecondByte); iRawSubDm++ {
			subStream := &track.SubStreamsRough[iRawSubDm-int(stateDataFirstByte)]
			subStream.Manager.FromBuf(stateDataArrayBuf[iRawSubDm*8:])
			subStream.Samples = make(map[int]interface{})

			_l.Printf("      - halfs %d: %+v", iRawSubDm, subStream.Manager)
			parseFramesRotationRaw(stateBuf, subStream, &track.DataBitMap, &track.Descr, true)
		}
	} else {
		track.DataBitMap = track.GetDataBitMap(&track.Descr, stateData)
		track.Stream.Samples = make(map[int]interface{})

		_l.Printf("       RAW %+v  flag %v", track.DataBitMap, track.Descr.FlagsProbably&1 != 0)
		if track.Descr.FlagsProbably&1 == 0 {
			_l.Printf("       RAW RAW %+v", track.DataBitMap)
			parseFramesRotationRaw(stateData, &track.Stream, &track.DataBitMap, &track.Descr, false)
		} else {
			shifts := track.GetShiftsArray(&track.Descr, stateData)
			_l.Printf("       RAW ADDITIVE %+v shifts: %v", track.DataBitMap, shifts)
			parseFramesRotationAdd(stateData, &track.Stream, &track.DataBitMap, &track.Descr, false, shifts)
		}
	}

	return track
}

func ParseSkinningAttributeTrackPosition(stateBuf []byte, _l *utils.Logger) *AnimStateSkinningAttributeTrack {
	track := &AnimStateSkinningAttributeTrack{}
	track.Descr.FromBuf(stateBuf)
	track.Stream.Manager.FromBuf(stateBuf[4:])
	stateData := stateBuf[(uint32(track.Descr.HowMany64kbWeNeedSkip)<<16)+uint32(track.Stream.Manager.OffsetToData):]

	_l.Printf(">>>>>>>> POSITION PARSING >>>>>>>>>>>>>>>>>>>>> POSITION PARSING >>>>>>>>>>>>>>>>")
	_l.Printf(">>>>>>>> STATE (baseDataIndex: %d, flags 0x%.2x, sm: %+v) >>>>>>>>>>>>>>>",
		track.Descr.BaseTargetDataIndex, track.Descr.FlagsProbably, track.Stream.Manager)

	_ = stateData
	if track.Stream.Manager.Count == 0 {
		stateDataFirstByte := stateData[0]
		stateDataSecondByte := stateData[1]
		stateDataArrayBuf := stateData[2:]

		bitMapOffset := stateData[2+int(stateDataSecondByte)*8:]
		track.DataBitMap = track.GetDataBitMap(&track.Descr, bitMapOffset)

		track.SubStreamsAdd = make([]AnimStateSubstream, stateDataFirstByte)
		for iAddSubDm := 0; iAddSubDm < int(stateDataFirstByte); iAddSubDm++ {
			subStream := &track.SubStreamsAdd[iAddSubDm]
			subStream.Manager.FromBuf(stateDataArrayBuf[iAddSubDm*8:])
			subStream.Samples = make(map[int]interface{})
			shifts := track.GetShiftsArray(&track.Descr, bitMapOffset)
			parseFramesPositionAdd(stateBuf, subStream, &track.DataBitMap, &track.Descr, true, shifts)
			//utils.LogDump(subStream)
		}

		track.SubStreamsRough = make([]AnimStateSubstream, stateDataSecondByte-stateDataFirstByte)
		for iRawSubDm := int(stateDataFirstByte); iRawSubDm < int(stateDataSecondByte); iRawSubDm++ {
			subStream := &track.SubStreamsRough[iRawSubDm-int(stateDataFirstByte)]
			subStream.Manager.FromBuf(stateDataArrayBuf[iRawSubDm*8:])
			subStream.Samples = make(map[int]interface{})
			parseFramesPositionRaw(stateBuf, subStream, &track.DataBitMap, &track.Descr, true)
			//utils.LogDump(subStream)
		}

	} else {
		track.DataBitMap = track.GetDataBitMap(&track.Descr, stateData)
		track.Stream.Samples = make(map[int]interface{})

		_l.Printf("       RAW %+v  flag %v", track.DataBitMap, track.Descr.FlagsProbably&1 != 0)
		if track.Descr.FlagsProbably&1 == 0 {
			parseFramesPositionRaw(stateData, &track.Stream, &track.DataBitMap, &track.Descr, false)
		} else {
			shifts := track.GetShiftsArray(&track.Descr, stateData)
			parseFramesPositionAdd(stateData, &track.Stream, &track.DataBitMap, &track.Descr, false, shifts)
		}
	}

	return track
}

func ParseSkinningAttributeTrackScale(stateBuf []byte, _l *utils.Logger) *AnimStateSkinningAttributeTrack {
	track := &AnimStateSkinningAttributeTrack{}
	track.Descr.FromBuf(stateBuf)
	track.Stream.Manager.FromBuf(stateBuf[4:])
	stateData := stateBuf[(uint32(track.Descr.HowMany64kbWeNeedSkip)<<16)+uint32(track.Stream.Manager.OffsetToData):]

	_l.Printf(">>>>>>>> SCALE PARSING >>>>>>>>>>>>>>>>>>>>> SCALE PARSING >>>>>>>>>>>>>>>>")
	_l.Printf(">>>>>>>> STATE (baseDataIndex: %d, flags 0x%.2x, sm: %+v) >>>>>>>>>>>>>>>",
		track.Descr.BaseTargetDataIndex, track.Descr.FlagsProbably, track.Stream.Manager)

	_ = stateData
	if track.Stream.Manager.Count == 0 {
		// green
		log.Printf("Got stream consisting of multiple entries")

		stateDataFirstByte := stateData[0]
		stateDataSecondByte := stateData[1]
		stateDataArrayBuf := stateData[2:]

		if stateDataFirstByte != 0 {
			panic(stateDataFirstByte)
		}

		bitMapOffset := stateData[2+int(stateDataSecondByte)*8:]
		track.DataBitMap = track.GetDataBitMap(&track.Descr, bitMapOffset)
		/*
			a.ScaleSubStreamsAdd = make([]AnimStateSubstream, stateDataFirstByte)
			for iAddSubDm := 0; iAddSubDm < int(stateDataFirstByte); iAddSubDm++ {
				subStream := &a.ScaleSubStreamsAdd[iAddSubDm]
				subStream.Manager.FromBuf(stateDataArrayBuf[iAddSubDm*8:])
				subStream.Samples = make(map[int]interface{})
				shifts := a.GetShiftsArray(&a.ScaleDescr, bitMapOffset)
				parseFramesScaleAdd(stateBuf, subStream, &a.ScaleDataBitMap, &a.ScaleDescr, true, shifts)
				//utils.LogDump(subStream)
			}*/

		track.SubStreamsRough = make([]AnimStateSubstream, stateDataSecondByte-stateDataFirstByte)
		for iRawSubDm := int(stateDataFirstByte); iRawSubDm < int(stateDataSecondByte); iRawSubDm++ {
			subStream := &track.SubStreamsRough[iRawSubDm-int(stateDataFirstByte)]
			subStream.Manager.FromBuf(stateDataArrayBuf[iRawSubDm*8:])
			subStream.Samples = make(map[int]interface{})
			parseFramesScaleRaw(stateBuf, subStream, &track.DataBitMap, &track.Descr, true)
			//utils.LogDump(subStream)
		}

	} else {
		// blue
		track.DataBitMap = track.GetDataBitMap(&track.Descr, stateData)
		track.Stream.Samples = make(map[int]interface{})

		_l.Printf("       RAW %+v  flag %v", track.DataBitMap, track.Descr.FlagsProbably&1 != 0)

		//if a.ScaleDescr.FlagsProbably&1 == 0 {
		parseFramesScaleRaw(stateData, &track.Stream, &track.DataBitMap, &track.Descr, false)
		/*} else {
			shifts := a.GetShiftsArray(&a.ScaleDescr, stateData)
			parseFramesScaleAdd(stateData, &a.ScaleStream, &a.ScaleDataBitMap, &a.ScaleDescr, false, shifts)
		}*/
	}

	return track
}

func ParseSkinningAttributeTrackAttachments(stateBuf []byte, _l *utils.Logger) *AnimStateSkinningAttributeTrack {
	track := &AnimStateSkinningAttributeTrack{}
	track.Descr.FromBuf(stateBuf)
	track.Stream.Manager.FromBuf(stateBuf[4:])

	_l.Printf("---- trackAttachments: %s", utils.SDump(track))

	return track
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

var scaleRawMultiplicator = math.Float32frombits(0x39C90FD8)

func parseFramesScaleRaw(buf []byte, subStream *AnimStateSubstream, bitMap *DataBitMap,
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
			frames[iFrame] = float32(uint32(binary.LittleEndian.Uint16(buf[offset:]))) * scaleRawMultiplicator
		}
		subStream.Samples[int(descr.BaseTargetDataIndex)+bitIndex] = frames
	})
}
