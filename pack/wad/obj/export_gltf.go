package obj

import (
	"log"

	"github.com/mogaika/god_of_war_browser/utils/gltfutils"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_mdl "github.com/mogaika/god_of_war_browser/pack/wad/mdl"
)

type GLTFObjectExported struct {
	JointNodes []uint32
	SkinId     uint32
}

func (tfoe *GLTFObjectExported) addModel(doc *gltf.Document, gmdle *file_mdl.GLTFModelExported) {
	for _, mesh := range gmdle.Meshes {
		for _, object := range mesh.Objects {
			doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, uint32(len(doc.Nodes)))
			doc.Nodes = append(doc.Nodes, &gltf.Node{
				Mesh: gltf.Index(object.GLTFMeshIndex),
				Skin: gltf.Index(tfoe.SkinId),
			})
		}
	}
}

func (o *Object) ExportGLTF(wrsrc *wad.WadNodeRsrc, gltfCacher *gltfutils.GLTFCacher) (*GLTFObjectExported, error) {
	doc := gltfCacher.Doc
	tfoe := &GLTFObjectExported{
		JointNodes: make([]uint32, len(o.Joints)),
	}
	defer gltfCacher.AddCache(wrsrc.Tag.Id, tfoe)

	matrices := make([][4][4]float32, len(o.Joints))
	for jointId := range o.Joints {
		joint := &o.Joints[jointId]

		rotation := o.GetQuaterionLocalRotationForJoint(jointId)

		node := &gltf.Node{
			Name:        joint.Name,
			Translation: o.Vectors4[jointId].Vec3(),
			Rotation:    rotation.V.Vec4(rotation.W),
			Scale:       o.Vectors6[jointId].Vec3(),
		}

		matrices[jointId] = [4][4]float32{
			joint.BindToJointMat.Row(0),
			joint.BindToJointMat.Row(1),
			joint.BindToJointMat.Row(2),
			joint.BindToJointMat.Row(3),
		}

		tfoe.JointNodes[jointId] = uint32(len(doc.Nodes))

		if joint.Parent != JOINT_CHILD_NONE {
			parentNode := doc.Nodes[tfoe.JointNodes[joint.Parent]]
			parentNode.Children = append(parentNode.Children, uint32(len(doc.Nodes)))
		}

		doc.Nodes = append(doc.Nodes, node)
	}

	gltfSkin := &gltf.Skin{
		Name:                wrsrc.Name(),
		InverseBindMatrices: gltf.Index(modeler.WriteAccessor(doc, gltf.TargetNone, matrices)),
		Joints:              tfoe.JointNodes,
	}
	tfoe.SkinId = uint32(len(doc.Skins))
	doc.Skins = append(doc.Skins, gltfSkin)

	doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, tfoe.JointNodes[0])

	// find joints created by model (part phase)
	for _, id := range wrsrc.Node.SubGroupNodes {
		node := wrsrc.Wad.GetNodeById(id)
		if instI, _, err := wrsrc.Wad.GetInstanceFromNode(node.Id); err == nil {
			switch inst := instI.(type) {
			case *file_mdl.Model:
				gmdle := gltfCacher.GetCachedOr(node.Tag.Id, func() interface{} {
					gmdle, err := inst.ExportGLTF(wrsrc.Wad.GetNodeResourceByTagId(node.Tag.Id), gltfCacher)
					if err != nil {
						log.Panicf("Error exporting model %q for object %q: %v", node.Tag.Name, wrsrc.Name(), err)
					}
					return gmdle
				}).(*file_mdl.GLTFModelExported)

				tfoe.addModel(doc, gmdle)
			}
		}
	}

	return tfoe, nil
}
