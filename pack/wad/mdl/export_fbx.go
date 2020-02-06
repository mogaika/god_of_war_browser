package mdl

import (
	"github.com/mogaika/god_of_war_browser/fbx"
	"github.com/mogaika/god_of_war_browser/fbx/cache"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_mat "github.com/mogaika/god_of_war_browser/pack/wad/mat"
	file_mesh "github.com/mogaika/god_of_war_browser/pack/wad/mesh"
)

type FbxExporter struct {
	ModelId uint64
}

func (m *Model) ExportFbx(wrsrc *wad.WadNodeRsrc, f *fbx.FBX, cache *cache.Cache) *FbxExporter {
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

	materials := make([]uint64, 0)

	for _, id := range wrsrc.Node.SubGroupNodes {
		node := wrsrc.Wad.GetNodeById(id)
		nodeResource := wrsrc.Wad.GetNodeResourceByNodeId(node.Id)
		if inst, _, err := wrsrc.Wad.GetInstanceFromNode(node.Id); err == nil {
			switch inst.(type) {
			case *file_mat.Material:
				mat := inst.(*file_mat.Material)
				var fbxMat *file_mat.FbxExporter
				if fbxMatI := cache.Get(node.Tag.Id); fbxMatI == nil {
					fbxMat = mat.ExportFbx(nodeResource, f, cache)
				} else {
					fbxMat = fbxMatI.(*file_mat.FbxExporter)
				}
				materials = append(materials, fbxMat.MaterialId)
			case *file_mesh.Mesh:
				mesh := inst.(*file_mesh.Mesh)
				var fbxMesh *file_mesh.FbxExporter
				if fbxMeshI := cache.Get(node.Tag.Id); fbxMeshI == nil {
					fbxMesh = mesh.ExportFbx(nodeResource, f, cache)
				} else {
					fbxMesh = fbxMeshI.(*file_mesh.FbxExporter)
				}

				for _, part := range fbxMesh.Parts {
					f.Connections.C = append(f.Connections.C, fbx.Connection{
						Type: "OO", Child: part.FbxModelId, Parent: model.Id,
					})

					for _, object := range part.Objects {
						f.Connections.C = append(f.Connections.C, fbx.Connection{
							Type: "OO", Child: materials[object.MaterialId], Parent: object.FbxModelId,
						})
					}
				}
			}
		}
	}

	f.Objects.Model = append(f.Objects.Model, model)

	return fe
}

func (m *Model) ExportFbxDefault(wrsrc *wad.WadNodeRsrc) *fbx.FBX {
	f := fbx.NewFbx()
	f.Objects.Model = make([]*fbx.Model, 0)

	fe := m.ExportFbx(wrsrc, f, cache.NewCache())
	f.Connections.C = append(f.Connections.C, fbx.Connection{
		Type: "OO", Parent: 0, Child: fe.ModelId,
	})

	f.CountDefinitions()

	return f
}
