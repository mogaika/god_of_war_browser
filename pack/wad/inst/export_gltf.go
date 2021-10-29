package inst

import (
	"log"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_obj "github.com/mogaika/god_of_war_browser/pack/wad/obj"
	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/mogaika/god_of_war_browser/utils/gltfutils"
	"github.com/pkg/errors"
	"github.com/qmuntal/gltf"
)

type GLTFInstanceExported struct {
	Node uint32
}

func (i *Instance) ExportGLTF(wrsrc *wad.WadNodeRsrc, gltfCacher *gltfutils.GLTFCacher) (*GLTFInstanceExported, error) {
	doc := gltfCacher.Doc
	tfoe := &GLTFInstanceExported{}
	defer gltfCacher.AddCache(wrsrc.Tag.Id, tfoe)

	scale := i.Rotation[3]
	q := utils.EulerToQuat(i.Rotation.Vec3().Mul((180.0 / math.Pi)))
	node := &gltf.Node{
		Name:        wrsrc.Name(),
		Translation: i.Position1.Vec3(),
		Rotation:    [4]float32{q.V[0], q.V[1], q.V[2], q.W},
		Scale:       mgl32.Vec3{scale, scale, scale},
	}

	tfoe.Node = uint32(len(doc.Nodes))
	doc.Nodes = append(doc.Nodes, node)

	objectNode := wrsrc.Wad.GetNodeByName(i.Object, wrsrc.Node.Id, false)
	if objectNode == nil {
		log.Printf("Failed to find node %q", i.Object)
	} else {
		objectI, _, err := wrsrc.Wad.GetInstanceFromNode(objectNode.Id)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to get isntance %q", i.Object)
		}
		subobject, err := objectI.(*file_obj.Object).ExportGLTF(wrsrc.Wad.GetNodeResourceByNodeId(objectNode.Id), gltfCacher)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to export object")
		}
		node.Children = append(node.Children, subobject.JointNodes[0])
	}

	return tfoe, nil
}
