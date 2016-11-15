package cxt

import (
	"fmt"
	"io"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/pack/wad/inst"
	"github.com/mogaika/god_of_war_browser/pack/wad/obj"
)

const CHUNK_MAGIC = 0x80000001
const FILE_SIZE = 0x34

type Chunk struct{}

func NewFromData(r io.ReaderAt) (*Chunk, error) {
	return &Chunk{}, nil
}

type Ajax struct {
	Instances []*inst.Instance
	Objects   map[string]interface{}
}

func (cxt *Chunk) Marshal(wad *wad.Wad, node *wad.WadNode) (interface{}, error) {
	ajax := Ajax{}

	for _, iSubNode := range node.SubNodes {
		subnode := wad.Node(iSubNode)
		if subnode.IsLink {
			subnode = subnode.ResolveLink()
			instance, err := wad.Get(subnode.Id)
			if err != nil {
				return nil, fmt.Errorf("Error getting instance: %v", err)
			}
			ajax.Instances = append(ajax.Instances, instance.(*inst.Instance))
		}
	}

	ajax.Objects = make(map[string]interface{}, int(len(ajax.Instances)))

	for _, instance := range ajax.Instances {
		if _, ok := ajax.Objects[instance.Object]; !ok {
			object := wad.FindNode(instance.Object, -1, node.Id)
			if object == nil {
				continue
				//return nil, fmt.Errorf("Cannot find object '%s'", instance.Object)
			}
			objectData, err := wad.Get(object.Id)
			if err != nil {
				return nil, fmt.Errorf("Cannot parse object '%s'", instance.Object)
			}
			ajax.Objects[object.Name], err = objectData.(*obj.Object).Marshal(object.Wad, object)
			if err != nil {
				return nil, fmt.Errorf("Error when marshaling '%s'", object.Name)
			}
		}
	}

	return ajax, nil
}

func init() {
	wad.SetHandler(CHUNK_MAGIC, func(w *wad.Wad, node *wad.WadNode, r io.ReaderAt) (wad.File, error) {
		return NewFromData(r)
	})
}
