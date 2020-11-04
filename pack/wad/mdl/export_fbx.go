package mdl

import (
	"fmt"
	"path/filepath"

	"github.com/mogaika/fbx/builders/bfbx73"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_mat "github.com/mogaika/god_of_war_browser/pack/wad/mat"
	file_mesh "github.com/mogaika/god_of_war_browser/pack/wad/mesh"
	"github.com/mogaika/god_of_war_browser/utils/fbxbuilder"
)

type FbxExporter struct {
	Models []*file_mesh.FbxExporter
}

func (m *Model) ExportFbx(wrsrc *wad.WadNodeRsrc, f *fbxbuilder.FBXBuilder) *FbxExporter {
	fe := &FbxExporter{
		Models: make([]*file_mesh.FbxExporter, 0),
	}
	defer f.AddCache(wrsrc.Tag.Id, fe)

	materials := make([]int64, 0)
	modelId := 0

	for _, id := range wrsrc.Node.SubGroupNodes {
		node := wrsrc.Wad.GetNodeById(id)
		nodeResource := wrsrc.Wad.GetNodeResourceByNodeId(node.Id)
		if inst, _, err := wrsrc.Wad.GetInstanceFromNode(node.Id); err == nil {
			switch inst.(type) {
			case *file_mat.Material:
				mat := inst.(*file_mat.Material)
				var fbxMat *file_mat.FbxExporter
				if fbxMatI := f.GetCached(node.Tag.Id); fbxMatI == nil {
					fbxMat = mat.ExportFbx(nodeResource, f)
				} else {
					fbxMat = fbxMatI.(*file_mat.FbxExporter)
				}
				materials = append(materials, fbxMat.MaterialId)
			case *file_mesh.Mesh:
				mesh := inst.(*file_mesh.Mesh)
				var fbxMesh *file_mesh.FbxExporter
				if fbxMeshI := f.GetCached(node.Tag.Id); fbxMeshI == nil {
					fbxMesh = mesh.ExportFbx(nodeResource, f)
				} else {
					fbxMesh = fbxMeshI.(*file_mesh.FbxExporter)
				}
				modelId += 1

				fe.Models = append(fe.Models, fbxMesh)

				for _, part := range fbxMesh.Parts {
					for _, object := range part.Objects {
						object.FbxModel.Properties[1] = fmt.Sprintf("m%d_%s", modelId, object.FbxModel.Properties[1])
						f.AddConnections(bfbx73.C("OO", materials[object.MaterialId], object.FbxModelId))
					}
				}
			}
		}
	}

	return fe
}

func (m *Model) ExportFbxDefault(wrsrc *wad.WadNodeRsrc) *fbxbuilder.FBXBuilder {
	f := fbxbuilder.NewFBXBuilder(filepath.Join(wrsrc.Wad.Name(), wrsrc.Name()))

	fe := m.ExportFbx(wrsrc, f)

	for _, model := range fe.Models {
		for _, part := range model.Parts {
			//f.AddConnections(bfbx73.C("OO", part.FbxModelId, 0))
			for _, object := range part.Objects {
				f.AddConnections(bfbx73.C("OO", object.FbxModelId, 0))
			}
		}
	}

	return f
}
