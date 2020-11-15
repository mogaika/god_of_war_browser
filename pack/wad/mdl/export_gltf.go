package mdl

import (
	"log"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_mat "github.com/mogaika/god_of_war_browser/pack/wad/mat"
	file_mesh "github.com/mogaika/god_of_war_browser/pack/wad/mesh"
	"github.com/mogaika/god_of_war_browser/utils/gltfutils"
	"github.com/qmuntal/gltf"
)

type GLTFModelExported struct {
	Meshes    []*file_mesh.GLTFMeshExported
	Materials []*file_mat.GLTFMaterialExported
}

func (m *Model) ExportGLTF(wrsrc *wad.WadNodeRsrc, gltfCacher *gltfutils.GLTFCacher) (*GLTFModelExported, error) {
	gmdle := &GLTFModelExported{
		Meshes:    make([]*file_mesh.GLTFMeshExported, 0),
		Materials: make([]*file_mat.GLTFMaterialExported, 0),
	}
	defer gltfCacher.AddCache(wrsrc.Tag.Id, gmdle)

	for _, id := range wrsrc.Node.SubGroupNodes {
		node := wrsrc.Wad.GetNodeById(id)
		nodeResource := wrsrc.Wad.GetNodeResourceByNodeId(node.Id)

		instI, _, err := wrsrc.Wad.GetInstanceFromNode(node.Id)
		if err == nil {
			switch inst := instI.(type) {
			case *file_mat.Material:
				gme := gltfCacher.GetCachedOr(node.Tag.Id, func() interface{} {
					gme, err := inst.ExportGLTF(nodeResource, gltfCacher)
					if err != nil {
						log.Panicf("Error exporting material %q for model %q: %v", node.Tag.Name, wrsrc.Name(), err)
					}
					return gme
				}).(*file_mat.GLTFMaterialExported)

				gmdle.Materials = append(gmdle.Materials, gme)
			case *file_mesh.Mesh:
				gmeshe := gltfCacher.GetCachedOr(node.Tag.Id, func() interface{} {
					gmeshe, err := inst.ExportGLTF(nodeResource, gltfCacher)
					if err != nil {
						log.Panicf("Error exporting mesh %q for model %q: %v", node.Tag.Name, wrsrc.Name(), err)
					}
					return gmeshe
				}).(*file_mesh.GLTFMeshExported)

				for _, object := range gmeshe.Objects {
					object.GLTFMesh.Primitives[0].Material = gltf.Index(gmdle.Materials[object.MaterialId].MaterialId)
				}

				gmdle.Meshes = append(gmdle.Meshes, gmeshe)
			}
		}
	}

	return gmdle, nil
}

func (m *Model) ExportGLTFDefault(wrsrc *wad.WadNodeRsrc) (*gltf.Document, error) {
	gltfCacher := gltfutils.NewCacher()
	doc := gltfCacher.Doc

	me, err := m.ExportGLTF(wrsrc, gltfCacher)
	if err != nil {
		return nil, err
	}

	for _, mesh := range me.Meshes {
		for _, object := range mesh.Objects {
			doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, uint32(len(doc.Nodes)))
			doc.Nodes = append(doc.Nodes, &gltf.Node{
				Name: object.GLTFMesh.Name,
				Mesh: gltf.Index(object.GLTFMeshIndex),
			})
		}
	}

	return doc, nil
}
