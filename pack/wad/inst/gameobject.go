package inst

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

const INSTANCE_MAGIC = 0x00020001
const FILE_SIZE = 0x5C

type Instance struct {
	Object    string
	Id        uint16
	Params    uint16
	Position1 mgl32.Vec4 // object translation. need transform object by this
	Rotation  mgl32.Vec4 // rotation of object (euler, rads)
	Position2 mgl32.Vec4 // world-relative position for visibility check mby>???
	Unk       [3]uint32
}

func NewFromData(fmat io.ReaderAt) (*Instance, error) {
	buf := make([]byte, FILE_SIZE)
	if _, err := fmat.ReadAt(buf, 0); err != nil {
		return nil, err
	}

	inst := &Instance{
		Object: utils.BytesToString(buf[0x4:0x1c]),
		Id:     binary.LittleEndian.Uint16(buf[0x1c:0x1e]),
		Params: binary.LittleEndian.Uint16(buf[0x1e:0x20]),
		Unk: [3]uint32{
			binary.LittleEndian.Uint32(buf[0x50:0x54]),
			binary.LittleEndian.Uint32(buf[0x54:0x58]),
			binary.LittleEndian.Uint32(buf[0x58:0x5C]),
		},
	}

	binary.Read(bytes.NewReader(buf[0x20:0x30]), binary.LittleEndian, &inst.Position1)
	binary.Read(bytes.NewReader(buf[0x30:0x40]), binary.LittleEndian, &inst.Rotation)
	binary.Read(bytes.NewReader(buf[0x40:0x50]), binary.LittleEndian, &inst.Position2)

	return inst, nil
}

func (inst *Instance) Marshal(wad *wad.Wad, node *wad.WadNode) (interface{}, error) {
	return inst, nil
}

func init() {
	wad.SetHandler(INSTANCE_MAGIC, func(w *wad.Wad, node *wad.WadNode, r *io.SectionReader) (wad.File, error) {
		return NewFromData(r)
	})
}
