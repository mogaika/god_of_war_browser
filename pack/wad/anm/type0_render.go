package anm

/*

type RenderSkinningInit struct {
	Rotation [][4]float32
	Position [][4]float32
}

type RenderedSkinningStream struct {
	Values [][4]float32
	Index  []int
}

type RenderedSkinningState struct {
	Rotation map[int]*RenderedSkinningStream
	Position map[int]*RenderedSkinningStream
}

func (rss *RenderedSkinningState) dataStoreGetStream(dataStore map[int]*RenderedSkinningStream, iStream int) *RenderedSkinningStream {
	if s, ok := dataStore[iStream]; ok {
		return s
	} else {
		s := &RenderedSkinningStream{
			Values: make([][4]float32, 0, 1),
			Index:  make([]int, 0, 1),
		}
		dataStore[iStream] = s
		return s
	}
}

func (rss *RenderedSkinningState) renderStream(
	frame int, stream *AnimStateSubstream, dataStore map[int]*RenderedSkinningStream, init [][4]float32) {

	sampleIndex := frame - int(stream.Manager.Offset)
	if sampleIndex < 0 || sampleIndex >= int(stream.Manager.Count+stream.Manager.DatasCount3) {
		return
	}

	if sampleIndex >= int(stream.Manager.Count) {
		sampleIndex = int(stream.Manager.Count) - 1
	}

	if _, isAdd := stream.Samples[-100]; isAdd {
		for iStream := range stream.Samples {
			if iStream < 0 {
				continue
			}

			samples := stream.Samples[iStream].([]float32)
			jointId, coordId := iStream/4, iStream%4

			renderStream := rss.dataStoreGetStream(dataStore, jointId)
			renderStream.Values[renderStream.getValuesIndexForFrame(frame, init[jointId])][coordId] += samples[sampleIndex]
		}
	} else {
		for iStream := range stream.Samples {
			samples := stream.Samples[iStream].([]float32)
			jointId, coordId := iStream/4, iStream%4

			renderStream := rss.dataStoreGetStream(dataStore, jointId)
			renderStream.Values[renderStream.getValuesIndexForFrame(frame, init[jointId])][coordId] = samples[sampleIndex]
		}
	}
}

func (rs *RenderedSkinningStream) getValuesIndexForFrame(frame int, init [4]float32) int {
	if len(rs.Values) == 0 {
		rs.Values = append(rs.Values, init)
		rs.Index = append(rs.Index, frame)
	} else {
		lastFrame := rs.Index[len(rs.Index)-1]

		if lastFrame > frame {
			panic("shoud not happen")
		} else if lastFrame < frame {
			rs.Values = append(rs.Values, rs.Values[len(rs.Values)-1])
			rs.Index = append(rs.Index, frame)
		}
	}

	return len(rs.Index) - 1
}

func RenderSkinningData(frames int, states []*AnimStateSkinningAttributeTrack, init RenderSkinningInit) *RenderedSkinningState {
	rss := &RenderedSkinningState{
		Rotation: make(map[int]*RenderedSkinningStream),
		Position: make(map[int]*RenderedSkinningStream),
	}

	for frame := 0; frame < frames; frame++ {
		for _, state := range states {
			if state.RotationStream.Manager.Count != 0 {
				rss.renderStream(frame, &state.RotationStream, rss.Rotation, init.Rotation)
			} else {
				for i := range state.RotationSubStreamsAdd {
					rss.renderStream(frame, &state.RotationSubStreamsAdd[i], rss.Rotation, init.Rotation)
				}
				for i := range state.RotationSubStreamsRough {
					rss.renderStream(frame, &state.RotationSubStreamsRough[i], rss.Rotation, init.Rotation)
				}
			}

			if state.PositionStream.Manager.Count != 0 {
				rss.renderStream(frame, &state.PositionStream, rss.Position, init.Position)
			} else {
				for i := range state.PositionSubStreamsAdd {
					rss.renderStream(frame, &state.PositionSubStreamsAdd[i], rss.Position, init.Position)
				}
				for i := range state.PositionSubStreamsRough {
					rss.renderStream(frame, &state.PositionSubStreamsRough[i], rss.Position, init.Position)
				}
			}
		}
	}

	return rss
}

*/
