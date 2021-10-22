package cam

import (
	"bytes"
	"encoding/binary"

	"github.com/mogaika/god_of_war_browser/pack/wad"

	"github.com/go-gl/mathgl/mgl32"
)

type Rail struct {
	Matrices []mgl32.Mat4
	Floats   []float32
}

func (r *Rail) FromData(data []byte) error {
	count := binary.LittleEndian.Uint32(data[0:])
	r.Matrices = make([]mgl32.Mat4, count)
	r.Floats = make([]float32, count)

	unk04 := binary.LittleEndian.Uint32(data[4:])
	unk08 := binary.LittleEndian.Uint32(data[8:])
	unk0c := binary.LittleEndian.Uint32(data[0xc:])
	if unk04 != 0 || unk08 != 0xffff_ffff || unk0c != 0xffff_ffff {
		panic([]uint32{unk04, unk08, unk0c})
	}

	if err := binary.Read(bytes.NewReader(data[8:8+count*0x40]), binary.LittleEndian, r.Matrices); err != nil {
		return err
	}
	if err := binary.Read(bytes.NewReader(data[8+count*0x40:]), binary.LittleEndian, r.Floats); err != nil {
		return err
	}
	return nil
}

func (r *Rail) Marshal(rsrc *wad.WadNodeRsrc) (interface{}, error) {
	return r, nil
}

/*
func init() {
	wad.SetTagHandler(112, func(rsrc *wad.WadNodeRsrc) (wad.File, error) {
		r := &Rail{}
		err := r.FromData(rsrc.Tag.Data)
		return r, err
	})
}
*/
