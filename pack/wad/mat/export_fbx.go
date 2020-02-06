package mat

import (
	"fmt"
	"log"

	"github.com/mogaika/god_of_war_browser/fbx"
	"github.com/mogaika/god_of_war_browser/fbx/cache"
	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_txr "github.com/mogaika/god_of_war_browser/pack/wad/txr"
	"github.com/mogaika/god_of_war_browser/utils"
)

type FbxExporter struct {
	MaterialId uint64
}

func (m *Material) ExportFbx(wrsrc *wad.WadNodeRsrc, f *fbx.FBX, cache *cache.Cache) *FbxExporter {
	fe := &FbxExporter{}
	defer cache.Add(wrsrc.Tag.Id, fe)

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

	material := &fbx.Material{
		Id:           fe.MaterialId,
		Name:         fmt.Sprintf("Material::%d_%v", wrsrc.Tag.Id, wrsrc.Tag.Name),
		Element:      "",
		Version:      102,
		ShadingModel: "lambert",
		MultiLayer:   0,
		Properties70: fbx.Properties70{
			P: []*fbx.Propertie70{
				&fbx.Propertie70{Name: "AmbientColor", Type: "Color", Idk: "A", Value: []float32{0, 0, 0}},
				&fbx.Propertie70{Name: "DiffuseColor", Type: "Color", Idk: "A", Value: color[:3]},
				&fbx.Propertie70{Name: "Emissive", Type: "Vector3D", Purpose: "Vector", Value: []float32{0, 0, 0}},
				&fbx.Propertie70{Name: "Ambient", Type: "Vector3D", Purpose: "Vector", Value: []float32{0, 0, 0}},
				&fbx.Propertie70{Name: "Diffuse", Type: "Vector3D", Purpose: "Vector", Value: color[:3]},
				&fbx.Propertie70{Name: "Opacity", Type: "double", Purpose: "Number", Value: color[3]},
			},
		},
	}

	if mainLayer.ParsedFlags.HaveTexture {
		n := wrsrc.Wad.GetNodeByName(mainLayer.Texture, wrsrc.Node.Id-1, false)
		if n == nil {
			log.Panicf("Error getting texture node '%s' for material '%s'", mainLayer.Texture, wrsrc.Tag.Name)
		}

		var textureFe *file_txr.FbxExporter
		if cachedTexture := cache.Get(n.Tag.Id); cachedTexture == nil {
			txr, _, err := wrsrc.Wad.GetInstanceFromNode(n.Id)
			if err != nil {
				log.Panicf("Error getting texture '%s' for material '%s': %v", n.Tag.Name, wrsrc.Tag.Name, err)
			}

			textureFe = txr.(*file_txr.Texture).ExportFbx(wrsrc.Wad.GetNodeResourceByNodeId(n.Id), f, cache)
		} else {
			textureFe = cachedTexture.(*file_txr.FbxExporter)
		}

		f.Connections.C = append(f.Connections.C, fbx.Connection{
			Type: "OP", Child: textureFe.TextureId, Parent: fe.MaterialId, Extra: []string{"DiffuseColor"},
		})
	}

	f.Objects.Material = append(f.Objects.Material, material)

	return fe
}
