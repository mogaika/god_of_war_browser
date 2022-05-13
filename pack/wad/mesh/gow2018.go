package mesh

import (
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils"
)

const GOW2018_MODEL_GEOMETRY_TAG_MAGIC = 0x0001000c
const GOW2018_MODEL_GEOMETRY_GPU_DATA_TAG = 29

type GOW2018Mesh struct {
}

func (m *GOW2018Mesh) Marshal(rsrc *wad.WadNodeRsrc) (interface{}, error) {
	return nil, nil
}

func NewGOW2018Mesh(bs *utils.BufStack) (*GOW2018Mesh, error) {
	return nil, nil
}
