package cxt

import (
	"log"
	"math"
	"path/filepath"

	"github.com/mogaika/fbx/builders/bfbx73"

	"github.com/mogaika/fbx"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_inst "github.com/mogaika/god_of_war_browser/pack/wad/inst"
	file_obj "github.com/mogaika/god_of_war_browser/pack/wad/obj"
	"github.com/mogaika/god_of_war_browser/utils/fbxbuilder"
)

type FbxExporter struct {
	FbxModelId int64

	Instances []*fbx.Node
}

func (cxt *Chunk) ExportFbx(wrsrc *wad.WadNodeRsrc, f *fbxbuilder.FBXBuilder) *FbxExporter {
	fe := &FbxExporter{
		Instances:  make([]*fbx.Node, 0),
		FbxModelId: f.GenerateId(),
	}
	defer f.AddCache(wrsrc.Tag.Id, fe)

	cxtModel := bfbx73.Model(fe.FbxModelId, wrsrc.Tag.Name+"\x00\x01Model", "Null").AddNodes(
		bfbx73.Version(232),
		bfbx73.Properties70(),
		bfbx73.Shading(true),
		bfbx73.Culling("CullingOff"),
	)

	nodeAttribute := bfbx73.NodeAttribute(f.GenerateId(), wrsrc.Tag.Name+"\x00\x01NodeAttribute", "Null").AddNodes(
		bfbx73.TypeFlags("Null"),
	)
	f.AddConnections(bfbx73.C("OO", nodeAttribute.Properties[0].(int64), fe.FbxModelId))
	f.AddObjects(cxtModel, nodeAttribute)

	for _, iSubNode := range wrsrc.Node.SubGroupNodes {
		instance, _, err := wrsrc.Wad.GetInstanceFromNode(iSubNode)
		if err != nil {
			// log.Panicf("Failed to load instance %s: %v", wrsrc.Wad.Nodes[iSubNode].Tag.Name, err)
			continue
		}

		gameInstance := instance.(*file_inst.Instance)

		pos := gameInstance.Position1.Vec3()
		rotation := gameInstance.Rotation.Vec3().Mul(180.0 / math.Pi)

		instModelId := f.GenerateId()
		instModel := bfbx73.Model(instModelId, wrsrc.Wad.Nodes[iSubNode].Tag.Name+"\x00\x01Model", "Null").AddNodes(
			bfbx73.Version(232),
			bfbx73.Properties70(),
			bfbx73.Shading(true),
			bfbx73.Culling("CullingOff"),
			bfbx73.Properties70().AddNodes(
				bfbx73.P("Lcl Translation", "Lcl Translation", "", "A+",
					float64(pos[0]), float64(pos[1]), float64(pos[2])),
				bfbx73.P("Lcl Rotation", "Lcl Rotation", "", "A+",
					float64(rotation[0]), float64(rotation[1]), float64(rotation[2])),
			),
		)

		nodeAttribute := bfbx73.NodeAttribute(
			f.GenerateId(), wrsrc.Wad.Nodes[iSubNode].Tag.Name+"\x00\x01NodeAttribute", "Null").AddNodes(
			bfbx73.TypeFlags("Null"),
		)

		f.AddConnections(bfbx73.C("OO", nodeAttribute.Properties[0].(int64), instModelId))
		f.AddObjects(instModel, nodeAttribute)

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
		if fbxObjectI := f.GetCached(objectNode.Tag.Id); fbxObjectI == nil {
			fbxObject = object.ExportFbx(wrsrc.Wad.GetNodeResourceByNodeId(objectNode.Id), f)
		} else {
			fbxObject = fbxObjectI.(*file_obj.FbxExporter)
		}

		f.AddConnections(
			bfbx73.C("OO", fbxObject.FbxModelId, instModelId),
			bfbx73.C("OO", instModelId, fe.FbxModelId),
		)
	}

	return fe
}

func (cxt *Chunk) ExportFbxDefault(wrsrc *wad.WadNodeRsrc) *fbxbuilder.FBXBuilder {
	f := fbxbuilder.NewFBXBuilder(filepath.Join(wrsrc.Wad.Name(), wrsrc.Name()))

	fe := cxt.ExportFbx(wrsrc, f)

	f.AddConnections(bfbx73.C("OO", fe.FbxModelId, 0))

	return f
}
