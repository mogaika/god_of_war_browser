package obj

import (
	"fmt"
	"math"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/mogaika/god_of_war_browser/fbx"
	"github.com/mogaika/god_of_war_browser/fbx/cache"
	"github.com/mogaika/god_of_war_browser/pack/wad"

	file_mdl "github.com/mogaika/god_of_war_browser/pack/wad/mdl"
)

type FbxExporter struct {
	FbxModelId uint64
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

func (o *Object) ExportFbx(wrsrc *wad.WadNodeRsrc, f *fbx.FBX, cache *cache.Cache) *FbxExporter {
	fe := &FbxExporter{
		FbxModelId: f.GenerateId(),
	}
	defer cache.Add(wrsrc.Tag.Id, fe)

	model := &fbx.Model{
		Id:      fe.FbxModelId,
		Name:    "Model::" + wrsrc.Tag.Name,
		Element: "Null",
		Version: 232,
		Shading: true,
		Culling: "CullingOff",
	}

	for _, id := range wrsrc.Node.SubGroupNodes {
		n := wrsrc.Wad.GetNodeById(id)
		if inst, _, err := wrsrc.Wad.GetInstanceFromNode(n.Id); err == nil {
			switch inst.(type) {
			case *file_mdl.Model:
				mdl := inst.(*file_mdl.Model)

				var exMdl *file_mdl.FbxExporter
				if exMdlI := cache.Get(n.Tag.Id); exMdlI == nil {
					exMdl = mdl.ExportFbx(wrsrc.Wad.GetNodeResourceByTagId(n.Tag.Id), f, cache)
				} else {
					exMdl = exMdlI.(*file_mdl.FbxExporter)
				}

				for _, submodel := range exMdl.Models {
					for _, part := range submodel.Parts {
						partModel := part.FbxModel

						joint := o.Joints[part.RawPart.JointId]

						q := mgl32.Mat4ToQuat(joint.RenderMat)
						euler := quatToEuler(q)
						pos := joint.RenderMat.Col(3).Vec3()
						scale := joint.RenderMat.Diag().Vec3().Mul(joint.RenderMat.Diag().W())

						partModel.Properties70.P = append(partModel.Properties70.P,
							&fbx.Propertie70{
								Name: "Lcl Translation", Type: "Lcl Translation", Purpose: "", Idk: "A+", Value: pos},
							&fbx.Propertie70{
								Name: "Lcl Rotation", Type: "Lcl Rotation", Purpose: "", Idk: "A+", Value: euler.Mul(180.0 / math.Pi)},
							&fbx.Propertie70{
								Name: "Lcl Scaling", Type: "Lcl Scaling", Purpose: "", Idk: "A+", Value: scale})

						partModel.Name = fmt.Sprintf("Model::%s_part%d", joint.Name, part.Part)
						f.Connections.C = append(f.Connections.C, fbx.Connection{
							Type: "OO", Parent: model.Id, Child: partModel.Id,
						})
					}
				}
			}
		}
	}

	f.Objects.Model = append(f.Objects.Model, model)

	return fe
}

func (o *Object) ExportFbxDefault(wrsrc *wad.WadNodeRsrc) *fbx.FBX {
	f := fbx.NewFbx()
	f.Objects.Model = make([]*fbx.Model, 0)

	fe := o.ExportFbx(wrsrc, f, cache.NewCache())

	f.Connections.C = append(f.Connections.C, fbx.Connection{
		Type: "OO", Parent: 0, Child: fe.FbxModelId,
	})

	f.CountDefinitions()

	return f
}
