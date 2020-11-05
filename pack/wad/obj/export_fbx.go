package obj

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/mogaika/god_of_war_browser/utils"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/mogaika/fbx"
	"github.com/mogaika/fbx/builders/bfbx73"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_mdl "github.com/mogaika/god_of_war_browser/pack/wad/mdl"
	file_mesh "github.com/mogaika/god_of_war_browser/pack/wad/mesh"
	"github.com/mogaika/god_of_war_browser/utils/fbxbuilder"
)

/*
	Model.LimbNode.Lcl Translation/Rotation/Scale = idle joint relative to parent
	Deformer.SubDeformer.Cluster.Transform  = bind global space -> local joint space (in our case mesh space == 0.0)
	Deformer.SubDeformer.Cluster.TransformLink = joint global space

	for us .Cluster.Transform = TransformLink.Inverse() because our mesh always at zero transformation
	transform link of parent + lcl transform =  transform link

= = = Anim KeyAttrFlags

1 << 2 |   # interpolation mode, 1 = constant, 2 = linear, 3 = cubic.
1 << 8 |   # tangent mode, 8 = auto, 9 = TCB, 10 = user, 11 = generic break,
1 << 13 |  # tangent mode, 12 = generic clamp, 13 = generic time independent,
1 << 14 |  # tangent mode, 13 + 14 = generic clamp progressive.

Cubic|TangeantUser|WeightedRight|WeightedNextLeft
0x03000408      50332680

Cubic|TangeantUser|WeightedRight
0x01000408      16778248

Cubic|TangeantUser
0x00000408      1032

Cubic|TangeantAuto|GenericTimeIndependent|GenericClampProgressive
0x00006108      24840


= = = Anim KeyAttrDataFloat

RightSlope:2.28906, NextLeftSlope:0, RightWeight:0.333333, NextLeftWeight:0.333333, RightVelocity:0, NextLeftVelocity:0;
1074954210,0,218434821,0,

RightAuto:0, NextLeftAuto:0;
0,0,218434821,0,

RightSlope:0.421768, NextLeftSlope:0.0552456, RightWeight:0.333333, NextLeftWeight:0.333333, RightVelocity:0, NextLeftVelocity:0;
1054339578,1029851450,218434821,0,

RightSlope:0.0552447, NextLeftSlope:0, RightWeight:0.333333, NextLeftWeight:0.333333, RightVelocity:0, NextLeftVelocity:0;
1029851214,0,218434821,0,

RightAuto:0, NextLeftAuto:0;
0,0,218434821,0

*/

type FbxExporterJoint struct {
	FbxLimbNode   *fbx.Node
	FbxLimbNodeId int64

	FbxNodeAttribute   *fbx.Node
	FbxNodeAttributeId int64
}

type FbxExporter struct {
	FbxModelId int64
	Joints     []FbxExporterJoint
}

func (fe *FbxExporter) AddMeshPartNoSkinning(o *Object, part *file_mesh.FbxExportPart, f *fbxbuilder.FBXBuilder) {
	for _, object := range part.Objects {
		f.AddConnections(bfbx73.C("OO", object.FbxModelId, fe.FbxModelId))
	}
}

func (fe *FbxExporter) AddMeshPartWithSkinning(o *Object, part *file_mesh.FbxExportPart, f *fbxbuilder.FBXBuilder) {
	for _, object := range part.Objects {
		skinDeformerId := f.GenerateId()
		skinDeformer := bfbx73.Deformer(skinDeformerId, "\x00\x01Deformer", "Skin")
		skinDeformer.AddNodes(
			bfbx73.Version(101),
			bfbx73.Link_DeformAcuracy(50.0),
			bfbx73.SkinningType("Linear"),
		)

		f.AddConnections(
			bfbx73.C("OO", object.FbxModelId, fe.FbxModelId),
			bfbx73.C("OO", skinDeformerId, object.FbxGeometryId),
		)
		f.AddObjects(skinDeformer)

		for jointID := range object.AffectedByJoints {
			indexes := make([]int32, 0, 64)
			weights := make([]float64, 0, 64)

			for iVertice, jointsForVertice := range object.VerticeToJoint {
				var weight float32
				if jointsForVertice[0] == jointID && jointsForVertice[1] == jointID {
					weight = 1.0
				} else if jointsForVertice[0] == jointID {
					weight = object.VerticeJointWeight[iVertice]
				} else if jointsForVertice[1] == jointID {
					weight = 1.0 - object.VerticeJointWeight[iVertice]
				} else {
					continue
				}

				if weight <= 0.000001 {
					continue
				}
				fmt.Println(jointsForVertice, object.VerticeJointWeight[iVertice], "selected j", jointID, "w", weight)

				indexes = append(indexes, int32(iVertice))
				weights = append(weights, float64(weight))
			}

			if len(indexes) == 0 {
				continue
			}

			transform := utils.FloatArray32to64(o.Joints[jointID].BindToJointMat[:])
			transformLink := utils.FloatArray32to64(o.Joints[jointID].BindWorldJoint[:])

			subDeformerId := f.GenerateId()
			subDeformer := bfbx73.Deformer(subDeformerId, "\x00\x01SubDeformer", "Cluster")

			subDeformer.AddNodes(
				bfbx73.Version(100),
				bfbx73.UserData("", ""),
				bfbx73.Indexes(indexes),
				bfbx73.Weights(weights),
				bfbx73.Transform(transform),
				bfbx73.TransformLink(transformLink),
			)

			f.AddConnections(
				bfbx73.C("OO", fe.Joints[jointID].FbxLimbNodeId, subDeformerId),
				bfbx73.C("OO", subDeformerId, skinDeformerId),
			)
			f.AddObjects(subDeformer)
		}
	}
}

func (o *Object) ExportFbx(wrsrc *wad.WadNodeRsrc, f *fbxbuilder.FBXBuilder) *FbxExporter {
	fe := &FbxExporter{
		FbxModelId: f.GenerateId(),
		Joints:     make([]FbxExporterJoint, len(o.Joints)),
	}
	defer f.AddCache(wrsrc.Tag.Id, fe)

	var position mgl32.Vec4
	var rotation mgl32.Vec3
	var scale = mgl32.Vec4{1.0, 1.0, 1.0, 1.0}

	if len(o.Joints) != 1 {
		bindPose := bfbx73.Pose(f.GenerateId(), "BIND_POSES\x00\x01Pose", "BindPose")
		bindPose.AddNodes(
			bfbx73.Type("BindPose"),
			bfbx73.Version(100),
			bfbx73.NbPoseNodes(int32(len(o.Joints))),
		)

		for iJoint := range o.Joints {
			objJoint := &o.Joints[iJoint]
			eJoint := &fe.Joints[iJoint]

			eJoint.FbxLimbNodeId = f.GenerateId()
			eJoint.FbxNodeAttributeId = f.GenerateId()

			pos := o.Vectors4[objJoint.Id]
			rotation := o.GetEulerLocalRotationForJoint(objJoint.Id)
			scale := o.Vectors6[objJoint.Id]

			eJoint.FbxLimbNode = bfbx73.Model(eJoint.FbxLimbNodeId, objJoint.Name+"\x00\x01Model", "LimbNode").AddNodes(
				bfbx73.Version(232),
				bfbx73.Properties70().AddNodes(
					bfbx73.P("InheritType", "enum", "", "", int32(1)),
					bfbx73.P("Lcl Translation", "Lcl Translation", "", "A",
						float64(pos[0]), float64(pos[1]), float64(pos[2])),
					bfbx73.P("Lcl Rotation", "Lcl Rotation", "", "A",
						float64(rotation[0]), float64(rotation[1]), float64(rotation[2])),
					bfbx73.P("Lcl Scaling", "Lcl Scaling", "", "A",
						float64(scale[0]), float64(scale[1]), float64(scale[2])),
				),
				bfbx73.Shading(false),
				bfbx73.Culling("CullingOff"),
			)

			eJoint.FbxNodeAttribute = bfbx73.NodeAttribute(eJoint.FbxNodeAttributeId, "\x00\x01NodeAttribute", "LimbNode").AddNodes(
				bfbx73.Properties70().AddNodes(
					bfbx73.P("Size", "double", "Number", "", float64(50.0)),
				),
				bfbx73.TypeFlags("Skeleton"),
			)
			f.AddConnections(bfbx73.C("OO", eJoint.FbxNodeAttributeId, eJoint.FbxLimbNodeId))

			bindPose.AddNodes(bfbx73.PoseNode().AddNodes(
				bfbx73.Node(eJoint.FbxLimbNodeId),
				bfbx73.Matrix(utils.FloatArray32to64(objJoint.BindWorldJoint[:])),
			))

			if objJoint.Parent == JOINT_CHILD_NONE {
				f.AddConnections(bfbx73.C("OO", eJoint.FbxLimbNodeId, fe.FbxModelId))
			} else {
				f.AddConnections(bfbx73.C("OO", eJoint.FbxLimbNodeId, fe.Joints[objJoint.Parent].FbxLimbNodeId))
			}

			f.AddObjects(eJoint.FbxLimbNode, eJoint.FbxNodeAttribute)
		}
		f.AddObjects(bindPose)
	} else {
		position = o.Vectors4[0]
		rotation = o.GetEulerLocalRotationForJoint(0)
		scale = o.Vectors6[0]
	}

	model := bfbx73.Model(fe.FbxModelId, wrsrc.Tag.Name+"\x00\x01Model", "Null").AddNodes(
		bfbx73.Version(232),
		bfbx73.Properties70().AddNodes(
			bfbx73.P("InheritType", "enum", "", "", int32(1)),
			bfbx73.P("DefaultAttributeIndex", "int", "Integer", "", int32(0)),
			bfbx73.P("Lcl Translation", "Lcl Translation", "", "A",
				float64(position[0]), float64(position[1]), float64(position[2])),
			bfbx73.P("Lcl Rotation", "Lcl Rotation", "", "A",
				float64(rotation[0]), float64(rotation[1]), float64(rotation[2])),
			bfbx73.P("Lcl Scaling", "Lcl Scaling", "", "A",
				float64(scale[0]), float64(scale[1]), float64(scale[2])),
		),
		bfbx73.Shading(true),
		bfbx73.Culling("CullingOff"),
	)

	//nodeAttribute := bfbx73.NodeAttribute(f.GenerateId(), wrsrc.Tag.Name+"\x00\x01NodeAttribute", "Null").AddNodes(
	nodeAttribute := bfbx73.NodeAttribute(f.GenerateId(), "\x00\x01NodeAttribute", "Null").AddNodes(
		bfbx73.TypeFlags("Null"),
	)

	f.AddConnections(bfbx73.C("OO", nodeAttribute.Properties[0].(int64), fe.FbxModelId))
	f.AddObjects(model, nodeAttribute)

	// find joints created by model (part phase)
	for _, id := range wrsrc.Node.SubGroupNodes {
		n := wrsrc.Wad.GetNodeById(id)
		if inst, _, err := wrsrc.Wad.GetInstanceFromNode(n.Id); err == nil {
			switch inst.(type) {
			case *file_mdl.Model:
				mdl := inst.(*file_mdl.Model)

				exMdl := f.GetCachedOr(n.Tag.Id, func() interface{} {
					return mdl.ExportFbx(wrsrc.Wad.GetNodeResourceByTagId(n.Tag.Id), f)
				}).(*file_mdl.FbxExporter)

				if len(o.Joints) != 1 {
					log.Println("Exporting model with skinning")
					for _, submodel := range exMdl.Models {
						for _, part := range submodel.Parts {
							fe.AddMeshPartWithSkinning(o, part, f)
						}
					}
				} else {
					log.Println("Exporting model no skinning")
					for _, submodel := range exMdl.Models {
						for _, part := range submodel.Parts {
							fe.AddMeshPartNoSkinning(o, part, f)
						}
					}
				}
			}
		}
	}

	return fe
}

func (o *Object) ExportFbxDefault(wrsrc *wad.WadNodeRsrc) *fbxbuilder.FBXBuilder {
	f := fbxbuilder.NewFBXBuilder(filepath.Join(wrsrc.Wad.Name(), wrsrc.Name()))

	fe := o.ExportFbx(wrsrc, f)

	f.AddConnections(bfbx73.C("OO", fe.FbxModelId, 0))

	return f
}
