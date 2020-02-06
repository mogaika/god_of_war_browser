package obj

import (
	"github.com/mogaika/god_of_war_browser/fbx"
	"github.com/mogaika/god_of_war_browser/fbx/cache"
	"github.com/mogaika/god_of_war_browser/pack/wad"

	file_mdl "github.com/mogaika/god_of_war_browser/pack/wad/mdl"
)

type FbxExporter struct {
	ModelId uint64
}

func (o *Object) ExportFbx(wrsrc *wad.WadNodeRsrc, f *fbx.FBX, cache *cache.Cache) *FbxExporter {
	fe := &FbxExporter{
		ModelId: f.GenerateId(),
	}
	defer cache.Add(wrsrc.Tag.Id, fe)

	model := &fbx.Model{
		Id:      fe.ModelId,
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
				f.Connections.C = append(f.Connections.C, fbx.Connection{
					Type: "OO", Parent: model.Id, Child: exMdl.ModelId,
				})
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
		Type: "OO", Parent: 0, Child: fe.ModelId,
	})

	f.CountDefinitions()

	return f
}
