package mat

import (
	"log"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	file_txr "github.com/mogaika/god_of_war_browser/pack/wad/txr"
	"github.com/mogaika/god_of_war_browser/utils/gltfutils"
	"github.com/qmuntal/gltf"
)

type GLTFMaterialExported struct {
	MaterialId uint32
}

func (m *Material) ExportGLTF(wrsrc *wad.WadNodeRsrc, gltfCacher *gltfutils.GLTFCacher) (*GLTFMaterialExported, error) {
	glme := &GLTFMaterialExported{}
	defer gltfCacher.AddCache(wrsrc.Tag.Id, glme)

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

	// TODO: remove texture alpha for reflection maps
	_ = removeTextureAlpha

	color := new([4]float32)
	*color = [4]float32(mainLayer.BlendColor)

	gltfMaterial := &gltf.Material{
		Name:        wrsrc.Name(),
		DoubleSided: true,
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			BaseColorFactor: color,
		},
	}

	if mainLayer.ParsedFlags.HaveTexture {
		n := wrsrc.Wad.GetNodeByName(mainLayer.Texture, wrsrc.Node.Id-1, false)
		if n == nil {
			log.Panicf("Error getting texture node %q for material %q", mainLayer.Texture, wrsrc.Name())
		}

		gte := gltfCacher.GetCachedOr(
			n.Tag.Id, func() interface{} {
				txr, _, err := wrsrc.Wad.GetInstanceFromNode(n.Id)
				if err != nil {
					log.Panicf("Error getting texture %q for material %q: %v", n.Tag.Name, wrsrc.Name(), err)
				}

				gte, err := txr.(*file_txr.Texture).ExportGLTF(wrsrc.Wad.GetNodeResourceByNodeId(n.Id), gltfCacher)
				if err != nil {
					log.Panicf("Error exporting texture %q for material %q: %v", n.Tag.Name, wrsrc.Name(), err)
				}

				return gte
			}).(*file_txr.GLTFTextureExported)

		gltfMaterial.PBRMetallicRoughness.BaseColorTexture = &gltf.TextureInfo{
			Index: gte.TextureIndex,
		}

	}

	glme.MaterialId = uint32(len(gltfCacher.Doc.Materials))
	gltfCacher.Doc.Materials = append(gltfCacher.Doc.Materials, gltfMaterial)

	return glme, nil
}
