package obj

import (
	"log"

	"github.com/go-gl/mathgl/mgl32"

	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_anm "github.com/mogaika/god_of_war_browser/pack/wad/anm"
	file_mdl "github.com/mogaika/god_of_war_browser/pack/wad/mdl"
	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/mogaika/god_of_war_browser/utils/gltfutils"
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

func (tfoe *GLTFObjectExported) addAnimation(o *Object, doc *gltf.Document, ganim *file_anm.Animations) error {
	dataTypeIndex := -1
	for dti, dt := range ganim.DataTypes {
		if dt.TypeId == file_anm.DATATYPE_SKINNING {
			dataTypeIndex = dti
			break
		}
	}
	if dataTypeIndex == -1 {
		return nil
	}

	var skinInit file_anm.RenderSkinningInit
	skinInit.Rotation = make([][4]float32, 0, len(o.Joints))
	for _, vec := range o.Vectors5 {
		skinInit.Rotation = append(skinInit.Rotation,
			[4]float32{float32(vec[0]), float32(vec[1]), float32(vec[2]), float32(vec[3])})
	}

	skinInit.Position = make([][4]float32, 0, len(o.Joints))
	for _, vec := range o.Vectors4 {
		skinInit.Position = append(skinInit.Position, vec)
	}

	for iGroup := range ganim.Groups {
		group := &ganim.Groups[iGroup]

		for iAct := range group.Acts {
			act := &group.Acts[iAct]
			descr := &act.StateDescrs[dataTypeIndex]

			data := descr.Data.([]*file_anm.AnimState0Skinning)
			if data == nil || len(data) == 0 {
				continue
			}

			gltfAnim := &gltf.Animation{
				Name:     group.Name + "_" + act.Name,
				Samplers: make([]*gltf.AnimationSampler, 0),
				Channels: make([]*gltf.Channel, 0),
			}
			doc.Animations = append(doc.Animations, gltfAnim)

			rendered := file_anm.RenderSkinningData(
				int(act.Duration/descr.FrameTime), data, skinInit)

			for iJoint, stream := range rendered.Rotation {
				input := make([]float32, 0, len(stream.Index))
				for _, frame := range stream.Index {
					input = append(input, float32(frame)*descr.FrameTime)
				}

				output := make([][4]float32, 0, len(stream.Values)*4)
				for _, v := range stream.Values {
					var q mgl32.Quat

					if o.Joints[iJoint].IsQuaterion {
						q = mgl32.Quat{
							V: mgl32.Vec3{v[0], v[1], v[2]}.Mul(quat_to_float),
							W: v[3] * quat_to_float,
						}
					} else {
						q = utils.EulerToQuat(utils.DegreeToRadiansV3(
							mgl32.Vec3{v[0], v[1], v[2]}.Mul(quat_to_float * 360.0)),
						)
					}

					q = q.Normalize()
					output = append(output, [4]float32{q.V[0], q.V[1], q.V[2], q.W})
				}

				if act.Name == "air360" {
					if !o.Joints[iJoint].IsQuaterion {
						if iJoint == 2 {
							log.Printf("++++ JOINT %d", iJoint)
							for i, v := range stream.Values {
								q := utils.EulerToQuat(utils.DegreeToRadiansV3(
									mgl32.Vec3{v[0], v[1], v[2]}.Mul(quat_to_float * 360.0),
								))
								log.Printf("[%d]: %v %v",
									stream.Index[i],
									stream.Values[i],
									utils.QuatToEuler(q).Mul(1.0/(quat_to_float*360.0)))
							}
						}
					}
					//utils.LogDump(input, output)
				}

				inputAccesor := modeler.WriteAccessor(doc, gltf.TargetNone, input)
				outputAccesor := modeler.WriteAccessor(doc, gltf.TargetNone, output)

				gltfAnim.Samplers = append(gltfAnim.Samplers, &gltf.AnimationSampler{
					Input:         &inputAccesor,
					Output:        &outputAccesor,
					Interpolation: gltf.InterpolationStep,
				})
				gltfAnim.Channels = append(gltfAnim.Channels, &gltf.Channel{
					Sampler: gltf.Index(uint32(len(gltfAnim.Samplers) - 1)),
					Target: gltf.ChannelTarget{
						Node: gltf.Index(tfoe.JointNodes[iJoint]),
						Path: gltf.TRSRotation,
					},
				})
			}

			for iJoint, stream := range rendered.Position {
				input := make([]float32, 0, len(stream.Index))
				for _, frame := range stream.Index {
					input = append(input, float32(frame)*descr.FrameTime)
				}

				inputAccesor := modeler.WriteAccessor(doc, gltf.TargetNone, input)
				outputAccesor := modeler.WriteAccessor(doc, gltf.TargetNone, stream.Values)

				gltfAnim.Samplers = append(gltfAnim.Samplers, &gltf.AnimationSampler{
					Input:         &inputAccesor,
					Output:        &outputAccesor,
					Interpolation: gltf.InterpolationStep,
				})
				gltfAnim.Channels = append(gltfAnim.Channels, &gltf.Channel{
					Sampler: gltf.Index(uint32(len(gltfAnim.Samplers) - 1)),
					Target: gltf.ChannelTarget{
						Node: gltf.Index(tfoe.JointNodes[iJoint]),
						Path: gltf.TRSTranslation,
					},
				})
			}
		}
	}

	return nil
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

			case *file_anm.Animations:
				if err := tfoe.addAnimation(o, doc, inst); err != nil {
					log.Panicf("Failed to export animations: %v", err)
				}
			}
		}
	}

	return tfoe, nil
}
