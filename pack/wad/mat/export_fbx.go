package mat

import (
	"fmt"
	"log"

	"github.com/mogaika/fbx/builders/bfbx73"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_txr "github.com/mogaika/god_of_war_browser/pack/wad/txr"
	"github.com/mogaika/god_of_war_browser/utils"
	"github.com/mogaika/god_of_war_browser/utils/fbxbuilder"
)

type FbxExporter struct {
	MaterialId int64
}

func (m *Material) ExportFbx(wrsrc *wad.WadNodeRsrc, f *fbxbuilder.FBXBuilder) *FbxExporter {
	fe := &FbxExporter{}
	defer f.AddCache(wrsrc.Tag.Id, fe)

	var mainLayer *Layer

	removeTextureAlpha := false

	for iLayer := range m.Layers {
		layer := &m.Layers[iLayer]

		if layer.ParsedFlags.RenderingStrangeBlended {
			removeTextureAlpha = true
			mainLayer = layer
			break
		} else if layer.ParsedFlags.RenderingUsual {
			mainLayer = layer
		} else if mainLayer == nil {
			mainLayer = layer
		}
	}

	color := utils.NewColorFloatA(mainLayer.BlendColor[:])
	for i := range color {
		color[i] *= m.Color[i]
	}

	// TODO: remove texture alpha for reflection maps
	_ = removeTextureAlpha

	fe.MaterialId = f.GenerateId()

	material := bfbx73.Material(fe.MaterialId, fmt.Sprintf("%d_%v\x00\x01Material", wrsrc.Tag.Id, wrsrc.Tag.Name), "").AddNodes(
		bfbx73.Version(102),
		bfbx73.ShadingModel("lambert"),
		bfbx73.MultiLayer(0),
		bfbx73.Properties70().AddNodes(
			bfbx73.P("AmbientColor", "Color", "", "A", float64(0), float64(0), float64(0)),
			bfbx73.P("DiffuseColor", "Color", "", "A", float64(color[0]), float64(color[1]), float64(color[2])),
			bfbx73.P("Emissive", "Vector3D", "Vector", "", float64(0), float64(0), float64(0)),
			bfbx73.P("Ambient", "Vector3D", "Vector", "", float64(0), float64(0), float64(0)),
			bfbx73.P("Diffuse", "Vector3D", "Vector", "", float64(color[0]), float64(color[1]), float64(color[2])),
			bfbx73.P("Opacity", "double", "Number", "", float64(color[3])),
		),
	)

	if mainLayer.ParsedFlags.HaveTexture {
		n := wrsrc.Wad.GetNodeByName(mainLayer.Texture, wrsrc.Node.Id-1, false)
		if n == nil {
			log.Panicf("Error getting texture node '%s' for material '%s'", mainLayer.Texture, wrsrc.Tag.Name)
		}

		var textureFe *file_txr.FbxExporter
		if cachedTexture := f.GetCached(n.Tag.Id); cachedTexture == nil {
			txr, _, err := wrsrc.Wad.GetInstanceFromNode(n.Id)
			if err != nil {
				log.Panicf("Error getting texture '%s' for material '%s': %v", n.Tag.Name, wrsrc.Tag.Name, err)
			}

			textureFe = txr.(*file_txr.Texture).ExportFbx(wrsrc.Wad.GetNodeResourceByNodeId(n.Id), f)
		} else {
			textureFe = cachedTexture.(*file_txr.FbxExporter)
		}

		f.AddConnections(bfbx73.C("OP", textureFe.TextureId, fe.MaterialId, "DiffuseColor"))
	}

	f.AddObjects(material)

	return fe
}
