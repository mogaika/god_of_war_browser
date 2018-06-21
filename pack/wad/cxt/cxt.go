package cxt

import (
	"fmt"

	"github.com/mogaika/god_of_war_browser/config"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/pack/wad/inst"
	"github.com/mogaika/god_of_war_browser/pack/wad/obj"
)

const CHUNK_MAGIC = 0x80000001
const FILE_SIZE = 0x34

type Chunk struct{}

func NewFromData(buf []byte) (*Chunk, error) {
	return &Chunk{}, nil
}

type Ajax struct {
	Instances []*inst.Instance
	Objects   map[string]interface{}
}

func (cxt *Chunk) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	ajax := Ajax{}

	for _, iSubNode := range wrsrc.Node.SubGroupNodes {
		instance, _, err := wrsrc.Wad.GetInstanceFromNode(iSubNode)
		if err != nil {
			continue
			//return nil, fmt.Errorf("Error getting instance: %v", err)
		}
		ajax.Instances = append(ajax.Instances, instance.(*inst.Instance))
	}

	ajax.Objects = make(map[string]interface{}, int(len(ajax.Instances)))

	for _, instance := range ajax.Instances {
		if _, ok := ajax.Objects[instance.Object]; !ok {
			object := wrsrc.Wad.GetNodeByName(instance.Object, wrsrc.Node.Id-1, false)
			if object == nil {
				continue
				//return nil, fmt.Errorf("Cannot find object '%s'", instance.Object)
			}
			objectData, _, err := wrsrc.Wad.GetInstanceFromNode(object.Id)
			if err != nil {
				return nil, fmt.Errorf("Cannot parse object '%s'", instance.Object)
			}
			ajax.Objects[object.Tag.Name], err = objectData.(*obj.Object).Marshal(wrsrc.Wad.GetNodeResourceByNodeId(object.Id))
			if err != nil {
				return nil, fmt.Errorf("Error when marshaling '%s'", object.Tag.Name)
			}
		}
	}

	return ajax, nil
}

func init() {
	wad.SetHandler(config.GOW1ps2, CHUNK_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data)
	})
}
