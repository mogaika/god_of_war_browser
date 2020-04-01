package obj

import (
	"fmt"
	"math"
	"path/filepath"

	"github.com/mogaika/fbx/builders/bfbx73"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils/fbxbuilder"

	file_mdl "github.com/mogaika/god_of_war_browser/pack/wad/mdl"
)

type FbxExporter struct {
	FbxModelId int64
}

func quatToEuler(q mgl32.Quat) (e mgl32.Vec3) {
	sinr_cosp := float64(2 * (q.W*q.X() + q.Y()*q.Z()))
	cosr_cosp := float64(1 - 2*(q.X()*q.X()+q.Y()*q.Y()))

	e[0] = float32(math.Atan2(sinr_cosp, cosr_cosp))

	sinp := float64(2 * (q.W*q.Y() - q.Z()*q.X()))
	if math.Abs(sinp) >= 1 {
		e[1] = math.Pi / 2
		if sinp < 0 {
			e[1] *= -1
		}
	} else {
		e[1] = float32(math.Asin(sinp))
	}

	siny_cosp := float64(2 * (q.W*q.Z() + q.X()*q.Y()))
	cosy_cosp := float64(1 - 2*(q.Y()*q.Y()+q.Z()*q.Z()))
	e[2] = float32(math.Atan2(siny_cosp, cosy_cosp))

	return e
}

func (o *Object) ExportFbx(wrsrc *wad.WadNodeRsrc, f *fbxbuilder.FBXBuilder) *FbxExporter {
	fe := &FbxExporter{
		FbxModelId: f.GenerateId(),
	}
	defer f.AddCache(wrsrc.Tag.Id, fe)

	model := bfbx73.Model(fe.FbxModelId, wrsrc.Tag.Name+"\x00\x01Model", "Null").AddNodes(
		bfbx73.Version(232),
		bfbx73.Shading(true),
		bfbx73.Culling("CullingOff"),
	)

	nodeAttribute := bfbx73.NodeAttribute(f.GenerateId(), wrsrc.Tag.Name+"\x00\x01NodeAttribute", "Null").AddNodes(
		bfbx73.TypeFlags("Null"),
	)

	f.AddConnections(bfbx73.C("OO", nodeAttribute.Properties[0].(int64), fe.FbxModelId))

	f.AddObjects(model, nodeAttribute)

	for _, id := range wrsrc.Node.SubGroupNodes {
		n := wrsrc.Wad.GetNodeById(id)
		if inst, _, err := wrsrc.Wad.GetInstanceFromNode(n.Id); err == nil {
			switch inst.(type) {
			case *file_mdl.Model:
				mdl := inst.(*file_mdl.Model)

				var exMdl *file_mdl.FbxExporter
				if exMdlI := f.GetCached(n.Tag.Id); exMdlI == nil {
					exMdl = mdl.ExportFbx(wrsrc.Wad.GetNodeResourceByTagId(n.Tag.Id), f)
				} else {
					exMdl = exMdlI.(*file_mdl.FbxExporter)
				}

				for _, submodel := range exMdl.Models {
					for _, part := range submodel.Parts {
						partModel := part.FbxModel

						joint := o.Joints[part.RawPart.JointId]

						q := mgl32.Mat4ToQuat(joint.RenderMat)
						pos := joint.RenderMat.Col(3).Vec3()
						rotation := quatToEuler(q).Mul(180.0 / math.Pi)
						scale := joint.RenderMat.Diag().Vec3().Mul(joint.RenderMat.Diag().W())

						// TODO: change float32 to float64
						partModel.GetOrAddNode(bfbx73.Properties70()).AddNodes(
							bfbx73.P("Lcl Translation", "Lcl Translation", "", "A+",
								float64(pos[0]), float64(pos[1]), float64(pos[2])),
							bfbx73.P("Lcl Rotation", "Lcl Rotation", "", "A+",
								float64(rotation[0]), float64(rotation[1]), float64(rotation[2])),
							bfbx73.P("Lcl Scaling", "Lcl Scaling", "", "A+",
								float64(scale[0]), float64(scale[1]), float64(scale[2])),
						)
						partModel.Properties[1] = fmt.Sprintf("%s_part%d\x00\x01Model", joint.Name, part.Part)

						f.AddConnections(
							bfbx73.C("OO", partModel.Properties[0].(int64), model.Properties[0].(int64)),
						)
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
