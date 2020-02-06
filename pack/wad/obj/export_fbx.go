package obj

import (
	"fmt"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/mogaika/god_of_war_browser/fbx"
	"github.com/mogaika/god_of_war_browser/fbx/cache"
	"github.com/mogaika/god_of_war_browser/pack/wad"

	file_mdl "github.com/mogaika/god_of_war_browser/pack/wad/mdl"
)

type FbxExporter struct {
	FbxModelId uint64
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

						_ = mgl32.Abs
						partModel.Properties70.P = append(partModel.Properties70.P,
							&fbx.Propertie70{
								Name: "Lcl Translation", Type: "Lcl Translation", Purpose: "", Idk: "A+", Value: o.Vectors4[joint.Id]},
							//&fbx.Propertie70{
							//		Name: "Lcl Rotation", Type: "Lcl Translation", Purpose: "", Idk: "A+", Value: o.Vectors5[joint.Id]},
							&fbx.Propertie70{
								Name: "Lcl Scaling", Type: "Lcl Translation", Purpose: "", Idk: "A+", Value: o.Vectors6[joint.Id]})

						partModel.Name += fmt.Sprintf("_JOINT%d", joint.Id)
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
