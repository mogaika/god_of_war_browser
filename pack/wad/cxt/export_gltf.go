package cxt

import (
	"github.com/pkg/errors"
	"github.com/qmuntal/gltf"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_inst "github.com/mogaika/god_of_war_browser/pack/wad/inst"
	"github.com/mogaika/god_of_war_browser/utils/gltfutils"
)

type GLTFChunkExported struct {
	Node uint32
}

func (cxt *Chunk) ExportGLTF(wrsrc *wad.WadNodeRsrc, gltfCacher *gltfutils.GLTFCacher) (*GLTFChunkExported, error) {
	doc := gltfCacher.Doc
	tfce := &GLTFChunkExported{}
	defer gltfCacher.AddCache(wrsrc.Tag.Id, tfce)

	node := &gltf.Node{
		Name: wrsrc.Name(),
	}

	tfce.Node = uint32(len(doc.Nodes))
	doc.Nodes = append(doc.Nodes, node)

	for _, iSubNode := range wrsrc.Node.SubGroupNodes {
		subNode := wrsrc.Wad.Nodes[iSubNode]

		if len(subNode.Tag.Data) != 0 {
			continue
		}

		instI, _, err := wrsrc.Wad.GetInstanceFromNode(iSubNode)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to get instance %q", subNode.Tag.Name)
		}

		subinst, err := instI.(*file_inst.Instance).ExportGLTF(wrsrc.Wad.GetNodeResourceByNodeId(iSubNode), gltfCacher)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to export instance")
		}

		node.Children = append(node.Children, subinst.Node)
	}

	doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, tfce.Node)

	return tfce, nil
}

func (cxt *Chunk) ExportGLTFDefault(wrsrc *wad.WadNodeRsrc) (*gltf.Document, error) {
	gltfCacher := gltfutils.NewCacher()
	doc := gltfCacher.Doc

	_, err := cxt.ExportGLTF(wrsrc, gltfCacher)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
