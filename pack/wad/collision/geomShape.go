package collision

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/mogaika/god_of_war_browser/pack/wad"
)

const MESH_TAG = 112

type GeomShapeVertex struct {
	Pos  [3]float32
	Norm [3]float32
}

type GeomShapeIndex struct {
	Indexes [3]uint16
	Flags   uint16
}

type GeomShape struct {
	Vertexes []GeomShapeVertex
	Indexes  []GeomShapeIndex
}

func NewGeomShapeFromData(r io.Reader) (*GeomShape, error) {
	var buf [16]byte
	if _, err := r.Read(buf[:]); err != nil {
		return nil, err
	}

	gs := &GeomShape{
		Vertexes: make([]GeomShapeVertex, binary.LittleEndian.Uint32(buf[:4])),
		Indexes:  make([]GeomShapeIndex, binary.LittleEndian.Uint32(buf[4:8])),
	}

	for i := range gs.Vertexes {
		var buf [24]byte
		if _, err := r.Read(buf[:]); err != nil {
			return nil, err
		}

		v := &gs.Vertexes[i]
		v.Pos[0] = math.Float32frombits(binary.LittleEndian.Uint32(buf[:4]))
		v.Pos[1] = math.Float32frombits(binary.LittleEndian.Uint32(buf[4:8]))
		v.Pos[2] = math.Float32frombits(binary.LittleEndian.Uint32(buf[8:12]))
		v.Norm[0] = math.Float32frombits(binary.LittleEndian.Uint32(buf[12:16]))
		v.Norm[1] = math.Float32frombits(binary.LittleEndian.Uint32(buf[16:20]))
		v.Norm[2] = math.Float32frombits(binary.LittleEndian.Uint32(buf[20:24]))
	}

	for i := range gs.Indexes {
		var buf [8]byte
		if n, err := r.Read(buf[:]); err != nil && n != 8 {
			return nil, err
		}

		x := &gs.Indexes[i]
		x.Indexes[0] = binary.LittleEndian.Uint16(buf[:2])
		x.Indexes[1] = binary.LittleEndian.Uint16(buf[2:4])
		x.Indexes[2] = binary.LittleEndian.Uint16(buf[4:6])
		x.Flags = binary.LittleEndian.Uint16(buf[6:8])
	}

	return gs, nil
}

func (gs *GeomShape) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	return gs, nil
}

func init() {
	/*wad.SetTagHandler(MESH_TAG, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewGeomShapeFromData(bytes.NewReader(wrsrc.Tag.Data))
	})*/
}
