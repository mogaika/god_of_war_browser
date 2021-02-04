package inst

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/mogaika/god_of_war_browser/config"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_obj "github.com/mogaika/god_of_war_browser/pack/wad/obj"
)

const INSTANCE_MAGIC_GOW2 = 0x00030001
const FILE_SIZE_GOW2 = 0x68

type InstanceGow2 struct {
	Id        uint16
	Params    uint16
	Position1 mgl32.Vec4
}

func NewGow2FromData(buf []byte) (*InstanceGow2, error) {
	inst := &InstanceGow2{
		Id:     binary.LittleEndian.Uint16(buf[0x1c:0x1e]),
		Params: binary.LittleEndian.Uint16(buf[0x1e:0x20]),
	}

	binary.Read(bytes.NewReader(buf[0x50:0x60]), binary.LittleEndian, &inst.Position1)

	return inst, nil
}

type AjaxGow2 struct {
	InstanceGow2
	Name   string
	Object interface{}
}

func (inst *InstanceGow2) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	oNId := wrsrc.Node.SubGroupNodes[0]
	var object interface{}

	o, _, err := wrsrc.Wad.GetInstanceFromNode(oNId)
	if err != nil {
		return nil, fmt.Errorf("Error when extracting node %d game obj info: %v", oNId, err)
	} else {
		switch o.(type) {
		case *file_obj.Object:
			object, err = o.(*file_obj.Object).Marshal(wrsrc.Wad.GetNodeResourceByNodeId(oNId))
			if err != nil {
				return nil, fmt.Errorf("Error when getting mdl: %v", err)
			}
		}
	}

	return &AjaxGow2{
		Name:         wrsrc.Name(),
		InstanceGow2: *inst,
		Object:       object,
	}, nil
}

func init() {
	wad.SetHandler(config.GOW2, INSTANCE_MAGIC_GOW2, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewGow2FromData(wrsrc.Tag.Data)
	})
}
