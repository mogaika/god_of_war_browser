package obj

import (
	"github.com/qmuntal/gltf"
	//	"github.com/qmuntal/gltf/modeler"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	//	"github.com/mogaika/god_of_war_browser/pack/wad/mdl"
)

type GLTFObjectExported struct {
	JointNodes []uint32
}

func (o *Object) ExportGLTF(wrsrc *wad.WadNodeRsrc, doc *gltf.Document) (*GLTFObjectExported, error) {
	tfoe := &GLTFObjectExported{
		JointNodes: make([]uint32, len(o.Joints)),
	}

	for jointId := range o.Joints {
		joint := &o.Joints[jointId]

		rotation := o.GetQuaterionLocalRotationForJoint(jointId)

		node := &gltf.Node{
			Name:        joint.Name,
			Translation: o.Vectors4[jointId].Vec3(),
			Rotation:    rotation.V.Vec4(rotation.W),
			Scale:       o.Vectors6[jointId].Vec3(),
		}

		tfoe.JointNodes[jointId] = uint32(len(doc.Nodes))
		doc.Nodes = append(doc.Nodes, node)
	}

	//	for jointId := range o.Joints {

	//}

	return tfoe, nil
}
