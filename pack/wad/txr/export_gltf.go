package txr

import (
	"bytes"

	"github.com/qmuntal/gltf/modeler"

	"github.com/mogaika/god_of_war_browser/pack/wad"
	"github.com/mogaika/god_of_war_browser/utils/gltfutils"
	"github.com/pkg/errors"
	"github.com/qmuntal/gltf"
)

type GLTFTextureExported struct {
	TextureIndex uint32
	ImageIndex   uint32
	SamplerIndex uint32
}

func (txr *Texture) ExportGLTF(wrsrc *wad.WadNodeRsrc, gltfCacher *gltfutils.GLTFCacher) (*GLTFTextureExported, error) {
	gte := &GLTFTextureExported{}
	defer gltfCacher.AddCache(wrsrc.Tag.Id, gte)

	doc := gltfCacher.Doc

	filterReducedNearest := (txr.Flags>>15)&1 != 0
	filterExpandedNearest := (txr.Flags>>16)&1 != 0
	clampHorisontal := (txr.Flags>>20)&3 == 1
	clampVertical := (txr.Flags>>22)&3 == 1

	sampler := &gltf.Sampler{
		Name: wrsrc.Name() + "_sampler",
	}

	if filterReducedNearest {
		sampler.MinFilter = gltf.MinNearest
	} else {
		sampler.MinFilter = gltf.MinLinear
	}

	if filterExpandedNearest {
		sampler.MagFilter = gltf.MagNearest
	} else {
		sampler.MagFilter = gltf.MagLinear
	}

	if clampHorisontal {
		sampler.WrapS = gltf.WrapClampToEdge
	} else {
		sampler.WrapS = gltf.WrapRepeat
	}

	if clampVertical {
		sampler.WrapT = gltf.WrapClampToEdge
	} else {
		sampler.WrapT = gltf.WrapRepeat
	}

	gte.SamplerIndex = uint32(len(doc.Samplers))
	doc.Samplers = append(doc.Samplers, sampler)

	ajaxI, err := txr.Marshal(wrsrc)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to marshal image %q", wrsrc.Name())
	}
	pngBytes := ajaxI.(*Ajax).Images[0].Image

	/*
		gte.ImageIndex = uint32(len(doc.Images))
		doc.Images = append(doc.Images, &gltf.Image{
			Name: wrsrc.Name() + "_image",
			URI:  "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes),
		})
	*/
	gte.ImageIndex, err = modeler.WriteImage(doc, wrsrc.Name()+"_image", "image/png", bytes.NewReader(pngBytes))
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to write gltf image")
	}

	gte.TextureIndex = uint32(len(doc.Textures))
	doc.Textures = append(doc.Textures, &gltf.Texture{
		Name:    wrsrc.Name(),
		Sampler: gltf.Index(gte.SamplerIndex),
		Source:  gltf.Index(gte.ImageIndex),
	})

	return gte, nil
}
