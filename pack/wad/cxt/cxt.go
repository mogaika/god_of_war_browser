package cxt

import (
	"fmt"

	"github.com/mogaika/god_of_war_browser/config"

	"github.com/mogaika/god_of_war_browser/pack/wad"
)

const CHUNK_MAGIC = 0x80000001
const FILE_SIZE = 0x34

type Chunk struct{}

func NewFromData(buf []byte) (*Chunk, error) {
	return &Chunk{}, nil
}

type Ajax struct {
	Name      string
	Instances []any
}

func (cxt *Chunk) Marshal(wrsrc *wad.WadNodeRsrc) (interface{}, error) {
	ajax := Ajax{Name: wrsrc.Name()}

	for _, iSubNode := range wrsrc.Node.SubGroupNodes {
		instance, _, err := wrsrc.Wad.GetInstanceFromNode(iSubNode)
		if err != nil {
			continue
			// return nil, fmt.Errorf("Error getting instance: %v", err)
		}

		instanceAjax, err := instance.Marshal(wrsrc.Wad.GetNodeResourceByNodeId(iSubNode))
		if err != nil {
			return nil, fmt.Errorf("Error marshaling instance: %v", err)
		}

		ajax.Instances = append(ajax.Instances, instanceAjax)
	}

	return ajax, nil
}

func init() {
	wad.SetHandler(config.GOW1, CHUNK_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data)
	})
	wad.SetHandler(config.GOW2, CHUNK_MAGIC, func(wrsrc *wad.WadNodeRsrc) (wad.File, error) {
		return NewFromData(wrsrc.Tag.Data)
	})
}
