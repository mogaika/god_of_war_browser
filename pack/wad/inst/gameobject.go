package inst

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/mogaika/god_of_war_browser/config"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_scr "github.com/mogaika/god_of_war_browser/pack/wad/scr"
	"github.com/mogaika/god_of_war_browser/utils"
)

const INSTANCE_MAGIC = 0x00020001
const FILE_SIZE = 0x5C

type Instance struct {
	Object    string
	Id        uint16
	Params    uint16
	Position1 mgl32.Vec4 // object translation. need transform object by this
	Rotation  mgl32.Vec4 // 3 elements for rotation of object (euler, rads). last element is scale.
	Position2 mgl32.Vec4 // world-relative position (of center of element probably)
	Unk       [3]uint32
}

func NewFromData(buf []byte) (*Instance, error) {
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

type Ajax struct {
	Instance
	Scripts []interface{}
	Name    string
	Object  interface{}
}

func (inst *Instance) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	scripts := make([]interface{}, 0)

	objNode := wrsrc.Wad.GetNodeByName(inst.Object, wrsrc.Node.Id-1, false)
	var obj any
	if objNode != nil {
		parsedObj, _, err := wrsrc.Wad.GetInstanceFromNode(objNode.Id)
		if err != nil {
			return nil, fmt.Errorf("Error when parsing object for game obj info: %v", err)
		}

		obj, err = parsedObj.Marshal(wrsrc.Wad.GetNodeResourceByNodeId(objNode.Id))
		if err != nil {
			return nil, fmt.Errorf("Error when marshaling '%s'", objNode.Tag.Name)
		}
	}

	for _, i := range wrsrc.Node.SubGroupNodes {
		n := wrsrc.Wad.GetNodeById(i)
		name := n.Tag.Name
		sn, _, err := wrsrc.Wad.GetInstanceFromNode(n.Id)
		if err != nil {
			continue
			// return nil, fmt.Errorf("Error when extracting node %d->%s game obj info: %v", i, name, err)
		} else {
			switch sn.(type) {
			case *file_scr.ScriptParams:
				scr, err := sn.Marshal(wrsrc.Wad.GetNodeResourceByNodeId(n.Id))
				if err != nil {
					return nil, fmt.Errorf("Error when getting script info %d-'%s': %v", i, name, err)
				}
				scripts = append(scripts, scr)
			}
		}
	}

	return &Ajax{
		Name:     wrsrc.Name(),
		Instance: *inst,
		Scripts:  scripts,
		Object:   obj,
	}, nil
}

func init() {
	wad.SetServerHandler(config.GOW1, INSTANCE_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data)
	})
}
