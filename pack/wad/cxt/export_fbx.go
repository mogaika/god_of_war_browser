package cxt

import (
	"log"
	"math"

	"github.com/mogaika/god_of_war_browser/fbx"
	"github.com/mogaika/god_of_war_browser/fbx/cache"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_inst "github.com/mogaika/god_of_war_browser/pack/wad/inst"
	file_obj "github.com/mogaika/god_of_war_browser/pack/wad/obj"
)

type FbxExporter struct {
	FbxModelId uint64

	Instances []*fbx.Model
}

func (cxt *Chunk) ExportFbx(wrsrc *wad.WadNodeRsrc, f *fbx.FBX, cache *cache.Cache) *FbxExporter {
	fe := &FbxExporter{
		Instances:  make([]*fbx.Model, 0),
		FbxModelId: f.GenerateId(),
	}
	defer cache.Add(wrsrc.Tag.Id, fe)

	cxtModel := &fbx.Model{
		Id:      fe.FbxModelId,
		Name:    "Model::CXT_" + wrsrc.Tag.Name,
		Element: "Null",
		Version: 232,
		Shading: true,
		Culling: "CullingOff",
	}
	f.Objects.Model = append(f.Objects.Model, cxtModel)

	for _, iSubNode := range wrsrc.Node.SubGroupNodes {
		instance, _, err := wrsrc.Wad.GetInstanceFromNode(iSubNode)
		if err != nil {
			// log.Panicf("Failed to load instance %s: %v", wrsrc.Wad.Nodes[iSubNode].Tag.Name, err)
			continue
		}

		gameInstance := instance.(*file_inst.Instance)

		instModel := &fbx.Model{
			Id:      f.GenerateId(),
			Name:    "Model::" + wrsrc.Wad.Nodes[iSubNode].Tag.Name,
			Element: "Null",
			Version: 232,
			Shading: true,
			Culling: "CullingOff",
		}
		f.Objects.Model = append(f.Objects.Model, instModel)

		instModel.Properties70.P = append(instModel.Properties70.P,
			&fbx.Propertie70{
				Name: "Lcl Translation", Type: "Lcl Translation", Purpose: "", Idk: "A+", Value: gameInstance.Position1.Vec3()},
			&fbx.Propertie70{
				Name: "Lcl Rotation", Type: "Lcl Rotation", Purpose: "", Idk: "A+", Value: gameInstance.Rotation.Vec3().Mul(180.0 / math.Pi)},
		)

		// try to find object now
		objectNode := wrsrc.Wad.GetNodeByName(gameInstance.Object, wrsrc.Node.Id, false)
		if objectNode == nil {
			log.Printf("Wasn't able to find node '%s'", gameInstance.Object)
			continue
		}
		objectI, _, err := wrsrc.Wad.GetInstanceFromNode(objectNode.Id)
		if err != nil {
			log.Panicf("can't get instance %s: %v", objectNode.Tag.Name, err)
		}
		object := objectI.(*file_obj.Object)

		var fbxObject *file_obj.FbxExporter
		if fbxObjectI := cache.Get(objectNode.Tag.Id); fbxObjectI == nil {
			fbxObject = object.ExportFbx(wrsrc.Wad.GetNodeResourceByNodeId(objectNode.Id), f, cache)
		} else {
			fbxObject = fbxObjectI.(*file_obj.FbxExporter)
		}

		f.Connections.C = append(f.Connections.C, fbx.Connection{
			Type: "OO", Parent: instModel.Id, Child: fbxObject.FbxModelId,
		})
		f.Connections.C = append(f.Connections.C, fbx.Connection{
			Type: "OO", Parent: cxtModel.Id, Child: instModel.Id,
		})
	}

	return fe
}

func (cxt *Chunk) ExportFbxDefault(wrsrc *wad.WadNodeRsrc) *fbx.FBX {
	f := fbx.NewFbx()

	fe := cxt.ExportFbx(wrsrc, f, cache.NewCache())

	f.Connections.C = append(f.Connections.C, fbx.Connection{
		Type: "OO", Parent: 0, Child: fe.FbxModelId,
	})

	f.CountDefinitions()

	return f
}
